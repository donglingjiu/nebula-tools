#!/usr/bin/env bash
# ============================================================
# Nova Tools v2.2.1 —— 全量落盘脚本 1/3（Go 后端）【bash 版】
# 用法: 在仓库根目录执行  bash setup-1.sh
# ============================================================
cd "$(dirname "$0")"

w(){ # w <相对路径>，内容走 stdin（无 BOM UTF-8，LF 行尾）
  local p="$1"; mkdir -p "$(dirname "$p")"
  cat > "$p"
  echo "  写入 $p  ($(wc -c < "$p" | tr -d ' ') 字节)"
}

echo "═══ Nova v2.2.1 脚本1/3：Go 后端 ═══"

# ---------- go.mod ----------
echo "── go.mod ──"
w go.mod <<'GO_MOD'
module nova

go 1.22

require github.com/skip2/go-qrcode v0.0.0-20200617113157-d469926327f9
GO_MOD

# ---------- store.go（终版全量） ----------
echo "── store.go（终版全量） ──"
w store.go <<'STORE_GO'
package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type User struct {
	Id        string    `json:"id"`
	Username  string    `json:"username"`
	Hash      string    `json:"hash"` // 落盘必需；对外 API 与 Export 均已抹除
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}
type Session struct {
	Token   string    `json:"token"`
	User    string    `json:"user"`
	Expires time.Time `json:"expires"`
}
type Hist struct {
	Id     int64     `json:"id"`
	User   string    `json:"user"`
	Tool   string    `json:"tool"`
	Input  string    `json:"input"`
	Output string    `json:"output"`
	Ms     int64     `json:"ms"`
	At     time.Time `json:"at"`
}
type Notice struct {
	Id    int64     `json:"id"`
	Title string    `json:"title"`
	Body  string    `json:"body"`
	At    time.Time `json:"at"`
}
type LogEntry struct {
	At     time.Time `json:"at"`
	User   string    `json:"user"`
	Method string    `json:"method"`
	Path   string    `json:"path"`
	Status int       `json:"status"`
	Ms     int64     `json:"ms"`
}
type Workflow struct {
	Id    int64    `json:"id"`
	User  string   `json:"user"`
	Name  string   `json:"name"`
	Steps []string `json:"steps"`
}

type DB struct {
	mu        sync.Mutex `json:"-"`
	path      string     `json:"-"`
	Users     map[string]*User
	Sessions  map[string]*Session
	Settings  map[string]string
	ToolState map[string]bool
	History   []Hist
	Favs      map[string]map[string]bool
	Logs      []LogEntry
	Notices   []Notice
	Workflows []Workflow
	NextId    int64
}

func Load() *DB {
	d := &DB{path: "data/nova.json",
		Users: map[string]*User{}, Sessions: map[string]*Session{},
		Settings: map[string]string{"site_name": "Nova Tools", "allow_register": "1",
			"ai_endpoint": "", "ai_key": "", "ai_model": "gpt-4o-mini"},
		ToolState: map[string]bool{}, Favs: map[string]map[string]bool{}}
	if b, err := os.ReadFile(d.path); err == nil {
		if json.Unmarshal(b, d) == nil {
			for id := range ToolRegistry {
				if _, ok := d.ToolState[id]; !ok {
					d.ToolState[id] = true
				}
			}
			return d
		}
	}
	for id := range ToolRegistry {
		d.ToolState[id] = true
	}
	return d
}

func (d *DB) Save() {
	d.mu.Lock()
	defer d.mu.Unlock()
	b, _ := json.MarshalIndent(d, "", " ")
	os.MkdirAll(filepath.Dir(d.path), 0o755)
	os.WriteFile(d.path+".tmp", b, 0o644)
	os.Rename(d.path+".tmp", d.path)
}

