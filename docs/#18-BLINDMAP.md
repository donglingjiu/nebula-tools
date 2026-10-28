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
