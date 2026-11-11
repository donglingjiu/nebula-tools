package main

// NT-1: feature pool with realtime on/off toggling (SSE). stdlib only.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

// Feature is one entry of the feature pool.
type Feature struct {
	ID          string    `json:"id"`
	Category    string    `json:"category"`
	Name        string    `json:"name"`
	Desc        string    `json:"desc"`
	Route       string    `json:"route"`
	Status      string    `json:"status"`
	GrayPercent int       `json:"grayPercent"`
	AI          bool      `json:"ai"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func ntSeedFeatures() []*Feature {
	mk := func(id, cat, name, desc, route, status string, gray int, ai bool) *Feature {
		return &Feature{ID: id, Category: cat, Name: name, Desc: desc, Route: route,
			Status: status, GrayPercent: gray, AI: ai, UpdatedAt: time.Now()}
	}
	return []*Feature{
		mk("doc-pdf2word", "document", "PDF 转 Word", "PDF 转可编辑 Word", "/pdf2word.html", "online", 100, false),
		mk("doc-word2pdf", "document", "Word 转 PDF", "Word 导出 PDF", "/word2pdf.html", "online", 100, false),
		mk("doc-pdf-tools", "document", "PDF 合并拆分", "PDF 合并/拆分/压缩", "/pdf.html", "online", 100, false),
		mk("doc-markdown", "document", "Markdown 编辑", "Markdown 编辑与导出", "/markdown.html", "online", 100, false),
		mk("doc-ocr", "document", "图片 OCR", "图片文字识别(AI)", "/pdf.html", "beta", 30, true),
		mk("sheet-excel", "spreadsheet", "Excel 工具箱", "Excel 清洗/转换/合并", "/excel.html", "online", 100, false),
		mk("sheet-csv-clean", "spreadsheet", "CSV 清洗", "CSV 去重/补全/标准化", "/excel.html", "online", 100, false),
		mk("sheet-merge", "spreadsheet", "多表合并", "多 Sheet/多文件合并", "/excel.html", "beta", 30, false),
		mk("data-json", "data", "JSON 格式化", "JSON 格式化/校验", "/json.html", "online", 100, false),
		mk("data-jsonx", "data", "JSON 高级", "JSON Path/对比/转表", "/jsonx.html", "online", 100, false),
		mk("data-regex", "data", "正则测试", "正则表达式调试", "/regex.html", "online", 100, false),
		mk("data-base64", "data", "Base64 编解码", "文本/文件 Base64", "/base64.html", "online", 100, false),
		mk("data-imgtool", "data", "图片处理", "压缩/裁剪/格式转换", "/imgtool.html", "online", 100, false),
		mk("ppt-outline", "ppt", "PPT 大纲生成", "AI 生成 PPT 大纲", "/ppt.html", "offline", 0, true),
		mk("ppt-beautify", "ppt", "PPT 美化", "AI 版式与配图建议", "/ppt.html", "offline", 0, true),
		mk("mind-mindmap", "mindmap", "脑图编辑", "节点编辑/导出", "/mindmap.html", "online", 100, false),
		mk("mind-ai-expand", "mindmap", "AI 思维扩散", "AI 补全与结构化脑图", "/mindmap.html", "beta", 30, true),
		mk("wf-ai-chat", "workflow", "AI 对话", "多模型对话工作台", "/ai.html", "online", 100, true),
		mk("wf-ai-workflow", "workflow", "AI 工作流", "可视化工作流编排", "/ai-studio.html", "online", 100, true),
		mk("ai-quickchat", "ai", "AI 快聊", "不建流程直问, SSE 流式", "/quickchat.html", "online", 95, true),
		mk("ai-materials", "ai", "材料库", "跨流程复用的参考资料", "/materials.html", "online", 95, true),
		mk("wf-flowchart", "workflow", "流程图", "流程图绘制", "/flowchart.html", "online", 100, false),
		mk("wf-pipeline-7", "workflow", "全流程 7 步", "头脑风暴到测试落地", "/ai-studio.html", "online", 100, true),
		mk("wf-decompose", "workflow", "智能工作流", "数据清洗步骤链, AI 拆解自然语言", "/workflows.html", "online", 100, true),
		mk("wf-metrics", "workflow", "监控看板", "QPS/延迟/错误率/LLM 调用", "/admin-metrics.html", "online", 100, false),
	}
}

type ntFeatureStore struct {
	mu    sync.RWMutex
	items map[string]*Feature
	subMu sync.Mutex
	subs  map[chan []byte]bool
}

var ntFeaturePool = newNTFeatureStore()

func newNTFeatureStore() *ntFeatureStore {
	s := &ntFeatureStore{items: map[string]*Feature{}, subs: map[chan []byte]bool{}}
	for _, f := range ntSeedFeatures() {
		s.items[f.ID] = f
	}
	return s
}

func (s *ntFeatureStore) broadcast(ev map[string]interface{}) {
	b, _ := json.Marshal(ev)
	s.subMu.Lock()
	for ch := range s.subs {
		select {
		case ch <- b:
		default:
		}
	}
	s.subMu.Unlock()
}

func ntWriteJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func ntAdminOK(r *http.Request) bool {
	tok := os.Getenv("ADMIN_TOKEN")
	if tok == "" {
		return true
	}
	return r.Header.Get("X-Admin-Token") == tok
}

func (s *ntFeatureStore) list(includeOffline bool) []*Feature {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Feature, 0, len(s.items))
	for _, f := range s.items {
		if includeOffline || f.Status != "offline" {
			cp := *f
			out = append(out, &cp)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (s *ntFeatureStore) hPublicList(w http.ResponseWriter, r *http.Request) {
	ntWriteJSON(w, 200, map[string]interface{}{"features": s.list(false)})
}

func (s *ntFeatureStore) hAdminList(w http.ResponseWriter, r *http.Request) {
	if !ntAdminOK(r) {
		ntWriteJSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}
	ntWriteJSON(w, 200, map[string]interface{}{"features": s.list(true)})
}

func (s *ntFeatureStore) hToggle(w http.ResponseWriter, r *http.Request) {
	if !ntAdminOK(r) {
		ntWriteJSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}
	if r.Method != http.MethodPost {
		ntWriteJSON(w, 405, map[string]string{"error": "POST required"})
		return
	}
	var req struct {
		ID          string `json:"id"`
		Status      string `json:"status"`
		GrayPercent *int   `json:"grayPercent"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
		ntWriteJSON(w, 400, map[string]string{"error": "bad request"})
		return
	}
	switch req.Status {
	case "online", "beta", "gray", "offline":
	default:
		ntWriteJSON(w, 400, map[string]string{"error": "invalid status"})
		return
	}
	s.mu.Lock()
	f, ok := s.items[req.ID]
	if !ok {
		s.mu.Unlock()
		ntWriteJSON(w, 404, map[string]string{"error": "feature not found"})
		return
	}
	f.Status = req.Status
	if req.GrayPercent != nil && *req.GrayPercent >= 0 && *req.GrayPercent <= 100 {
		f.GrayPercent = *req.GrayPercent
	}
	f.UpdatedAt = time.Now()
	cp := *f
	s.mu.Unlock()
	s.broadcast(map[string]interface{}{"type": "update", "feature": &cp})
	ntWriteJSON(w, 200, map[string]interface{}{"ok": true, "feature": &cp})
}

func (s *ntFeatureStore) hStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	ch := make(chan []byte, 16)
	s.subMu.Lock()
	s.subs[ch] = true
	s.subMu.Unlock()
	defer func() {
		s.subMu.Lock()
		delete(s.subs, ch)
		s.subMu.Unlock()
	}()
	fmt.Fprint(w, "data: {\"type\":\"hello\"}\n\n")
	fl.Flush()
	for {
		select {
		case b := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", b)
			fl.Flush()
		case <-time.After(25 * time.Second):
			fmt.Fprint(w, ": keepalive\n\n")
			fl.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// FeatureWrap routes feature-pool APIs before the app's own router.
func FeatureWrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/features":
			ntFeaturePool.hPublicList(w, r)
		case "/api/features/stream":
			ntFeaturePool.hStream(w, r)
		case "/api/pool/features":
			ntFeaturePool.hAdminList(w, r)
		case "/api/pool/features/toggle":
			ntFeaturePool.hToggle(w, r)
		default:
			next.ServeHTTP(w, r)
		}
	})
}
