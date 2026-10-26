package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type llmStore struct{}

func newLLMStore() *llmStore {
	if novaDB != nil {
		if err := dbProfilesMigrate(filepath.Join("data", "llm_profiles.json")); err != nil {
			log.Printf("[db] profiles migrate soft-fail: %v", err)
		}
	}
	return &llmStore{}
}

func dbProfilesMigrate(jsonPath string) error {
	if novaDB == nil {
		return nil
	}
	var n int
	if err := novaDB.QueryRow("SELECT COUNT(*) FROM profiles").Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	b, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil
	}
	var disk struct {
		Profiles []LLMProfile `json:"profiles"`
	}
	if err := json.Unmarshal(b, &disk); err != nil {
		return err
	}
	if len(disk.Profiles) == 0 {
		return nil
	}
	ps := disk.Profiles
	sort.Slice(ps, func(i, j int) bool { return ps[i].CreatedAt < ps[j].CreatedAt })
	tx, err := novaDB.Begin()
	if err != nil {
		return err
	}
	for i := range ps {
		p := ps[i]
		id := strings.TrimSpace(p.ID)
		if id == "" {
			id = fmt.Sprintf("p%d", time.Now().UnixNano()+int64(i))
		}
		ca := p.CreatedAt
		if ca == 0 {
			ca = time.Now().UnixMilli()
		}
		en := 0
		if p.Enabled {
			en = 1
		}
		if _, err := tx.Exec("INSERT INTO profiles(id,name,base_url,model,api_key,enabled,priority,created_at) VALUES(?,?,?,?,?,?,?,?)",
			id, p.Name, p.BaseURL, p.Model, p.APIKey, en, 100+i, ca); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	_ = os.Rename(jsonPath, jsonPath+".imported-bak")
	_ = os.Chmod(jsonPath+".imported-bak", 0o444)
	log.Printf("[db] profiles migrated: %d rows from %s", len(ps), jsonPath)
	return nil
}

func dbProfilesList(masked bool) []LLMProfile {
	out := []LLMProfile{}
	if novaDB == nil {
		return out
	}
	rows, err := novaDB.Query("SELECT id,name,base_url,model,api_key,enabled,created_at FROM profiles ORDER BY priority, created_at")
	if err != nil {
		log.Printf("[db] profiles list fail: %v", err)
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var p LLMProfile
		var en int
		if err := rows.Scan(&p.ID, &p.Name, &p.BaseURL, &p.Model, &p.APIKey, &en, &p.CreatedAt); err != nil {
			continue
		}
		p.Enabled = en != 0
		p.APIKey = decSecret(p.APIKey)
		out = append(out, p)
	}
	if masked {
		for i := range out {
			out[i].APIKey = maskSecret(out[i].APIKey)
		}
	}
	return out
}

func dbProfilesGet(id string) (LLMProfile, bool) {
	if novaDB == nil || strings.TrimSpace(id) == "" {
		return LLMProfile{}, false
	}
	var p LLMProfile
	var en int
	err := novaDB.QueryRow("SELECT id,name,base_url,model,api_key,enabled,created_at FROM profiles WHERE id=?", strings.TrimSpace(id)).
		Scan(&p.ID, &p.Name, &p.BaseURL, &p.Model, &p.APIKey, &en, &p.CreatedAt)
	if err != nil {
		return LLMProfile{}, false
	}
	p.Enabled = en != 0
	p.APIKey = decSecret(p.APIKey)
	return p, true
}

func dbProfilesUpsert(p LLMProfile) LLMProfile {
	if strings.TrimSpace(p.ID) == "" {
		p.ID = fmt.Sprintf("p%d", time.Now().UnixNano())
		p.CreatedAt = time.Now().UnixMilli()
	}
	if novaDB == nil {
		return p
	}
	var oldKey string
	var oldCA int64
	err := novaDB.QueryRow("SELECT api_key, created_at FROM profiles WHERE id=?", p.ID).Scan(&oldKey, &oldCA)
	if err == nil {
		if strings.TrimSpace(p.APIKey) == "" {
			p.APIKey = decSecret(oldKey)
		}
		p.CreatedAt = oldCA
		en := 0
		if p.Enabled {
			en = 1
		}
		if _, err := novaDB.Exec("UPDATE profiles SET name=?, base_url=?, model=?, api_key=?, enabled=? WHERE id=?",
			p.Name, p.BaseURL, p.Model, encSecret(p.APIKey), en, p.ID); err != nil {
			log.Printf("[db] profile update fail: %v", err)
		}
		return p
	}
	if p.CreatedAt == 0 {
		p.CreatedAt = time.Now().UnixMilli()
	}
	pri := 100
	if err := novaDB.QueryRow("SELECT COALESCE(MAX(priority),99)+1 FROM profiles").Scan(&pri); err != nil {
		pri = 100
	}
	en := 0
	if p.Enabled {
		en = 1
	}
	if _, err := novaDB.Exec("INSERT INTO profiles(id,name,base_url,model,api_key,enabled,priority,created_at) VALUES(?,?,?,?,?,?,?,?)",
		p.ID, p.Name, p.BaseURL, p.Model, encSecret(p.APIKey), en, pri, p.CreatedAt); err != nil {
		log.Printf("[db] profile insert fail: %v", err)
	}
	return p
}

func dbProfilesDel(id string) bool {
	if novaDB == nil || strings.TrimSpace(id) == "" {
		return false
	}
	res, err := novaDB.Exec("DELETE FROM profiles WHERE id=?", strings.TrimSpace(id))
	if err != nil {
		return false
	}
	n, _ := res.RowsAffected()
	return n > 0
}

func (s *llmStore) list() []LLMProfile {
	return dbProfilesList(true)
}

func (s *llmStore) listChat() []LLMProfile {
	return dbProfilesList(false)
}

func (s *llmStore) upsert(p LLMProfile) LLMProfile {
	return dbProfilesUpsert(p)
}

func (s *llmStore) del(id string) bool {
	return dbProfilesDel(id)
}

func (s *llmStore) get(id string) (LLMProfile, bool) {
	return dbProfilesGet(id)
}
