# Changelog

## v2.9.7 - 2026-10-23
### Changed
- B3 收官微刀：--nv-s9 并入 --nv-s10（9px 已于 v2.9.5 归一为 10px，与 s10 同值成孪生）——删定义、nv-mt9 消费点改引 var(--nv-s10)、verify 尺度锚点随迁 s10；全站 var(--nv-s9) 外部引用为零（v2.9.5 侦察实证），零视觉变更
### Closed
- B3 战役全队列清零：拆甲（14 拆 12 豁免，棘轮 27 终档 12）+ 9px 归一 + 孪生合并 + #2d3d60 归一（--nv-ink）

## v2.9.6 - 2026-10-23
### Changed
- B3 队列终项：#2d3d60 归一——新增令牌 --nv-ink:#2d3d60，规则值改引 var(--nv-ink)（计算值零变化，纯归一）；与正文墨 --nv-text(#1b2437) 分工并存（ink=表格强调墨）
- verify.sh 第 27 项「墨色令牌在位」（grep -o 数次，防行数次陷阱）
### Closed
- B3 队列清零：9px 归一（v2.9.5 采纳 10px，实际影响面=tokens 内 1 处消费点）与 #2d3d60 双结案

## v2.9.5 - 2026-10-23
### Changed
- B3 队列3 裁决：9px 归一 A/B 采纳 10px——--nv-s9 一行改值，verify「尺度tokens在位」锚点同步；A/B 截图对比存档

## v2.9.4 - 2026-10-23
### Changed
- B3 批3改判：组件甲豁免入册——nv-opt-input(7)、nv-searchbar-full(2)、nv-hero-strip(2)、nv-c-ink(1) 十二处保留；威胁裁决实测 66 个潜在高优先级对手（含 .searchbar input、.x strong 家族等结构性实锤，另有扫描器逗号组/标签限定符两类假阳已自首记档），组件甲是 B2 忠实翻译内联优先级的载体而非技术债——拆之必改渲染
- B2 注释措辞归一（字面提及改为「强优先」），棘轮终档 12：体检「拆甲基线立账」转职为防回潮卫兵（tokens 甲数只许等于 12，新增即 FAIL）
### Closed
- B3 拆甲审计终审结案：工具甲 14 处全拆（批1八刀+批2六刀，全站截图零差异实证），组件甲 12 处豁免立碑（威胁扫描呈堂）；raw 27→19→13→12 终档

## v2.9.3 - 2026-10-23
### Changed
- B3 拆甲批2：nv-flex1、nv-py4、nv-ico32、nv-ico31、nv-fs15、nv-fs13 六类脱甲（6 处移除，棘轮 19→13）；前置证据：六类元素级内联零碰撞 + 六个组件样式表零定义冲突 + 静态装载 tokens 恒居末位或唯一

## v2.9.2 - 2026-10-23
### Changed
- B3 拆甲批1：nv-mt12/mt10/mt9/mt8、nv-m0、nv-mlauto、nv-grid3、nv-grid2 八类脱甲（8 处移除，棘轮 27→19）；前置证据：tokens 内零竞争（注释剥离复核）+ 内联零碰撞（元素级复核）
### Process
- 普查修正：27 raw = 26 实甲 + 1 注释提及（棘轮按 raw 计，偏移恒定）；census 14→18 类（nv-opt-input 等复合选择器首轮漏检，宽网补全）；批2/批3 排程入册
- 军规 #6 附件：双引号内的命令替换使内层单引号失效（! 照样引爆，实测两次）；安全形态为顶层赋值再测试
- 军规 #10：多步写盘必须单块分阶段断言（依赖块前置断言），禁止无闸串联

## v2.9.1 - 2026-10-23
### Added
- B3 开工：拆甲审计第一阶段——全景侦察 + 基线立账 27 处（只降不升）；批1终审：120字符窗误报经元素级复核洗清 nv-mt12/nv-m0，批1满编 8 处零碰撞零竞争；注释剥离证实全部甲均居于 .nv-* 家族名下
### Fixed
- 面板打磨三连：兜底项显示原样输入；组合期 input 静默（候选条开着不重渲染）；compositionend 补刀（部分 Linux IME 结束组合不派发 input——实测截图发现）
### Added
- verify.sh 第 24/25/26 项
### Process
- 军规 #7 扩展（禁用百分号格式化构造任何落盘文本，一律拼接）；军规 #8（verify.sh 写盘当场 bash -n 自验）；军规 #9（环境基线 Python 3.6：capture_output 等 3.7+ API 禁用，subprocess 一律 stdout/stderr=PIPE）

