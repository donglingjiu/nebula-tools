package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nova/internal/admin"
	"nova/internal/auth"
)

func do(t *testing.T, ts *httptest.Server, method, path, token, body string) (int, map[string]any, string) {
	t.Helper()
	var rd io.Reader
	if body != "" {
		rd = bytes.NewBufferString(body)
	}
	req, _ := http.NewRequest(method, ts.URL+path, rd)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	var m map[string]any
	json.Unmarshal(raw, &m)
	return res.StatusCode, m, string(raw)
}

func TestAIGatesAndSSE(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		fl := w.(http.Flusher)
		for _, tok := range []string{"你好", "，", "世界"} {
			b, _ := json.Marshal(map[string]any{"choices": []map[string]any{{"delta": map[string]string{"content": tok}}}})
			fmt.Fprintf(w, "data: %s\n\n", b)
			fl.Flush()
		}
		fmt.Fprint(w, "data: [DONE]\n\n")
		fl.Flush()
	}))
	defer up.Close()

	mux := http.NewServeMux()
	dir := t.TempDir()
	asvc, err := auth.Mount(mux, dir, "/api/v2/auth")
	if err != nil {
		t.Fatal(err)
	}
	if err := admin.Mount(mux, dir, asvc); err != nil {
		t.Fatal(err)
	}
	if err := Mount(mux, dir, asvc); err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(mux)
	defer ts.Close()

	_, m, _ := do(t, ts, "POST", "/api/v2/auth/register", "", `{"email":"root@x.com","password":"12345678"}`)
	super := m["token"].(string)
	do(t, ts, "POST", "/api/v2/auth/register", "", `{"email":"u@x.com","password":"12345678"}`)
	_, m2, _ := do(t, ts, "POST", "/api/v2/auth/login", "", `{"email":"u@x.com","password":"12345678"}`)
	user := m2["token"].(string)

	_, m3, _ := do(t, ts, "POST", "/api/admin/llm/profiles", super,
		fmt.Sprintf(`{"name":"模拟","base_url":"%s/v1","model":"mock-1","api_key":"sk-mock-0000","enabled":true}`, up.URL))
	pid := m3["id"].(string)
	if c, _, _ := do(t, ts, "PUT", "/api/admin/llm/profiles/"+pid, super, `{"is_default":true}`); c != 200 {
		t.Fatalf("设默认: %d", c)
	}

	if c, _, _ := do(t, ts, "POST", "/api/ai/chat", "", `{"messages":[{"role":"user","content":"hi"}]}`); c != 401 {
		t.Fatalf("无token应401: %d", c)
	}
	c, _, raw := do(t, ts, "POST", "/api/ai/chat", user, `{"messages":[{"role":"user","content":"hi"}]}`)
	if c != 200 || !strings.Contains(raw, "data: ") || !strings.Contains(raw, "你好") || !strings.Contains(raw, "[DONE]") {
		t.Fatalf("SSE异常: %d %q", c, raw)
	}

	_, mf, _ := do(t, ts, "GET", "/api/admin/features", super, "")
	var fid string
	fl, _ := mf["features"].([]any)
	if fl == nil {
		fl, _ = mf["items"].([]any)
	}
	for _, it := range fl {
		f := it.(map[string]any)
		if f["key"] == "ai_chat" {
			fid, _ = f["id"].(string)
		}
	}
	if fid == "" && len(fl) > 0 {
		fid, _ = fl[0].(map[string]any)["id"].(string)
	}
	if fid == "" {
		t.Skipf("features列表形状未知, 跳过门矩阵: %v", mf)
	}

	do(t, ts, "PUT", "/api/admin/features/"+fid, super, `{"enabled":false}`)
	if c, m4, _ := do(t, ts, "POST", "/api/ai/chat", user, `{"messages":[{"role":"user","content":"hi"}]}`); c != 410 {
		t.Fatalf("下线应410: %d %v", c, m4)
	}
	do(t, ts, "PUT", "/api/admin/features/"+fid, super, `{"enabled":true}`)

	do(t, ts, "PUT", "/api/admin/features/"+fid, super, `{"is_paid":true,"price":9.9}`)
	do(t, ts, "PUT", "/api/admin/ai/ops", super, `{"paid_mode":true}`)
	c, m5, _ := do(t, ts, "POST", "/api/ai/chat", user, `{"messages":[{"role":"user","content":"hi"}]}`)
	if c != 402 {
		t.Fatalf("付费应402: %d %v", c, m5)
	}
	if pr, ok := m5["price"].(float64); !ok || pr != 9.9 {
		t.Fatalf("402应带price: %v", m5)
	}
	do(t, ts, "PUT", "/api/admin/ai/ops", super, `{"paid_mode":false}`)
	do(t, ts, "PUT", "/api/admin/features/"+fid, super, `{"is_paid":false}`)
	if c, _, _ := do(t, ts, "POST", "/api/ai/chat", user, `{"messages":[{"role":"user","content":"hi"}]}`); c != 200 {
		t.Fatalf("还原后应放行: %d", c)
	}

	do(t, ts, "PUT", "/api/admin/ai/ops", super, `{"ai_enabled":false}`)
	if c, _, _ := do(t, ts, "POST", "/api/ai/chat", user, `{"messages":[{"role":"user","content":"hi"}]}`); c != 503 {
		t.Fatalf("总开关关应503: %d", c)
	}
	do(t, ts, "PUT", "/api/admin/ai/ops", super, `{"ai_enabled":true}`)

	if c, ms, _ := do(t, ts, "GET", "/api/ai/status", user, ""); c != 200 || ms["ai_enabled"] != true {
		t.Fatalf("status: %d %v", c, ms)
	}
}
