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

待核二针：conversations PATCH/DELETE（api() 助手 method 对 grep 不可见）；EP 映射全集。

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
