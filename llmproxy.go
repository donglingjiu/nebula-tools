package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func llmWriteJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(v)
}

func llmReadJSON(r *http.Request, v interface{}) error {
	b, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func llmTrunc(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n]) + "..."
	}
	return s
}

func llmKeyHint(k string) string {
	k = strings.TrimSpace(k)
	if k == "" {
		return "(未设置)"
	}
	if len(k) <= 8 {
		return "****"
	}
	return k[:4] + "****" + k[len(k)-4:]
}

func llmMask(p LLMProfile) map[string]interface{} {
	host := ""
	if u, err := url.Parse(strings.TrimSpace(p.BaseURL)); err == nil {
		host = u.Host
	}
	return map[string]interface{}{
		"id": p.ID, "name": p.Name, "baseUrl": p.BaseURL, "model": p.Model,
		"enabled": p.Enabled, "createdAt": p.CreatedAt,
		"keyHint": llmKeyHint(p.APIKey), "host": host,
	}
}

func llmNormalizeBase(b string) (string, error) {
	b = strings.TrimSpace(b)
	b = strings.TrimRight(b, "/")
	if !strings.HasPrefix(b, "http://") && !strings.HasPrefix(b, "https://") {
		return "", errors.New("baseUrl 必须以 http:// 或 https:// 开头")
	}
	if !strings.HasSuffix(b, "/chat/completions") {
		b = b + "/chat/completions"
	}
	u, err := url.Parse(b)
	if err != nil || u.Host == "" {
		return "", errors.New("baseUrl 无效")
	}
	return b, nil
}

func llmCallU(client *http.Client, p LLMProfile, msgs []map[string]string, maxTokens int) (string, int64, int, error) {
	if strings.TrimSpace(p.BaseURL) == "" || strings.TrimSpace(p.APIKey) == "" {
		return "", 0, 0, errors.New("档案不完整: baseUrl 与 apiKey 必填")
	}
	if strings.TrimSpace(p.Model) == "" {
		return "", 0, 0, errors.New("档案不完整: model 必填")
	}
	endpoint, err := llmNormalizeBase(p.BaseURL)
	if err != nil {
		return "", 0, 0, err
	}
	if maxTokens <= 0 {
		maxTokens = 1024
	}
	payload := map[string]interface{}{
		"model": p.Model, "messages": msgs, "max_tokens": maxTokens, "temperature": 0.2,
	}
	jb, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(jb))
	if err != nil {
		return "", 0, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)
	t0 := time.Now()
	resp, err := client.Do(req)
	ms := int(time.Since(t0).Milliseconds())
	if err != nil {
		return "", 0, ms, err
	}
	defer resp.Body.Close()
	rb, rerr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if rerr != nil {
		return "", 0, ms, rerr
	}
	if resp.StatusCode >= 300 {
		return "", 0, ms, errors.New("HTTP " + fmt.Sprint(resp.StatusCode) + ": " + llmTrunc(string(rb), 180))
	}
	var out struct {
		Usage struct {
			PromptTokens     int64 `json:"prompt_tokens"`
			CompletionTokens int64 `json:"completion_tokens"`
			TotalTokens      int64 `json:"total_tokens"`
		} `json:"usage"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rb, &out); err != nil {
		return "", 0, ms, errors.New("响应非 JSON: " + llmTrunc(string(rb), 120))
	}
	if out.Error != nil && out.Error.Message != "" {
		return "", 0, ms, errors.New(out.Error.Message)
	}
	if len(out.Choices) == 0 || strings.TrimSpace(out.Choices[0].Message.Content) == "" {
		return "", 0, ms, errors.New("模型返回空内容")
	}
	return out.Choices[0].Message.Content, out.Usage.TotalTokens, ms, nil
}

func registerLLMRoutes(mux *http.ServeMux, s *srv) {

	mux.HandleFunc("GET /api/llm/profiles", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "gone: /api/llm retired by OPT-017/018; use /api/ai/*", http.StatusGone)
	})

	mux.HandleFunc("POST /api/llm/profiles", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "gone: /api/llm retired by OPT-017/018; use /api/ai/*", http.StatusGone)
	})

	mux.HandleFunc("POST /api/llm/profiles/delete", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "gone: /api/llm retired by OPT-017/018; use /api/ai/*", http.StatusGone)
	})

	mux.HandleFunc("POST /api/llm/test", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "gone: /api/llm retired by OPT-017/018; use /api/ai/*", http.StatusGone)
	})

	mux.HandleFunc("POST /api/llm/chat", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "gone: /api/llm retired by OPT-017/018; use /api/ai/*", http.StatusGone)
	})
}

func llmCall(client *http.Client, p LLMProfile, msgs []map[string]string, maxTokens int) (string, int, error) {
	c, _, ms, err := llmCallU(client, p, msgs, maxTokens)
	return c, ms, err
}

func llmLastUser(msgs []map[string]string) string {
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i]["role"] == "user" {
			return msgs[i]["content"]
		}
	}
	return ""
}
