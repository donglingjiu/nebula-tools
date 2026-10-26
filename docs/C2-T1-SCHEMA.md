# T1 Schema v1（migration 0001，驱动无关：modernc 与 mattn 均可执行）
CREATE TABLE IF NOT EXISTS schema_migrations(
  version INTEGER PRIMARY KEY,
  applied_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE TABLE IF NOT EXISTS settings(
  key TEXT PRIMARY KEY,
  value TEXT
);
CREATE TABLE IF NOT EXISTS profiles(
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  base_url TEXT NOT NULL,
  model TEXT NOT NULL,
  api_key TEXT NOT NULL,
  enabled INTEGER NOT NULL DEFAULT 1,
  priority INTEGER NOT NULL DEFAULT 100,
  created_at INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS conversations(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL DEFAULT '新对话',
  profile_id TEXT,
  user_id TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE TABLE IF NOT EXISTS messages(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  conversation_id INTEGER NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
  role TEXT NOT NULL CHECK(role IN ('user','assistant','system')),
  content TEXT NOT NULL,
  tokens INTEGER,
  branch_of INTEGER,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_messages_conv ON messages(conversation_id, id);
CREATE INDEX IF NOT EXISTS idx_profiles_fallback ON profiles(enabled, priority);
INSERT OR IGNORE INTO schema_migrations(version) VALUES (1);
## v1.1 修订（T1 编码对齐，2026-10-26）
- profiles.id 改 TEXT 主键：与现 JSON 字符串 ID 对齐，T2 迁入零 ID 折损；created_at 用 INTEGER 纪元（同源 JSON），conversations.profile_id 随之 TEXT。
- api_key 列存 encSecret 密文（与现 JSON 同法，读侧 decSecret），0600 权限+密文双保险。

## 落地细则
- NOVA_DB 默认 data/nova.db；start-nova.sh 注入 NOVA_DB=${NOVA_DB:-/home/work/fanguohui/data/nova.db}
- 权限铁律：api_key 入库后 nova.db 必须 chmod 600（继承保险柜义务）；体检候选「DB权限600」
- JSON 迁入：首启 profiles 表空且 llm_profiles.json 存在→按序 priority=100,101... enabled 保留 JSON 原值（缺省视为 1）→JSON 暂留原地，T2 读切换同发再改名 .imported-bak 只读备份
- 预留列对账：priority/enabled（兜底链）、tokens（成本）、branch_of（分支）、user_id（多用户）＝PRD v2 四钩齐活
- 体检预告：会话持久化在位＋DB权限600，随 T1 入册 27→29

## T2 执行注记（2026-10-26）
- 迁入触发点：newLLMStore 启动即调 dbProfilesMigrate（novaDB 就绪前提）——表空且 JSON 在才迁、事务包裹、成功即改名 .imported-bak 并 chmod 444。
- 读切换：llmStore 各方法内部改查 SQLite，api_key 列密文承袭（enc 落库/dec 读取）；新增 listChat 供兜底链取真钥（顺修掩码 401 隐患）；list 排序改 priority 升序。
- decSecret 容错明文：JSON 里明文钥迁入后读侧直返不破，下次 upsert 落库自动升级密文。
- 体检随 T2 增至 30 项（档案库SQLite化）。
- 测试退役：llmsec_test.go 随 JSON 载体退役（llmStore.Profiles 字段已随 T2 消失），at-rest 密文断言语义移植至 dbprofiles_test.go——新增 raw nova.db/-wal 文件字节级明文扫描，密文列与文件字节双保险。

## T3 执行注记（2026-10-26）
- 会话 API 落地：六路由（POST/GET /api/conversations、PATCH/DELETE /{id}、GET/POST /{id}/messages），s.user(r) 鉴权、user_id 取 u.Username；消息读写前置 dbConvOwn 归属检查，跨用户一律 404。
- 原子性：dbMsgAppend 事务内三步一体（验会话在→插消息→bump updated_at）；删除走外键级联（DSN 已带 foreign_keys(1)），会话删则消息自清。
- 预留：tokens/branch_of 列 T3 恒 NULL（T4/T8 接 usage 统计时启用）；profile_id 待 T5/T6 前端接线填充。
- 体检随 T3 增至 31 项（会话API在位，二进制路由字节串探针）。

## T4 执行注记（2026-10-26）
- tokens 列启用：dbMsgAppend 变长参数 tokens ...int64（零参=NULL 向后兼容，>0 才落库）；llmCallU 解析 usage.total_tokens，chat 带 conversationId 时事务落双消息。
- 新路由 GET /api/usage/summary（registerUsageRoute）：按日聚合 messages，user 隔离经 conversations 子查询。
- 体检 31→32（用量API在位）。
