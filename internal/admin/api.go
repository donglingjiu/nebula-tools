package admin

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nova/internal/auth"
)

type Service struct {
	st   *Store
	aead cipher.AEAD
}

// Mount 挂载 /api/admin/*; asvc 提供三级角色门; ops 写仅 super
func Mount(mux *http.ServeMux, dir string, asvc *auth.Service) error {
	if asvc == nil {
		return errors.New("auth service 为 nil, admin 路由未挂载")
	}
	st, err := OpenStore(dir)
	if err != nil {
		return err
	}
	aead, err := loadAEAD(dir)
	if err != nil {
		return err
	}
	s := &Service{st: st, aead: aead}
	mux.HandleFunc("GET /api/admin/llm/profiles", asvc.RequireRole("admin", s.listProfiles))
	mux.HandleFunc("POST /api/admin/llm/profiles", asvc.RequireRole("admin", s.createProfile))
	mux.HandleFunc("PUT /api/admin/llm/profiles/{id}", asvc.RequireRole("admin", s.updateProfile))
	mux.HandleFunc("DELETE /api/admin/llm/profiles/{id}", asvc.RequireRole("admin", s.deleteProfile))
	mux.HandleFunc("GET /api/admin/ai/ops", asvc.RequireRole("admin", s.getOps))
	mux.HandleFunc("PUT /api/admin/ai/ops", asvc.RequireRole("super", s.putOps))
	mux.HandleFunc("GET /api/admin/features", asvc.RequireRole("admin", s.listFeatures))
	mux.HandleFunc("PUT /api/admin/features/{id}", asvc.RequireRole("admin", s.putFeature))
	return nil
}

// loadAEAD 复用 auth 的 data/secret.key 派生 AES-256-GCM 密钥
func loadAEAD(dir string) (cipher.AEAD, error) {
	b, err := os.ReadFile(filepath.Join(dir, "secret.key"))
	if err != nil || len(b) < 16 {
		return nil, errors.New("secret.key 缺失(应先挂载 auth)")
	}
	k := sha256.Sum256(b)
	blk, err := aes.NewCipher(k[:])
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(blk)
}

func (s *Service) encKey(plain string) string {
	if plain == "" {
		return ""
	}
	n := make([]byte, s.aead.NonceSize())
	_, _ = rand.Read(n)
	ct := s.aead.Seal(nil, n, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(append(n, ct...))
}

func (s *Service) decKey(enc string) string {
	raw, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return ""
	}
	ns := s.aead.NonceSize()
	if len(raw) < ns {
		return ""
	}
	pt, err := s.aead.Open(nil, raw[:ns], raw[ns:], nil)
	if err != nil {
		return ""
	}
	return string(pt)
}

func keyTail(k string) string {
	if len(k) <= 4 {
		return "****"
	}
	return "****" + k[len(k)-4:]
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"message": msg})
}

func (s *Service) listProfiles(w http.ResponseWriter, r *http.Request) {
	s.st.mu.Lock()
	defer s.st.mu.Unlock()
	arr := make([]map[string]any, 0, len(s.st.Profiles))
	for _, p := range s.st.Profiles {
		kt := keyTail(s.decKey(p.APIKeyEnc))
		arr = append(arr, map[string]any{
			"id": p.ID, "name": p.Name, "base_url": p.BaseURL, "model": p.Model,
			"enabled": p.Enabled, "created_at": p.CreatedAt, "key_tail": kt, "keyTail": kt, "api_key_mask": kt,
		})
	}
	writeJSON(w, 200, map[string]any{"profiles": arr})
}

func (s *Service) createProfile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		BaseURL string `json:"base_url"`
		Model   string `json:"model"`
		APIKey  string `json:"api_key"`
		Enabled *bool  `json:"enabled"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || strings.TrimSpace(req.Name) == "" ||
		strings.TrimSpace(req.Model) == "" || req.APIKey == "" {

		return
	}
	en := true
	if req.Enabled != nil {
		en = *req.Enabled
	}
	id := newID()
	s.st.mu.Lock()
	s.st.Profiles = append(s.st.Profiles, Profile{
		ID: id, Name: strings.TrimSpace(req.Name), BaseURL: strings.TrimSpace(req.BaseURL),
		Model: strings.TrimSpace(req.Model), APIKeyEnc: s.encKey(req.APIKey),
		Enabled: en, CreatedAt: time.Now().Format(time.RFC3339),
	})
	err := s.st.saveLocked()
	s.st.mu.Unlock()
	if err != nil {
		fail(w, 500, "保存失败")
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "id": id})
}

func (s *Service) updateProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Name    *string `json:"name"`
		BaseURL *string `json:"base_url"`
		Model   *string `json:"model"`
		APIKey  *string `json:"api_key"`
		Enabled *bool   `json:"enabled"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		fail(w, 400, "参数错误")
		return
	}
	s.st.mu.Lock()
	defer s.st.mu.Unlock()
	for i := range s.st.Profiles {
		if s.st.Profiles[i].ID == id {
			if req.Name != nil {
				s.st.Profiles[i].Name = *req.Name
			}
			if req.BaseURL != nil {
				s.st.Profiles[i].BaseURL = *req.BaseURL
			}
			if req.Model != nil {
				s.st.Profiles[i].Model = *req.Model
			}
			if req.APIKey != nil && *req.APIKey != "" {
				s.st.Profiles[i].APIKeyEnc = s.encKey(*req.APIKey)
			}
			if req.Enabled != nil {
				s.st.Profiles[i].Enabled = *req.Enabled
			}
			if err := s.st.saveLocked(); err != nil {
				fail(w, 500, "保存失败")
				return
			}
			writeJSON(w, 200, map[string]any{"ok": true})
			return
		}
	}
	fail(w, 404, "Profile 不存在")
}

