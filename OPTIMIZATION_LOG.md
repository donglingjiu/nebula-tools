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
