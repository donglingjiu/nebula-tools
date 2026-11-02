# 7-3 阶段复盘：问题 / 解决 / 亮点（2025 工具域战役）

> 范围：B-1 盘点 → ①C 接线 → ②A 留档 → ③ 测试矩阵 → ④ 卫生收官
> commit 链：7b4bfe0(B-1盘点+②A) → 494bf98(①C接线) → b80286b(③矩阵) → 97181db(④卫生)

## 一、遇到过的问题与解决（P1-P8）

| # | 问题 | 根因 | 解决 | 沉淀规约 |
|---|---|---|---|---|
| P1 | E2E 五连全 FAIL（E1 状态码 000） | 9091 E2E 实例从未启动，后续 curl 级联落空 | 取证 PORT env 机制（main.go:114）后补起实例段；守卫先拦 | E2E 前置健康检查必须先于断言执行 |
| P2 | `base64: command not found` | 内网机无 base64 CLI，PNG fixture 未生成 | python3 base64 模块生成 | 内网工具链不可假设，解释器 stdlib 优先 |
| P3 | 守卫判据自堵（E2E 全过后仍拦） | 判据写 `nova-fx>=5`，patch 内该词仅 1 次（sec.id） | 判据与实际计数对齐（fx-≥5 + /api/t/≥6 + nova-fx≥1） | 守卫判据设计期必须 grep 实数，不拍脑袋 |
| P4 | dev.totp 测试失败 "input 填 Base32 密钥" | 假设"生成型工具零输入可跑"，违背契约 | RFC6238 惯用密钥 JBSWY3DPEHPK3PXP 注入用例 | 用例输入按契约（NR_SCHEMA）填，不按直觉省略 |
| P5 | FAIL E4:ffd8ff（断言误报） | magic ffd8ff 恰是 JPEG 本体=功能成功；挂的是 MIME 断言，后端直拼 image/jpg（非规范但兼容） | grep 响应头取证坐实 → 断言修 image/(jpeg\|jpg)；记录观察项 | 功能真值与规范符合性分开断言，先取证再定性 |
| P6 | git push 443 SSL_ERROR_SYSCALL 反复 | 内网出口 github 抖动 | `git push \|\| git push` 重试链（三次内成功） | 网络抖动是常态；本地 commit 先落袋不丢 |
| P7 | workspace IIFE 局部作用域不可复用 | runOne/dl/outName 困在 600 行 app.js 作用域内 | 自包含 IIFE 复刻 ~15 行（401→toast→跳转 + blob 下载 + CD 解析） | 复刻成熟模式远比全局化重构安全 |
| P8 | /api/t/* 三工具零前端入口（孤岛） | 后端注册路由但从未接线 | tools.html 传统文件工具区：select+file+opts 表单+Content-Type 分流 | 后端能力盘点必须前端接线对账（双轨五维） |

## 二、亮点

1. **三族工具域全双通**：/api/run 23 文本（试炼场）+ /api/t/* 3 文件（本段接线）+ /api/fx/* 5 PDF（workspace）——31 端点前后端全链路，孤岛清零
2. **三锁测试矩阵**（tools_test.go，0.021s）：快照锁（len==27 漂移即红）+ 正向锁（每个注册 id 必有 case）+ 反向锁（case 必在 Registry）——**新增工具不带测试 = go test 直接红**
3. **守卫链零误放行**：E2E mark 计数 + patch grep 计数 + go build/test 与链；两次拦截皆真问题（E2E 未起、totp 契约错）
4. **E2E 实例三隔离**：独立二进制名（-o /tmp/nova-e2e73，防 pkill 误杀 8080）+ 独立 data dir + PORT env（源码唯一注入点先取证）
5. **断言选型规约**：强断言特征串优先取 registry desc 亲供值（QR dataURI/chmod rwx/子网掩码）；实现细节未知降级长度断言；excel 用同库 excelize 构造 fixture，pdf/word 只测分发可达——**测试锁分发链不锁三方库**
6. **收官清理三判据**：跟踪态定删除方式（git ls-files）+ 进程占用定保留边界（ps）+ ignore 覆盖定补口需求（check-ignore）——9 个 .bak 清 1.1MB，运行实例零受扰
7. **B-1 意外收获**：三族定谳（原判"双轨"漏计 /api/fx）；公开工具箱无门≠缺陷（DEPLOY.md 设计意图考古，②A 留档）

## 三、遗留观察项（下阶段议题）

1. image/jpg 非规范 MIME——filetools.go fxSend 一行修（mime 映射 jpg→jpeg），功能无损纯规范问题
2. /api/upload 文件轨 4 工具零前端——与 /api/fx 功能域重叠，需拍板：合并 / 并存 / 下线
3. 传统文件工具区发现性——当前在 tools.html 页面底部区块，是否进主工具卡导航

## 四、方法论沉淀（35 条规约精选）

- 证据链驱动：每段执行前取证（证N编号），判读后再写 patch——不猜契约
- 契约先行：前端表单/测试用例的 opts 键值全部来自后端源码取证，不按直觉
- 假阴性防线：守卫判据宁可错拦不可错放；拦截必回贴取证再放行
- 收官清理三判据：物理删除仅对"未跟踪+价值已收割+无进程依赖"三者同时成立的文件
