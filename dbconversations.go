package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type convRow struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	ProfileID string `json:"profileId,omitempty"`
	UserID    string `json:"userId"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type msgRow struct {
	ID        int64  `json:"id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	Tokens    *int64 `json:"tokens,omitempty"`
	CreatedAt string `json:"createdAt"`
}

func dbConvCreate(user, title string) (int64, error) {
	if novaDB == nil {
		return 0, errors.New("db not ready")
	}
	t := strings.TrimSpace(title)
	if t == "" {
		t = "New Chat"
	}
	res, err := novaDB.Exec("INSERT INTO conversations(title, user_id) VALUES(?,?)", t, user)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func dbConvList(user string) []convRow {
	out := []convRow{}
	if novaDB == nil {
		return out
	}
	rows, err := novaDB.Query("SELECT id,title,profile_id,user_id,created_at,updated_at FROM conversations WHERE user_id=? ORDER BY updated_at DESC, id DESC", user)
	if err != nil {
		log.Printf("[db] conv list fail: %v", err)
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var c convRow
		var pid sql.NullString
		if err := rows.Scan(&c.ID, &c.Title, &pid, &c.UserID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			continue
		}
		c.ProfileID = pid.String
		out = append(out, c)
	}
	return out
}

func dbConvOwn(id int64, user string) bool {
	if novaDB == nil {
		return false
	}
	var n int
	if err := novaDB.QueryRow("SELECT COUNT(*) FROM conversations WHERE id=? AND user_id=?", id, user).Scan(&n); err != nil {
		return false
	}
	return n == 1
}

func dbConvRename(id int64, user, title string) error {
	if novaDB == nil {
		return errors.New("db not ready")
	}
	t := strings.TrimSpace(title)
	if t == "" {
		return errors.New("title required")
	}
	res, err := novaDB.Exec("UPDATE conversations SET title=?, updated_at=datetime('now') WHERE id=? AND user_id=?", t, id, user)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("not found or forbidden")
	}
	return nil
}

func dbConvDelete(id int64, user string) error {
	if novaDB == nil {
		return errors.New("db not ready")
	}
	res, err := novaDB.Exec("DELETE FROM conversations WHERE id=? AND user_id=?", id, user)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("not found or forbidden")
	}
	return nil
}

func dbMsgAppend(convID int64, role, content string, tokens ...int64) (int64, error) {
	if novaDB == nil {
		return 0, errors.New("db not ready")
	}
	if role != "user" && role != "assistant" && role != "system" {
		return 0, errors.New("bad role")
	}
	c := strings.TrimSpace(content)
	if c == "" {
		return 0, errors.New("content required")
	}
	tx, err := novaDB.Begin()
	if err != nil {
		return 0, err
	}
	var n int
	if err := tx.QueryRow("SELECT COUNT(*) FROM conversations WHERE id=?", convID).Scan(&n); err != nil || n != 1 {
		_ = tx.Rollback()
		return 0, errors.New("conversation not found")
	}
	var tokPtr any
	if len(tokens) > 0 && tokens[0] > 0 {
		tokPtr = tokens[0]
	}
	res, err := tx.Exec("INSERT INTO messages(conversation_id, role, content, tokens) VALUES(?,?,?,?)", convID, role, c, tokPtr)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	if _, err := tx.Exec("UPDATE conversations SET updated_at=datetime('now') WHERE id=?", convID); err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func dbMsgList(convID int64) []msgRow {
	out := []msgRow{}
	if novaDB == nil {
		return out
	}
	rows, err := novaDB.Query("SELECT id,role,content,tokens,created_at FROM messages WHERE conversation_id=? ORDER BY id", convID)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var m msgRow
		if err := rows.Scan(&m.ID, &m.Role, &m.Content, &m.Tokens, &m.CreatedAt); err != nil {
			continue
		}
		out = append(out, m)
	}
	return out
}