func (s *Service) deleteProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.st.mu.Lock()
	defer s.st.mu.Unlock()
	for i := range s.st.Profiles {
		if s.st.Profiles[i].ID == id {
			s.st.Profiles = append(s.st.Profiles[:i], s.st.Profiles[i+1:]...)
			if err := s.st.saveLocked(); err != nil {
				fail(w, 500, "保存失败")
				return
			}
			writeJSON(w, 200, map[string]any{"ok": true})
			return
		}
	}
	fail(w, 404, "Profile 不存在")
}

func (s *Service) getOps(w http.ResponseWriter, r *http.Request) {
	s.st.mu.Lock()
	o := s.st.Ops
	s.st.mu.Unlock()
	writeJSON(w, 200, map[string]any{
		"switches": map[string]any{"ai_enabled": o.AIEnabled, "paid_mode": o.PaidMode, "default_profile_id": o.DefaultProfileID},
		"today":    map[string]any{"tokens": 0, "calls": 0, "success_rate": 100, "avg_ms": 0},
		"recent":   []any{},
	})
}

func (s *Service) putOps(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AIEnabled        *bool   `json:"ai_enabled"`
		PaidMode         *bool   `json:"paid_mode"`
		DefaultProfileID *string `json:"default_profile_id"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		fail(w, 400, "参数错误")
		return
	}
	s.st.mu.Lock()
	defer s.st.mu.Unlock()
	if req.DefaultProfileID != nil && *req.DefaultProfileID != "" {
		found := false
		for _, p := range s.st.Profiles {
			if p.ID == *req.DefaultProfileID {
				found = true
				break
			}
		}
		if !found {
			fail(w, 400, "默认模型不存在")
			return
		}
	}
	if req.AIEnabled != nil {
		s.st.Ops.AIEnabled = *req.AIEnabled
	}
	if req.PaidMode != nil {
		s.st.Ops.PaidMode = *req.PaidMode
	}
	if req.DefaultProfileID != nil {
		s.st.Ops.DefaultProfileID = *req.DefaultProfileID
	}
	if err := s.st.saveLocked(); err != nil {
		fail(w, 500, "保存失败")
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

func (s *Service) listFeatures(w http.ResponseWriter, r *http.Request) {
	s.st.mu.Lock()
	defer s.st.mu.Unlock()
	arr := make([]map[string]any, 0, len(s.st.Features))
	for _, f := range s.st.Features {
		arr = append(arr, map[string]any{
			"id": f.ID, "key": f.Key, "name": f.Name, "category": f.Category,
			"enabled": f.Enabled, "is_paid": f.IsPaid, "price": f.Price,
			"calls_7d": 0, "success_rate": 100,
		})
	}
	writeJSON(w, 200, map[string]any{"features": arr})
}

func (s *Service) putFeature(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Enabled *bool    `json:"enabled"`
		IsPaid  *bool    `json:"is_paid"`
		Price   *float64 `json:"price"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		fail(w, 400, "参数错误")
		return
	}
	s.st.mu.Lock()
	defer s.st.mu.Unlock()
	for i := range s.st.Features {
		if s.st.Features[i].ID == id {
			if req.Enabled != nil {
				s.st.Features[i].Enabled = *req.Enabled
			}
			if req.IsPaid != nil {
				s.st.Features[i].IsPaid = *req.IsPaid
			}
			if req.Price != nil {
				s.st.Features[i].Price = *req.Price
			}
			if err := s.st.saveLocked(); err != nil {
				fail(w, 500, "保存失败")
				return
			}
			writeJSON(w, 200, map[string]any{"ok": true})
			return
		}
	}
	fail(w, 404, "功能不存在")
}
