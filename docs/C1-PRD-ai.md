# C1-PRD：ai 工作台接通真实模型（v1 草案，待评审）

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
