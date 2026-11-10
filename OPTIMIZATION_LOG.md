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
- ui() 尾部注入「 标准七步」「↻ 历史流程」两按钮 + nbsw-hist 容器
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
- ai-studio.js：→▶ 字体兼容；flow/run state=error 时提示「上游 AI 未配置」替代误导性「生成中」
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
2. 终端粘贴通道把 4 字节 emoji（/ 非 BMP）搅成非法 UTF-8（0xED CESU 痕迹）-> 文件本身非法，Python 严格解码的修复工具自己也跑不了；grep/sed 字节级操作全程无感
### 修复
- 字节模式：marker 正则修复 + surrogateescape 剥离全部非法字节 + 严格解码复验 + 侍服内容复验
### 规约（即刻生效）
- 一切经终端 cat 粘贴的脚本/代码：禁用全部 4 字节 emoji；符号仅用 BMP（▶ ↻ ✓ -）；每次改 JS 后必须过严格 UTF-8 解码门

## [2026-11-06] OPT-013 死壳清扫 + hero 接线（P2 收尾）

### 侦破
- ai-studio.html 静态壳 = G5 时代多层克隆（hero/状态条/nbst 两步卡/阶段面板/产出汇总栏），全部无接线
- 活工作台为 ai-studio.js 纯 JS 构建（静态 html 无任何 nbsw-* id），与死壳三层叠放造成混淆
- ai-chat.html 零 API 引用 + 「阶段5接入」陈年占位 = 纯死壳

### 变更
- hero btnStart -> runStd7（死按钮变标准七步真入口）
- 死壳 CSS 隐藏，带护栏：候选 id 凡被 live js getElementById 引用一律豁免（防误杀活容器）
- ai-chat.html 跳转化 -> ai-studio

### 缺点 / 已知限制
- hero 单选「单阶段逐步跑」成装饰（点击一律 std7）；隐藏后若留空壳容器再微调


## [2026-11-06] OPT-013d 死壳终局 + 工具链复盘（合并勘误 013b/c）

### 死壳歼灭战全程
- OPT-013（dd755aa/dfeb450）：getElementById 护栏隐藏静态死壳 + ai-chat 跳转；hero 接线补丁因 bytes 字面量含中文未落地（事故1）
- OPT-013b（7513459）：marker 回溯隐藏 hero 容器；但取 id 窗口 300 字符抓到子元素 id，三处空放（事故2）；日志勘误因文件非法 UTF-8 未写成（事故3）
- OPT-013c（e067930）：日志修复 12 非法字节 + 精确选择器点射（aside[流程阶段]/.meter-wrap/#ideaInput/#btnStart）生效；去重再翻车于 py3.6 零宽 split（事故4）；底部死卡确认为内联脚本构建，整块摘除被 auth 护栏正确拒绝
- OPT-013d（本次）：w.innerHTML 前缀注入 display:none 击杀死卡（auth 零接触）；手工分段去重；TODO 失实行纠正
### 工具链规约（累计）
- 终端粘贴链路：4 字节 emoji 禁入一切 heredoc（js/py/md/json 通杀）
- py3.6：re.split 零宽 lookahead 不可用，分段用 finditer 手工切
- bytes 字面量只能 ASCII；所有 mutation 带硬门（存在性 grep 不等于语法/内容校验）

## [2026-11-06] OPT-014 flowChat 模型 override（补齐 TODO 尾巴）

