package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newSrv(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	s, err := Mount(mux, t.TempDir(), "/api/auth")
	if err != nil {
		t.Fatal(err)
	}
	for _, role := range []string{"user", "admin", "super"} {
		mux.HandleFunc("GET /probe/"+role, s.RequireRole(role, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	}
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return ts
}

func call(t *testing.T, ts *httptest.Server, method, path, token, body string) (int, map[string]any) {
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

func TestAuthFlow(t *testing.T) {
	ts := newSrv(t)
	// 1) 首个注册用户 = super；邮箱归一化
	code, m := call(t, ts, "POST", "/api/auth/register", "", `{"email":"Admin@Example.com","password":"12345678"}`)
	if code != 200 || m["user"].(map[string]any)["role"] != "super" {
		t.Fatalf("register1: %d %v", code, m)
	}
	tok := m["token"].(string)
	// 2) 重复注册 409 / 弱密码 400 / 坏邮箱 400
	if c, _ := call(t, ts, "POST", "/api/auth/register", "", `{"email":"admin@example.com","password":"12345678"}`); c != 409 {
		t.Fatalf("dup: %d", c)
	}
	if c, _ := call(t, ts, "POST", "/api/auth/register", "", `{"email":"u2@example.com","password":"1234"}`); c != 400 {
		t.Fatalf("weakpw: %d", c)
	}
	if c, _ := call(t, ts, "POST", "/api/auth/register", "", `{"email":"nope","password":"12345678"}`); c != 400 {
		t.Fatalf("badmail: %d", c)
	}
	// 3) 第二个用户 = user
	code, m2 := call(t, ts, "POST", "/api/auth/register", "", `{"email":"u2@example.com","password":"12345678"}`)
	if code != 200 || m2["user"].(map[string]any)["role"] != "user" {
		t.Fatalf("register2: %d %v", code, m2)
	}
	// 4) 登录：错密码 401 / 对密码 200
	if c, _ := call(t, ts, "POST", "/api/auth/login", "", `{"email":"admin@example.com","password":"wrong-pass"}`); c != 401 {
		t.Fatalf("badpw: %d", c)
	}
	code, m3 := call(t, ts, "POST", "/api/auth/login", "", `{"email":"admin@example.com","password":"12345678"}`)
	if code != 200 || m3["token"] == "" {
		t.Fatalf("login: %d %v", code, m3)
	}
	// 5) me：合法 token / 非法 token
	code, m4 := call(t, ts, "GET", "/api/auth/me", tok, "")
	if code != 200 || m4["user"].(map[string]any)["email"] != "admin@example.com" {
		t.Fatalf("me: %d %v", code, m4)
	}
	if c, _ := call(t, ts, "GET", "/api/auth/me", "bad.token.here", ""); c != 401 {
		t.Fatalf("me bad: %d", c)
	}
	// 6) 角色门：super 通吃三门；user 只通 user 门
	for _, p := range []string{"/probe/user", "/probe/admin", "/probe/super"} {
		if c, _ := call(t, ts, "GET", p, tok, ""); c != 200 {
			t.Fatalf("probe %s: %d", p, c)
		}
	}
	tok2 := m2["token"].(string)
	if c, _ := call(t, ts, "GET", "/probe/user", tok2, ""); c != 200 {
		t.Fatalf("u2 user: %d", c)
	}
	if c, _ := call(t, ts, "GET", "/probe/admin", tok2, ""); c != 403 {
		t.Fatalf("u2 admin: %d", c)
	}
	// 7) 限流：打满 20 次后 429
	for i := 0; i < 20; i++ {
		call(t, ts, "POST", "/api/auth/login", "", `{"email":"n@x.com","password":"12345678"}`)
	}
	if c, _ := call(t, ts, "POST", "/api/auth/login", "", `{"email":"n@x.com","password":"12345678"}`); c != 429 {
		t.Fatalf("ratelimit: %d", c)
	}
}