func pwHash(p string) string { h := sha256.Sum256([]byte("nova$salt$" + p)); return hex.EncodeToString(h[:]) }
func newToken() string       { b := make([]byte, 24); rand.Read(b); return hex.EncodeToString(b) }
func (d *DB) nextId() int64  { d.NextId++; return d.NextId }

func (d *DB) UserCount() int { d.mu.Lock(); defer d.mu.Unlock(); return len(d.Users) }

func (d *DB) Register(username, pw string) (*User, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if username == "" || len(pw) < 6 {
		return nil, errors.New("用户名不能为空，密码至少 6 位")
	}
	if _, ok := d.Users[username]; ok {
		return nil, errors.New("用户已存在")
	}
	if len(d.Users) > 0 && d.Settings["allow_register"] != "1" {
		return nil, errors.New("管理员已关闭注册")
	}
	role := "user"
	if len(d.Users) == 0 {
		role = "admin"
	}
	u := &User{Id: username, Username: username, Hash: pwHash(pw), Role: role, CreatedAt: time.Now()}
	d.Users[username] = u
	return u, nil
}

func (d *DB) CheckPw(username, pw string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	u := d.Users[username]
	return u != nil && u.Hash == pwHash(pw)
}
func (d *DB) SetPw(username, pw string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if u := d.Users[username]; u != nil && len(pw) >= 6 {
		u.Hash = pwHash(pw)
	}
}
func (d *DB) SetRole(name, role string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if u := d.Users[name]; u != nil {
		u.Role = role
	}
}
func (d *DB) DelUser(name string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.Users, name)
	delete(d.Favs, name)
	for t, s := range d.Sessions {
		if s.User == name {
			delete(d.Sessions, t)
		}
	}
}

func (d *DB) NewSession(username string) string {
	d.mu.Lock()
	defer d.mu.Unlock()
	t := newToken()
	d.Sessions[t] = &Session{Token: t, User: username, Expires: time.Now().Add(7 * 24 * time.Hour)}
	return t
}
func (d *DB) SessionUser(token string) *User {
	d.mu.Lock()
	defer d.mu.Unlock()
	s := d.Sessions[token]
	if s == nil || time.Now().After(s.Expires) {
		return nil
	}
	return d.Users[s.User]
}
func (d *DB) Logout(token string) { d.mu.Lock(); delete(d.Sessions, token); d.mu.Unlock() }

func (d *DB) SetTool(id string, on bool) { d.mu.Lock(); d.ToolState[id] = on; d.mu.Unlock() }
func (d *DB) ToolOn(id string) bool      { d.mu.Lock(); defer d.mu.Unlock(); return d.ToolState[id] }

