# TODO

> 完成打 x；新需求追加到对应分组。

## 部署与交付（2026-11-05 用户提出：之前支持过二进制/Docker 部署，需补齐并优化）
- [ ] Dockerfile 多阶段构建（golang builder -> scratch/distroless，CGO_ENABLED=0）
- [ ] docker-compose.yml（端口/ADMIN_TOKEN/数据卷示例）
- [ ] Makefile：build / docker-build / docker-push / release 目标
- [ ] GitHub Actions release.yml：打 tag 自动产出 linux/darwin/windows 二进制 + GHCR 镜像
- [ ] README 部署章节：环境变量清单（ADMIN_TOKEN、各模型 API Key）

## NT 迭代
- [x] NT-1 功能池 Feature Flag + SSE 实时上下线（代码落地，CI 验证中）
- [ ] NT-2 六家模型 API Key 配置（llm-profiles 页 + llmproxy 流式透传）
- [ ] NT-3 全流程 7 步（pipeline.html + 状态机，消费 stage-gates.yaml 与 7 步模板）
- [ ] NT-4 监控报警（metrics.go 按 llmops-metrics 口径 + admin 大屏）
- [ ] 功能池持久化（内存 -> store.go/PostgreSQL，重启不丢状态）
- [ ] 功能池 21 -> 60+（PPT 模块补齐、更多文档/表格 AI 增强）

## 基础设施
- [ ] 真实鉴权（JWT + RBAC，替换 ADMIN_TOKEN dev 放行）
- [ ] 真实文件转换后端（替换演示数据）
- [ ] E2E 浏览器自动化 + 负载测试
- [ ] 功能池与现有 ToolRegistry/ToolOn 及 T-07 rollout 管理侧合并去重（Feature.Percent/putFeature/listFeatures）
