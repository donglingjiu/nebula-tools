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

## [2026-11-05] OPT-006c /api/history 匿名 200 空数组（P1 收尾）

### 变更
- GET /api/history：未登录由 401 改为 200 + []（语义：未登录即无历史），消除首页 app.js:527 的控制台 401 红字
- 仅动该 handler 一行（awk 锚定 HistOf 边界），其余 401 鉴权语义不变

### 取舍
- 服务端修而非前端：app.js 为 vhash 版本化压缩产物，改服务端一行风险最小
- e2e 测试走登录态会话，不受影响；脚本内置 vet/build/test 三道闸，失败自动回滚

### 缺点 / 已知限制
- /api/favs 匿名仍 401（控制台未观察到噪音，按最小改动原则未动，TODO 记观察项）
- 若未来有前端需要区分"未登录"与"空历史"，需另加 /api/auth/me 判断（前端已有该调用）

## [2026-11-05] OPT-006d 修复两处前端损伤（回归 1 处 + 既有 1 处）

### 变更
- web/index.html：<a<a href= → <a href=，「立即体验」恢复可点（回归修复）
- web/register.html：p2 = document.( → p2 = (，消除 63:39 语法错误，注册按钮恢复（既有损伤）

### 根因复盘
- 回归：nt6 清残留时替换串误写 </span><a，与紧随的完整 <a 拼成非法标签 a<a；smoke 只断言"坏模式清零"未断言"好模式就位"
- 既有：register.html 历史替换在 ( 前误插 document.，整段脚本 parse 失败导致按钮失效

### 经验
1. HTML 修复断言必须双向：坏模式清零 + 好模式（<a href= 精确匹配）就位
2. 前端页面无编译期检查，语法错误只在浏览器控制台暴露——全页 sweep 纳入 smoke

## [2026-11-05] OPT-007 用户管理 CLI（cmd/pwreset）

### 变更
- internal/auth/store.go 增导出方法：SetPassword/SetRole/Remove/List（复用内部锁与 saveLocked）
- 新增 cmd/pwreset：list/setpw/verify/rm/promote 五子命令，复用项目自身哈希函数保证格式一致
- 清理 smoke_ 测试账号；重置 admin@example.com 密码并验证（HTTP 登录 200）

### 取舍
- 不做文件手术：store 内存常驻，直接改 users.json 会被下次保存覆盖；必须走工具 + 重启
- 本机 gcc 过老无法编 cgo 运行时，全部构建/运行统一 CGO_ENABLED=0（与主程序一致）

### 缺点 / 已知限制
- pwreset 无鉴权（本机执行），生产环境应限制数据目录的主机访问权限
- 管理操作后建议重启服务，避免与服务端内存副本的并发写窗口

## [2026-11-06] OPT-008 标准7步模板并入 flow 引擎（P2-1 步骤 A）

### 变更
- internal/ai/preset.go：std7 模板表（移植 NT-3 pipePrompt，角色+输出结构逐字保留）+ flowStartStd + flowList + truncRunes
- flow.go flowCtx：每步输出 4000 rune 截断护栏（7 步长链防 token 膨胀）
- 路由：POST /api/ai/flow/start7、GET /api/ai/flow/list（锚点镜像插入）
- flow_test.go：TestTruncRunes

### 取舍
- 上一步结论走 advanceOne 已有的 system 消息（前序步骤输出），模板零插值 → advanceOne 零改动
- pipeline 两阶段（前端直发 AI）无门禁无计费，不配当基座；其模板+历史列表语义已全部后移

### 缺点 / 已知限制
- flowCtx 汇总全部已完成步骤（非仅上一步），长流程上下文更丰富但 token 更多，护栏 4000/步
- 旧 /api/pipeline* 端点暂留，C 步（前端切换 + 过渡期后清理）另行提交

## [2026-11-06] OPT-008b 路由加固 + smoke 诊断改进

### 变更
- start7/list 补包 asvc.RequireRole("user", ...)（与既有 flow 路由完全一致）
- smoke 登录改为「raw 响应打印 + tr 去空白后 grep 提取」，杜绝错误响应被当 token 的长度检查漏洞

### 经验
1. 锚点镜像必须镜像整行语义（含中间件包裹），只取前缀会漏掉横切关注点
2. 管理操作（pwreset setpw）必须在服务停止后执行，运行中内存副本会覆盖文件写入

## [2026-11-06] OPT-009 ai-studio 前端接入标准七步 + 历史流程（P2-1 步骤 B）

### 变更
- web/assets/js/ai-studio.js（纯增量，闭包内）：runStd7（start7→flow/run 一键，失败提示逐步兜底）+ loadHist（flow/list 前10条，点击载入 renderFlow 并设 nbswFlowId，追问/导出/推进一步全复用）
- ui() 尾部注入「��� 标准七步」「↻ 历史流程」两按钮 + nbsw-hist 容器
- vhash bump 防 JS 缓存

### 取舍
- 独立按钮而非第三 radio：不触碰 createRun 与 select innerHTML（raw 字节不可信），零回归面
- 新代码插在闭包内 function ui() 前，复用 api/note/handleErr/renderFlow，不重复造轮子

### 缺点 / 已知限制
- 一键跑 7 步×真实 AI 耗时较长，超时/付费墙时降级为「推进一步」
- 历史列表返回完整 Flow（含全部输出），流量偏大；>10 条不分页
- styled hero 与页尾组件的双 UI 打通记 TODO

## [2026-11-06] OPT-010 pipeline 跳转化 + 收尾打磨（P2-1 步骤 C，双台归一完成）

### 变更
- pipeline.html → 自动跳转页（meta refresh + location.replace → /ai-studio.html），文案说明承接关系
- feature 卡片 /pipeline.html → /ai-studio.html（Go 源锚点验证）
- ai-studio.js：���→▶ 字体兼容；flow/run state=error 时提示「上游 AI 未配置」替代误导性「生成中」
- sweep 加 -L 修正 301 误报

### 验收（浏览器实测）
- 标准七步按钮 / 历史列表 / 打开渲染 7 步 / 追问导出接管 全通过
- AI 步骤 error 根因 = 上游占位符 api.example.com（配置问题非代码），机械链路完整验证

### 缺点 / 已知限制
- 旧 /api/pipeline* API 保留过渡，前端已无调用方
- aiworkflow.html（P2-2）待盘点裁决

### 勘误（OPT-010 补，nt15b）
- nt15 提交步骤路径笔误（web/ai-studio.js ≠ web/assets/js/ai-studio.js）致 git add fatal，变更滞留工作区；nt15b 修正提交
- feature 卡片为裸字符串 "/pipeline.html"（非 href=），nt15 探测模式漏改；nt15b 改 feature.go 并过 gate

## [2026-11-06] OPT-011b key 清洗 + 上游验证（glm）

### 诊断链（nt16b→c→d）
1. nt16c C 步全量解析定位：net/http invalid header = key 含 \r（剪贴板 CRLF 粘贴，read 只吃 \n）
2. 错误 URL=open.bigmodel.cn → default_profile_id 生效（占位符嫌疑排除）
3. curl exit=60 → 该域 TLS 证书验证失败（git/github 正常 → 疑似企业网对 AI 域名的选择性拦截）
4. nt16d：粘贴清洗（tr -d 空白）重建 profile → 默认切换 → 按 TLS 结果条件回落 mock

### 结果
- 真实调用 glm-4-flash 验证通过（nt16d 清洗 key 后）
### 经验
- 一切粘贴输入的 secret 必须在入口清洗空白字符；HTTP 头非法字符错误 = 数据脏不是网络断

## [2026-11-06] OPT-012 aiworkflow 三独有全量移植进 flow 引擎（P2-2, option A）

### 变更
- admin/profile_lookup.go：ResolveProfileByID(锁模式同默认解析)+ProfileAvailable+ListEnabledProfiles(仅id/name/model)
- flow.go：Flow.ProfileID 字段 + advanceOne override（flowRun/flowAdvance 共用咽喉, 已验证锚点）
- preset.go：start7 增 material(第一步注入, 8000 rune 上限)/profile_id(创建时校验启用)；flowEditStep(done 步骤产出编辑, running 可编辑=人工把关核心)；flowProfiles
- ai-studio.js：模型下拉+材料上传(.txt/.md/.csv)+步骤编辑面板(下拉选done步→textarea→保存→后续步骤吃新产出)
- aiworkflow.html 跳转化 + feature 卡改路由；legacy /api/llm 保留(AI 对话页仍在用, P3 处理)

### 取舍
- override 只打 advanceOne 已验证锚点, flowChat 未动（聊天用默认模型, 记 TODO）
- 材料仅注入第一步 prompt, 靠上下文链向后传递, 不落库
- 编辑不设 running 拦截: 改完继续正是功能灵魂, 并发编辑竞态接受(单用户工具)

### 缺点 / 已知限制
- flowChat 模型 override 待补；材料>8000字截断；编辑面板用 textarea 无 markdown 预览

## [2026-11-06] OPT-012c 二进制修复 + 粘贴通道规约

### 事故（双层）
1. nt17 marker sed 坏分支产生裸星号行（L225）-> IIFE 语法爆炸
2. 终端粘贴通道把 4 字节 emoji（���/��� 非 BMP）搅成非法 UTF-8（0xED CESU 痕迹）-> 文件本身非法，Python 严格解码的修复工具自己也跑不了；grep/sed 字节级操作全程无感
### 修复
- 字节模式：marker 正则修复 + surrogateescape 剥离全部非法字节 + 严格解码复验 + 侍服内容复验
### 规约（即刻生效）
- 一切经终端 cat 粘贴的脚本/代码：禁用全部 4 字节 emoji；符号仅用 BMP（▶ ↻ ✓ -）；每次改 JS 后必须过严格 UTF-8 解码门
