package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nova/internal/admin"
)

type stepIn struct {
	Name   string `json:"name"`
	Prompt string `json:"prompt"`
}

type FlowStep struct {
	Name   string `json:"name"`
	Prompt string `json:"prompt"`
	Output string `json:"output,omitempty"`
	Status string `json:"status"`
}

type Flow struct {
	ID        string              `json:"id"`
	Title     string              `json:"title"`
	State     string              `json:"state"`
	CreatedAt string              `json:"created_at"`
	Steps     []FlowStep          `json:"steps"`
	Chats     []map[string]string `json:"chats,omitempty"`
}

type flowStore struct {
	Flows map[string]*Flow `json:"flows"`
}

func (s *Service) loadFlows() (*flowStore, error) {
	st := &flowStore{Flows: map[string]*Flow{}}
	b, err := os.ReadFile(filepath.Join(s.dir, "flows.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return st, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(b, st); err != nil {
		return nil, err
	}
	if st.Flows == nil {
		st.Flows = map[string]*Flow{}
	}
	return st, nil
}

func (s *Service) saveFlows(st *flowStore) error {
	b, err := json.MarshalIndent(st, "", " ")
	if err != nil {
		return err
	}
	p := filepath.Join(s.dir, "flows.json")
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, b, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

func newFlowID() string {
	return fmt.Sprintf("%x", time.Now().UnixNano())
}

// collectChat 非流式调上游取全文(工作台步骤/对话用)
func collectChat(prof admin.Profile, key string,
	msgs []map[string]string) (string, error) {
	payload, _ := json.Marshal(map[string]any{
		"model": prof.Model, "messages": msgs, "stream": false,
	})
	req, err := http.NewRequest("POST",
		strings.TrimRight(prof.BaseURL, "/")+"/chat/completions",
		bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("上游返回 %d: %s",
			resp.StatusCode, strings.TrimSpace(string(b)))
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", errors.New("上游无返回")
	}
	return out.Choices[0].Message.Content, nil
}

// flowCtx 汇总已完成步骤输出作为上下文
func flowCtx(f *Flow) string {
	var sb strings.Builder
	for _, st := range f.Steps {
		if st.Status == "done" && st.Output != "" {
			sb.WriteString("[" + st.Name + "] " + st.Output + "\n")
		}
	}
	return strings.TrimSpace(sb.String())
}

// gateNow 门链+默认Profile解析(AI消耗端点共用)
func (s *Service) gateNow() (admin.Profile, string, *gateErr) {
	if _, _, ge := s.resolveFeature(""); ge != nil {
		return admin.Profile{}, "", ge
	}
	prof, key, err := admin.ResolveDefaultProfile(s.dir)
	if err != nil {
		return admin.Profile{}, "", &gateErr{code: 503, msg: err.Error()}
	}
	return prof, key, nil
}

func (s *Service) getFlow(id string) (*flowStore, *Flow, error) {
	st, err := s.loadFlows()
	if err != nil {
		return nil, nil, err
	}
	f, ok := st.Flows[id]
	if !ok {
		return nil, nil, errors.New("流程不存在")
	}
	return st, f, nil
}

func (s *Service) flowStart(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string   `json:"title"`
		Steps []stepIn `json:"steps"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil ||
		strings.TrimSpace(req.Title) == "" || len(req.Steps) == 0 {
		fail(w, 400, "参数错误: 需要 title 与 steps")
		return
	}
	st, err := s.loadFlows()
	if err != nil {
		fail(w, 500, "读取流程失败")
		return
	}
	f := &Flow{ID: newFlowID(), Title: req.Title,
		State:     "created",
		CreatedAt: time.Now().Format(time.RFC3339)}
	for _, si := range req.Steps {
		f.Steps = append(f.Steps, FlowStep{
			Name: si.Name, Prompt: si.Prompt, Status: "pending"})
	}
	st.Flows[f.ID] = f
	if err := s.saveFlows(st); err != nil {
		fail(w, 500, "保存流程失败")
		return
	}
	writeJSON(w, 200, f)
}

// advanceOne 执行当前待跑步骤; 返回nil步=已全部完成
func (s *Service) advanceOne(f *Flow) (*FlowStep, *gateErr) {
	idx := -1
	for i := range f.Steps {
		if f.Steps[i].Status == "pending" {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil, nil
	}
	prof, key, ge := s.gateNow()
	if ge != nil {
		return nil, ge
	}
	msgs := []map[string]string{}
	if ctx := flowCtx(f); ctx != "" {
		msgs = append(msgs, map[string]string{
			"role":    "system",
			"content": "前序步骤输出:\n" + ctx,
		})
	}
	msgs = append(msgs, map[string]string{
		"role": "user", "content": f.Steps[idx].Prompt,
	})
	out, err := collectChat(prof, key, msgs)
	if err != nil {
		f.Steps[idx].Status = "error"
		f.Steps[idx].Output = err.Error()
		return &f.Steps[idx], nil
	}
	f.Steps[idx].Output = out
	f.Steps[idx].Status = "done"
	f.State = "running"
	all := true
	for _, st := range f.Steps {
		if st.Status != "done" {
			all = false
		}
	}
	if all {
		f.State = "done"
	}
	return &f.Steps[idx], nil
}

func (s *Service) flowAdvance(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID string `json:"id"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.ID == "" {
		fail(w, 400, "参数错误: 需要 id")
		return
	}
	st, f, err := s.getFlow(req.ID)
	if err != nil {
		fail(w, 404, err.Error())
		return
	}
	step, ge := s.advanceOne(f)
	if ge != nil {
		writeGate(w, ge)
		return
	}
	if err := s.saveFlows(st); err != nil {
		fail(w, 500, "保存流程失败")
		return
	}
	if step == nil {
		writeJSON(w, 200, map[string]any{"done": true, "flow": f})
		return
	}
	writeJSON(w, 200, map[string]any{
		"done": false, "step": step, "flow": f})
}

func (s *Service) flowRun(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID string `json:"id"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.ID == "" {
		fail(w, 400, "参数错误: 需要 id")
		return
	}
	st, f, err := s.getFlow(req.ID)
	if err != nil {
		fail(w, 404, err.Error())
		return
	}
	for {
		step, ge := s.advanceOne(f)
		if ge != nil {
			writeGate(w, ge)
			return
		}
		if step == nil || step.Status == "error" {
			break
		}
	}
	if err := s.saveFlows(st); err != nil {
		fail(w, 500, "保存流程失败")
		return
	}
	writeJSON(w, 200, f)
}

func (s *Service) flowChat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID      string `json:"id"`
		Message string `json:"message"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil ||
		req.ID == "" || strings.TrimSpace(req.Message) == "" {
		fail(w, 400, "参数错误: 需要 id 与 message")
		return
	}
	st, f, err := s.getFlow(req.ID)
	if err != nil {
		fail(w, 404, err.Error())
		return
	}
	prof, key, ge := s.gateNow()
	if ge != nil {
		writeGate(w, ge)
		return
	}
	msgs := []map[string]string{}
	if ctx := flowCtx(f); ctx != "" {
		msgs = append(msgs, map[string]string{
			"role":    "system",
			"content": "流程「" + f.Title + "」步骤输出:\n" + ctx,
		})
	}
	for _, m := range f.Chats {
		msgs = append(msgs, m)
	}
	msgs = append(msgs, map[string]string{
		"role": "user", "content": req.Message,
	})
	reply, err := collectChat(prof, key, msgs)
	if err != nil {
		fail(w, 502, err.Error())
		return
	}
	f.Chats = append(f.Chats,
		map[string]string{"role": "user", "content": req.Message},
		map[string]string{"role": "assistant", "content": reply})
	if err := s.saveFlows(st); err != nil {
		fail(w, 500, "保存流程失败")
		return
	}
	writeJSON(w, 200, map[string]any{"reply": reply, "flow": f})
}

func (s *Service) flowExport(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	_, f, err := s.getFlow(id)
	if err != nil {
		fail(w, 404, err.Error())
		return
	}
	var sb strings.Builder
	sb.WriteString("# " + f.Title + "\n\n")
	sb.WriteString("- 状态: " + f.State + "\n")
	sb.WriteString("- 创建: " + f.CreatedAt + "\n\n")
	for _, st := range f.Steps {
		sb.WriteString("## " + st.Name + " [" + st.Status + "]\n\n")
		sb.WriteString("**提示词**: " + st.Prompt + "\n\n")
		if st.Output != "" {
			sb.WriteString("**输出**:\n\n" + st.Output + "\n\n")
		}
	}
	if len(f.Chats) > 0 {
		sb.WriteString("## 对话记录\n\n")
		for _, m := range f.Chats {
			who := "用户"
			if m["role"] == "assistant" {
				who = "AI"
			}
			sb.WriteString("**" + who + "**: " + m["content"] + "\n\n")
		}
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition",
		"attachment; filename=flow-"+f.ID+".md")
	w.WriteHeader(200)
	w.Write([]byte(sb.String()))
}
