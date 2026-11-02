# 案例：600行巨型app.js上的孤岛后端接线（nebula-tools 7-3）

> 来源：nebula-tools 工具域战役 · /api/t/* 三文件工具接入 tools.html
> 调用时机：给已有应用新增后端能力入口，且既有 JS 为巨型 IIFE 局部作用域时

## 场景
app.js 614 行，workspace 的 runOne/dl/outName 困在 IIFE 局部作用域；后端 /api/t/* 三路由（CSV清洗/图片处理/Excel剖析）功能完整但零前端入口。

## 可复用模式
1. **锚点前置 patch + 双 assert**：python3 inline（pathlib）——`assert anchor in s`（锚点漂移检测）+ `assert 'nova-fx' not in s`（防重入），patch 插在稳定注释锚点之前
2. **自包含 IIFE 而非全局重构**：复刻成熟模式 ~15 行（401→toast→延时跳 login、blob+URL.createObjectURL 下载、Content-Disposition filename 双正则解析）——不动 600 行既有作用域，零回归风险
3. **页面守卫对齐**：`if(!/tools\.html/.test(location.pathname))return;` 与既有区块同款
4. **schema 驱动按需渲染**：opts 表单按端点映射（FX_SCHEMA）+ file input accept 联动（FX_ACC）——对齐既有 NR_SCHEMA 模式
5. **Content-Type 分流**：同端点混合响应不按端点硬编码——`ct.indexOf("json")>=0` 走 JSON 展示，否则 blob 下载（xlsx 默认剖析展示、format=md 自动落下载分支）
6. **错误解析对齐后端契约**：后端 fxE 返回 {"error":msg} → 前端 `r.json().catch(()=>({}))` 兜底

## 陷阱
- helper 命名加前缀（fxel/fxt/fxesc）——多 IIFE 作用域不冲突但 grep 排查易混
- 无浏览器环境的验证口径：API 层 E2E 五连 + grep 计数守卫，页面渲染靠计数旁证

## 效果
+70 行自包含块落地，孤岛三工具全链通；E2E 五连（401门/登录/csv去重/jpg magic/xlsx剖析）全 PASS。