func (d *DB) AddHist(user, tool, in, out string, ms int64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.History = append(d.History, Hist{Id: d.nextId(), User: user, Tool: tool, Input: in, Output: out, Ms: ms, At: time.Now()})
	if len(d.History) > 3000 {
		d.History = d.History[len(d.History)-3000:]
	}
}
func (d *DB) HistOf(user string) []Hist {
	d.mu.Lock()
	defer d.mu.Unlock()
	var out []Hist
	for i := len(d.History) - 1; i >= 0 && len(out) < 100; i-- {
		if d.History[i].User == user {
			out = append(out, d.History[i])
		}
	}
	return out
}
func (d *DB) ToggleFav(user, tool string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.Favs[user] == nil {
		d.Favs[user] = map[string]bool{}
	}
	d.Favs[user][tool] = !d.Favs[user][tool]
	return d.Favs[user][tool]
}
func (d *DB) FavList(user string) []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	var out []string
	for t, on := range d.Favs[user] {
		if on {
			out = append(out, t)
		}
	}
	sort.Strings(out)
	return out
}
func (d *DB) AddLog(e LogEntry) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Logs = append(d.Logs, e)
	if len(d.Logs) > 500 {
		d.Logs = d.Logs[len(d.Logs)-500:]
	}
}
func (d *DB) AddNotice(title, body string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Notices = append(d.Notices, Notice{Id: d.nextId(), Title: title, Body: body, At: time.Now()})
}
func (d *DB) DelNotice(id int64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	var n []Notice
	for _, x := range d.Notices {
		if x.Id != id {
			n = append(n, x)
		}
	}
	d.Notices = n
}
func (d *DB) NoticesList() []Notice {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := append([]Notice(nil), d.Notices...)
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}
func (d *DB) SaveWorkflow(user, name string, steps []string) int64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	id := d.nextId()
	d.Workflows = append(d.Workflows, Workflow{Id: id, User: user, Name: name, Steps: steps})
	return id
}
func (d *DB) WorkflowsOf(user string) []Workflow {
	d.mu.Lock()
	defer d.mu.Unlock()
	var o []Workflow
	for _, w := range d.Workflows {
		if w.User == user {
			o = append(o, w)
		}
	}
	return o
}
func (d *DB) Analytics(user string) (map[string]int, map[string]int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	tool := map[string]int{}
	day := map[string]int{}
	for _, h := range d.History {
		if user != "" && h.User != user {
			continue
		}
		tool[h.Tool]++
		day[h.At.Format("01-02")]++
	}
	return tool, day
}
func (d *DB) Stats() (users, total, today int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	td := time.Now().Format("2006-01-02")
	users, total = len(d.Users), len(d.History)
	for _, h := range d.History {
		if h.At.Format("2006-01-02") == td {
			today++
		}
	}
	return
}
func (d *DB) GetSettingSafe(k string) string { d.mu.Lock(); defer d.mu.Unlock(); return d.Settings[k] }
func (d *DB) SetSetting(k, v string)         { d.mu.Lock(); d.Settings[k] = v; d.mu.Unlock() }
func (d *DB) LogsLast(n int) []LogEntry {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := []LogEntry{}
	for i := len(d.Logs) - 1; i >= 0 && len(out) < n; i-- {
		out = append(out, d.Logs[i])
	}
	return out
}
func (d *DB) Export() map[string]any {
	d.mu.Lock()
	defer d.mu.Unlock()
	users := map[string]map[string]any{}
	for k, v := range d.Users {
		users[k] = map[string]any{"id": v.Id, "username": v.Username, "role": v.Role, "created_at": v.CreatedAt, "hash": "(已抹除)"}
	}
	return map[string]any{"users": users, "settings": d.Settings, "modules": d.ToolState,
		"history": d.History, "notices": d.Notices, "workflows": d.Workflows}
}
STORE_GO

# ---------- main.go（终版全量，含 newRouter 重构） ----------
echo "── main.go（终版全量） ──"
w main.go <<'MAIN_GO'
package main

import (
	"embed"
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

//go:embed web
var webFS embed.FS

var Version = "v2.2.1"
var started = time.Now()

type srv struct{ db *DB }

func J(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(v)
}
func E(w http.ResponseWriter, code int, msg string) { w.WriteHeader(code); J(w, map[string]string{"error": msg}) }
func body(r *http.Request, v any) error             { return json.NewDecoder(io.LimitReader(r.Body, 8<<20)).Decode(v) }

type sw struct {
	http.ResponseWriter
	code int
}

func (s *sw) WriteHeader(c int) { s.code = c; s.ResponseWriter.WriteHeader(c) }

func (s *srv) user(r *http.Request) *User {
	tok := ""
	if c, err := r.Cookie("nova_tok"); err == nil {
		tok = c.Value
	}
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		tok = strings.TrimPrefix(h, "Bearer ")
	}
	return s.db.SessionUser(tok)
}

func (s *srv) admin(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := s.user(r)
		if u == nil {
			E(w, 401, "未登录")
			return
		}
		if u.Role != "admin" {
			E(w, 403, "需要管理员权限")
			return
		}
		h(w, r)
	}
}