### 变更
- flow.go flowChat：gateNow 块后插入同款 override（锚点: gateNow+writeGate 四行组合全文件唯一; f 在 gateNow 前已获取）
- 与 advanceOne 共用 ResolveProfileByID; 失败静默回落默认（行为与 advanceOne 一致）
### 验证
- mock(:18787) 功能探针（curl 实请求看 [mock] 回显）+ 绑定流程 advance/chat 双路; chat 用 [mock]+NBX014 双校验防回落假阳性
### 事故与规约（nt22/nt22b/nt22c 三幕）
- 现象: mock 两次"启动失败"但 mock.log 全空; fuser 曾在 18787 杀出两个陈年 PID
- 诊断: ss 检测不可靠（缺失/异常）-> 全部端口检查为假阴性, 而 mock 实际一直活着
- 规约: 端口/服务活性一律用功能探针（对端口发真实请求验证响应内容）, ss/固定 sleep 均不可作为生死判据; 后台启动 poll 上限 10s
### 已知限制
- ResolveProfileByID 失败（profile 运行期被删/禁用）静默回落默认模型, 创建时已校验, 属边角

## [2026-11-06] OPT-015 独立快聊页 quickchat（relaySSE 首个前端消费者）

### 变更
- web/quickchat.html：自包含页（无 core.js 依赖），fetch 流式读 SSE + delta 累积渲染；token 多级 fallback；Enter 发送/Shift+Enter 换行；历史仅内存
- feature.go：AI 快聊卡（从 ai-studio 卡 morph, build 失败回滚保护）
### 侦察
- Part A 先照亮盲点（chat handler 实现/token key/卡片结构）再动手; token key 注入 + 常见键兜底链
### 已知限制
- /api/ai/chat 用默认模型（不支持 per-session 换模型）; 历史不落库（快聊定位）; token 存储键以冒烟实测为准

## [2026-11-06] OPT-015b 快聊可发现性（工具注册表 + 导航）

### 背景
- OPT-015 页面本体 E2E 通过（用户实测真 GLM 流式回复, 控制台干净）
- 但入口缺失: tools.html 搜索 0 结果 / Quick Switch 无项 / hero 搜索无果
### 变更
- 工具注册表 morph 插入 AI 快聊行（build 硬门+回滚）
- core.js 导航 morph 插入 AI 快聊项 + 全站 core.js vhash bump（缓存穿透）
- /api/features 实测裁决 features.html refs=0 之谜（SPA 静态 grep 假阴性）
### 已知限制
- Quick Switch 若有独立索引, 视侦察结果下轮补

## [2026-11-06] OPT-015c 快聊入口修复（tools 类别正名 + ID 去撞 + QS 源定位）

### 诊断（nt24 侦察裁决）
- /api/tools refs=0 根因: nt23 morph 保留模板行 category=workflow, 而 tools 端点不收录 workflow 类
- 隐患: morph 行沿用 id wf-ai-workflow, 与真工作台行撞车
- features.html refs=0 之谜裁决: /api/features 实测含 quickchat=SPA 静态 grep 假阴性
### 变更
- feature.go: quickchat 行正名 id=ai-quickchat, category 归入 AI 助手类, weight 95
- Quick Switch 源文件 morph 插入 AI 快聊项（守卫式, 源由 A1 侦察定位）+ vhash bump
### 已知限制
- 若 tools 端点为白名单类别且无 ai 类, 需按 A2/A3 输出下轮调整

## [2026-11-06] OPT-015d 快聊入口合围（ToolRegistry + app.js routes）

### 侦察裁决（nt24/25 遗案全解）
- /api/tools = ToolRegistry 独立注册表（与 feature pool 并行); 条目含 on 开关, 非管理员仅见 on=true
- 顶部导航 + Quick Switch 同源 = app.js routes 数组; 「AI 对话」href=chat.html（nt25 锚点落空根因）
- 路由注册为 Go1.22 方法前缀「GET /api/tools」（此前引号锚定 grep 漏检）
- tools 筛选 chips（AI 助手/文档处理/数据分析/开发工具）对应 ToolRegistry.cat 值域
### 变更
- ToolRegistry 注册 ai-quickchat（克隆现有条目样式, cat/icon/kind 取自 AI 族条目实测值）
- app.js routes 插入 quickchat 项（顶部导航位 + QS 自动命中）+ tools.html/app.js href 映射补行（若形态可克隆）
### 验证
- 匿名 /api/tools 可见 = 默认开启实锤; admin cookie 探针区分注册/开启两态

