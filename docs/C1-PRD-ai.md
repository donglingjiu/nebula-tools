# C1-PRD：ai 工作台接通真实模型（v2 定稿 · 2026-10-23 评审通过）

## 背景与现状
- 模型层已交付（llm.html）：/api/llm/chat 同源代理、多档案、Key 服务端存储（data/llm_profiles.json 权限 600，浏览器零接触）、连通测试 1252ms；
- ai.html 为完整交互壳（对话 UI + IME 卫兵），自述「真实模型尚未接入」；
- 参考库 ai-toolsBox 提供功能模式（多 Profile 兜底/受限 DSL/隐私分级），其 localStorage 明文存储不采纳——以 SQLite 服务端代之（用户架构决策）。

## 目标（本期做）
1. 对话接通 /api/llm/chat（含流式输出）；
2. 会话与消息持久化：data/nova.db（SQLite），NOVA_DB 环境变量可配；
3. 故障切换精细规则：仅可重试错误（网络/超时/408/429/5xx/格式错误）切下档；400/401/403 不切；流式已开始不切；默认关闭可在 UI 开启；
4. 档案表单 URL 安全清洗（origin+path、剥离凭据、强制 https、localhost 例外）；
5. 一键迁移：导出（nova.db + llm_profiles.json 打包）/ 导入（导入端点校验后落盘）。

## 非目标（本期不做）
- 多用户/鉴权体系；analytics 数据口径（另案）；workflows 定时触发（C1-workflows）。

## 数据模型（草案）
- conversations(id, title, profile_id, created_at, updated_at)；
- messages(id, conversation_id, role, content, created_at)。

## 验收标准
A. 发送消息经 /api/llm/chat 收到真实模型回复（流式）；
B. 重启服务/换浏览器后历史对话完整可见（SQLite 持久化）；
C. 主档案 Key 置错后，可重试错误自动切备用档案且对话不中断；401 不切换并明确报错；
D. 导出包在第二台机导入后对话+档案完整可用。

## 编码前置侦察
- app.js el() 语义单一性确认（ai-toolsBox el() 双语义惨案教训，60+ 处失效）；
- nova-server 现有路由风格与 llm 端点实现（复用其代理与档案读取逻辑）。

## 开放问题
- llm_profiles.json 是否并入 nova.db（建议：并入，迁移端点顺带完成）；
- 上下文窗口策略（全量历史 vs 滑动窗口）——编码阶段实测后定。

## 评审决议（2026-10-23）
- 验收标准 A-D：通过，按原文执行；
- llm_profiles.json：并入 nova.db（profiles 表），旧 JSON 首启自动迁入后保留只读备份；迁移端点顺带完成；
- 预留逻辑：按「预留逻辑（写库即埋钩）」节执行，后续优化零迁移成本。

## 市面对标（2026-10-23 检索）
LobeChat=平台化天花板（30+Provider/知识库RAG/分支对话/插件/MCP/TTS/多用户）【外部情报】；LibreChat=预设动态切换/分支/临时聊天/MCP；CherryStudio=联网搜索/MCP/参数调节简陋；OpenWebUI=自托管框架/配置复杂；Chatbox=对话基本功扎实/高级功能付费；客户端评估五维度=流式+Markdown、一键切模型、历史与同步、成本可见、隐私。
最优解定位：C1 做「最好的轻量自托管对话端」——Chatbox/LibreChat 级基本功 + 自家差异化（Key 服务端 600 保险柜、SQLite 持久化、兜底错误分类学），不追 RAG/插件/多用户等平台化重装备。
三档裁决：C1 采纳=A-D+流式Markdown+档案管理并入SQLite+双向迁移+URL清洗（+增补候选:usage落库回显）；预留钩子见下节；延后不追=RAG/插件MCP/语音/绘图/分支UI/多用户/云同步。

## 预留逻辑（写库即埋钩，后续优化零迁移成本）
- 框架级：schema_migrations(version,applied_at)——建表即上迁移框架；settings(key,value) KV——兜底开关/窗口策略等应用配置住这里，不再新增零散 JSON；
- 列级：profiles.priority/enabled（兜底链排序）；messages.tokens（成本面板预留）、messages.branch_of（分支对话预留）；conversations.user_id（多用户预留，默认 NULL）；
- 格式级：导出包顶层 schema_version 字段，导入端点校验版本、拒绝高于当前——未来格式升级留版本位；
- 增补候选（评审备注）：响应 usage 落 messages.tokens 并在 UI 回显 token 数（成本可见维度，代价≈0，编码阶段顺手则做）。

## C2 指向
拆解清单见 docs/C2-TASKS-ai.md；编码每轮沿用三闸（bash -n / verify / 截图对照），新体检项随对应轮次落 verify.sh。
