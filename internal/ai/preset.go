package ai

// OPT-008/012: 标准7步模板(自NT-3移植) + 材料注入 + 按流程选模型 + 步骤产出编辑。
// 复用 flow 引擎: 服务端AI/门禁付费墙/上下文链/追问/导出。
// 上一步结论经 advanceOne 注入的 system 消息传递; truncRunes 为长链护栏。

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"nova/internal/admin"
)

type stdStep struct {
	Key   string
	Title string
	Tmpl  string
}

var std7 = []stdStep{
	{"brainstorm", "头脑风暴", "你是产品头脑风暴专家。请基于原始想法进行发散。\n【原始想法】{{IDEA}}\n输出 markdown：\n## 候选方向（3~4 个）\n| # | 方向 | 目标用户 | 预期价值 | 信心(高/中/低) |\n## 关键假设与风险\n每条假设附最小验证方式\n## 结论\n推荐进入下一步的 1~2 个方向及理由（200 字内）"},
	{"assessment", "需求评估", "你是需求评估专家，对上一步选定的方向做可行性评估。\n【上一步结论】（见系统消息「前序步骤输出」最后一段）\n输出 markdown：\n## 假设清单（按风险降序）\n| # | 假设 | 验证方式 | 成本(低/中/高) | 结论 |\n## 优先级矩阵\n影响力×不确定性，列出四象限代表项\n## 评估结论\n继续/转向/放弃 + 一句话理由；若继续，明确 MVP 边界"},
	{"analysis", "需求分析", "你是资深产品经理，输出可开发的需求文档。\n【上一步结论】（见系统消息「前序步骤输出」最后一段）\n输出 markdown：\n## 用户与场景\n角色-场景-痛点\n## 功能需求（MoSCoW）\nMust/Should/Could/Won't 表格\n## 非功能需求\n性能/安全/合规，逐条可验证\n## 验收标准\n每条 Must 对应一条可测试验收"},
	{"solution", "方案设计", "你是架构师，基于需求做方案设计。\n【上一步结论】（见系统消息「前序步骤输出」最后一段）\n输出 markdown：\n## 备选方案对比\n>=2 个方案，表格：维度/方案A/方案B\n## 技术选型与理由\n## 风险与缓解\n每条风险配缓解措施\n## 里程碑建议\nM1 可用 / M2 完整 / M3 增强"},
	{"breakdown", "功能点拆解", "你是项目经理，把方案拆解为可执行任务。\n【上一步结论】（见系统消息「前序步骤输出」最后一段）\n输出 markdown：\n## 功能点拆解\n| # | 功能点 | 用户故事 | 优先级 | 估时(人日) | 依赖 |\n## 里程碑切分\nM1（1~2 周）/ M2 / M3 各自包含的功能点"},
	{"implementation", "实现计划", "你是全栈工程师，基于功能点拆解生成本轮实现计划与自查清单（真实变更由人工在执行中补充）。\n【上一步结论】（见系统消息「前序步骤输出」最后一段）\n输出 markdown：\n## 本轮实现范围\n选定进入 M1 的功能点及顺序\n## 模块变更计划\n| # | 模块 | 计划变更 | 自验证方式 |\n## 代码红线自查\n- [ ] 输入校验 - [ ] 错误处理 - [ ] 并发安全 - [ ] 敏感信息不落日志\n## 遗留问题"},
	{"testing", "测试落地", "你是测试工程师，基于实现计划输出测试方案与执行记录。\n【上一步结论】（见系统消息「前序步骤输出」最后一段）\n输出 markdown：\n## 测试场景清单\n| # | 场景 | 步骤 | 预期 | 结果(待执行) |\n## 缺陷记录\n| # | 描述 | 严重级 | 状态 |\n## 上线检查单\n逐条可勾选"},
}

func truncRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…(截断)"
}

