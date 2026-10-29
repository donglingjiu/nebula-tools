// Package admin —— 后台域: LLM密钥(密文存储)/功能池/AI运营总控
// R1: 独立JSON存储(data/admin.json); R2 迁移SQLite
package admin

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type Profile struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	BaseURL   string `json:"base_url"`
	Model     string `json:"model"`
	APIKeyEnc string `json:"api_key_enc"` // AES-GCM 密文(b64), 明文不出库
	Enabled   bool   `json:"enabled"`
	CreatedAt string `json:"created_at"`
}

type Feature struct {
	ID       string  `json:"id"`
	Key      string  `json:"key"`
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Enabled  bool    `json:"enabled"`
	IsPaid   bool    `json:"is_paid"`
	Price    float64 `json:"price"`
}

type Ops struct {
	AIEnabled        bool   `json:"ai_enabled"`
	PaidMode         bool   `json:"paid_mode"`
	DefaultProfileID string `json:"default_profile_id"`
}

type Store struct {
	mu       sync.Mutex
	path     string
	Profiles []Profile `json:"profiles"`
	Features []Feature `json:"features"`
	Ops      Ops       `json:"ops"`
}

func newID() string { b := make([]byte, 8); _, _ = rand.Read(b); return hex.EncodeToString(b) }

func OpenStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	s := &Store{path: filepath.Join(dir, "admin.json"), Ops: Ops{AIEnabled: true}}
	b, err := os.ReadFile(s.path)
	if err == nil {
		if json.Unmarshal(b, s) != nil {
			return nil, errors.New("admin.json 损坏")
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if s.Features == nil {
		s.Features = []Feature{
			{ID: newID(), Key: "ai_studio", Name: "AI 工作台", Category: "AI", Enabled: true},
			{ID: newID(), Key: "ai_chat", Name: "AI 对话", Category: "AI", Enabled: true},
			{ID: newID(), Key: "doc_tools", Name: "文档工具集", Category: "文档", Enabled: true},
			{ID: newID(), Key: "excel_insight", Name: "Excel 洞察", Category: "表格", Enabled: true},
			{ID: newID(), Key: "data_insight", Name: "数据洞察", Category: "数据", Enabled: true},
			{ID: newID(), Key: "ppt_tools", Name: "PPT 助手", Category: "PPT", Enabled: true},
			{ID: newID(), Key: "mind_tools", Name: "脑图", Category: "脑图", Enabled: true},
		}
		if err := s.saveLocked(); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *Store) saveLocked() error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