## [2026-11-06] OPT-015e 快聊进入 tools.html 40 格清单

### 架构裁决（nt26 收尾）
- ToolRegistry(/api/tools) = 程序化文本工具注册表（23 个, cats data/dev/gen, kind:text, 条目如 data.csv2json）——快聊是页面工具, 语义不符, 有意不注册
- tools.html 40 格网格 + 4 chips（AI 助手/文档处理/数据分析/开发工具）来自独立静态清单（40 != 23, chips 文案两表皆无）
- 已实锤生效: app.js routes 注入 -> 顶部导航第 3 位 + Quick Switch 默认列表（用户截图验证）
### 变更
- 40 格清单克隆式插入 ai-quickchat（AI 需求分析卡为模板, 规划中->可用）+ cache-bust + go:embed 重建

## [2026-11-06] OPT-015f 快聊进入 tools.html 精选工具网格

### nt27 复盘（双事故）
- 清单真身: app.js toolsPage() 内 toolsData（5 张 AI 规划卡, href 空 -> 全部"规划中"）+ word2pdf 族 + mini 工具族, 三段拼 40 格
- 事故1: grep|head -1 误选 requirements.html（同名展示页）-> 文件定位必须用最强锚（chip 文案与 word2pdf 同文件者才是真身）
- 事故2: 裸 raise SystemExit 退出码 0, FAIL 门被绕过 -> Python 失败路径必须显式非零退出
### 变更
- toolsData 头部插入 ["AI 快聊","不建流程直问 · SSE 流式回复","zap","violet","可用","quickchat","ai"]（首个 map 的 href 规则自动成链 -> AI 助手类唯一"可用+可点击"卡）
- 全部工具徽标与 toolsCount 40 -> 41; app.js cache-bust + go:embed 重建

## [2026-11-06] OPT-016a 材料库：六路侦察 + 存储层脚手架

### 侦察（输出即档案）
- 材料注入链 / 8000 截断点 / Flow struct 落盘格式 / start7 请求体 / api.go 路由注册块 / toolCard 徽标逻辑 / /api/llm 预扫——A1~A7 全量留痕, OPT-016b 按此定锚
### 变更
- 新增 internal/ai/materials.go（Material CRUD, 一材料一 JSON 落 s.dir/materials/, 单条 2MB 上限, 全量不截断）
- 零接触既有文件; build/vet/utf-8 三门, 失败即删回滚; 本轮 inert（无路由引用）
### staging
- OPT-016b: 路由接线（materials CRUD 四端点）+ start7/advance 注入改造（材料库引用替换内联截断）+ ai-studio 前端（保存到库/从库引入）+ toolCard 徽标清理（若 A6 证实多余规划中）+ E2E

## [2026-11-06] OPT-016b 材料库后端接通（CRUD 路由 + start7 注入）

### 变更
- materials.go: +net/http, +matList/matCreate/matGet/matDelete 四 handler（writeJSON/fail 助手, 与 preset.go 风格一致）
- api.go Mount: flow/export 行后插 4 条路由（RequireRole user, 克隆现有样式）
- preset.go flowStartStd: req +MaterialIDs []; 内联 material 行为原样保留; 其后新增库材料注入（GetMaterial 回显, 单条 12000 rune, 累计 30000 上限, 缺失 id 静默跳过）
- 注入语义 = 创建时快照: 删材料不影响已建流程
### 验证
- mock 全链 E2E: create/list/get → start7(material_ids) → advance step0 回显含 [mock]+NBX016 → chat 回归 → delete/get 404
- 锚点全部 strip-matched（tab 无关）+ 计数断言 + build/vet/test 三门 + 全量备份回滚
### 已知限制
- 材料为共享池（无 per-user 归属）; flowChat 未注入材料（待 OPT-016d, 需 Flow 持久化 MaterialIDs）
- tools.html 快聊卡「可用+规划中」双徽标来自 app.js 后处理（R1 已定位行号, nt31 处理）

