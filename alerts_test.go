package main

import (
"encoding/json"
"net/http"
"net/http/httptest"
"testing"
"time"
)

func TestSendAlert(t *testing.T) {
got := make(chan map[string]any, 1)
hook := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
var m map[string]any
json.NewDecoder(r.Body).Decode(&m)
got <- m
}))
defer hook.Close()

sendAlert(hook.URL, "pdf_merge", "alice", "boom")
select {
case m := <-got:
if m["event"] != "tool_failed" || m["tool"] != "pdf_merge" || m["user"] != "alice" || m["error"] != "boom" {
t.Fatalf("payload mismatch: %v", m)
}
case <-time.After(2 * time.Second):
t.Fatal("webhook not received within 2s")
}

sendAlert("", "x", "y", "z") // 空配置必须静默返回，不阻塞不panic
}
