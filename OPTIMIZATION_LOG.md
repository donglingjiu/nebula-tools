# 迭代优化日志（OPTIMIZATION_LOG）

> 每次优化追加一条。新条目置顶。

---

## [2026-11-05] OPT-001 功能池 Feature Flag + SSE 实时上下线

### 变更
- 新增 feature.go：21 个种子功能 x 6 分类 + 4 个 API（公开列表/SSE 流/管理列表/实时 toggle）
- 新增 FeatureWrap 中间件 + init() 双兼容路由；main.go 自动 sed 接线
- 新增 web/features.html（用户端）+ web/admin/features.html（管理端），暖色治愈系
- 新增 .github/workflows/build.yml（CI 即编译验证器）+ scripts/validate.sh + TODO.md

### 对标来源
- Dify 灰度发布/功能开关（online/beta/gray/offline 四态 + 灰度百分比）
- Coze 插件上下线管理（管理端实时操作，前台秒级生效）

### 优点
1. 后台 toggle 后 SSE 广播，前台秒级感知无需刷新
2. 纯 stdlib 零依赖；CI 云端编译，开发机无 Go 环境也能闭环
3. 网络容错：git fail-fast（lowSpeedLimit）+ push 五次重试

### 缺点 / 已知限制
1. 内存存储重启丢状态（TODO 已记持久化项）
2. ADMIN_TOKEN 未设时管理接口全放行（dev 模式）
3. 开发机老 gcc 需 CGO_ENABLED=0（已固化为默认构建方式）
4. gin/chi/echo 路由需手动接 FeatureWrap（检测到会提示）
5. 种子 21 个功能，60+ 目标分批补齐（TODO 已记）

### 影响面
- NT-2/NT-4 复用同一 wrap/init/SSE 模式与 CI 流水线

## [2026-11-05] OPT-001b 接线修复 + 冲突保护

### 变更
- main.go 接线：secMiddleware(newRouter(db)) -> secMiddleware(FeatureWrap(newRouter(db)))，改用老 sed 兼容的字面替换（开发机 sed 无 -E）
- 冲突保护：检测到其他 Go 文件已占用 api/features 时自动改挂 /api/pool/*，避免遮蔽本地 T-07 rollout 体系
- TODO 新增：功能池与 ToolRegistry/ToolOn、T-07 rollout 合并去重

### 优点
1. 不侵入 newRouter 内部，一行包装即生效；auth 仍由 secMiddleware 统一管辖
2. 与本地已有的工具开关/rollout 体系共存互不破坏

### 缺点 / 已知限制
1. 若 secMiddleware 对 /api/* 统一鉴权，功能池管理接口需带会话或豁免（见 smoke 401 时处理）
2. 功能池与 ToolOn 双体系并存，待合并

## [2026-11-05] OPT-002 六家模型预设一键填充（NT-2 收尾）

### 变更
- web/llm.html 注入自包含预设区块（ntp- 前缀零冲突，幂等注入 </body> 前）：6 家厂商卡片点击自动填 名称/BaseUrl/Model + 独立保存/试一句按钮，直调既有 /api/llm/profiles 与 /api/llm/chat
- TODO.md 核销：Dockerfile/docker-compose/release/README 部署章节均已在库（此前 16 个本地提交），NT-2 标记完成

### 对标来源
- Dify 模型供应商预设（点选厂商→填 Key 即用，免手填 endpoint）
- Coze 插件配置卡（卡片化参数预设）

### 优点
1. 后端 llmproxy.go（CRUD/test/chat+故障切换）与 dbprofiles.go（SQLite+Key 加密+掩码）复用既有实现，零后端改动
2. 注入式增强不碰现有表单/JS，可重复执行；Anthropic 协议差异已标注（官方 /v1/messages → 走 OpenAI 兼容网关）
3. 六家预设：OpenAI/DeepSeek/智 GLM/Kimi/Qwen 兼容模式/Anthropic(OpenRouter)

### 缺点 / 已知限制
1. 试一句按钮固定走候选链第一个可用档案，未按档案单独测试（既有列表区已有单档案测试则够用）
2. OpenRouter 预设仅为 Anthropic 兼容示例，用户需自备其 Key；官方直连需等 /v1/messages 协议适配（TODO）
3. 若后台对 /api/llm/* 强制登录，未登录保存会 401（页面已提示先登录）

### 影响面
- NT-3 全流程 7 步可直接消费 /api/llm/chat（多档案 fallback 已就绪）
- NT-4 监控的 AI token 统计挂在 llmCall 返回 usage 上（dbprofiles/aiops 已有 AICall 三件套）

## [2026-11-05] OPT-003 全流程 7 步（NT-3）

### 变更
- 新增 pipeline.go：7 步状态机（todo/running/done + current 推进）+ JSON 落盘 data/pipelines/ + chatbody 组装端点（把上一步产物嵌入 prompt）
- 新增 web/pipeline.html：stepper + AI 生成/保存/完成 + 全流程导出 MD，暖色治愈系
- main.go 接线 secMiddleware(PipelineWrap(FeatureWrap(newRouter(db))))；功能池 wf-pipeline-7 种子切 online

### 对标来源
- product-expert-team 7 步产物模板（头脑风暴→评估→分析→方案→拆解→实现→测试）
- Dify 工作流节点逐步执行 / Coze 多轮产物沉淀

### 优点
1. 两段式 AI：后端零 AI 依赖（组装 chatbody），前端直调 /api/llm/chat 复用多档案 fallback，天然支持重新生成与人工修订
2. 产物 JSON 落盘，导出 MD 一步到位；7 步 prompt 内置上下文链（每步自动带上一步全文，超 4000 字截断）
3. 功能池卡片直达，与 NT-1/NT-2 形成完整链路：功能池 → 模型档案 → 全流程

### 缺点 / 已知限制
1. prompt 组装为固定模板，暂未开放自定义（P2 配置化）
2. 步间上下文只带上一步全文（非全部历史），长流程可能丢早期信息
3. 无并发编辑保护（单管理员场景够用）
4. AI 生成质量依赖所配模型（glm-4-flash 实测可用）

### 影响面
- NT-4 监控可在此链路上统计每步 token/耗时（chatbody→chat 已有明确埋点位）
