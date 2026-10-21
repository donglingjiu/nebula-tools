package main

import (
"os"
"strings"
"testing"
)

func TestLLMKeyEncryptedAtRest(t *testing.T) {
_ = os.Remove("data/llm_profiles.json")
t.Cleanup(func() { _ = os.Remove("data/llm_profiles.json") })

s := newLLMStore()
s.upsert(LLMProfile{Name: "prov", BaseURL: "https://api.example.com", APIKey: "sk-supersecret-2123", Model: "m1", Enabled: true})

b, err := os.ReadFile("data/llm_profiles.json")
if err != nil {
t.Fatal(err)
}
if strings.Contains(string(b), "sk-supersecret-2123") || !strings.Contains(string(b), "enc1:") {
t.Fatal("落盘必须是密文")
}

// 重启回读：解密还原 + 出口打码
s2 := newLLMStore()
if len(s2.Profiles) != 1 || s2.Profiles[0].APIKey != "sk-supersecret-2123" {
t.Fatalf("解密回读失败: %+v", s2.Profiles)
}
if got := s2.list()[0].APIKey; got != "****2123" {
t.Fatalf("出口未打码: %q", got)
}

// 空密钥更新：保旧逻辑在解密域照常工作，二次落盘仍无明文
s2.upsert(LLMProfile{ID: s2.Profiles[0].ID, Name: "prov2"})
if s2.Profiles[0].APIKey != "sk-supersecret-2123" {
t.Fatal("空 key 更新丢了旧密钥")
}
b2, _ := os.ReadFile("data/llm_profiles.json")
if strings.Contains(string(b2), "sk-supersecret-2123") {
t.Fatal("二次保存泄漏明文")
}
}