func (s *srv) issue(w http.ResponseWriter, username string) string {
	tok := s.db.NewSession(username)
	http.SetCookie(w, &http.Cookie{Name: "nova_tok", Value: tok, Path: "/", HttpOnly: true, MaxAge: 7 * 24 * 3600, SameSite: http.SameSiteLaxMode})
	s.db.Save()
	return tok
}

func (s *srv) currentName(r *http.Request) string {
	if u := s.user(r); u != nil {
		return u.Username
	}
	return ""
}

func (s *srv) logMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}
		sww := &sw{w, 200}
		t := time.Now()
		next.ServeHTTP(sww, r)
		name := ""
		if u := s.user(r); u != nil {
			name = u.Username
		}
		s.db.AddLog(LogEntry{At: time.Now(), User: name, Method: r.Method, Path: r.URL.Path, Status: sww.code, Ms: time.Since(t).Milliseconds()})
	})
}

func main() {
	db := Load()
	go func() {
		for {
			time.Sleep(30 * time.Second)
			db.Save()
		}
	}()
	log.Printf("Nova Tools %s 已启动 → http://localhost:8080", Version)
	log.Fatal(http.ListenAndServe(":8080", newRouter(db)))
}

func newRouter(db *DB) http.Handler {
	s := &srv{db}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		J(w, map[string]any{"ok": true, "app": "nova", "version": Version, "db_driver": "json", "users": db.UserCount(), "uptime_sec": int(time.Since(started).Seconds())})
	})
	mux.HandleFunc("POST /api/auth/register", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if body(r, &req) != nil {
			E(w, 400, "参数错误")
			return
		}
		u, err := db.Register(req.Username, req.Password)
		if err != nil {
			E(w, 400, err.Error())
			return
		}
		s.issue(w, u.Username)
		J(w, map[string]any{"ok": true, "username": u.Username, "role": u.Role})
	})
	mux.HandleFunc("POST /api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if body(r, &req) != nil {
			E(w, 400, "参数错误")
			return
		}
		if !db.CheckPw(req.Username, req.Password) {
			E(w, 401, "用户名或密码错误")
			return
		}
		u := db.SessionUser(s.issue(w, req.Username))
		J(w, map[string]any{"ok": true, "username": req.Username, "role": u.Role})
	})
	mux.HandleFunc("POST /api/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		if c, err := r.Cookie("nova_tok"); err == nil {
			db.Logout(c.Value)
			db.Save()
		}
		J(w, map[string]any{"ok": true})
	})
	mux.HandleFunc("GET /api/auth/me", func(w http.ResponseWriter, r *http.Request) {
		u := s.user(r)
		if u == nil {
			E(w, 401, "未登录")
			return
		}
		J(w, map[string]any{"username": u.Username, "role": u.Role, "created_at": u.CreatedAt})
	})
	mux.HandleFunc("POST /api/auth/password", func(w http.ResponseWriter, r *http.Request) {
		u := s.user(r)
		if u == nil {
			E(w, 401, "未登录")
			return
		}
		var req struct {
			Old string `json:"old"`
			New string `json:"new"`
		}
		if body(r, &req) != nil || !db.CheckPw(u.Username, req.Old) {
			E(w, 400, "原密码错误")
			return
		}
		if len(req.New) < 6 {
			E(w, 400, "新密码至少 6 位")
			return
		}
		db.SetPw(u.Username, req.New)
		db.Save()
		J(w, map[string]any{"ok": true})
	})

	mux.HandleFunc("GET /api/tools", func(w http.ResponseWriter, r *http.Request) {
		u := s.user(r)
		isAdmin := u != nil && u.Role == "admin"
		ids := make([]string, 0, len(ToolRegistry))
		for id := range ToolRegistry {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		list := []map[string]any{}
		for _, id := range ids {
			t := ToolRegistry[id]
			on := db.ToolOn(id)
			if !on && !isAdmin {
				continue
			}
			list = append(list, map[string]any{"id": t.Id, "name": t.Name, "icon": t.Icon, "cat": t.Cat, "desc": t.Desc, "kind": t.Kind, "on": on})
		}
		J(w, list)
	})
	mux.HandleFunc("POST /api/run/{tool}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("tool")
		if _, ok := ToolRegistry[id]; !ok || !db.ToolOn(id) {
			E(w, 403, "模块未开启或不存在")
			return
		}
		var req struct {
			Input string            `json:"input"`
			Opts  map[string]string `json:"opts"`
		}
		if body(r, &req) != nil {
			E(w, 400, "参数错误")
			return
		}
		t0 := time.Now()
		out, err := runTool(id, req.Input, req.Opts)
		ms := time.Since(t0).Milliseconds()
		if err != nil {
			J(w, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		if u := s.user(r); u != nil {
			db.AddHist(u.Username, id, req.Input, out, ms)
		}
		J(w, map[string]any{"ok": true, "output": out, "ms": ms})
	})
	mux.HandleFunc("POST /api/upload/{tool}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("tool")
		if _, ok := ToolRegistry[id]; !ok || !db.ToolOn(id) {
			E(w, 403, "模块未开启或不存在")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 32<<20)
		f, _, err := r.FormFile("file")
		if err != nil {
			E(w, 400, "请选择文件")
			return
		}
		defer f.Close()
		b, err := io.ReadAll(f)
		if err != nil {
			E(w, 400, "读取文件失败")
			return
		}
		opts := map[string]string{}
		if sv := r.FormValue("opts"); sv != "" {
			if err := json.Unmarshal([]byte(sv), &opts); err != nil {
				E(w, 400, "opts 参数错误")
				return
			}
		}
		t0 := time.Now()
		out, err := runToolFile(id, b, opts)
		ms := time.Since(t0).Milliseconds()
		if err != nil {
			J(w, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		if u := s.user(r); u != nil {
			db.AddHist(u.Username, id, "(文件 "+strconv.Itoa(len(b))+" 字节)", out, ms)
		}
		J(w, map[string]any{"ok": true, "output": out, "ms": ms})
	})
	mux.HandleFunc("GET /api/history", func(w http.ResponseWriter, r *http.Request) {
		u := s.user(r)
		if u == nil {
			E(w, 401, "未登录")
			return
		}
		J(w, db.HistOf(u.Username))
	})
	mux.HandleFunc("POST /api/fav/{tool}", func(w http.ResponseWriter, r *http.Request) {
		u := s.user(r)
		if u == nil {
			E(w, 401, "未登录")
			return
		}
		J(w, map[string]any{"on": db.ToggleFav(u.Username, r.PathValue("tool"))})
	})
	mux.HandleFunc("GET /api/favs", func(w http.ResponseWriter, r *http.Request) {
		u := s.user(r)
		if u == nil {
			E(w, 401, "未登录")
			return
		}
		J(w, db.FavList(u.Username))
	})
	mux.HandleFunc("GET /api/notices", func(w http.ResponseWriter, r *http.Request) {
		J(w, db.NoticesList())
	})
	mux.HandleFunc("POST /api/ai/chat", func(w http.ResponseWriter, r *http.Request) {
		u := s.user(r)
		if u == nil {
			E(w, 401, "请先登录后使用 AI 助手")
			return
		}
		var req struct {
			Messages []map[string]any `json:"messages"`
		}
		if body(r, &req) != nil {
			E(w, 400, "参数错误")
			return
		}
		ep := db.GetSettingSafe("ai_endpoint")
		if ep == "" {
			J(w, map[string]any{"ok": true, "reply": "（演示模式）管理员尚未配置 AI 端点。请管理员进入 后台 → 设置 填入 OpenAI 兼容接口地址与密钥，即可开启真实对话。"})
			return
		}
		pr, _ := json.Marshal(map[string]any{"model": db.GetSettingSafe("ai_model"), "messages": req.Messages, "stream": false})
		hr, _ := http.NewRequest("POST", ep, strings.NewReader(string(pr)))
		hr.Header.Set("Content-Type", "application/json")
		if k := db.GetSettingSafe("ai_key"); k != "" {
			hr.Header.Set("Authorization", "Bearer "+k)
		}
		resp, err := (&http.Client{Timeout: 60 * time.Second}).Do(hr)
		if err != nil {
			E(w, 502, "AI 端点不可达: "+err.Error())
			return
		}
		defer resp.Body.Close()
		var ar struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
			E(w, 502, "AI 响应解析失败")
			return
		}
		if ar.Error != nil {
			E(w, 502, "AI 端点返回错误: "+ar.Error.Message)
			return
		}
		reply := ""
		if len(ar.Choices) > 0 {
			reply = ar.Choices[0].Message.Content
		}
		J(w, map[string]any{"ok": true, "reply": reply})
	})
	mux.HandleFunc("GET /api/workflows", func(w http.ResponseWriter, r *http.Request) {
		u := s.user(r)
		if u == nil {
			E(w, 401, "未登录")
			return
		}
		J(w, db.WorkflowsOf(u.Username))
	})
	mux.HandleFunc("POST /api/workflows", func(w http.ResponseWriter, r *http.Request) {
		u := s.user(r)
		if u == nil {
			E(w, 401, "未登录")
			return
		}
		var req struct {
			Name  string   `json:"name"`
			Steps []string `json:"steps"`
		}
		if body(r, &req) != nil || req.Name == "" || len(req.Steps) == 0 {
			E(w, 400, "需要名称和至少一个步骤")
			return
		}
		for _, st := range req.Steps {
			t, ok := ToolRegistry[st]
			if !ok || t.Kind != "text" {
				E(w, 400, "工作流仅支持文本工具: "+st)
				return
			}
		}
		id := db.SaveWorkflow(u.Username, req.Name, req.Steps)
		db.Save()
		J(w, map[string]any{"ok": true, "id": id})
	})
	mux.HandleFunc("POST /api/workflows/{id}/run", func(w http.ResponseWriter, r *http.Request) {
		u := s.user(r)
		if u == nil {
			E(w, 401, "未登录")
			return
		}
		var req struct {
			Input string `json:"input"`
		}
		body(r, &req)
		idStr := r.PathValue("id")
		var wf *Workflow
		wfs := db.WorkflowsOf(u.Username)
		for i := range wfs {
			if strconv.FormatInt(wfs[i].Id, 10) == idStr {
				wf = &wfs[i]
			}
		}
		if wf == nil {
			E(w, 404, "工作流不存在")
			return
		}
		t0 := time.Now()
		cur := req.Input
		trace := []map[string]any{}
		for _, st := range wf.Steps {
			out, err := runTool(st, cur, nil)
			if err != nil {
				J(w, map[string]any{"ok": false, "error": "步骤 " + st + " 失败: " + err.Error(), "trace": trace})
				return
			}
			trace = append(trace, map[string]any{"tool": st, "output": out})
			cur = out
		}
		db.AddHist(u.Username, "workflow:"+wf.Name, req.Input, cur, time.Since(t0).Milliseconds())
		J(w, map[string]any{"ok": true, "trace": trace, "output": cur})
	})
	mux.HandleFunc("GET /api/analytics", func(w http.ResponseWriter, r *http.Request) {
		u := s.user(r)
		if u == nil {
			E(w, 401, "未登录")
			return
		}
		tool, day := db.Analytics(u.Username)
		J(w, map[string]any{"byTool": tool, "byDay": day})
	})

	// ---------- 管理后台 ----------
	mux.HandleFunc("GET /api/admin/overview", s.admin(func(w http.ResponseWriter, r *http.Request) {
		users, total, today := db.Stats()
		on := 0
		for id := range ToolRegistry {
			if db.ToolOn(id) {
				on++
			}
		}
		J(w, map[string]any{"users": users, "runs_total": total, "runs_today": today, "modules_on": on, "modules_total": len(ToolRegistry), "uptime_sec": int(time.Since(started).Seconds())})
	}))
	mux.HandleFunc("GET /api/admin/users", s.admin(func(w http.ResponseWriter, r *http.Request) {
		type uo struct {
			Username  string    `json:"username"`
			Role      string    `json:"role"`
			CreatedAt time.Time `json:"created_at"`
		}
		names := []string{}
		for n := range db.Users {
			names = append(names, n)
		}
		sort.Strings(names)
		out := []uo{}
		for _, n := range names {
			u := db.Users[n]
			out = append(out, uo{n, u.Role, u.CreatedAt})
		}
		J(w, out)
	}))
	mux.HandleFunc("POST /api/admin/users/{name}/role", s.admin(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Role string `json:"role"`
		}
		if body(r, &req) != nil || (req.Role != "admin" && req.Role != "user") {
			E(w, 400, "角色只能是 admin / user")
			return
		}
		if r.PathValue("name") == s.currentName(r) {
			E(w, 400, "不能修改自己的角色")
			return
		}
		db.SetRole(r.PathValue("name"), req.Role)
		db.Save()
		J(w, map[string]any{"ok": true})
	}))
	mux.HandleFunc("POST /api/admin/users/{name}/reset", s.admin(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Password string `json:"password"`
		}
		if body(r, &req) != nil || len(req.Password) < 6 {
			E(w, 400, "新密码至少 6 位")
			return
		}
		db.SetPw(r.PathValue("name"), req.Password)
		db.Save()
		J(w, map[string]any{"ok": true})
	}))
	mux.HandleFunc("DELETE /api/admin/users/{name}", s.admin(func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("name") == s.currentName(r) {
			E(w, 400, "不能删除自己")
			return
		}
		db.DelUser(r.PathValue("name"))
		db.Save()
		J(w, map[string]any{"ok": true})
	}))
	mux.HandleFunc("POST /api/admin/modules/{id}", s.admin(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			On bool `json:"on"`
		}
		if body(r, &req) != nil {
			E(w, 400, "参数错误")
			return
		}
		db.SetTool(r.PathValue("id"), req.On)
		db.Save()
		J(w, map[string]any{"ok": true})
	}))
	mux.HandleFunc("GET /api/admin/settings", s.admin(func(w http.ResponseWriter, r *http.Request) {
		J(w, db.Settings)
	}))
	mux.HandleFunc("POST /api/admin/settings", s.admin(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]string
		if body(r, &req) != nil {
			E(w, 400, "参数错误")
			return
		}
		allow := map[string]bool{"site_name": true, "allow_register": true, "ai_endpoint": true, "ai_key": true, "ai_model": true}
		for k, v := range req {
			if allow[k] {
				db.SetSetting(k, v)
			}
		}
		db.Save()
		J(w, map[string]any{"ok": true})
	}))
	mux.HandleFunc("GET /api/admin/logs", s.admin(func(w http.ResponseWriter, r *http.Request) {
		J(w, db.LogsLast(200))
	}))
	mux.HandleFunc("GET /api/admin/analytics", s.admin(func(w http.ResponseWriter, r *http.Request) {
		tool, day := db.Analytics("")
		J(w, map[string]any{"byTool": tool, "byDay": day})
	}))
	mux.HandleFunc("POST /api/admin/notices", s.admin(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Title string `json:"title"`
			Body  string `json:"body"`
		}
		if body(r, &req) != nil || req.Title == "" {
			E(w, 400, "需要标题")
			return
		}
		db.AddNotice(req.Title, req.Body)
		db.Save()
		J(w, map[string]any{"ok": true})
	}))
	mux.HandleFunc("DELETE /api/admin/notices/{id}", s.admin(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err == nil {
			db.DelNotice(id)
			db.Save()
		}
		J(w, map[string]any{"ok": true})
	}))
	mux.HandleFunc("GET /api/admin/export", s.admin(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", "attachment; filename=nova-export.json")
		J(w, db.Export())
	}))

	// ---------- 静态资源（embed 进二进制，部署只需一个可执行文件） ----------
	sub, _ := fs.Sub(webFS, "web")
	fileServer := http.FileServer(http.FS(sub))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			b, _ := fs.ReadFile(sub, "index.html")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(b)
			return
		}
		fileServer.ServeHTTP(w, r)
	})

	return s.logMW(mux)
}
MAIN_GO

