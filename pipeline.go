package main

// NT-3: full-flow 7-step pipeline. Self-contained: JSON storage under
// data/pipelines/. Zero touch to novaDB / router internals.

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const pipeDir = "data/pipelines"

const pipeP = "/api/pipeline"

var pipeKeys = []string{"brainstorm", "assessment", "analysis", "solution", "breakdown", "implementation", "testing"}
var pipeTitles = []string{"头脑风暴", "需求评估", "需求分析", "方案设计", "功能点拆解", "实现计划", "测试落地"}

type PipeStep struct {
	Index     int       `json:"index"`
	Key       string    `json:"key"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	Output    string    `json:"output"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type PipeRun struct {
	ID        string      `json:"id"`
	Title     string      `json:"title"`
	Idea      string      `json:"idea"`
	Current   int         `json:"current"`
	Steps     []*PipeStep `json:"steps"`
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt time.Time   `json:"updatedAt"`
}

var pipeMu sync.Mutex

func pipeNewRun(title, idea string) *PipeRun {
	id := fmt.Sprintf("p%d%04d", time.Now().UnixNano()/int64(time.Millisecond), rand.Intn(10000))
	r := &PipeRun{ID: id, Title: title, Idea: idea, Current: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	for i := 1; i <= 7; i++ {
		r.Steps = append(r.Steps, &PipeStep{Index: i, Key: pipeKeys[i-1], Title: pipeTitles[i-1], Status: "todo"})
	}
	return r
}

func pipeSaveRun(r *PipeRun) error {
	if err := os.MkdirAll(pipeDir, 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(pipeDir, r.ID+".json"), b, 0644)
}

func pipeLoadRun(id string) (*PipeRun, error) {
	b, err := os.ReadFile(filepath.Join(pipeDir, filepath.Base(id)+".json"))
	if err != nil {
		return nil, err
	}
	var r PipeRun
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func pipeListRuns() []*PipeRun {
	entries, _ := os.ReadDir(pipeDir)
	var out []*PipeRun
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		if r, err := pipeLoadRun(strings.TrimSuffix(e.Name(), ".json")); err == nil {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}

func pipeContext(run *PipeRun, step int) string {
	if step <= 1 {
		return run.Idea
	}
	for _, s := range run.Steps {
		if s.Index == step-1 {
			out := strings.TrimSpace(s.Output)
			if out == "" {
				return run.Idea + "\n\n(上一步暂无产物，请基于原始想法合理推导)"
			}
			if len(out) > 4000 {
				out = out[:4000] + "…(截断)"
			}
			return out
		}
	}
	return run.Idea
}

func pipePrompt(step int, run *PipeRun) string {
	ctx := pipeContext(run, step)
	switch step {
	case 1:
		return "你是产品头脑风暴专家。请基于原始想法进行发散。\n【原始想法】" + ctx +
			"\n输出 markdown：\n## 候选方向（3~4 个）\n| # | 方向 | 目标用户 | 预期价值 | 信心(高/中/低) |\n## 关键假设与风险\n每条假设附最小验证方式\n## 结论\n推荐进入下一步的 1~2 个方向及理由（200 字内）"
	case 2:
		return "你是需求评估专家，对上一步选定的方向做可行性评估。\n【上一步结论】" + ctx +
			"\n输出 markdown：\n## 假设清单（按风险降序）\n| # | 假设 | 验证方式 | 成本(低/中/高) | 结论 |\n## 优先级矩阵\n影响力×不确定性，列出四象限代表项\n## 评估结论\n继续/转向/放弃 + 一句话理由；若继续，明确 MVP 边界"
	case 3:
		return "你是资深产品经理，输出可开发的需求文档。\n【评估结论】" + ctx +
			"\n输出 markdown：\n## 用户与场景\n角色-场景-痛点\n## 功能需求（MoSCoW）\nMust/Should/Could/Won't 表格\n## 非功能需求\n性能/安全/合规，逐条可验证\n## 验收标准\n每条 Must 对应一条可测试验收"
	case 4:
		return "你是架构师，基于需求做方案设计。\n【需求文档】" + ctx +
			"\n输出 markdown：\n## 备选方案对比\n>=2 个方案，表格：维度/方案A/方案B\n## 技术选型与理由\n## 风险与缓解\n每条风险配缓解措施\n## 里程碑建议\nM1 可用 / M2 完整 / M3 增强"
	case 5:
		return "你是项目经理，把方案拆解为可执行任务。\n【方案设计】" + ctx +
			"\n输出 markdown：\n## 功能点拆解\n| # | 功能点 | 用户故事 | 优先级 | 估时(人日) | 依赖 |\n## 里程碑切分\nM1（1~2 周）/ M2 / M3 各自包含的功能点"
	case 6:
		return "你是全栈工程师，基于功能点拆解生成本轮实现计划与自查清单（真实变更由人工在执行中补充）。\n【功能点拆解】" + ctx +
			"\n输出 markdown：\n## 本轮实现范围\n选定进入 M1 的功能点及顺序\n## 模块变更计划\n| # | 模块 | 计划变更 | 自验证方式 |\n## 代码红线自查\n- [ ] 输入校验 - [ ] 错误处理 - [ ] 并发安全 - [ ] 敏感信息不落日志\n## 遗留问题"
	default:
		return "你是测试工程师，基于实现计划输出测试方案与执行记录。\n【实现计划】" + ctx +
			"\n输出 markdown：\n## 测试场景清单\n| # | 场景 | 步骤 | 预期 | 结果(待执行) |\n## 缺陷记录\n| # | 描述 | 严重级 | 状态 |\n## 上线检查单\n逐条可勾选"
	}
}

func pipeHNew(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ntWriteJSON(w, 405, map[string]string{"error": "POST required"})
		return
	}
	var req struct{ Title, Idea string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ntWriteJSON(w, 400, map[string]string{"error": "bad json"})
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		req.Title = "未命名流程"
	}
	run := pipeNewRun(strings.TrimSpace(req.Title), strings.TrimSpace(req.Idea))
	pipeMu.Lock()
	err := pipeSaveRun(run)
	pipeMu.Unlock()
	if err != nil {
		ntWriteJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	ntWriteJSON(w, 200, map[string]interface{}{"ok": true, "run": run})
}

func pipeHList(w http.ResponseWriter, r *http.Request) {
	runs := pipeListRuns()
	if runs == nil {
		runs = []*PipeRun{}
	}
	ntWriteJSON(w, 200, map[string]interface{}{"ok": true, "runs": runs})
}

func pipeHGet(w http.ResponseWriter, r *http.Request) {
	run, err := pipeLoadRun(r.URL.Query().Get("id"))
	if err != nil {
		ntWriteJSON(w, 404, map[string]string{"error": "run not found"})
		return
	}
	ntWriteJSON(w, 200, map[string]interface{}{"ok": true, "run": run})
}

func pipeHSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ntWriteJSON(w, 405, map[string]string{"error": "POST required"})
		return
	}
	var req struct {
		ID     string `json:"id"`
		Index  int    `json:"index"`
		Output string `json:"output"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
		ntWriteJSON(w, 400, map[string]string{"error": "bad request"})
		return
	}
	pipeMu.Lock()
	defer pipeMu.Unlock()
	run, err := pipeLoadRun(req.ID)
	if err != nil {
		ntWriteJSON(w, 404, map[string]string{"error": "run not found"})
		return
	}
	if req.Index < 1 || req.Index > 7 {
		ntWriteJSON(w, 400, map[string]string{"error": "index out of range"})
		return
	}
	st := run.Steps[req.Index-1]
	st.Output = req.Output
	st.UpdatedAt = time.Now()
	switch req.Status {
	case "todo", "running", "done":
		st.Status = req.Status
	}
	if st.Status == "" {
		st.Status = "running"
	}
	if st.Status == "done" && req.Index >= run.Current && run.Current < 7 {
		run.Current = req.Index + 1
	}
	run.UpdatedAt = time.Now()
	if err := pipeSaveRun(run); err != nil {
		ntWriteJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	ntWriteJSON(w, 200, map[string]interface{}{"ok": true, "run": run})
}

func pipeHDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ntWriteJSON(w, 405, map[string]string{"error": "POST required"})
		return
	}
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
		ntWriteJSON(w, 400, map[string]string{"error": "bad request"})
		return
	}
	pipeMu.Lock()
	defer pipeMu.Unlock()
	_ = os.Remove(filepath.Join(pipeDir, filepath.Base(req.ID)+".json"))
	ntWriteJSON(w, 200, map[string]interface{}{"ok": true})
}

func pipeHChatBody(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	idx := 0
	fmt.Sscanf(q.Get("index"), "%d", &idx)
	if idx < 1 || idx > 7 {
		ntWriteJSON(w, 400, map[string]string{"error": "index 1..7 required"})
		return
	}
	run, err := pipeLoadRun(q.Get("id"))
	if err != nil {
		ntWriteJSON(w, 404, map[string]string{"error": "run not found"})
		return
	}
	ntWriteJSON(w, 200, map[string]interface{}{
		"profileId": "",
		"messages": []map[string]string{
			{"role": "system", "content": "你是 Nova Tools 全流程助手，输出结构化 markdown，直接给内容不要客套。"},
			{"role": "user", "content": pipePrompt(idx, run)},
		},
		"maxTokens": 2000,
	})
}

// PipelineWrap routes pipeline APIs before the app's own router.
func PipelineWrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case pipeP + "/new":
			pipeHNew(w, r)
		case pipeP + "/list":
			pipeHList(w, r)
		case pipeP + "/get":
			pipeHGet(w, r)
		case pipeP + "/save":
			pipeHSave(w, r)
		case pipeP + "/delete":
			pipeHDelete(w, r)
		case pipeP + "/chatbody":
			pipeHChatBody(w, r)
		default:
			next.ServeHTTP(w, r)
		}
	})
}

// init registers on DefaultServeMux as fallback.
func init() {
	http.HandleFunc(pipeP+"/new", pipeHNew)
	http.HandleFunc(pipeP+"/list", pipeHList)
	http.HandleFunc(pipeP+"/get", pipeHGet)
	http.HandleFunc(pipeP+"/save", pipeHSave)
	http.HandleFunc(pipeP+"/delete", pipeHDelete)
	http.HandleFunc(pipeP+"/chatbody", pipeHChatBody)
}
