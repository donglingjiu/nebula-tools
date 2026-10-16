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
