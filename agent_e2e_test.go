package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestAgentRunE2E(t *testing.T) {
	tc := start(t)

	d, err := openDBAt(filepath.Join(t.TempDir(), "nova.db"))
	if err != nil {
		t.Fatal(err)
	}
	savedDB := novaDB
	novaDB = d
	t.Cleanup(func() { novaDB = savedDB; d.Close() })

	mockCalls := 0
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockCalls++
		w.Header().Set("Content-Type", "application/json")
		if mockCalls == 1 {
			fmt.Fprint(w, `{"choices":[{"message":{"content":"{\"tool\":\"dev.json\",\"input\":\"{\\\"a\\\":1}\"}"}}],"usage":{"total_tokens":10}}`)
			return
		}
		fmt.Fprint(w, `{"choices":[{"message":{"content":"JSON已格式化完毕,缩进2空格"}}],"usage":{"total_tokens":8}}`)
	}))
	defer mock.Close()

	tc.ok("POST", "/api/auth/register", map[string]string{"username": "agent1", "password": "Ag123456"})
	tc.ok("POST", "/api/llm/profiles", map[string]any{
		"name": "mock", "baseUrl": mock.URL, "apiKey": "k", "model": "mock", "enabled": true,
	})

	code, m := tc.do("POST", "/api/agent/run", map[string]any{"input": "帮我格式化 JSON", "maxRounds": 3})
	if code != 200 {
		t.Fatalf("code=%d body=%v", code, m)
	}
	if v, _ := m["ok"].(bool); !v {
		t.Fatalf("ok=false err=%v", m["error"])
	}
	if mockCalls < 2 {
		t.Fatalf("mockCalls=%d want>=2 (tool-call+终答两轮)", mockCalls)
	}
	calls, _ := m["calls"].([]any)
	if len(calls) < 1 {
		t.Fatalf("calls empty: %v", m["calls"])
	}
	first, _ := calls[0].(map[string]any)
	if first["tool"] != "dev.json" {
		t.Fatalf("first.tool=%v want dev.json", first["tool"])
	}
	if okc, _ := first["ok"].(bool); !okc {
		t.Fatalf("first tool call failed: %v", first["output"])
	}
	ans, _ := m["answer"].(string)
	if !strings.Contains(ans, "格式化") {
		t.Fatalf("answer=%q", ans)
	}
	if cid, _ := m["conversationId"].(float64); cid <= 0 {
		t.Fatalf("conversationId=%v", m["conversationId"])
	}
	if rd, _ := m["rounds"].(float64); rd < 2 {
		t.Fatalf("rounds=%v want>=2", rd)
	}
}
