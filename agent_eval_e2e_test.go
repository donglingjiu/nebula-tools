package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAgentEvalE2E(t *testing.T) {
	tc := start(t)

	kw2tool := map[string]string{
		"格式化": "dev.json", "CSV 表格": "data.json2csv", "转成 JSON": "data.csv2json",
		"MD5": "dev.hash", "多少行": "dev.count", "差异": "dev.diff",
		"全大写": "dev.case", "URL 编码": "dev.encode",
	}
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var req struct {
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		_ = json.Unmarshal(b, &req)
		last := ""
		if len(req.Messages) > 0 {
			last = req.Messages[len(req.Messages)-1].Content
		}
		if strings.HasPrefix(last, "工具结果: ") {
			fmt.Fprint(w, `{"choices":[{"message":{"content":"已完成"}}],"usage":{"total_tokens":8}}`)
			return
		}
		for kw, tool := range kw2tool {
			if strings.Contains(last, kw) {
				fmt.Fprintf(w, `{"choices":[{"message":{"content":"{\"tool\":\"%s\",\"input\":\"x\"}"}}],"usage":{"total_tokens":10}}`, tool)
				return
			}
		}
		fmt.Fprint(w, `{"choices":[{"message":{"content":"未匹配"}}],"usage":{"total_tokens":10}}`)
	}))
	defer mock.Close()

	tc.ok("POST", "/api/auth/register", map[string]string{"username": "admin", "password": "admin123"})
	dbProfilesUpsert(LLMProfile{Name: "mock", BaseURL: mock.URL, APIKey: "k", Model: "mock", Enabled: true})

	code, m := tc.do("POST", "/api/agent/eval", map[string]any{})
	if code != 200 {
		t.Fatalf("code=%d body=%v", code, m)
	}
	total, _ := m["total"].(float64)
	if total != 8 {
		t.Fatalf("total=%v want 8", total)
	}
	match, _ := m["toolMatch"].(float64)
	if match != 8 {
		t.Fatalf("toolMatch=%v want 8, results=%v", match, m["results"])
	}
}
