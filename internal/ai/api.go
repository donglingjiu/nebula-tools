package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"nova/internal/admin"
	"nova/internal/auth"
)

// defaultFeatureKey 无 feature_key 时的默认功能键; 种子无此键时回落第一个启用功能
const defaultFeatureKey = "ai_chat"

type Service struct {
	dir string
}

type gateErr struct {
	code  int
	msg   string
	price float64
}

// Mount 挂载 /api/ai/*; 410 下线门/402 付费门/503 总开关门在 chat 关口生效
func Mount(mux *http.ServeMux, dir string, asvc *auth.Service) error {
	if asvc == nil {
		return errors.New("auth service 为 nil, ai 路由未挂载")
	}
	s := &Service{dir: dir}
	mux.HandleFunc("POST /api/ai/chat", asvc.RequireRole("user", s.chat))
	mux.HandleFunc("GET /api/ai/status", asvc.RequireRole("user", s.status))
	mux.HandleFunc("POST /api/ai/flow/start", asvc.RequireRole("user", s.flowStart))
	mux.HandleFunc("POST /api/ai/flow/advance", asvc.RequireRole("user", s.flowAdvance))
	mux.HandleFunc("POST /api/ai/flow/run", asvc.RequireRole("user", s.flowRun))
	mux.HandleFunc("POST /api/ai/flow/chat", asvc.RequireRole("user", s.flowChat))
	mux.HandleFunc("GET /api/ai/flow/export", asvc.RequireRole("user", s.flowExport))
	return nil
}

// checkGate 门槛链: 总开关503 → 功能404 → 下线410 → 付费402(带price)
func checkGate(st *admin.Store, key string) *gateErr {
	if !st.GetOps().AIEnabled {
		return &gateErr{code: 503, msg: "AI 总开关已关闭, 请联系管理员"}
	}
	f, ok := st.FeatureByKey(key)
	if !ok {
		return &gateErr{code: 404, msg: "功能不存在: " + key}
	}
	if !f.Enabled {
		return &gateErr{code: 410, msg: "功能已下线: " + f.Name}
	}
	if f.IsPaid && st.GetOps().PaidMode {
		return &gateErr{code: 402, msg: fmt.Sprintf("「%s」为付费功能(¥%.2f), 请开通后使用", f.Name, f.Price), price: f.Price}
	}
	return nil
}

// resolveFeature 解析请求功能键(空则默认, 默认缺失回落首个启用)
func (s *Service) resolveFeature(want string) (*admin.Store, string, *gateErr) {
	st, err := admin.OpenStore(s.dir)
	if err != nil {
		return nil, "", &gateErr{code: 500, msg: "读取配置失败"}
	}
	key := want
	if key == "" {
		key = defaultFeatureKey
		if _, ok := st.FeatureByKey(key); !ok {
			for _, f := range st.FeaturesAll() {
				if f.Enabled {
					key = f.Key
					break
				}
			}
		}
	}
	if ge := checkGate(st, key); ge != nil {
		return nil, "", ge
	}
	return st, key, nil
}

func (s *Service) status(w http.ResponseWriter, r *http.Request) {
	st, err := admin.OpenStore(s.dir)
	if err != nil {
		fail(w, 500, "读取配置失败")
		return
	}
	ops := st.GetOps()
	arr := make([]map[string]any, 0, len(st.FeaturesAll()))
	for _, f := range st.FeaturesAll() {
		arr = append(arr, map[string]any{
			"key": f.Key, "name": f.Name, "enabled": f.Enabled,
			"is_paid": f.IsPaid, "price": f.Price,
		})
	}
	def := map[string]any{}
	if prof, _, err := admin.ResolveDefaultProfile(s.dir); err == nil {
		def = map[string]any{"name": prof.Name, "model": prof.Model}
	}
	writeJSON(w, 200, map[string]any{
		"ai_enabled": ops.AIEnabled, "paid_mode": ops.PaidMode,
		"default": def, "features": arr,
	})
}

func (s *Service) chat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FeatureKey string              `json:"feature_key"`
		Messages   []map[string]string `json:"messages"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || len(req.Messages) == 0 {
		fail(w, 400, "参数错误: 需要 messages 数组")
		return
	}
	if _, _, ge := s.resolveFeature(req.FeatureKey); ge != nil {
		writeGate(w, ge)
		return
	}
	prof, key, err := admin.ResolveDefaultProfile(s.dir)
	if err != nil {
		fail(w, 503, err.Error())
		return
	}
	relaySSE(w, r, prof, key, req.Messages)
}

// relaySSE 转发上游 OpenAI 兼容 SSE 流(逐行透传 data:/event:/注释/空行)
func relaySSE(w http.ResponseWriter, r *http.Request, prof admin.Profile, key string, msgs []map[string]string) {
	payload, _ := json.Marshal(map[string]any{
		"model": prof.Model, "messages": msgs, "stream": true,
	})
	up, err := http.NewRequestWithContext(r.Context(), "POST",
		strings.TrimRight(prof.BaseURL, "/")+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		fail(w, 500, "构造上游请求失败")
		return
	}
	up.Header.Set("Content-Type", "application/json")
	up.Header.Set("Authorization", "Bearer "+key)
	up.Header.Set("Accept", "text/event-stream")
	resp, err := http.DefaultClient.Do(up)
	if err != nil {
		fail(w, 502, "上游不可达: "+err.Error())
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		fail(w, 502, fmt.Sprintf("上游返回 %d: %s", resp.StatusCode, strings.TrimSpace(string(b))))
		return
	}
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	fl, _ := w.(http.Flusher)
	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "data:") ||
			strings.HasPrefix(line, "event:") || strings.HasPrefix(line, ":") {
			fmt.Fprintln(w, line)
			if fl != nil {
				fl.Flush()
			}
		}
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"message": msg})
}

func writeGate(w http.ResponseWriter, ge *gateErr) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(ge.code)
	if ge.price > 0 {
		json.NewEncoder(w).Encode(map[string]any{"message": ge.msg, "price": ge.price})
	} else {
		json.NewEncoder(w).Encode(map[string]string{"message": ge.msg})
	}
}
