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

func dbMsgAppend(convID int64, role, content string) (int64, error) {
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
	res, err := tx.Exec("INSERT INTO messages(conversation_id, role, content) VALUES(?,?,?)", convID, role, c)
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
