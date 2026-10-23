# C2 拆解清单：ai 工作台接通真实模型（PRD v2 定稿对应，待开工排序）

## 侦察项（T0，编码前必答）
- R1 nova-server Go 侧 SQLite 驱动选型：modernc.org/sqlite（纯 Go 无 CGO，交叉编译友好）vs mattn/go-sqlite3（需 CGO）——以现有 build.sh 交叉编译链实测定夺；
- R2 app.js el() 语义单一性确认（ai-toolsBox 双语义惨案教训）；
- R3 /api/llm/chat 现有路由与档案读取逻辑梳理（复用面盘点）。

## 服务端（每任务独立可验）
- T1 SQLite 落地：NOVA_DB 变量 + 首启建库 + schema_migrations/settings 框架表（验收→B 之基座）；
- T2 profiles 并入：建 profiles 表（含 priority/enabled），llm_profiles.json 首启自动迁入留只读备份（验收→决策2）；
- T3 对话持久化：conversations/messages 表（含 tokens/branch_of/user_id 预留列）+ chat 端点写库（验收→B）；
- T4 兜底精细规则：错误分类（可重试/不可重试/流式已开始）+ 按 priority 切档（验收→C）；
- T5 迁移端点：/api/export 打包（带 schema_version）+ /api/import 校验落盘（验收→D）。

## 前端
- T6 ai.html 接流式对话 + Markdown 渲染（验收→A）；
- T7 会话列表（重命名/删除/切换）+ 历史回放（验收→B 之 UI 面）；
- T8 兜底开关（settings 读写）+ 档案表单 URL 安全清洗（PRD 目标4）+ usage 回显（若 T3 顺手埋列则加）。

## 闸门与体检
- 每任务一发一验：bash -n / verify.sh / 截图对照；
- T3-T5 落地后 verify.sh 新增体检项：「会话持久化在位」「导入导出在位」（随对应轮次入册，体检条数 27→29 预期）；
- 开工序建议：R1-R3 → T1 → T2 → T3+T6 → T4 → T7 → T5+T8（A-D 依序点亮）。
