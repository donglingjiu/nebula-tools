# nebula-tools 演进 TODO

> 来源：web-delivery 阶段2 需求简报 v1.0。逐版本消化，完成即勾选并附 PR 链接。

## R1 · AI 引擎 + 账户底座（进行中）
- [ ] AI 全流程工作台 /ai-studio（7 阶段）
- [ ] 真实注册/登录/三级权限（普通/管理员/超管）
- [ ] LLM 全局 Profile 管理 + 默认 key 兜底
- [ ] AI 付费开关（超管一键）+ token 计量
- [ ] 功能池最小开关机制（AI 模块上下线）
## R2 · 功能池运营 + 监控报警
- [ ] T-05 报警渠道：邮件 / Webhook / 钉钉 / 飞书
- [ ] T-06 用户行为分析 + 功能池全量运营看板（QPS/延迟/错误率/工具调用成败/token/活跃）
- [ ] 功能灰度：按用户百分比放量
## R3 · 工具轻量扩展
- [ ] 文档：在线查看 + 格式转换 + 合并拆分
- [ ] PPT：AI 生成大纲 + 模板出片
- [ ] T-10 表格/数据轻处理与可视化
## Backlog（远期）
- [ ] T-01 AI 全流程：更多技术栈模板选择
- [ ] T-02 AI 产出可运行项目（zip/在线预览，重）
- [ ] T-03 第三方登录 GitHub/Google/微信
- [ ] T-04 用户 BYOK 个人 key 配置界面
- [ ] T-07 文档富文本协同编辑（重）
- [ ] T-08 PPT 完整画布编辑器（极重）
- [ ] T-09 脑图 XMind 级编辑（重）
- [ ] T-11 多实例水平扩展：会话外置 Redis、data/ 迁对象存储
- [ ] T-12 K8s 部署清单 + HPA
- [ ] T-13 异步任务队列（长 AI 任务）
- [ ] T-14 分库分表与读写分离评估

## 过程规约（持续追加）
- 交付命令块内严禁字面三反引号；代码需要三反引号时用 String.fromCharCode(96,96,96) 构造（教训：批次1 core.js 正则含三反引号，截断外层代码围栏导致粘贴损坏）
- 每批文件写入后必须执行自检：行数统计 + heredoc 标记泄漏检查 + JS 语法检查（node --check，无 node 则跳过）+ go build，全部通过才算交付完成

## 过程规约（续）
- 交付块内禁止 4 字节 emoji（会被终端粘贴替换为乱码）；一律用 ASCII/中文替代
- 大文件粘贴易随机丢字：写完必跑哨兵 grep 核对关键行；修复用 sed 定点改行，不整块重贴
- 本机构建统一 CGO_ENABLED=0 go build（服务器 gcc 过旧与 cgo 不兼容；项目纯 Go 无影响）

## 本地开发回路规约
- 本地验证一律走 ./dev.sh（web/app/build/smoke/stop/check）；构建唯一方式 CGO_ENABLED=0，不把物理机 gcc 等当依赖
- 前端迭代 = ./dev.sh web + 浏览器刷新（0 编译）；Go 迭代 = ./dev.sh app（go run 秒级）
- web 模式端口被占自动 +1；stop 只杀 python 静态服务，不碰正在运行的 Go 服务
- dev 与发布解耦：CI 产物/docker/K8s 仅阶段7部署门使用；目标机只做"拉取+启动"（后续一键脚本）
- JS 语法门禁：浏览器 F12 控制台零红错；可选 nvm 装 node（$HOME 内）后升级为 node --check
- 括号原始计数为信息项非门禁（正则字符类会污染计数）

## 过程规约（续2）
- 单个 heredoc 尽量控制在 120 行内；同一文件若二次损坏，改用 base64 分块传输（解码失败会报错，不会静默损坏）

