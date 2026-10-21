package main

import (
"bytes"
"encoding/json"
"errors"
"fmt"
"io"
"net/http"
"net/url"
"os"
"path/filepath"
"sort"
"strings"
"sync"
"time"
)

type LLMProfile struct {
ID        string `json:"id"`
Name      string `json:"name"`
BaseURL   string `json:"baseUrl"`
APIKey    string `json:"apiKey"`
Model     string `json:"model"`
Enabled   bool   `json:"enabled"`
CreatedAt int64  `json:"createdAt"`
}

type llmStore struct {
mu       sync.Mutex
path     string
Profiles []LLMProfile `json:"profiles"`
}

func newLLMStore() *llmStore {
s := &llmStore{path: filepath.Join("data", "llm_profiles.json")}
b, err := os.ReadFile(s.path)
if err == nil {
_ = json.Unmarshal(b, s)
		for i := range s.Profiles {
			s.Profiles[i].APIKey = decSecret(s.Profiles[i].APIKey)
		}
}
if s.Profiles == nil {
s.Profiles = []LLMProfile{}
}
return s
}

func (s *llmStore) flushLocked() {
	disk := llmDisk{Profiles: make([]LLMProfile, len(s.Profiles))}
	for i, p := range s.Profiles {
		p.APIKey = encSecret(p.APIKey)
		disk.Profiles[i] = p
	}
	b, _ := json.MarshalIndent(disk, "", "  ")
_ = os.MkdirAll(filepath.Dir(s.path), 0o755)
_ = os.WriteFile(s.path, b, 0o600)
}

func (s *llmStore) list() []LLMProfile {
s.mu.Lock()
defer s.mu.Unlock()
out := make([]LLMProfile, len(s.Profiles))
copy(out, s.Profiles)
sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt < out[j].CreatedAt })
	for i := range out {
		out[i].APIKey = maskSecret(out[i].APIKey)
	}
return out
}

func (s *llmStore) upsert(p LLMProfile) LLMProfile {
s.mu.Lock()
defer s.mu.Unlock()
if strings.TrimSpace(p.ID) == "" {
p.ID = fmt.Sprintf("p%d", time.Now().UnixNano())
p.CreatedAt = time.Now().UnixMilli()
}
for i := range s.Profiles {
if s.Profiles[i].ID == p.ID {
if strings.TrimSpace(p.APIKey) == "" {
p.APIKey = s.Profiles[i].APIKey
}
p.CreatedAt = s.Profiles[i].CreatedAt
s.Profiles[i] = p
s.flushLocked()
return p
}
}
if p.CreatedAt == 0 {
p.CreatedAt = time.Now().UnixMilli()
}
s.Profiles = append(s.Profiles, p)
s.flushLocked()
return p
}

func (s *llmStore) del(id string) bool {
s.mu.Lock()
defer s.mu.Unlock()
for i := range s.Profiles {
if s.Profiles[i].ID == id {
s.Profiles = append(s.Profiles[:i], s.Profiles[i+1:]...)
s.flushLocked()
return true
}
}
return false
}

func (s *llmStore) get(id string) (LLMProfile, bool) {
s.mu.Lock()
defer s.mu.Unlock()
for i := range s.Profiles {
if s.Profiles[i].ID == id {
return s.Profiles[i], true
}
}
return LLMProfile{}, false
}

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

func llmCall(client *http.Client, p LLMProfile, msgs []map[string]string, maxTokens int) (string, int, error) {
if strings.TrimSpace(p.BaseURL) == "" || strings.TrimSpace(p.APIKey) == "" {
return "", 0, errors.New("档案不完整: baseUrl 与 apiKey 必填")
}
if strings.TrimSpace(p.Model) == "" {
return "", 0, errors.New("档案不完整: model 必填")
}
endpoint, err := llmNormalizeBase(p.BaseURL)
if err != nil {
return "", 0, err
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
return "", 0, err
}
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", "Bearer "+p.APIKey)
t0 := time.Now()
resp, err := client.Do(req)
ms := int(time.Since(t0).Milliseconds())
if err != nil {
return "", ms, err
}
defer resp.Body.Close()
rb, rerr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
if rerr != nil {
return "", ms, rerr
}
if resp.StatusCode >= 300 {
return "", ms, errors.New("HTTP " + fmt.Sprint(resp.StatusCode) + ": " + llmTrunc(string(rb), 180))
}
var out struct {
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
return "", ms, errors.New("响应非 JSON: " + llmTrunc(string(rb), 120))
}
if out.Error != nil && out.Error.Message != "" {
return "", ms, errors.New(out.Error.Message)
}
if len(out.Choices) == 0 || strings.TrimSpace(out.Choices[0].Message.Content) == "" {
return "", ms, errors.New("模型返回空内容")
}
return out.Choices[0].Message.Content, ms, nil
}