func convIDFromPath(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func registerConvRoutes(mux *http.ServeMux, s *srv) {
	registerUsageRoute(mux, s)
	mux.HandleFunc("POST /api/conversations", func(w http.ResponseWriter, r *http.Request) {
		u := s.user(r)
		if u == nil {
			E(w, 401, "login required")
			return
		}
		var req struct {
			Title string `json:"title"`
		}
		if body(r, &req) != nil {
			E(w, 400, "bad request")
			return
		}
		id, err := dbConvCreate(u.Username, req.Title)
		if err != nil {
			E(w, 500, err.Error())
			return
		}
		J(w, map[string]any{"ok": true, "conversationId": id})
	})

	mux.HandleFunc("GET /api/conversations", func(w http.ResponseWriter, r *http.Request) {
		u := s.user(r)
		if u == nil {
			E(w, 401, "login required")
			return
		}
		J(w, map[string]any{"ok": true, "conversations": dbConvList(u.Username)})
	})

	mux.HandleFunc("PATCH /api/conversations/{id}", func(w http.ResponseWriter, r *http.Request) {
		u := s.user(r)
		if u == nil {
			E(w, 401, "login required")
			return
		}
		id, err := convIDFromPath(r)
		if err != nil {
			E(w, 400, "bad id")
			return
		}
		var req struct {
			Title string `json:"title"`
		}
		if body(r, &req) != nil {
			E(w, 400, "bad request")
			return
		}
		if err := dbConvRename(id, u.Username, req.Title); err != nil {
			E(w, 404, err.Error())
			return
		}
		J(w, map[string]any{"ok": true})
	})

	mux.HandleFunc("DELETE /api/conversations/{id}", func(w http.ResponseWriter, r *http.Request) {
		u := s.user(r)
		if u == nil {
			E(w, 401, "login required")
			return
		}
		id, err := convIDFromPath(r)
		if err != nil {
			E(w, 400, "bad id")
			return
		}
		if err := dbConvDelete(id, u.Username); err != nil {
			E(w, 404, err.Error())
			return
		}
		J(w, map[string]any{"ok": true})
	})

	mux.HandleFunc("POST /api/conversations/{id}/branch", func(w http.ResponseWriter, r *http.Request) {
		u := s.user(r)
		if u == nil {
			E(w, 401, "login required")
			return
		}
		id, err := convIDFromPath(r)
		if err != nil {
			E(w, 400, "bad id")
			return
		}
		var req struct {
			MsgID int64 `json:"msgId"`
		}
		if body(r, &req) != nil || req.MsgID <= 0 {
			E(w, 400, "bad request")
			return
		}
		nid, err := dbConvBranch(id, req.MsgID, u.Username)
		if err != nil {
			E(w, 404, err.Error())
			return
		}
		J(w, map[string]any{"ok": true, "conversationId": nid})
	})

	mux.HandleFunc("GET /api/conversations/{id}/messages", func(w http.ResponseWriter, r *http.Request) {
		u := s.user(r)
		if u == nil {
			E(w, 401, "login required")
			return
		}
		id, err := convIDFromPath(r)
		if err != nil {
			E(w, 400, "bad id")
			return
		}
		if !dbConvOwn(id, u.Username) {
			E(w, 404, "not found")
			return
		}
		J(w, map[string]any{"ok": true, "messages": dbMsgList(id)})
	})

	mux.HandleFunc("POST /api/conversations/{id}/messages", func(w http.ResponseWriter, r *http.Request) {
		u := s.user(r)
		if u == nil {
			E(w, 401, "login required")
			return
		}
		id, err := convIDFromPath(r)
		if err != nil {
			E(w, 400, "bad id")
			return
		}
		if !dbConvOwn(id, u.Username) {
			E(w, 404, "not found")
			return
		}
		var req struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}
		if body(r, &req) != nil {
			E(w, 400, "bad request")
			return
		}
		if req.Role == "" {
			req.Role = "user"
		}
		mid, err := dbMsgAppend(id, req.Role, req.Content)
		if err != nil {
			E(w, 400, err.Error())
			return
		}
		J(w, map[string]any{"ok": true, "messageId": mid})
	})
}