## v2.9.0 - 2026-10-23
### Fixed
- IME 组合期守卫（双点位）：命令面板与 AI 输入框的 Enter/↑↓ 处理器补 isComposing||keyCode===229——中文输入法选词的 Enter 不再误触导航/发送（截图实测发现）；守卫置顶，组合期全部按键交还输入法
### Added
- verify.sh 第 23 项（IME卫兵在位 ×2）

## v2.8.9 - 2026-10-23
### Added
- 命令面板兜底搜索项：输入非空时末尾追加「在工具大全中搜索『q』」→ tools.html?q=词（落地即过滤）；空结果不再死胡同；verify.sh 第 22 项（面板兜底搜索）

## v2.8.8 - 2026-10-23
### Added
- B3 开工：spacing/typography 尺度 tokens 化——--nv-s4/8/9/10/12/15/19 与 --nv-fs13/15 入 :root，.nv-py4/.nv-mt*/.nv-fs* 改引变量（值不变，零视觉变化）；缓存戳 v286→v288；verify.sh 第 21 项（尺度tokens在位）
### Corrected
- v2.8.7 账目更正：「全部工具 40」经 DOM 实数证实准确（前判 28 系侦察脚本对 toolsData 分段截断失真）；计数已运行时化
### Decision
- icon 容器家族（11 种尺寸各 ×1）：按 C 方案保留单体不归一；#2d3d60 token 化 A/B 挂账
### Process
- 两条新军规：build 闸门必须双检（产物在盘 + changelog 在册）；REMOTE-CONFIRMED 仅在真实新提交推送后打印

## v2.8.7 - 2026-10-23
### Fixed
- ?q= 深链修复：synthetic input 事件默认 bubbles:false + 派发早于监听挂载 → 预填生效但过滤从不触发；现 bubbles:true + setTimeout 延迟派发
- 「全部工具」计数纠偏：静态硬编码 40 与实际卡数漂移（源码三组数组初判 28），chip 与 toolsCount 改为运行时 DOM 实数
### Added
- verify.sh 第 20 项：深链派发冒泡静态守卫

## v2.8.6 - 2026-10-23
### Refactor
- 内联收编第四批（收官批）：重复≥2 模式全清（font-size:15px×2、font-size:13px×2、color:#2d3d60×2 → .nv-fs15/.nv-fs13/.nv-c-ink）；棘轮 117→111
- B2 收编战役收官：app.js 内联 164→111（累计 -53）；缓存戳 v284→v286；剩余单体（多为 28~62px 方形图标容器家族）移交 B3 逐个判决（组件类+尺寸变量方向）

## v2.8.5 - 2026-10-23
### Refactor
- 内联收编第三批（组件级）：试炼场参数输入框 → .nv-opt-input、搜索条全宽 → .nv-searchbar-full、英雄渐变条 → .nv-hero-strip；棘轮 120→117（战役累计 164→117）

## v2.8.4 - 2026-10-23
### Refactor
- 内联收编第二批（标签感知合并）：18 处 → 8 个工具类（py4/mt12/mt10/mt9/mt8/ico32/ico31/mlauto）；棘轮 138→120；tokens 缓存戳 v280→v284
### Fixed
- 塌缩脚本升级：class 与 style 被其他属性分隔时合并进既有 class（DOUBLE-CLASS 卫兵首次拦截成功，文件零损伤）；a030d55 为半程提交（tokens+bump+棘轮），本版补齐塌缩与 changelog

## v2.8.3 - 2026-10-23
### Refactor
- 内联 style 收编第一批：flex:1×11、margin:0×4、三列网格×9、两列网格×2 → .nv-flex1/.nv-m0/.nv-grid3/.nv-grid2（忠实翻译+过渡护甲，共 26 处）；verify.sh 增至 17 项