func registerLLMRoutes(mux *http.ServeMux) {
st := newLLMStore()
client := &http.Client{Timeout: 90 * time.Second}

mux.HandleFunc("GET /api/llm/profiles", func(w http.ResponseWriter, r *http.Request) {
ps := st.list()
arr := make([]map[string]interface{}, 0, len(ps))
for _, p := range ps {
arr = append(arr, llmMask(p))
}
llmWriteJSON(w, map[string]interface{}{"ok": true, "profiles": arr})
})

mux.HandleFunc("POST /api/llm/profiles", func(w http.ResponseWriter, r *http.Request) {
var in LLMProfile
if err := llmReadJSON(r, &in); err != nil {
llmWriteJSON(w, map[string]interface{}{"ok": false, "error": "请求体无效"})
return
}
in.Name = strings.TrimSpace(in.Name)
if in.Name == "" {
llmWriteJSON(w, map[string]interface{}{"ok": false, "error": "名称必填"})
return
}
if _, err := llmNormalizeBase(in.BaseURL); err != nil {
llmWriteJSON(w, map[string]interface{}{"ok": false, "error": err.Error()})
return
}
in.BaseURL = strings.TrimRight(strings.TrimSpace(in.BaseURL), "/")
in.Model = strings.TrimSpace(in.Model)
if in.Model == "" {
llmWriteJSON(w, map[string]interface{}{"ok": false, "error": "model 必填(如 glm-4-flash)"})
return
}
p := st.upsert(in)
llmWriteJSON(w, map[string]interface{}{"ok": true, "profile": llmMask(p)})
})

mux.HandleFunc("POST /api/llm/profiles/delete", func(w http.ResponseWriter, r *http.Request) {
var in struct {
ID string `json:"id"`
}
if err := llmReadJSON(r, &in); err != nil || in.ID == "" {
llmWriteJSON(w, map[string]interface{}{"ok": false, "error": "缺少 id"})
return
}
llmWriteJSON(w, map[string]interface{}{"ok": st.del(in.ID)})
})

mux.HandleFunc("POST /api/llm/test", func(w http.ResponseWriter, r *http.Request) {
var in struct {
ID string `json:"id"`
}
if err := llmReadJSON(r, &in); err != nil {
llmWriteJSON(w, map[string]interface{}{"ok": false, "error": "请求体无效"})
return
}
p, found := st.get(in.ID)
if !found {
llmWriteJSON(w, map[string]interface{}{"ok": false, "error": "档案不存在"})
return
}
msgs := []map[string]string{{"role": "user", "content": "请只回复两个字:正常"}}
content, ms, err := llmCall(client, p, msgs, 16)
if err != nil {
llmWriteJSON(w, map[string]interface{}{"ok": false, "error": llmTrunc(err.Error(), 200), "ms": ms})
return
}
llmWriteJSON(w, map[string]interface{}{"ok": true, "ms": ms, "content": llmTrunc(content, 60)})
})

mux.HandleFunc("POST /api/llm/chat", func(w http.ResponseWriter, r *http.Request) {
var in struct {
ProfileID string              `json:"profileId"`
Messages  []map[string]string `json:"messages"`
MaxTokens int                 `json:"maxTokens"`
}
if err := llmReadJSON(r, &in); err != nil || len(in.Messages) == 0 {
llmWriteJSON(w, map[string]interface{}{"ok": false, "error": "messages 必填"})
return
}
if len(in.Messages) > 40 {
in.Messages = in.Messages[len(in.Messages)-40:]
}
for i := range in.Messages {
if in.Messages[i]["role"] == "" {
in.Messages[i]["role"] = "user"
}
in.Messages[i]["content"] = llmTrunc(in.Messages[i]["content"], 16000)
}
var cands []LLMProfile
if in.ProfileID != "" {
if p, ok := st.get(in.ProfileID); ok {
cands = append(cands, p)
}
}
for _, p := range st.list() {
if p.Enabled && p.ID != in.ProfileID {
cands = append(cands, p)
}
}
if len(cands) == 0 {
llmWriteJSON(w, map[string]interface{}{"ok": false, "error": "没有可用档案,请先在模型设置中添加并启用"})
return
}
var errs []string
for _, p := range cands {
content, ms, err := llmCall(client, p, in.Messages, in.MaxTokens)
if err == nil {
llmWriteJSON(w, map[string]interface{}{"ok": true, "content": content, "profile": p.Name, "ms": ms})
return
}
errs = append(errs, p.Name+": "+llmTrunc(err.Error(), 120))
}
llmWriteJSON(w, map[string]interface{}{"ok": false, "error": strings.Join(errs, " | ")})
})
}
