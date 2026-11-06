# TODO

> 完成打 x；新需求追加到对应分组。

## 部署与交付（2026-11-05 用户提出：之前支持过二进制/Docker 部署，需补齐并优化）
- [x] Dockerfile 多阶段构建（已存在，见 README 部署A / Dockerfile）
- [x] docker-compose.yml（已存在）
- [ ] Makefile：build / docker-build / docker-push / release 目标
- [x] GitHub Actions release：tag 自动三平台二进制+GHCR 双架构（已存在，见 README 发版）
- [x] README 部署章节（DEPLOY.md + README 部署A/B 已存在）

## NT 迭代
- [x] NT-1 功能池 Feature Flag + SSE 实时上下线（代码落地，CI 验证中）
- [x] NT-2 六家模型 API Key 配置（后端 llmproxy+dbprofiles 既有；llm.html 增 6 家预设一键填充）
- [x] NT-3 全流程 7 步（pipeline.go 状态机+落盘 + pipeline.html + AI 两段式消费 /api/llm/chat）
- [x] NT-4 监控看板（metrics.go+MetricsWrap 最外层接线+admin-metrics.html 5s 轮询，对齐 llmops-metrics.md）
- [ ] 功能池持久化（内存 -> store.go/PostgreSQL，重启不丢状态）
- [ ] 功能池 21 -> 60+（PPT 模块补齐、更多文档/表格 AI 增强）

## 基础设施
- [ ] 真实鉴权（JWT + RBAC，替换 ADMIN_TOKEN dev 放行）
- [ ] 真实文件转换后端（替换演示数据）
- [ ] E2E 浏览器自动化 + 负载测试
- [ ] 功能池与现有 ToolRegistry/ToolOn 及 T-07 rollout 管理侧合并去重（Feature.Percent/putFeature/listFeatures）
- [ ] Makefile（build.sh 已覆盖 build/dev/smoke/verify，可按需补 Makefile 薄封装）
- [x] 首页横幅修复（破损标签残留清除 + 尾部重复去重，见 OPT-006）
- [x] /api/history 匿名返回 200 空数组（语义：未登录=无历史，消除控制台 401）
- [ ] ai-studio.html 与 pipeline.html 双流程工作台合并去重
- [ ] pipeline 导出 MD 应包含未保存的当前 textarea 内容
- [ ] AI token 级统计挂 llmCall usage（对齐 llmops-metrics.md tokens_in/out，接 aiops AICall 三件套）
- [ ] /api/metrics 未鉴权（生产需 ADMIN_TOKEN 校验或反代门禁；当前开发网可用）
- [ ] 观察：若 /api/favs 匿名 401 也出现在控制台，同法处理（当前仅 history 有噪音）
- [x] sweep 已加 -L（301 误报修正）
- [ ] 登录页 GitHub/Google/微信 按钮为装饰未接后端（2026-11-05 截图），接 OAuth 或移除
- [ ] P2-1 双工作台合并（侦察已完成，方案定稿待拍板）：以 ai-studio 为基座（JWT/付费墙 402·410/追问 flow/chat/流式/导出 zip+md 齐备，首页+导航正式入口），移植 pipeline 独有三件套：①7 步精选提示词模板（pipeline.go pipePrompt）②历史列表 ③data/pipelines 持久化。三步走：A) internal/ai/flow.go 加 standard7 模板 + flow/list 端点 B) ai-studio.js/html 加"标准 7 步"入口 C) pipeline.html 变跳转页、feature.go 卡片改路由、旧 API 留一版本后清理。动手前先读 flow.go 全文（9.8KB，含 flow_test.go 7KB）
- [ ] P2-2 第三工作台盘点：web/aiworkflow.html（17.6KB）与 ai-studio 重叠度未知，随 P2-1 一并裁决去留
- [x] P2-1.A 后端标准7步：start7/list 端点 + truncRunes 护栏（OPT-008）
- [x] P2-1.B 前端标准七步：ai-studio 注入入口+历史列表（OPT-009）
- [x] ai-studio 双 UI 统一：静态死壳+hero容器+内联死卡全灭（hero 接线方案作废: 控件本就未渲染, OPT-013a-d）
- [x] P2-1.C pipeline.html 跳转 ai-studio + feature 卡片改路由（OPT-010）
- [x] AI Profile 配置：glm-4-flash 接入（OPT-011/glm）
- [x] P2-2 aiworkflow 全量移植：edit-step/材料/按流程选模型 + 跳转化（OPT-012）
- [x] flowChat 模型 override（同 advanceOne 补丁, mock 功能探针双校验, OPT-014）
- [ ] 材料入库存储（现仅注入第一步 prompt, >8000 字截断）
- [ ] P3 候选：flowChat 模型 override / 独立快聊页(relaySSE) / 材料入库存储 / hero 单选第二项语义 / 旧 /api/llm 下线盘点