## [2026-11-06] OPT-016c 材料库前端（独立页 + 导航 + ai-studio 接线）

### 变更
- web/materials.html：自包含材料库页（建/列/看全文/复制ID/删除/带此材料开流程），nb_token，风格沿用 quickchat
- feature.go: ai-materials 卡（克隆 ai-quickchat 行）; app.js routes: 材料库导航项（克隆 quickchat 行, icon fileText）+ cache-bust
- ai-studio.js（守卫式）: start7 body 注入 material_ids（来源 ?material_id=，用后 replaceState 清参）——锚点 JSON.stringify+idea 唯一方接线，否则跳过留 nt32
### 验收
- materials.html CRUD 手测; 顶部导航出现「材料库」; ai-studio.html?material_id=xxx 建流程后第一步产出应引用材料内容
### 遗留
- app.js L611 上下文（A2）裁决双徽标成因 -> nt32 处理; flowChat 材料注入（OPT-016d, 可选）

## [2026-11-06] OPT-016 收官：start7 接线 + badges() 拦截修复 + 材料库卡入列

### nt31 复盘（侦察裁决）
- A1: ai-studio.js L147 start7 为 api() 对象字面量（非 JSON.stringify）——nt31 守卫按设计跳过; 本轮以精确锚 material: (window.nbswMaterial || '') 接线
- A2: badges() 后处理系统全貌——MAP 已知工具追加状态徽标+href; MAP 外一律追加"规划中"+ preventDefault 点击拦截; 双徽标本是设计常态, 真隐患是快聊卡被拦截不可点
### 变更
- ai-studio.js: start7 body +material_ids（?material_id= 一次性消费, replaceState 清参）
- app.js badges(): else 分支加"数据徽标已标可用则信任之"短路——快聊/材料库卡保住 href 与可点性, 规划中卡行为不变
- app.js toolsData: 材料库卡入列（AI 助手类第二张可用卡）+ 计数 41→42
### 验证
- served 资产探针: badge-fix 标记 / toolsData 条目 / material_ids 各就位; node --check（若有）; go:embed 重建
### 浏览器验收项
- tools.html 快聊卡点击直达（不再弹规划中）; 材料库→带此材料开流程→标准七步→第一步产出引用材料
### 遗留（可选）
- OPT-016d: flowChat 追问注入材料（需 Flow 持久化 MaterialIDs）

## [2026-11-06] OPT-016c+ material_id 反填：idea 预填材料标题

### 背景（用户验收反馈）
- 带此材料开流程后仍需手输想法 -> 体验断点
### 变更
- ai-studio.js 追加 prefill IIFE（append-only 零接触）: 载入时读 ?material_id= -> GET /api/ai/materials/get -> idea 框反填材料标题（仅空框才填 + 派发 input 事件兼容监听型页面）, DOMContentLoaded 兜底
- 语义分工: 想法=命题（标题反填, 可改）; 材料=资料（内容经 material_ids 注入第一步, nt30 mock E2E 已验）
### 验证
- served 资产探针 + node --check（若有）; 浏览器: 材料库->带此材料开流程->idea 框已填标题

## [2026-11-06] OPT-017b /api/llm 下线第一梯队

### 裁决（nt34/35 侦察）
- internal/ai 直连上游(prof.BaseURL/chat/completions), 与 /api/llm 零耦合——下线不伤工作台
- 零前端调用端点: test / profiles/delete -> 本轮 410
- 活调用方: workflows.html(chat x2, 匿名) + llm.html(孤儿页, 零导航引用)
- 档案池: llmproxy 走 novaDB profiles 表(迁库后原 JSON 改名 .imported-bak); admin 同源性待 C 轮定案
- 安全备注: /api/llm/chat 匿名可调 = 烧默认档案 token 的口子, auth 包裹方案待 C4(s.user 实现)定案
### 变更
- llmproxy.go: test + profiles/delete 两 handler 替换为 410 门(http.StatusGone)
- pipeline.go: 移除过时 /api/llm/chat 注释(NT-3 两阶段管道已被 start7 体系取代, 只删旧不捏造新)
### 验证
- POST 两端点 -> 410; GET profiles 仍 200; chat 空报文 -> messages 必填(路由存活, 零上游消耗)

