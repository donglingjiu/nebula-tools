package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

type agentReq struct {
	ProfileID      string `json:"profileId"`
	Input          string `json:"input"`
	MaxRounds      int    `json:"maxRounds"`
	MaxTokens      int    `json:"maxTokens"`
	ConversationID int64  `json:"conversationId"`
}

type agentCall struct {
	Tool   string `json:"tool"`
	Input  string `json:"input"`
	Output string `json:"output"`
	OK     bool   `json:"ok"`
	MS     int    `json:"ms"`
}

func agentToolPrompt() string {
	var b strings.Builder
	b.WriteString("你是 Nova 工具代理,可调用下列工具完成任务:\n")
	ids := make([]string, 0, len(ToolRegistry))
	for id := range ToolRegistry {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		b.WriteString("- " + ToolRegistry[id].Id + ": " + ToolRegistry[id].Desc + "\n")
	}
	b.WriteString("\n调用规则:\n1. 需要调用工具时,只输出一行JSON(不要markdown围栏): {\"tool\":\"工具Id\",\"input\":\"工具输入文本\",\"opts\":{}}\n2. 拿到工具结果后判断是否继续;无需再调用时,直接输出给用户的最终文字回答(非JSON)")
	return b.String()
}

func agentParseToolCall(s string) (tool, input string, opts map[string]string, ok bool) {
	t := strings.TrimSpace(s)
	if !strings.HasPrefix(t, "{") {
		i := strings.Index(t, "{")
		if i <= 0 {
			return "", "", nil, false
		}
		t = t[i:]
	}
	var c struct {
		Tool  string            `json:"tool"`
		Input string            `json:"input"`
		Opts  map[string]string `json:"opts"`
	}
	if err := json.Unmarshal([]byte(t), &c); err != nil || strings.TrimSpace(c.Tool) == "" {
		return "", "", nil, false
	}
	return c.Tool, c.Input, c.Opts, true
}

func agentExec(db *DB, tool, input string, opts map[string]string) (string, bool, int) {
	if _, exists := ToolRegistry[tool]; !exists {
		return "工具不存在: " + tool, false, 0
	}
	if !db.ToolOn(tool) {
		return "模块未开启: " + tool, false, 0
	}
	t0 := time.Now()
	out, err := runTool(tool, input, opts)
	ms := int(time.Since(t0).Milliseconds())
	if err != nil {
		return "执行失败: " + err.Error(), false, ms
	}
	return out, true, ms
}

// agentRunOnce 抽取自 /api/agent/run 循环核: 单profile完整跑一轮agent
// convID>0 落轨迹, eval传0纯跑. 返回 answer/calls/rounds/tokens/ms/err
func agentRunOnce(db *DB, client *http.Client, p LLMProfile, input string, maxRounds, maxTokens int, convID int64) (string, []agentCall, int, int64, int64, error) {
	msgs := []map[string]string{
		{"role": "system", "content": agentToolPrompt()},
		{"role": "user", "content": input},
	}
	var calls []agentCall
	answer, es := "", ""
	rounds := 0
	var tokTotal int64
	t0 := time.Now()
	for rd := 0; rd < maxRounds; rd++ {
		content, tok, _, err := llmCallU(client, p, msgs, maxTokens)
		if err != nil {
			es = err.Error()
			break
		}
		rounds++
		tokTotal += tok
		tool, input2, opts, isCall := agentParseToolCall(content)
		if !isCall {
			answer = content
			break
		}
		msgs = append(msgs, map[string]string{"role": "assistant", "content": content})
		out, okc, ms2 := agentExec(db, tool, input2, opts)
		calls = append(calls, agentCall{Tool: tool, Input: llmTrunc(input2, 500), Output: llmTrunc(out, 4000), OK: okc, MS: ms2})
		if convID > 0 {
			tag := "fail"
			if okc {
				tag = "ok"
			}
			_, _ = dbMsgAppend(convID, "assistant", fmt.Sprintf("[tool] %s %s (%dms)\n%s", tool, tag, ms2, llmTrunc(out, 2000)))
		}
		msgs = append(msgs, map[string]string{"role": "user", "content": "工具结果: " + llmTrunc(out, 8000)})
	}
	ms := time.Since(t0).Milliseconds()
	if rounds == 0 {
		return "", calls, 0, 0, ms, fmt.Errorf("%s", llmTrunc(es, 200))
	}
	if answer == "" {
		if convID > 0 {
			_, _ = dbMsgAppend(convID, "assistant", fmt.Sprintf("轮次耗尽(%d轮,%d次工具调用)", rounds, len(calls)))
		}
		return "", calls, rounds, tokTotal, ms, fmt.Errorf("maxRounds 耗尽")
	}
	if convID > 0 {
		_, _ = dbMsgAppendP(convID, "assistant", answer, p.Name, tokTotal)
	}
	return answer, calls, rounds, tokTotal, ms, nil
}

func registerAgentRoutes(mux *http.ServeMux, s *srv, db *DB) {
	st := newLLMStore()
	client := &http.Client{Timeout: 90 * time.Second}

	mux.HandleFunc("POST /api/agent/run", func(w http.ResponseWriter, r *http.Request) {
		u := s.user(r)
		if u == nil {
			E(w, 401, "请先登录")
			return
		}
		var in agentReq
		if body(r, &in) != nil || strings.TrimSpace(in.Input) == "" {
			E(w, 400, "input 必填")
			return
		}
		if in.MaxRounds <= 0 || in.MaxRounds > 8 {
			in.MaxRounds = 4
		}
		var cands []LLMProfile
		if in.ProfileID != "" {
			if p, found := st.get(in.ProfileID); found {
				cands = append(cands, p)
			}
		}
		for _, p := range st.listChat() {
			if p.Enabled && p.ID != in.ProfileID {
				cands = append(cands, p)
			}
		}
		if len(cands) == 0 {
			E(w, 422, "没有可用档案,请先在模型设置中添加并启用")
			return
		}
		convID := in.ConversationID
		if convID > 0 && !dbConvOwn(convID, u.Username) {
			E(w, 403, "conversation not found")
			return
		}
		if convID == 0 {
			if cid, err := dbConvCreate(u.Username, "agent: "+llmTrunc(in.Input, 40)); err == nil {
				convID = cid
			}
		}
		if convID > 0 {
			_, _ = dbMsgAppend(convID, "user", llmTrunc(in.Input, 4000))
		}
		var errs []string
		for _, p := range cands {
			answer, calls, rounds, tokTotal, ms, aerr := agentRunOnce(db, client, p, in.Input, in.MaxRounds, in.MaxTokens, convID)
			if aerr != nil {
				if rounds == 0 {
					errs = append(errs, p.Name+": "+llmTrunc(aerr.Error(), 120))
					continue
				}
				J(w, map[string]any{"ok": false, "error": aerr.Error(), "conversationId": convID, "rounds": rounds, "calls": calls, "profile": p.Name})
				return
			}
			J(w, map[string]any{"ok": true, "conversationId": convID, "rounds": rounds, "calls": calls, "answer": answer, "profile": p.Name, "tokens": tokTotal, "ms": ms})
			return
		}
		E(w, 502, strings.Join(errs, " | "))
	})
}
