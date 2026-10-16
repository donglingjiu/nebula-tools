# 部署手册

## A. GitHub 构建便携包（主链路）
1. 网页上传/推送全部代码到 GitHub 仓库（main 分支）
2. 仓库顶部 **Actions** 标签 → 左侧 **Release Portable**
3. 右侧灰色 **Run workflow** 下拉 → Branch 选 `main` → 版本号填 `v1.0.0` → 绿色 **Run workflow**
4. 等 3~6 分钟变绿 ✓ → 点进本次运行 → **页面最底部 Artifacts**：
   - `nebula-windows-amd64` → Windows 包
   - `nebula-darwin-arm64` → Apple Silicon Mac 包
   - `nebula-linux-amd64` → Linux 包
5. 解压两层：artifact zip → 里面的 `nebula-portable-v1.0.0-*-full.zip`（或 core）
6. **Windows**：放到纯英文路径 → 双击 `install.bat`（SmartScreen 选"更多信息→仍要运行"）
   **Mac/Linux**：`tar xzf nebula-portable-*.tar.gz && cd nebula-portable-* && bash install.sh`
7. 浏览器打开 http://localhost:8080

## B. 验证清单
- [ ] 前台页脚：`v1.0.0 · 构建链路 Go 1.22.x · 随包组件 MySQL 8.0.36 / Python 3.11.9 · 构建于 …`
- [ ] 右上角 ⚙️ → admin/admin123 登录成功
- [ ] 模块管理：任关一个 → 前台卡片消失、接口 403
- [ ] 数据画像 / 告警中心 / 依赖与版本 / 数据库 五个页签可打开
- [ ] full 包：后台"数据库"页显示 `mysql`；core 包显示 `sqlite`

## C. 云部署
- **Fly.io**：`flyctl launch`（读 fly.toml，含持久卷 + 健康检查）
- **Railway**：新建项目 → Deploy from GitHub（读 railway.json）

## D. 故障速查
| 现象 | 处置 |
|---|---|
| Actions 页没有 Run workflow | yml 必须在 main 分支且路径为 `.github/workflows/release.yml` |
| 构建红 ❌ | 点开日志看红的第一行；贴给 AI 按行修 |
| install.bat 闪退 | 路径含中文/空格 → 挪纯英文路径；闪退后看 data\mysql.log 与 nebula.log |
| full 包 MySQL 起不来 | 安装器已自动回退 SQLite，功能不受影响；Windows 缺 VC++ 运行库时安装它即可 |
| 8080 被占用 | `set PORT=9090` 后运行 start.bat（Mac/Linux: `PORT=9090 bash install.sh`） |
