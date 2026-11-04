package main

import (
	"embed"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed tasks.jsonl
var evalTasksFS embed.FS

type evalTask struct {
	ID         string `json:"id"`
	Input      string `json:"input"`
	ExpectTool string `json:"expectTool"`
}

type evalResult struct {
	TaskID     string `json:"taskId"`
	OK         bool   `json:"ok"`
	ToolCalled string `json:"toolCalled"`
	ExpectTool string `json:"expectTool"`
	ToolMatch  bool   `json:"toolMatch"`
	Answer     string `json:"answer"`
	Rounds     int    `json:"rounds"`
	Tokens     int64  `json:"tokens"`
	MS         int    `json:"ms"`
	Error      string `json:"error,omitempty"`
}

func evalLoadTasks() []evalTask {
	b, err := evalTasksFS.ReadFile("tasks.jsonl")
	if err != nil {
		return nil
	}
	var out []evalTask
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var t evalTask
		if json.Unmarshal([]byte(line), &t) == nil && t.ID != "" && t.ExpectTool != "" {
			out = append(out, t)
		}
	}
	return out
}

// evalHistory — 跑批快照历史(上限50轮), 持久副作用可见
func evalHistoryPath(db *DB) string {
	return filepath.Join(filepath.Dir(db.path), "eval-history.json")
}

func evalHistoryLoad(db *DB) []map[string]any {
	b, err := os.ReadFile(evalHistoryPath(db))
	if err != nil {
		return nil
	}
	var h []map[string]any
	if json.Unmarshal(b, &h) != nil {
		return nil
	}
	return h
}

func evalHistoryAppend(db *DB, res map[string]any) {
	snap := map[string]any{
		"ts":        time.Now().Format("2006-01-02 15:04:05"),
		"total":     res["total"],
		"toolMatch": res["toolMatch"],
		"pct":       res["pct"],
		"ms":        res["ms"],
	}
	h := append(evalHistoryLoad(db), snap)
	if len(h) > 50 {
		h = h[len(h)-50:]
	}
	b, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(evalHistoryPath(db)), 0755)
	_ = os.WriteFile(evalHistoryPath(db), b, 0644)
}

func evalRunAll(db *DB, client *http.Client, p LLMProfile) map[string]any {
	tasks := evalLoadTasks()
	results := make([]evalResult, 0, len(tasks))
	match, total := 0, 0
	t0 := time.Now()
	for _, t := range tasks {
		r := evalResult{TaskID: t.ID, ExpectTool: t.ExpectTool}
		answer, calls, rounds, tokens, ms, err := agentRunOnce(db, client, p, t.Input, 4, 2000, 0)
		if err != nil {
			r.Error = err.Error()
			results = append(results, r)
			total++
			continue
		}
		r.OK = true
		r.Answer = llmTrunc(answer, 300)
		r.Rounds = rounds
		r.Tokens = tokens
		r.MS = int(ms)
		if len(calls) > 0 {
			r.ToolCalled = calls[0].Tool
			r.ToolMatch = calls[0].Tool == t.ExpectTool
			if r.ToolMatch {
				match++
			}
		}
		results = append(results, r)
		total++
	}
	pct := 0
	if total > 0 {
		pct = match * 100 / total
	}
	return map[string]any{
		"total": total, "toolMatch": match, "pct": pct,
		"ms": time.Since(t0).Milliseconds(), "results": results,
	}
}

func registerEvalRoutes(mux *http.ServeMux, s *srv, db *DB) {
	mux.HandleFunc("POST /api/agent/eval", s.admin(func(w http.ResponseWriter, r *http.Request) {
		st := newLLMStore()
		var in struct {
			ProfileID string `json:"profileId"`
		}
		_ = body(r, &in)
		var cands []LLMProfile
		if in.ProfileID != "" {
			if p, found := st.get(in.ProfileID); found {
				cands = append(cands, p)
			}
		}
		if len(cands) == 0 {
			for _, p := range st.listChat() {
				if p.Enabled {
					cands = append(cands, p)
				}
			}
		}
		if len(cands) == 0 {
			E(w, 422, "没有可用档案")
			return
		}
		client := &http.Client{Timeout: 90 * time.Second}
		res := evalRunAll(db, client, cands[0])
		evalHistoryAppend(db, res)
		J(w, res)
	}))
	mux.HandleFunc("GET /api/agent/eval/history", s.admin(func(w http.ResponseWriter, r *http.Request) {
		J(w, map[string]any{"ok": true, "history": evalHistoryLoad(db)})
	}))
}