## [2026-11-06] OPT-017c /api/llm 收官：登录门 + profiles 410 + llm.html 移除 + 测试解耦

### C 轮定案（nt36 输出）
- 双池实锤: llmproxy 走 novaDB profiles 表(dbProfiles*); internal/ai 走 admin JSON store(OpenStore+mutex)——互不相通, /api/llm 岛屿下线不伤工作台
- admin 自有完整管理面: /api/admin/llm/profiles CRUD(RequireRole admin)
- s.user(main.go): cookie nova_tok 与 Bearer 双通道; A1 实测 login 无 Set-Cookie, Bearer 为承重通道
- llm.html 166 行孤儿页(feature.go/app.js 零导航引用), 是 /api/llm/profiles 唯一前端消费者; DB 池仅一条档案(智谱 GLM)
### nt37 停摆复盘（测试门拦截, 生产零影响）
- 报错 e2e_test.go:128 为助手 Fatalf 行号假象; 真凶 agent_e2e_test.go:27 与 agent_eval_e2e_test.go:49 以 POST /api/llm/profiles 播种测试档案
- 教训入库: nt34 A9 只 grep 了单文件 e2e_test.go 得假阴性; 覆盖普查必须 grep 裸词全量(nt38 T2 即如此, 一发命中)
- 测试从不调 /api/llm/chat, 唯一依赖 = 播种动作
### 变更（nt37 + nt40 合并落地）
- llmproxy.go: chat 匿名 401(JSON, 门在候选解析前, 零上游消耗); GET/POST profiles 410(池冻结, chat 回退读池不受影响)
- agent_e2e/agent_eval_e2e: 播种改同包直调 dbProfilesUpsert(U3 纯委托实锤等价; U4 测试夹具 novaDB 就绪; 返回值原本未消费)——测试永久脱钩已下线端点
- web/llm.html 删除; usage/chat/workflows 三页导航「模型设置」链接同步清除; workflows.html 两处 chat fetch 加 Bearer(nb_token)
- secmw 限流分支与 metrics llm 统计保留(chat 存活, 继续有效); pipeline.go 过时注释已于 nt36 更正
### /api/llm 终态
- 仅剩 POST /api/llm/chat(需登录); test/delete/profiles(GET,POST) 全 410; 管理面归零(admin 控制台接管)
- 可选后续: workflows.html AI 规划迁移 /api/ai/chat(SSE + feature_key), 归入工作台门禁/配额体系

## [2026-11-06] OPT-018 workflows 迁移 /api/ai/chat（SSE）

### 依据（nt41 侦察）
- /api/ai/chat: {feature_key?, messages} -> gateCheck(登录+功能门+402 已购放行) -> 默认档案 -> relaySSE 透传 OpenAI 兼容 data: 行
- feature_key 空则 defaultFeatureKey(ai_chat), 种子无此键回落第一个启用功能——前端省略 feature_key 最稳
- 原调用吃同步 JSON content 字段; maxTokens /api/ai/chat 不支持——意图门靠 prompt 约束, indexOf 判定兼容
### 变更
- workflows.html: 新增 aiChatSSE 助手(Bearer + 流读取 + delta.content 累积 + [DONE] 跳过; 非 200 解析 {message} 报错, 兼容 401/402 门)
- aiIntentGate 两次调用改走助手; plan 解析/AIT 过滤/渲染/persist 逐字保留; 页面 /api/llm/chat 引用归零
### 意义
- /api/llm 最后的前端消费者迁走——nt43 可对 chat 落 410, 全家族退役
### 验证
- served 探针: aiChatSSE=3, llm/chat=0, /api/ai/chat=1; 匿名 401; 带 token SSE data: 行正常

