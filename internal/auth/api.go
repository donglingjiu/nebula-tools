package auth

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Service struct{ store *Store }

// Mount 挂载 /api/auth/* 路由并打开存储；dir=数据目录
func Mount(mux *http.ServeMux, dir, prefix string) (*Service, error) {
	if err := initSecret(dir); err != nil {
		return nil, err
	}
	st, err := OpenStore(dir)
	if err != nil {
		return nil, err
	}
	s := &Service{store: st}
	mux.HandleFunc("POST "+prefix+"/register", s.register)
	mux.HandleFunc("POST "+prefix+"/login", s.login)
	mux.HandleFunc("GET "+prefix+"/me", s.me)
	return s, nil
}

var emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

type creds struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type limiter struct {
	mu sync.Mutex
	m  map[string][]time.Time
}

var lim = &limiter{m: map[string][]time.Time{}}

func (l *limiter) allow(key string, max int, win time.Duration) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	arr := l.m[key][:0]
	for _, t := range l.m[key] {
		if now.Sub(t) < win {
			arr = append(arr, t)
		}
	}
	if len(arr) >= max {
		l.m[key] = arr
		return false
	}
	l.m[key] = append(arr, now)
	return true
}

func clientIP(r *http.Request) string {
	h, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return h
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"message": msg})
}

func pubUser(u *User) map[string]any {
	return map[string]any{"email": u.Email, "role": u.Role, "created_at": u.CreatedAt}
}

func (s *Service) register(w http.ResponseWriter, r *http.Request) {
	if !lim.allow("reg:"+clientIP(r), 20, time.Minute) {
		fail(w, 429, "请求过于频繁，请稍后再试")
		return
	}
	var c creds
	if json.NewDecoder(r.Body).Decode(&c) != nil {
		fail(w, 400, "参数错误")
		return
	}
	c.Email = strings.ToLower(strings.TrimSpace(c.Email))
	if !emailRe.MatchString(c.Email) {
		fail(w, 400, "邮箱格式不正确")
		return
	}
	if len(c.Password) < 8 {
		fail(w, 400, "密码至少 8 位")
		return
	}
	u, err := s.store.Create(c.Email, c.Password)
	if err == ErrExists {
		fail(w, 409, "该邮箱已注册，请直接登录")
		return
	}
	if err != nil {
		fail(w, 500, "注册失败")
		return
	}
	tok, err := signToken(u, 7*24*time.Hour)
	if err != nil {
		fail(w, 500, "注册失败")
		return
	}
	writeJSON(w, 200, map[string]any{"token": tok, "user": pubUser(u)})
}

func (s *Service) login(w http.ResponseWriter, r *http.Request) {
	if !lim.allow("login:"+clientIP(r), 20, time.Minute) {
		fail(w, 429, "请求过于频繁，请稍后再试")
		return
	}
	var c creds
	if json.NewDecoder(r.Body).Decode(&c) != nil {
		fail(w, 400, "参数错误")
		return
	}
	u, ok := s.store.ByEmail(strings.ToLower(strings.TrimSpace(c.Email)))
	if !ok || u.Disabled || !VerifyPassword(u.Hash, c.Password) {
		fail(w, 401, "邮箱或密码错误")
		return
	}
	tok, err := signToken(&u, 7*24*time.Hour)
	if err != nil {
		fail(w, 500, "登录失败")
		return
	}
	writeJSON(w, 200, map[string]any{"token": tok, "user": pubUser(&u)})
}

func (s *Service) me(w http.ResponseWriter, r *http.Request) {
	u := s.authenticate(r)
	if u == nil {
		fail(w, 401, "请先登录")
		return
	}
	writeJSON(w, 200, map[string]any{"user": pubUser(u)})
}

var roleRank = map[string]int{"user": 1, "admin": 2, "super": 3}

type ctxKey int

const userKey ctxKey = 1

func (s *Service) authenticate(r *http.Request) *User {
	c, err := parseToken(bearer(r))
	if err != nil {
		return nil
	}
	u, ok := s.store.ByEmail(c.Email)
	if !ok || u.Disabled {
		return nil
	}
	return &u
}

// RequireRole 角色门：admin 含 super；super 仅超管。后续 admin 接口复用
func (s *Service) RequireRole(min string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := s.authenticate(r)
		if u == nil {
			fail(w, 401, "请先登录")
			return
		}
		if roleRank[u.Role] < roleRank[min] {
			fail(w, 403, "权限不足")
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), userKey, u)))
	}
}

// UserFrom 在 RequireRole 保护的 handler 内取当前用户
func (s *Service) UserFrom(r *http.Request) *User {
	u, _ := r.Context().Value(userKey).(*User)
	return u
}

func bearer(r *http.Request) string {
	return strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
}
