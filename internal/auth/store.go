// Package auth —— 认证域：用户存储/密码哈希/JWT/HTTP 层
// R1 决策：JSON 文件存储（零依赖）；R2 换 SQLite(modernc.org/sqlite) 时仅改本文件
package auth

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const pbkdf2Iters = 210000

var ErrExists = errors.New("邮箱已注册")

type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"` // user | admin | super
	Hash      string `json:"hash"` // pbkdf2-sha256$iter$salt$key
	CreatedAt string `json:"created_at"`
	Disabled  bool   `json:"disabled"`
}

type Store struct {
	mu    sync.Mutex
	path  string
	users []User
}

func OpenStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	s := &Store{path: filepath.Join(dir, "users.json")}
	b, err := os.ReadFile(s.path)
	if err == nil {
		if json.Unmarshal(b, &s.users) != nil {
			return nil, errors.New("users.json 损坏")
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

func (s *Store) saveLocked() error {
	b, err := json.MarshalIndent(s.users, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if os.WriteFile(tmp, b, 0o600) != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func newID() string { b := make([]byte, 8); _, _ = rand.Read(b); return hex.EncodeToString(b) }

func HashPassword(pw string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key, err := pbkdf2.Key(sha256.New, pw, salt, pbkdf2Iters, 32)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("pbkdf2-sha256$%d$%s$%s", pbkdf2Iters, hex.EncodeToString(salt), hex.EncodeToString(key)), nil
}

func VerifyPassword(hash, pw string) bool {
	parts := strings.Split(hash, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" {
		return false
	}
	iters, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}
	salt, e1 := hex.DecodeString(parts[2])
	want, e2 := hex.DecodeString(parts[3])
	if e1 != nil || e2 != nil {
		return false
	}
	key, err := pbkdf2.Key(sha256.New, pw, salt, iters, len(want))
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(key, want) == 1
}

func (s *Store) Count() int { s.mu.Lock(); defer s.mu.Unlock(); return len(s.users) }

func (s *Store) ByEmail(email string) (User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, u := range s.users {
		if u.Email == email {
			return u, true
		}
	}
	return User{}, false
}

// Create 注册；首个用户自动 super
func (s *Store) Create(email, pw string) (*User, error) {
	h, err := HashPassword(pw)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.users {
		if s.users[i].Email == email {
			return nil, ErrExists
		}
	}
	role := "user"
	if len(s.users) == 0 {
		role = "super"
	}
	u := &User{ID: newID(), Email: email, Role: role, Hash: h, CreatedAt: time.Now().Format(time.RFC3339)}
	s.users = append(s.users, *u)
	if err := s.saveLocked(); err != nil {
		return nil, err
	}
	return u, nil
}
