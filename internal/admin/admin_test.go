package admin

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nova/internal/auth"
)

func do(t *testing.T, ts *httptest.Server, method, path, token, body string) (int, map[string]any) {
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
	m := map[string]any{}
	json.NewDecoder(res.Body).Decode(&m)
	return res.StatusCode, m
}

func TestAdminFlow(t *testing.T) {
	mux := http.NewServeMux()
	dir := t.TempDir()
	asvc, err := auth.Mount(mux, dir, "/api/auth")
	if err != nil {
		t.Fatal(err)
	}
	if err := Mount(mux, dir, asvc); err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	_, m := do(t, ts, "POST", "/api/auth/register", "", `{"email":"root@x.com","password":"12345678"}`)
	super := m["token"].(string)
	_, m2 := do(t, ts, "POST", "/api/auth/register", "", `{"email":"u@x.com","password":"12345678"}`)
	user := m2["token"].(string)

	code, m := do(t, ts, "POST", "/api/admin/llm/profiles", super,
		`{"name":"主用","base_url":"https://api.x.com/v1","model":"gpt-x","api_key":"sk-secret-1234","enabled":true}`)
	if code != 200 {
		t.Fatalf("create: %d %v", code, m)
	}
	id := m["id"].(string)

	code, m = do(t, ts, "GET", "/api/admin/llm/profiles", super, "")
	ps := m["profiles"].([]any)
	if code != 200 || len(ps) != 1 {
		t.Fatalf("list: %d", code)
	}
	blob, _ := json.Marshal(ps[0])
	if strings.Contains(string(blob), "sk-secret-1234") {
		t.Fatal("明文key泄漏")
	}
	if !strings.Contains(string(blob), "****1234") {
		t.Fatal("key_tail缺失")
	}

	if c, _ := do(t, ts, "GET", "/api/admin/llm/profiles", user, ""); c != 403 {
		t.Fatalf("user应403: %d", c)
	}
	if c, _ := do(t, ts, "GET", "/api/admin/llm/profiles", "", ""); c != 401 {
		t.Fatalf("无token应401: %d", c)
	}

	if c, _ := do(t, ts, "PUT", "/api/admin/ai/ops", super, `{"ai_enabled":false,"paid_mode":true}`); c != 200 {
		t.Fatalf("putOps: %d", c)
	}
	_, m = do(t, ts, "GET", "/api/admin/ai/ops", super, "")
	sw := m["switches"].(map[string]any)
	if sw["ai_enabled"] != false || sw["paid_mode"] != true {
		t.Fatal("ops未生效")
	}
	if c, _ := do(t, ts, "PUT", "/api/admin/ai/ops", user, `{"ai_enabled":true}`); c != 403 {
		t.Fatalf("user写ops应403: %d", c)
	}

	_, m = do(t, ts, "GET", "/api/admin/features", super, "")
	fs := m["features"].([]any)
	if len(fs) != 7 {
		t.Fatalf("种子功能应7条: %d", len(fs))
	}
	fid := fs[0].(map[string]any)["id"].(string)
	if c, _ := do(t, ts, "PUT", "/api/admin/features/"+fid, super, `{"enabled":false,"is_paid":true,"price":9.9}`); c != 200 {
		t.Fatalf("putFeature: %d", c)
	}
	if c, _ := do(t, ts, "DELETE", "/api/admin/llm/profiles/"+id, super, ""); c != 200 {
		t.Fatalf("delete: %d", c)
	}
}
