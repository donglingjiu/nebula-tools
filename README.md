# 🪐 Nebula Tools

零依赖便携工具站：12 个工具模块 + 后台管理（模块开关/数据画像/告警中心/依赖版本/一键切库）。

## 便携包部署（推荐）
GitHub Actions 运行页底部 Artifacts 下载对应平台包：
- **core**：SQLite 零依赖，解压执行 `install.sh` / 双击 `install.bat`
- **full**：额外随包 MySQL + Python，首次安装自动初始化（失败自动回退 SQLite）

## 源码运行
    go run ./cmd/server        # http://localhost:8080

默认管理员 admin / admin123。版本标识由 scripts/components.env 统一维护，CI 自动写入。