type usageDay struct {
	Day    string `json:"day"`
	Msgs   int64  `json:"msgs"`
	Tokens int64  `json:"tokens"`
}

type usageProfile struct {
	Profile string `json:"profile"`
	Msgs    int    `json:"msgs"`
	Tokens  int64  `json:"tokens"`
}

func dbUsageProfiles(user string) []usageProfile {
	out := []usageProfile{}
	if novaDB == nil {
		return out
	}
	rows, err := novaDB.Query("SELECT COALESCE(NULLIF(profile,''),'(未记录)') AS p, COUNT(*) AS n, COALESCE(SUM(tokens),0) AS t FROM messages WHERE conversation_id IN (SELECT id FROM conversations WHERE user_id=?) AND role='assistant' GROUP BY p ORDER BY t DESC LIMIT 8", user)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var u usageProfile
		if err := rows.Scan(&u.Profile, &u.Msgs, &u.Tokens); err != nil {
			continue
		}
		out = append(out, u)
	}
	return out
}

func dbMsgAppendP(convID int64, role, content, profile string, tokens ...int64) (int64, error) {
	mid, err := dbMsgAppend(convID, role, content, tokens...)
	if err != nil || profile == "" || novaDB == nil {
		return mid, err
	}
	_, _ = novaDB.Exec("UPDATE messages SET profile=? WHERE id=?", profile, mid)
	return mid, err
}

func dbConvBranch(srcID, untilMsg int64, user string) (int64, error) {
	if novaDB == nil {
		return 0, errors.New("db not ready")
	}
	if !dbConvOwn(srcID, user) {
		return 0, errors.New("not found")
	}
	var n int
	if err := novaDB.QueryRow("SELECT COUNT(*) FROM messages WHERE id=? AND conversation_id=?", untilMsg, srcID).Scan(&n); err != nil {
		return 0, err
	}
	if n != 1 {
		return 0, errors.New("bad branch point")
	}
	var title string
	if err := novaDB.QueryRow("SELECT title FROM conversations WHERE id=?", srcID).Scan(&title); err != nil {
		return 0, err
	}
	newID, err := dbConvCreate(user, title+" ⎇")
	if err != nil {
		return 0, err
	}
	res, err := novaDB.Exec("INSERT INTO messages(conversation_id, role, content, tokens, branch_of, profile, created_at) SELECT ?, role, content, tokens, id, profile, created_at FROM messages WHERE conversation_id=? AND id<=?", newID, srcID, untilMsg)
	if err != nil {
		_ = dbConvDelete(newID, user)
		return 0, err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		_ = dbConvDelete(newID, user)
		return 0, errors.New("no messages to branch")
	}
	return newID, nil
}

func dbUsageSummary(user string) []usageDay {
	out := []usageDay{}
	if novaDB == nil {
		return out
	}
	rows, err := novaDB.Query("SELECT strftime('%Y-%m-%d',created_at,'localtime') AS d, COUNT(*) AS n, COALESCE(SUM(tokens),0) AS t FROM messages WHERE conversation_id IN (SELECT id FROM conversations WHERE user_id=?) GROUP BY d ORDER BY d DESC LIMIT 30", user)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var u usageDay
		if err := rows.Scan(&u.Day, &u.Msgs, &u.Tokens); err != nil {
			continue
		}
		out = append(out, u)
	}
	return out
}

func registerUsageRoute(mux *http.ServeMux, s *srv) {
	mux.HandleFunc("GET /api/usage/summary", func(w http.ResponseWriter, r *http.Request) {
		u := s.user(r)
		if u == nil {
			E(w, 401, "login required")
			return
		}
		J(w, map[string]any{"ok": true, "days": dbUsageSummary(u.Username)})
	})

	mux.HandleFunc("GET /api/usage/profiles", func(w http.ResponseWriter, r *http.Request) {
		u := s.user(r)
		if u == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"ok":false,"error":"login required"}`))
			return
		}
		llmWriteJSON(w, map[string]interface{}{"ok": true, "profiles": dbUsageProfiles(u.Username)})
	})
}