// flowStartStd POST /api/ai/flow/start7 {title, idea, material?, profile_id?}
func (s *Service) flowStartStd(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string   `json:"title"`
		Idea        string   `json:"idea"`
		Material    string   `json:"material,omitempty"`
		MaterialIDs []string `json:"material_ids,omitempty"`
		ProfileID   string   `json:"profile_id,omitempty"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil ||
		strings.TrimSpace(req.Idea) == "" {
		fail(w, 400, "参数错误: 需要 idea")
		return
	}
	u := s.auth.UserFrom(r)
	if u == nil {
		fail(w, 401, "未登录")
		return
	}
	if req.ProfileID != "" && !admin.ProfileAvailable(s.dir, req.ProfileID) {
		fail(w, 400, "指定的模型 Profile 不可用")
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = "标准七步: " + truncRunes(strings.TrimSpace(req.Idea), 30)
	}
	st, err := s.loadFlows()
	if err != nil {
		fail(w, 500, "读取流程失败")
		return
	}
	f := &Flow{
		ID:        newFlowID(),
		UserID:    u.ID,
		Title:     title,
		State:     "created",
		CreatedAt: time.Now().Format(time.RFC3339),
		ProfileID: req.ProfileID,
	}
	for i, sd := range std7 {
		p := sd.Tmpl
		if i == 0 {
			p = strings.Replace(p, "{{IDEA}}", strings.TrimSpace(req.Idea), 1)
			if m := strings.TrimSpace(req.Material); m != "" {
				p += "\n\n【参考资料】\n" + truncRunes(m, 8000)
			}
			used := 0
			for _, mid := range req.MaterialIDs {
				mat, merr := s.GetMaterial(mid)
				if merr != nil || mat == nil {
					continue
				}
				if used >= 30000 {
					p += "\n\n(库材料已达注入上限, 其余未附)"
					break
				}
				body := truncRunes(mat.Content, 12000)
				p += "\n\n【库材料·" + mat.Title + "】\n" + body
				used += len([]rune(body))
			}
		}
		f.Steps = append(f.Steps, FlowStep{Name: sd.Title, Prompt: p, Status: "pending"})
	}
	st.Flows[f.ID] = f
	if err := s.saveFlows(st); err != nil {
		fail(w, 500, "保存流程失败")
		return
	}
	writeJSON(w, 200, f)
}

// flowList GET /api/ai/flow/list
func (s *Service) flowList(w http.ResponseWriter, r *http.Request) {
	u := s.auth.UserFrom(r)
	if u == nil {
		fail(w, 401, "未登录")
		return
	}
	st, err := s.loadFlows()
	if err != nil {
		fail(w, 500, "读取流程失败")
		return
	}
	out := []*Flow{}
	for _, f := range st.Flows {
		if canAccess(f, u) {
			out = append(out, f)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt > out[j].CreatedAt })
	if out == nil {
		out = []*Flow{}
	}
	writeJSON(w, 200, out)
}

// flowEditStep POST /api/ai/flow/edit-step {id, index, output}
// OPT-012 独有①: 已完成步骤产出可编辑, 后续步骤经 flowCtx 消费新内容(human-in-the-loop)
func (s *Service) flowEditStep(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID     string `json:"id"`
		Index  int    `json:"index"`
		Output string `json:"output"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.ID == "" {
		fail(w, 400, "参数错误: 需要 id/index/output")
		return
	}
	u := s.auth.UserFrom(r)
	if u == nil {
		fail(w, 401, "未登录")
		return
	}
	st, err := s.loadFlows()
	if err != nil {
		fail(w, 500, "读取流程失败")
		return
	}
	f, ok := st.Flows[req.ID]
	if !ok || !canAccess(f, u) {
		fail(w, 404, "流程不存在")
		return
	}
	if req.Index < 0 || req.Index >= len(f.Steps) {
		fail(w, 400, "步骤序号非法")
		return
	}
	if f.Steps[req.Index].Status != "done" {
		fail(w, 400, "仅已完成步骤可编辑")
		return
	}
	f.Steps[req.Index].Output = strings.TrimSpace(req.Output)
	if err := s.saveFlows(st); err != nil {
		fail(w, 500, "保存流程失败")
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "flow": f})
}

// flowProfiles GET /api/ai/flow/profiles — 启用模型安全列表(仅 id/name/model)
func (s *Service) flowProfiles(w http.ResponseWriter, r *http.Request) {
	u := s.auth.UserFrom(r)
	if u == nil {
		fail(w, 401, "未登录")
		return
	}
	writeJSON(w, 200, admin.ListEnabledProfiles(s.dir))
}
