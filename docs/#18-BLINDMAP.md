# #18 盲区地图（E 站收官）

> 测绘基准 v2.14.1（main 6950387）。供给面=Go HandleFunc 在册 54 路；需求面=JS 四口径合流（api 单引号 6 型·fetch 双引号 5·fetch 单引号 5·fetch-direct 10 处）。最大发现：admin-refs 探针零命中——管理台十壳全挂线全有案，却零取数：挂线不等于通电。

## 表一盲端点（在册·无前端客户）

| 端点族 | 路 | 现状 | 处置 |
| --- | --- | --- | --- |
| /api/admin/** | 13 | 十壳挂 app.js、dispatch 全有案，app.js 零 /api/admin 引用——纯前端演示壳 | 牌位 T7：登录墙+真接线 |
| /api/auth/password | 1 | profile 改密无客户端 | 牌位：随 profile 真接线 |
| /api/analytics | 1 | analytics 壳不取数 | 牌位：随 admin 系 |
| /api/notices | 1 | 通知中心无客户端 | 牌位 |
| /api/favs、/api/fav/{tool} | 2 | 收藏无客户端 | 牌位：小而美候选 |
| /api/ai/chat | 1 | ai 壳不取数 | 牌位 |
| /api/llm/test | 1 | 连通性测试无客户端 | 牌位：随 admin 系 |
| /api/t/*（csv-clean/img-process/xlsx-profile） | 3 | 疑被 /api/run/{tool} 取代；e2e 仍测（内部客） | 复核后可退役（另案） |
| /api/upload/{tool} | 1 | 文件队列走 fx/EP 映射，未见直调 | 复核 |

待核二针：已裁决——见文末「增补：末针裁决」。

## 表二盲页

- 死页 0：口供②终审——十八张零 JS 页全为自包含静态页，无 #app 空壳；workflows/workspace 洗白（无 #app）
- 注记：llm.html 无 JS 引用（静态页，导航可达）
- 绊线：verify「app页普查棘轮」=15（web/*.html 顶层口径，eq 精确等值）——新增 #app 页必须当轮有意识地挂线或改基线

## 表三假面（data-toast 演示钮）

- 23 枚/16 行，全在 app.js：Google/GitHub 登录演示、「即将上线」转换类型、admin 壳内演示钮
- 处置：逐枚挂 T7 牌位；接真端点同轮摘面——一面一账

## 附：测绘自省（三条探针洞，案十六元层回响）

1. 口径漏单引号——me/logout 漏网，四口径合流修正
2. 普查路径漏 web/admin/——admin-shells 补针
3. 「零 JS=死页」想当然——#app×script 交叉终审，死页清零

## 增补：末针裁决（2026-10-28·第五口径收旗）

- 末针一（conversations）：DELETE 有客（chat.js:19 delConv——api() 助手的 method 参数对四口径隐形）；PATCH 无任何口径命中——盲名单 +1，入牌位
- 末针二（EP 全集）：EP=5 键（detail-Word/Excel、work-Word/Excel/PDF → fx 的 to-docx/to-xlsx/merge）
- compress/split 裁决：五口径全零——pdf-compress/pdf-split 无任何前端客，表一新增两路；app.js toast 文案「已支持合并/拆分/压缩」与 EP 路由能力偏差——承诺面大于路由面，摘面或补路由（T7 随件）
- admin 壳裁决维持：五口径含内联合计零 /api/admin 命中——13 路盲名单坐实，「挂线≠通电」定稿
- ODB 第二游离物定谳：b64595b ≠ 现行 app.js（7b480432）——124KB「Nova 静态原型」为 app.js 祖先版本，自家迭代沉积（良性），与 gin 快照无涉
- 测绘自省第 4 条：口径五——HTML 内联 script（admin-refs 初查只扫 app.js，42 张 HTML 的内联块从未入普查）——盲图自己埋的「待核」旗收旗后，普查闭合

## 增补二：fx UI 层定谳与快修（2026-10-28 · 第 60 发）

- 卡层真相：converter-options 与 data-detail-format 卡片均由 app.js 模板串动态渲染（workspacePage/detailPage）——SPA 壳的 HTML 静态 grep 天然盲；上轮「全 HTML 零命中」是壳层空、非卡层无
- 全栈对照（修前）：五真（detail-Word/Excel、work-Word/Excel/PDF）+ compress 附属暗道（合并特支 sws[2] 开关附加 opts.compress）+ 三假卡（detail-PPT、detail-Image、work-Summary——有卡无键，点击落「即将上线」toast，属有预告的体面假面）+ 两路无 UI（fx/compress、fx/split——路由在、EP 无键、卡无位，toast「已支持拆分/压缩」承诺大于 UI 路由面）
- 本轮快修（additive-only）：EP+2 键、卡+2（压缩/拆分）、ext 三态（docx/xlsx/默认 .pdf）、label 映射（fx 错误提示不再冒「PDF→Excel」名）——compress/split 转全栈接线，toast 承诺当轮兑现；style 棘轮 111 原样（新增卡零内联样式）
- 假面处置：三假卡入 T7 牌位——detail-PPT/Image 需新路由或摘卡；work-Summary 宜接 /api/ai/chat 或移出转换卡组——一面一账
- paas 定谳：llm.html 两处命中均文档文案（placeholder 示例+脚注典型 Base URL），非调用——无害
- 口径自省（续）：六=动态拼接（api('/api/llm/'+path)——chat/usage/llm 三份助手）；七=SPA 模板串（卡片与文案活在 app.js 反引号里）——静态普查三层盲区至此闭合
- 卫生挂签：web/assets 下 app.js.bak~bak6 六代备份与 nova-v4/v5/v6.css 多代样式共存——删除决策缓行另案（防误伤在役引用）

## 增补三：W0 撤案与 W1b 落刀（2026-10-28 · 第 61 发）

- W0 撤案：live 五路 401 + 静态双证（s.admin 包装、u.Role!="admin" 守卫、e2e 十三管理流程）——服务端墙一直在岗；自省第 6 条定稿：「后端本有 auth」由怀疑转实锤，且是活体探针证实的
- me 带 role：main.go me handler 系内联闭包（无具名函数——上轮 HANDLER=func 伪名根源），:173 响应体含 username/role/created_at——W1b 纯 web 可当轮落刀，二进制不动
- 首用户 bootstrap：store.go 首注册者即 admin；allow_register 可关注册——账号权属与测试路径依据
- TN 图定谳：工具名展示图早备 fx:pdf-split/compress——fx接线棘轮=4 即 EP2+TN2，「展示名先于接线」；FB 兜底假说证伪备案
- W1b（additive-only）：十壳 data-admin="true"→data-admin-req="true"（身份硬编码降格为门槛声明）；app.js ADMIN_REQ 异步门——admin 页先打 /api/auth/me，role==admin 注旗 boot，否则 ../login.html；tokens 注入器改读 body.dataset（作用域无关，贴合作者本意的 ../ 前缀）；代价：admin 页首帧白屏一个 me 往返（可接受，骨架屏缓行）
- 棘轮×2：admin旗硬编码=0、adminReq=10——客户端半边有绊线；服务端半边由 e2e 全流程覆盖在案
- 伪影二连（glob 抢答 + 引号字形不合）→ 案二十三；0 旗必问干草堆，正则必贴字形

## 增补四：W1b 活体验收与传输伪影（2026-10-28 · 第 62 发）

- 活体定谳：/admin/index.html 返回 301 重定向——浏览器真实路径以 Location/followed 回显为准；W1c 开案：循重定向补刀活体验收
- 五探针四绿在案：app.js gate=2、ratchet=111、me 401、盘面双棘轮 PASS——JS 门与 API 墙活体俱在，悬而未决的只是 HTML body 的传输 provenance
- 棘轮辖区注记：verify 双棘轮系盘面管辖（grep web/admin/*.html），活体传输层不在其辖——盘面绿+活体红不矛盾，是两层各说各话

## 增补五：W1c 收案与服务模型定谳（2026-10-28 · 第 63 发）

- W1c 收案：循 301 补刀——/admin/ code=200、body md5 与盘面新版一致（a137f8a1）、data-admin-req 活体命中；analytics.html 直达同验——W1b 客户端半墙端到端生效（浏览器级最后一步留你亲验：未登录开 /admin/ 应落 login.html）
- 服务模型定谳（精化第 60 发「磁盘直服坐实」注记）：main.go //go:embed web 默认烤进二进制（部署=单可执行），NOVA_WEB_DIR 环境变量覆盖为 http.Dir 磁盘直服（dev 模式）——现役 38593 environ 实读携带 NOVA_WEB_DIR——env 门控磁盘直服实锤；重启配方漏 env 即回退烤盘旧资产，NOVA_WEB_DIR 系承重墙（重启配方据此为契约件）
- 301 定性：Go FileServer index 规范化重定向（/admin/index.html → ./ → /admin/；/index.html → / 同理）——标准行为零改动；案二十四脚注：谜底良性收案，教训保留（传输层三证 301/空体/Location 本可一轮定谳）
- 决策点移交：首注册者即 admin（store.go bootstrap）——管理台真接线联测前，第一枚账号注册权属由你定夺（建议亲手注册）

## 增补六：DB 非空与部署契约入册（2026-10-28 · 第 64 发）

- 注册探针回读 role=user——DB 非空，bootstrap admin 已被领走（首用户 (?) · census users_total=0，时间戳可判是昨夜服务首启后注册还是更早遗留）；占位账号「<你的名>」入库（密码字面已知）——权属确认后阶梯清理或随清库消解
- 注册门：allow_register=未见键——实读自 census settings；真 admin 落定后建议 POST /api/admin/settings 关门（e2e:272 同款）
- 断链注记：上轮 DEPLOY.md 缺件（①回显早有预报「没有那个文件或目录」，未读成信号）——FileNotFoundError 断在第一写之前，工作区零污染（receipt 实证），门闸按设计拦下第 64 发；本轮改 create-if-missing 落刀——断在写前的失败是最便宜的失败
- DEPLOY.md 重启契约入册（承重条款+标准配方+自检）；verify 43 棘轮「重启契约在册」上岗
- W2 图纸在案：usersPage:167（演示数据全量）、GET /api/admin/users 形状 [{username,role,created_at}]、三 mutator 全 e2e 覆盖——下轮纯 web 零 Go 落刀；列瘦身（套餐/状态/AI 调用三列无数据源——裁到 username/role/created_at 或挂演示牌位）

## 增补七：十月十七悬案破案与占位号终局（2026-10-28 · 第 65 发）

- 破案：census 揪出 10-17 admin（role=admin · 2026-10-17T14:41:50）——DB 系 11 天在役库（History 416·Logs 500·Sessions 136·ToolState 27），「DB 非空」是文件历史非昨夜抢注；增补六「首用户 (?)」至此有名有姓
- A 外科执行战报：只清 Users（四段全保）·.bak 在册（原 10-17 admin 备份可考）·fresh bootstrap 生效——但占位符第三次被字面执行（<真名>/<真密码> 入库为唯一 admin）
- **阶梯终局（本发核心战果）**：login 占位→register 真（用户亲赋真值，具体值不入册）→promote→login 真→delete 占位→关门 allow_register=0——LADDER-OK sole-admin=admin，占位号消解，真号即唯一管理员，注册门已关
- 系统性对策三件套（全部实战验证）：①REFUSE 守卫首战告捷（用户首版赋值含 <> 被拒，二次真值通过）；②凭据前置赋值（真值由用户先赋 RN/RP，脚本本体占位字面仅用于靶账号）；③子壳包裹（两次守卫 exit 均未杀会话——案二十六修法当轮兑现）
- 案二十五入册：判决正则字段近视——weight 找 Runs/Workflows/LLMKeys 而 schema 实文 History/Logs/Sessions/ToolState，? 误读为轻；幸清库脚本外科边界独立自守（只清 Users），重数据无伤
- 案二十六入册：顶层 exit 陷阱——交互壳裸跑 exit 1 连 SSH 会话一起杀；凡含 exit 的执行段必须整体包 ( ) 子壳
- 案二十七入册：walrus 语法门·解析期中止·零副作用——:= 需 Python≥3.8，本机 3.7 拒解析——SyntaxError 炸在 parse 期、先于任何执行：文件零写入、哨兵零生成、commit 零发生，重跑完全幂等（本次即验证：同段原样重跑成功）
- 挂账待考：data/ 并见 nova.db（45K·10-26）+shm（32K）+wal（745K）与 nova.json（233K 在写）——store.go:74 字面系 nova.json，db 系另一实现或遗留组件；与本刀无关（阶梯走 API 层），列入后续考古队列
- 复跑双拦：旧定位器 miss data/（零伤）·死 PID readlink 空（零伤）——guard 再两次兑现；写死 PID 时代结束：后续一律 pgrep -f nova-server-linux 动态解析

## 增补八：T7-W2 users 真接线（2026-10-28 · 第 66 发）

- 王冠易主收官后首件施工：usersPage 全量换真——fetch /api/admin/users → [{username,role,created_at}] → 渲染；演示数字（128,436/86,421/1,206）→ 真数（users.length / admin count / 门状态）；userRows 假数据全量退役
- 列瘦身：套餐/状态/AI 调用三列无数据源全裁——改为用户/角色/注册日期/操作四列（全有源）；表头同步
- 四钮真钮化（data-toast 假面退役）：查看（toast 显示三元组）· role 改派（POST /role，admin↔user 双向，含 me 自我保护显示）· 重置密码（POST /reset，prompt 输入≥6位）· 删除（DELETE，confirm 二次确认）——全部端点 e2e 覆盖在案
- 附加：导出 CSV 前端生成（username,role,created_at）· 注册门开关（POST /settings allow_register 切换）· me 驱动「当前登录（自我保护）」标识（后端不能删自己守卫的前端镜像）
- 事件机制：bindUsersEvents 单次绑定（__usersBound 幂等守卫）+ closest 委托——避免重复 init 时多绑
- me precheck 扩展：ADMIN_REQ 分支里 me 成功且 adminPage==='users' 时 refreshUsers+refreshDoor——数据在 boot 前到位
- 棘轮×2：users演示数据=0（128,436/userRows/演示账户全退役）· users真接线≥12（ADMIN_USERS_LIVE/refreshUsers/adminUserAction 等标识在案）——W2 完成的绊线
- 兼容性纪律：Python 3.6.4 本机实测——heredoc 全段无 walrus、无 f-string =、无海象——案二十七修法沿用

## 增补九：W2 收尾——弹药库清除与口径校准（2026-10-28 · 第 67 发）

- 换段边界教训：外科刀只切了 usersPage 渲染函数，userRows 全局数据常量在换段外幸存——渲染退役不等于数据退役；另 128,436（adminDashboard metric 同伙）与「演示账户」字样各 1 处散落——共 3 处残留触发演示数据棘轮 FAIL
- 修法：userRows 定义段整删（弹药库清空）· 128,436→0（假数退位，dashboard 真接线留 W3）· 演示账户→真实账户；三词 count 归零，style 111 无伤
- 棘轮口径校准：users真接线 12→10——grep -c 数行数、v.count 数次数，写阈值前必先实测口径（案二十八）
- 案二十八入册：拍脑袋阈值×口径错位——棘轮值必须来自实测（先跑 grep -c 拿读数，再把读数写进 eq），不是来自「感觉该有 12 个」

## 增补十：W2 语法热修与语法门上岗（2026-10-28 · 第 68 发）

- 浏览器首访立功：users.html 白屏卡死（boot 前水帖）+ 控制台 app.js:221:132 SyntaxError——六十七发 grep 棘轮全绿照不出的真伤，用户亲验第一眼就揪出
- 病根：W2 heredoc 非原始 Python 串嵌 JS，join(\n) 的 \n 被解释成真换行写进文件——JS 单引号串断行、V8 parse 期即死、整页 app.js 不执行；伤处 2 处（rows join + CSV blob 头），修法=补成 \n 两字符转义；全文 quote-LF-quote 清零、style 111 无伤
- 案二十九：grep 棘轮测文本不测语法——语法只有 parser 说了算；verify 新增 JS 语法门（node --check 优先，缺 node 则 quote-LF-quote 扫描兜底）——文本棘轮+语法门双层合围才算数
- W3 脚本自身 TypeError 断在写前（%s 格式化撞上健康面板 CSS 的 100%; ）零副作用——且重审发现其表格行拼接缺 + 号，即便格式化成功也是坏 JS；W3 下轮重发：弃 %-format 改 token 替换、JS 段一律 r-string、代码段不过 N()

## 增补十一：语法门 v2——正则态上岗与会话复活注记（2026-10-28 · 第 70 发）

- 案三十一定谳：v1 门首巡抓 line 11 col 117 系假阳——V8 首错次序为证（上轮全文件首报 221，若 11 行真断必先报 11；本轮控制台零错）；真凶是正则字面量：escapeHtml 的 replace(/'/g,...) 模式——正则内引号把 v1 词法器踢进字符串态，后续真换行被冤判。正则是 JS 第三种串，词法器不认它就会冤枉好人
- v2 修法：正则态上岗——prev-token 启发（值后为除·op 后为正则·return/typeof 等关键词后为正则）·正则内裸换行=bad·跳过类处理并入跳字；夹具扩编至双向七枚（esc 全家桶/模板/除链/关键词正则 → 好样本；串内换行/未闭合/正则内换行×2 → 坏样本）——门上岗前先过门的仪式自此固化
- 会话复活注记：A 清库清 Users 未清 Sessions——旧 admin 的会话键在同名新 admin 复活后自动再挂，浏览器旧 cookie 直接获得新账号身份（本例无害且便利）；工法挂账：清用户应连 Sessions 一并清，否则同名账号继承旧钥匙（Go 侧卫生项，候 W 系列后议）
- login 跳转定谳：login.html 检 me=200 即送回首页——「已登录」标准行为非 bug；用户持活会话直开 /admin/users.html 即见 W2 真表
- W3 现场在档（/tmp/nv-w3pre.txt）：logs 形状实测为 at/user/method/path/status/ms——下轮硬编码映射（事件=path·范围=user·时间=at·状态=status），自适应猜测彻底退役；JS 段一律 r-string、弃 %-format（案二十七/二十九遗训）

## 增补十二：admin 壳全接线——首页拷贝病（2026-10-28 · 第 71 发）

- 浏览器第三庭：地址栏停在 /admin/users.html、渲染的是完整营销首页、控制台零错——非跳转（跳转会改地址）非崩溃（崩溃会报错），是壳原地启动常规应用画了 home：users.html 系首页拷贝，data-admin-req/data-admin-page 旗标从未存在，app.js 无旗标 → 常规 boot → 默认 home
- W2 盲区定谳：接线层（app.js）与壳层（HTML 旗标）双层，curl 验了 200+app.js 字节+端点，从没验过壳的属性——此前白屏反而掩护了它（app.js 没跑，壳病无从暴露）；浏览器法庭三连审（断串→热修无罪→壳病），每层真相都是 F12 先看到
- 修法：以 index 壳 body 标签为模板（旗标拼写与 index 严格一致），全 admin/*.html 缺旗标者整签替换、admin-page 按文件名赋值；全壳 census 后置断言（data-admin-req 全员在册）；棘轮「admin壳全接线」=grep -L 计无旗标文件数=0
- 案三十/三十一补入 B3 台账（上轮 A30 断言拦下坏锚点——案三十被引用却从未落刀；守卫零伤，先落锚再引用的纪律入册）
- 挂账：侧栏曝出 admin/usage.html 链接（底部状态栏）——壳清单待对照，缺件则 404；W4 队列核对导航全链接

## 增补十三：壳病翻案——诊断撤回与侦查重启（2026-10-28 · 第 72 发）

- census 判决：增补十二的「首页拷贝病」**撤回**——十壳生来全员有旗标（data-admin-req + data-admin-page 按文件名赋值），SHELLS-FIXED=(none)，本次会话无任何 commit 触碰 web/admin/*.html——壳从未带病，a-regret：诊断随 aac2d7e 入库在先、证据在后
- 案三十二（结论预支·侦查与台账同脚本）：上轮脚本把手术、census、台账写在一起，台账条目无条件落笔——若 SHELLS-FIXED 为空，诊断即假而台账照写；修法：文档写入必须以侦查结果为条件（fixed 非空才写诊断，空则写「待查」）——案二十八「先测后写」在文档域的复发，制度化只做了一半（棘轮做了，散文没做）
- 侦查员自误：上轮 ④ 探针用错文件名（notices——服务端实为 notifications.html）——404 有一半是探针的错；真缺件仅 usage.html（W4 导航审计队列）
- 现均线索：旗标在 + 控制台零错 + 营销首页原地渲染——嫌疑收敛为 ①boot() 的 admin 分支钥匙与壳旗标错位（line 112 currentAdminPage 常量的用途是关键）②me 对浏览器会话的身份判定；等待 boot 原文 + 控制台五行实验（L1-L5）定谳，证据齐前零代码手术