## [2026-11-06] OPT-017 终态：/api/llm 全家族退役（五路由 410, 手术式）

### nt43/nt44 失败复盘（构建门两次拦截, 回滚均完整, 生产零影响）
- nt43: 普查已打印共享助手的外部消费者, 但整文件重写未消费普查结果, 误删 llmTrunc/llmCallU/llmWriteJSON——教训: 普查必须作为突变守门人
- nt44: 结构体迁移写成复制而非移动, LLMProfile 同包重声明——教训: move = copy + remove, 缺一半即废
### 变更（手术式, 全部助手保留）
- llmproxy.go: 五 handler 主体替换为 410 门(registerLLMRoutes 签名不变); st/client 两行移除; LLMProfile 结构体真迁移(dbprofiles.go 获得声明, llmproxy.go 移除, 单声明断言入库)
- 共享助手全数保留: agent/eval/dbconversations 仍消费 llmTrunc/llmCallU/llmWriteJSON/llmCall 等
- secmw.go: /api/llm/ apiRL 专项限流分支移除; metrics llm 统计保留(死分支无害)
- 后续可选: 孤儿助手(llmMask/llmKeyHint/llmNormalizeBase/llmReadJSON/llmLastUser)动态普查驱动清创
### /api/llm 终态
- GET/POST profiles、profiles/delete、test、chat 五路由全 410(含带凭证请求)
- 唯一在册 AI 通道: /api/ai/chat(登录 + 功能门 + SSE); workflows 已迁(OPT-018)
### 验证
- 五端点 410; /api/ai/chat 匿名 401 / Bearer SSE data 行正常; 五页面 200; 三门全绿

## [2026-11-06] OPT-016d flowChat 注入持久化材料（nt47→nt49→nt50 三轮闭环）

### nt47 停摆（守门断言拦截, 零写入, 树无污染）
- 锚点 `for _, m := range f.Chats {` 全仓两处(flowChat 历史循环 + flowExport 导出循环); 教训: 锚点唯一性是全仓属性, 侦察必须全局计数
### nt49 停摆（计数门误标定; 突变已完成, 停在构建前）
- 突变全部成功([OK]x2, gofmt 通过=语法有效), 计数门 expect=3 实际=4: 助手文档注释行同样命中裸词 grep
- 准确账本: preset=3(注释+定义+调用) + flow=1(调用) = 4; "def+2 calls" 口算漏了注释
- 教训入库: 计数门必须用精确模式分项检查(定义/调用/持久化各自 grep), 裸词计数只做记录; 停摆点在突变后时, 恢复走"续跑验证"而非重放突变
### 变更
- Flow struct + MaterialIDs []string; flowStartStd 持久化 req.MaterialIDs
- 注入循环抽为 Service.appendMaterialCtx(预算语义逐字保留: 单条 12000 rune / 总 30000 rune, 失效 ID 宽容跳过), start7 与 flowChat 双处复用
- flowChat 在 flowCtx 之后、历史之前追加材料 system 消息(len==0 零改动); 前端零改动(flow id 已带)
### 已知边界
- 多轮追问 = 材料全文 + Chats 历史累积注入, token 随轮次增长(预算上限兜底; 对话压缩为独立后续 OPT)
### 验证
- 三门绿; start7 带 material_ids 零上游; flowList 确认持久化; flowChat 材料专属问题一次小调用注入生效; 无材料流程追问零回归

## [2026-11-06] OPT-017 后传①：llmproxy 孤儿清创落账（metrics 拆至下轮）

