# 部署手册

## A. GitHub 构建便携包（主链路）
1. 代码推送到 GitHub 仓库（main 分支）
2. 仓库顶部 **Actions** → 左侧 **Release Portable**
3. 右侧灰色 **Run workflow** 下拉 → Branch `main` → 版本 `v1.0.0` → 绿色 **Run workflow**
4. 等 3~6 分钟变绿 ✓ → 点进运行 → **页面最底部 Artifacts** 下载对应平台包
5. 解压两层：artifact zip → 里面的 `nebula-portable-*-full.zip`（或 core）
6. Windows：纯英文路径双击 `install.bat`（SmartScreen 选"仍要运行"）
   Mac/Linux：`tar xzf *.tar.gz && cd nebula-portable-* && bash install.sh`
7. 浏览器 http://localhost:8080

## B. 验证清单
- 页脚版本条：v1.0.0 · 构建链路 Go 1.22.x · 随包组件 MySQL 8.0.36 / Python 3.11.9
- ⚙️ admin/admin123 登录 → 模块开关生效 → 五个后台页签可用
- full 包后台显示 mysql；core 包显示 sqlite

## C. 云部署
Fly.io：`flyctl launch`；Railway：Deploy from GitHub（读 railway.json）

## D. 故障速查
| 现象 | 处置 |
|---|---|
| Actions 无 Run workflow | yml 必须在 main 分支 `.github/workflows/release.yml` |
| 构建红 ❌ | 点开日志看红的第一行 |
| install.bat 闪退 | 路径含中文/空格 → 挪纯英文路径；看 data\mysql.log |
| 8080 占用 | `set PORT=9090`（Win）/ `PORT=9090 bash install.sh` |
