# TODO

> 完成打 x；新需求追加到对应分组。

## 部署与交付（2026-11-05 用户提出：之前支持过二进制/Docker 部署，需补齐并优化）
- [x] Dockerfile 多阶段构建（已存在，见 README 部署A / Dockerfile）
- [x] docker-compose.yml（已存在）
- [ ] Makefile：build / docker-build / docker-push / release 目标
- [x] GitHub Actions release：tag 自动三平台二进制+GHCR 双架构（已存在，见 README 发版）
- [x] README 部署章节（DEPLOY.md + README 部署A/B 已存在）

## NT 迭代
- [x] NT-1 功能池 Feature Flag + SSE 实时上下线（代码落地，CI 验证中）
- [x] NT-2 六家模型 API Key 配置（后端 llmproxy+dbprofiles 既有；llm.html 增 6 家预设一键填充）
- [ ] NT-3 全流程 7 步（pipeline.html + 状态机，消费 stage-gates.yaml 与 7 步模板）
- [ ] NT-4 监控报警（metrics.go 按 llmops-metrics 口径 + admin 大屏）
- [ ] 功能池持久化（内存 -> store.go/PostgreSQL，重启不丢状态）
- [ ] 功能池 21 -> 60+（PPT 模块补齐、更多文档/表格 AI 增强）

## 基础设施
- [ ] 真实鉴权（JWT + RBAC，替换 ADMIN_TOKEN dev 放行）
- [ ] 真实文件转换后端（替换演示数据）
- [ ] E2E 浏览器自动化 + 负载测试
- [ ] 功能池与现有 ToolRegistry/ToolOn 及 T-07 rollout 管理侧合并去重（Feature.Percent/putFeature/listFeatures）
- [ ] Makefile（build.sh 已覆盖 build/dev/smoke/verify，可按需补 Makefile 薄封装）