## v2.8.2 - 2026-10-23
### Fixed
- 徽章路由 JS 补落：v2.8.1 仅落 .tag.plan 样式，toolCard 的 可用/规划中 路由因断言笔误未写盘（卫兵拦截零损伤）；现规划中徽章转灰
- word2pdf.html 补接 tokens（v2.8.1 因 head 结构跳过），独立页 19/19，verify.sh 增至 16 项
### Corrected
- v2.8.1 changelog「19 个独立页」实为 18，本版补齐并更正

## v2.8.1 - 2026-10-23
### Added
- B2 第一批：19 个独立工具页批量接入 nova-tokens.css（滚动条/选区/焦点环全站统一），verify.sh 增至 14 项
- 徽章规范：可用=info 蓝、非可用=plan 灰（.tag.plan 入 tokens）

## v2.8.0 - 2026-10-23
### Added
- nova-tokens.css 设计系统单一真源：清爽（冷白底+细边框）/好用（focus-visible 焦点环+悬浮反馈）/高级感（三档分层阴影+统一圆角+细滚动条+选区色），app.js 注入器全站最后加载，11 个主航页面生效
### Fixed
- 命令面板 Enter 失效：纯 NR 工具命中时无 active 项（现首位自动 active + Enter 兜底 items[0]）；verify.sh 增至 13 项

## v2.7.4 - 2026-10-22
### Fixed
- 命令面板页面段缺 .join('') 导致卡片间出现字面逗号（v2.7.0 起）

## v2.7.3 - 2026-10-22
### Fixed
- 命令面板修复：v2.7.0 双重 .join('') 致 renderCommandItems TypeError，浮层永不显示；点击绑定 $ 包裹误用改 $$ 直调（磁盘覆盖秒级上线，断言卫兵拦截一次错误替换）

## v2.7.2 - 2026-10-22
### Added
- NOVA_WEB_DIR 磁盘优先覆盖：设置环境变量时静态资源优先读磁盘（改 JS 免编译免重启刷新即生效），未设置时保持单文件内嵌部署

## v2.7.1 - 2026-10-22
### Fixed
- Mac Chrome 保留 Cmd+K（页面无法拦截）→ 新增 / 键呼出命令面板（输入框内不误触发），Cmd+K 保留兼容 Safari，底部提示更新

## v2.7.0 - 2026-10-22
### Added
- Ctrl+K 命令面板接入 14 个试炼场工具：搜索工具名/id，本页直达自动选中，跨页深链 tools.html?tool=<id>（带下拉填充重试）

## v2.6.2 - 2026-10-22
### Fixed
- 前端接线真正落地（v2.6.1 因补丁断言拦截未改 app.js，实为仅 CHANGELOG 的空版本）：4 新卡入网格、暗道直通、NR_SCHEMA 补 cron count（计数 36→40）

## v2.6.1 - 2026-10-22
### Fixed
- 仅变更日志版本：前端接线补丁被断言正确拦截（锚点错误），app.js 未改动

## v2.6.0 - 2026-10-22
### Added
- P0-2 工具扩容四件套：dev.cron（5 段解析+未来运行时间）、dev.chmod（八进制↔符号双向）、dev.subnet（IPv4 CIDR）、dev.yaml（YAML↔JSON 自动识别，引入 gopkg.in/yaml.v3）
- （仅后端；网格接线与 cron count 参数栏实际落于 v2.6.2）

## v2.5.3 - 2026-10-22
### Added
- 真机试炼场按工具渲染参数栏（NR_SCHEMA）：hmac key/alg、case mode、totp code、lorem paras

## v2.5.2 - 2026-10-22
### Added
- 工具大全网格新增 HMAC/TOTP/JWT/Slug/Lorem 5 卡；字数统计、大小写转换升级后端版；卡片点击直通试炼场（计数 31→36）

## v2.5.1 - 2026-10-22
### Fixed
- e2e 工具数断言跟随 ToolRegistry，新增工具不再改测试

## v2.5.0 - 2026-10-22
### Added
- P0-1 微工具带 7 件（hmac/totp/jwt/case/slug/lorem/count），extraRunners 钩子 tools.go 仅 +3 行；新工具默认开启