## 过程规约（续3）
- check 增加 HTML 通用结构哨兵：web/*.html 必含 </html>；含 <header 必含 </header>（教训：admin 两页 </header>/属性引号被截断，存在性哨兵抓不到）

## 过程规约（续4）
- 损伤判定以文件内容为准（grep/哨兵/check），终端回显可能自行丢字，不可作为判据（教训：本轮回显异常的文件实际完好，真正损伤藏在重复粘贴覆盖的新文件里）
- 同一写入/修复块只执行一次；重跑前先确认上一次是否已成功，避免覆盖

## 过程规约（续5）
- 导航侧栏不允许死链：页面未实现时先上占位页（含规划说明），accept B 才能真实闭合
- 认证类接口（login/register）401 走 silent + 页面级提示，全局"登录已过期"仅用于会话失效场景

## 阶段5进度
- 5-1 完成(修复后)：internal/auth 四件套 + httptest 全流程单测
- 关键发现：项目为 nova v2.2.1 成熟库，已含 SQLite/db.go/username 用户体系/secmw 中间件/ai-chat/llm-profiles；e2e_test 锁定旧 auth 行为
- 5-2 方向：共存接入（不破坏旧站与 e2e），待 R1-R5 侦察输出定具体方案

## 过程规约（续6）
- 粘贴内容(含Go)一律禁用TAB缩进：交互heredoc中TAB触发readline补全→目录列表炸屏+行截断；落盘后 go fmt 统一
- 长python块易断行：替换串尽量短，patch后必须grep回验
- 集成既有系统前先侦察(路由/中间件/存储/e2e约束)，禁止盲patch main.go

## 过程规约（续7）
- 同文件多个rep的guard禁用"前一个rep会注入的关键词"(自踩: guard='querySelector'被rep1注入导致rep2误跳过)
- 账号/密码等口径变更必须在当轮公告并保持唯一(教训: admin12345→admin123456两轮口径不一致引发误判)
- 旧站SW(scope /)会拦截新页面: 本站页面统一注销; 诊断要看sw.js的fetch策略

## 过程规约（续8）
- 交互终端一切curl必须带 -m 超时与 --noproxy '*': 无超时curl在对端不响应时永久挂住终端
- 超长命令块分段粘贴(每段≤20行), 上段输出确认后再续; 交互块禁用裸 exit(编译失败会关掉终端, 教训: store.go未用导入)
- 卡住三板斧: Ctrl+C → ps+tail日志 → health(-m 3); 用输出定位, 不猜
- 发新Go代码前先本地go vet/test过一遍再交付(自检缺口: 交付了带未用导入的代码)

## 阶段5进度(续2)
- 5-3b 修复：长粘贴断链致 api 丢 .go 后缀/test 丢失/main 未接线 → 全部重写补齐，矩阵全绿
- 5-3 完成：admin 8路由接真；index 横幅修复
- 5-4 待做：AI 接线（SSE chat/flow + 410 下线门/402 付费门）

## 过程规约（续9）
- 行级patch必须先grep原行并保真其结构后缀(教训: FIXS重写哨兵行丢掉for列表收口"]:")；改dev.sh后立即 bash -n + check 双验证通过才可commit
- grep -c统计行数非次数: 单行JSON计数用 grep -o ... | wc -l
- 本机gcc过老: Go命令统一 CGO_ENABLED=0(裸build的cgo报错属环境噪声)

## 过程规约（续10）
- web/ 经 go:embed 编入二进制: 改前端必须 ./dev.sh build + 重启才生效(教训: 截图旧文案误判为patch失败)
- git push 偶发 SSL_ERROR_SYSCALL 为瞬时抖动, 原命令重试即可

## 过程规约（续10）
- web/ 经 go:embed 编入二进制: 改前端必须 ./dev.sh build + 重启才生效(教训: 截图旧文案误判为patch失败)
- git push 偶发 SSL_ERROR_SYSCALL 为瞬时抖动, 原命令重试即可

## 阶段5进度(续3)
- 5-3f: 密钥列掩码垫片(过渡方案; 原生渲染代码存档/tmp/akr.txt待重构并入); 后端is_default契约补齐("设默认"按钮接真)
- 5-4 待做: AI 接线(SSE chat/flow + 410下线门/402付费门)

## 过程规约（续11）
- 含前端修改的批次, 收尾段必须包含 build+restart+线上curl验证(再犯: 段4漏build致shim未上线)
- MutationObserver内改DOM会自触发死循环: 垫片类代码禁用观察者, 改用重试+online事件
- ERR_INTERNET_DISCONNECTED=客户端断网请求未达服务端, 与服务端无关; 服务端状态以curl为准

## 过程规约（续12）
- 验证grep必须先在源文件跑对照组(文件=1而线上=0才是真未上线; 教训: "addEventListener('online')"带右括号匹配不了"addEventListener('online',"), 白警报一轮)
- Chrome错误类型即定位: ERR_INTERNET_DISCONNECTED=浏览器离线(或DevTools"离线"复选框勾着), ERR_CONNECTION_REFUSED才是服务端没起; 判服务死活用 curl -m 5 health, 不看浏览器

## 过程规约（续12）
- 验证grep必须先在源文件跑对照组(文件=1而线上=0才是真未上线; 教训: "addEventListener('online')"带右括号匹配不了"addEventListener('online',"), 白警报一轮)
- Chrome错误类型即定位: ERR_INTERNET_DISCONNECTED=浏览器离线(或DevTools"离线"复选框勾着), ERR_CONNECTION_REFUSED才是服务端没起; 判服务死活用 curl -m 5 health, 不看浏览器