### nt51 复盘（gofmt 门拦截, 回滚完整, 零生产影响）
- llmproxy 侧成功: 动态普查正确识别 5 真孤儿(llmCall/llmLastUser/llmMask/llmKeyHint/llmReadJSON), 保留 4 活符号(llmWriteJSON/llmTrunc/llmCallU 外部消费者; llmNormalizeBase 内部消费者 llmCallU)
- metrics 侧失败: 逐行删除不感知括号配对——`if len(s.llmMs) > 500 {` 被删而其闭括号行保留成孤儿括号; llmAvg/llmP95 声明行不含 PAT 成悬空变量; gofmt 解析失败拦截
- 教训入库: 块删除必须括号感知; 标识符模式要覆盖派生变量名; 结构未过目的文件不做逐行手术
### 本轮变更
- llmproxy.go: 5 孤儿函数删除(fixpoint 普查驱动, 双零才删); registerLLMRoutes 门文件与 4 个共享助手原样
### 拆分决策
- metrics llm 移除独立成轮(nt53): 先 sed 全形侦察自增区/统计区两区块, 再块感知切除

## [2026-11-06] OPT-017 后传②：metrics llm 死分支双侧移除

### nt51→nt52 关键更正（M6 侦察发现）
- nt51 P3 「web 零消费」为误判: admin-metrics.html 有 4 张 LLM 卡片消费 d.llm.calls/errors/avg_ms/p95_ms
- 误判根因: P3 模式清单是 Go 标识符+扁平 snake 键(llm_ms/llm_p95), 实际 JSON 为嵌套 "llm":{avg_ms,p95_ms}, 前端访问形态 d.llm.* 天然漏网
- 教训入库: web 消费普查必须覆盖对象访问形态(d.KEY/.KEY), 不能只查扁平键名
### 移除依据（语义腐坏, 非仅死码）
- 自增分支 if path=="/api/llm/chat" 家族退役后只统计 410 门命中(把门击记成「LLM 错误」), 真实 AI 流量走 /api/ai/chat 不经此分支——保留=永久假数据
- 前端 (d.llm&&d.llm.calls)||0 优雅降级: 只删后端页面不炸但四卡冻结在误导性 0——双侧一致移除才是诚实收官
### 变更（括号感知, 锚点全部 nt52 M 组侦察验证唯一）
- metrics.go 四处: mStore 三字段(llmCalls/llmErrs/llmMs) / record() 内 if path=="/api/llm/chat" 整块 / snapshot() 内 llmAvg,llmP95 声明+if n:=len(s.llmMs) 整块 / return map 的 "llm" 行
- web/admin-metrics.html: 4 张 LLM 卡片行删除
- 核心指标(req_total/top/recent/rpm)原样; sort.Slice 仍被 tops 消费
### 可选后续
- 统计改造为跟踪 /api/ai/chat(SSE 路径, 需区分流式)——feature 非 清创, 独立 OPT

## [2026-11-06] OPT-019 flowChat 历史预算压缩（确定性截断, nt55→nt56 续跑闭环）

### nt55 停摆（单测校准门, 突变后拦截, 零生产影响）
- TestCompressChatsHugeSingle 断言 n<=budget 失败: 实际 105=100+5, truncRunes 截断时追加可见标记
- 判定: 测试校准失误非逻辑缺陷——compressChats 委托 truncRunes 与 appendMaterialCtx 材料截断同语义, 标记让模型知晓内容被截; 保留 helper, 测试改 exact-form(got==truncRunes(原文,budget))+前缀保全双断言
- 教训入库: 断言既有助手输出形状时用 exact-form 对照, 不猜精确数值
### 变更
- compressChats: 从最新往旧按 rune 预算保留; 单条超限走 truncRunes(标记可见); 窗口修剪非 user 开头; 截断前置 system「早期对话已省略」
- histBudgetRunes = 8000(截断标记个位数超限可忽略); 仅压缩送模视图, 落盘历史全量保留(flowExport 完整导出); 前端零改动
- compress_test.go: 3 用例(预算内全保/超限部分保+user起始+最新存活/单条巨消息 exact-form)
### 验证
- 单测 3 绿; 三门绿; 双轮探针 chat#2 复述 chat#1 指定词 = 历史经压缩路径注入实证; f.Chats 全仓仅剩 flowExport 一处消费(回归门)
