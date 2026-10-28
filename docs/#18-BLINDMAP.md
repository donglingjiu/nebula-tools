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
