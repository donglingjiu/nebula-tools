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

## [2026-11-05] OPT-004 监控看板（NT-4）

### 变更
- 新增 metrics.go：MetricsWrap 最外层中间件（计数含鉴权拒绝）+ /api/metrics 快照端点（req_total/rpm/error_rate/LLM calls+avg+p95/top8 路径/最近 60 条）
- 新增 web/admin-metrics.html：8 指标卡 + Top 路径 + 最近请求表，5 秒轮询，暖色系
- main.go 接线 MetricsWrap(secMiddleware(PipelineWrap(FeatureWrap(...))))；功能池新增"监控看板"卡片
- SSE /stream 路径计数但豁免延迟统计（长连接不污染 P95）

### 对标来源
- AutoDev references/llmops-metrics.md（指标 schema、三层聚合、报警阈值 error_rate>5%）
- Dify LLMOps 看板（token/延迟/成本）；Grafana RED 方法（Rate/Error/Duration）

### 优点
1. 零依赖内存实现，快照 JSON 一发即用，后续可对接 Prometheus exporter
2. 最外层包装连 401/404 都计入，错误率真实；LLM P95 直接支撑容量评估（NT-3 实测单次 30s）
3. 看板补齐 admin-monitor.html 里"规划中"的看板位，与其报警渠道形成闭环

### 缺点 / 已知限制
1. 内存态，重启清零（TODO 已记持久化/导出）
2. token 级统计未挂（需动 llmproxy，P2 接 AICall 三件套）
3. 单实例视角，多副本部署需聚合层

## [2026-11-05] OPT-004b 修复 /api/metrics 404（看板空白根因）

### 变更
- MetricsWrap 内 metrics 路径分支由 next.ServeHTTP（透传给内层，无人处理→404）改为 metHJSON 直接服务
- admin-metrics.html LLM P95 卡片修正运算优先级（||0+"ms" → Math.round(...)+"ms"）
- TODO 记入：metrics 端点生产环境需鉴权门禁

### 根因复盘
- MetricsWrap 设计意图是"自身路径自己服务、不计数"，但误写成透传；内层 mux（newRouter）无此路由 → 404
- 前端 fetch 的 catch 静默吞错，放大为"页面空白无提示"；后续前端统一加 console.warn

### 经验
1. Wrap 链里"短路服务"与"透传"必须显式区分；smoke 里对新增端点本身要做 200 断言（本次 smoke 恰好暴露，价值验证）
2. 静默 catch 适合轮询页面，但首次加载失败应有可见提示（后续改进）

## [2026-11-05] OPT-006 首页横幅修复（残留清除 + 尾部去重）

### 变更
- web/index.html：清除 </span> 后破损标签残留（font-weight:700;..."> ），横幅不再渲染样式文本
- 尾部去重：body 后整段内容重复两份（两个 </body></html>），截断保留第一份，文档结构恢复合法
- 修复幂等（已修复则跳过）+ 失败自动回滚备份

### 根因复盘
- 残留：疑似历史横幅 sed 替换只覆盖旧链接标签前半段，样式尾部成为可见文本
- 重复：注入式追加横幅/导航未做幂等保护，多次注入叠加
- 两者同源：对生成文件做非幂等文本注入

### 经验
1. 注入 HTML 必须"检测已存在再操作"，替换目标须精确到完整标签边界
2. go:embed 下改 web 文件需重编译重启才生效（已纳入本次 smoke）

## [2026-11-05] OPT-006b 修复 metrics.go 结构体 tag（CI vet 红根因）

### 变更
- mPathStat.TotalMs 的 tag `json:"-"` 补回丢失的闭合引号（nt4 heredoc 转义笔误），vet 恢复通过

### 根因复盘
- unquoted heredoc + 反引号转义写 struct tag，闭合引号丢失；编译/测试/运行均不受影响，唯 vet 静态检查拦截
- 发布流水线（tidy+vet+test）在 16 个提交合入后首次跑 vet，被本次新代码触发

### 经验
1. heredoc 生成含反引号的 Go 代码极易出错，struct tag 优先用 awk printf %c(96) 重建整行
2. vet 是 CI 里唯一拦住此问题的关卡，继续保留 continue-on-error=false