# ---------- tools.go 智能补丁（幂等三态） ----------
echo "── tools.go 智能补丁 ──"
if [ -f tools.go ]; then
  if [ ! -f tools2.go ]; then
    echo "  ⚠ 未找到 tools2.go（v2.1 应已有），补丁引用 runTool2/runToolFile2 将无法编译，请先确认"
  fi
  # ① runTool fallthrough → runTool2
  if grep -q 'runTool2(' tools.go; then
    echo "  · ① 已对接 runTool2，跳过"
  elif grep -q '未知工具' tools.go; then
    sed -i 's|.*未知工具.*|\treturn runTool2(id, in, opts)|' tools.go
    if grep -q 'runTool2(id, in, opts)' tools.go; then echo "  · ① fallthrough → runTool2 完成"; else echo "  · ① ⚠ 替换未命中，请人工检查 tools.go"; fi
  else
    echo "  · ① ⚠ 未找到「未知工具」锚点，请人工检查 tools.go"
  fi
  # ② runToolFile 签名加 opts（并同步单测调用）
  if grep -q 'func runToolFile(.*opts' tools.go; then
    echo "  · ② runToolFile 已含 opts，跳过"
  else
    sed -i 's|func runToolFile(id string, b \[\]byte) (string, error) {|func runToolFile(id string, b []byte, opts map[string]string) {|' tools.go
    if grep -q 'func runToolFile(.*opts' tools.go; then
      echo "  · ② runToolFile 签名已加 opts"
      if [ -f tools_test.go ] && grep -q 'runToolFile(' tools_test.go && ! grep -q 'runToolFile(.*opts' tools_test.go; then
        sed -i -E 's/(runToolFile\([^)]*\))/\1, nil)/g' tools_test.go
        echo "  · ②b tools_test.go 中的 runToolFile 调用已补 nil"
      fi
    else
      echo "  · ② ⚠ 签名替换未命中，请人工检查 tools.go"
    fi
  fi
  # ③ runToolFile fallthrough → runToolFile2
  if grep -q 'runToolFile2(' tools.go; then
    echo "  · ③ 已对接 runToolFile2，跳过"
  elif grep -q '未知文件工具' tools.go; then
    sed -i 's|.*未知文件工具.*|\treturn runToolFile2(id, b, opts)|' tools.go
    if grep -q 'runToolFile2(id, b, opts)' tools.go; then echo "  · ③ fallthrough → runToolFile2 完成"; else echo "  · ③ ⚠ 替换未命中，请人工检查 tools.go"; fi
  else
    echo "  · ③ ⚠ 未找到「未知文件工具」锚点，请人工检查 tools.go"
  fi
else
  echo "  ⚠ 未找到 tools.go，请确认当前目录是仓库根"
fi

# ---------- 数据兼容清理 ----------
if [ -f data/nova.json ]; then
  rm -f data/nova.json
  echo "已删除旧 data/nova.json（json 键名协议已变更，旧文件作废，首次注册将重建 admin）"
fi

# ---------- 环境自检 ----------
echo "── 自检 ──"
for f in tools2.go tools_test.go web/index.html; do
  if [ -f "$f" ]; then echo "  ✓ $f 存在（保留不动）"; else echo "  ⚠ $f 不存在 —— tools2.go 缺失会编译失败；web/index.html 缺失由脚本3补齐"; fi
done

echo
echo "═══ 脚本1/3 完成。请继续执行 bash setup-2.sh ═══"
