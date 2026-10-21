package main

import (
"encoding/json"
"net/http"
"net/http/httptest"
"testing"
"time"
)

func TestAlertOnToolFailure(t *testing.T) {
got := make(chan map[string]any, 1)
hook := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
var m map[string]any
json.NewDecoder(r.Body).Decode(&m)
got <- m
}))
defer hook.Close()

tc := start(t)
tc.ok("POST", "/api/auth/register", map[string]string{"username": "admin", "password": "admin123"})
tc.ok("POST", "/api/admin/settings", map[string]string{"alert_webhook": hook.URL})

code, r := tc.do("POST", "/api/run/dev.json", map[string]string{"input": "{bad"})
if code != 200 || r["ok"] != false {
t.Fatalf("run 应失败: %d %v", code, r)
}
select {
case m := <-got:
if m["event"] != "tool_failed" || m["tool"] != "dev.json" || m["user"] != "admin" {
t.Fatalf("alert payload mismatch: %v", m)
}
case <-time.After(2 * time.Second):
t.Fatal("webhook 未在 2s 内收到")
}
}
