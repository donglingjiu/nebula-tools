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
