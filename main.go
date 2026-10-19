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
    registerLLMRoutes(mux)

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

	
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ai.html", http.StatusFound)
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
	registerFileTools(mux, db)
	registerToolkit(mux, db)
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
