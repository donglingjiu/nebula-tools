# Changelog

## v2.8.4 - 2026-10-23
### Refactor
- 内联收编第二批（标签感知合并）：18 处 → 8 个工具类（py4/mt12/mt10/mt9/mt8/ico32/ico31/mlauto）；棘轮 138→120；tokens 缓存戳 v280→v284
### Fixed
- 塌缩脚本升级：class 与 style 被其他属性分隔时合并进既有 class（DOUBLE-CLASS 卫兵首次拦截成功，文件零损伤）；a030d55 为半程提交（tokens+bump+棘轮），本版补齐塌缩与 changelog

## v2.8.3 - 2026-10-23
### Refactor
- 内联 style 收编第一批：flex:1×11、margin:0×4、三列网格×9、两列网格×2 → .nv-flex1/.nv-m0/.nv-grid3/.nv-grid2（忠实翻译+过渡护甲，共 26 处）；verify.sh 增至 17 项

## v2.8.2 - 2026-10-23
### Fixed
- 徽章路由 JS 补落：v2.8.1 仅落 .tag.plan 样式，toolCard 的 可用/规划中 路由因断言笔误未写盘（卫兵拦截零损伤）；现规划中徽章转灰
- word2pdf.html 补接 tokens（v2.8.1 因 head 结构跳过），独立页 19/19，verify.sh 增至 16 项
### Corrected
- v2.8.1 changelog「19 个独立页」实为 18，本版补齐并更正

## v2.8.1 - 2026-10-23
### Added
- B2 第一批：19 个独立工具页批量接入 nova-tokens.css（滚动条/选区/焦点环全站统一），verify.sh 增至 14 项
- 徽章规范：可用=info 蓝、非可用=plan 灰（.tag.plan 入 tokens）

## v2.8.0 - 2026-10-23
### Added
- nova-tokens.css 设计系统单一真源：清爽（冷白底+细边框）/好用（focus-visible 焦点环+悬浮反馈）/高级感（三档分层阴影+统一圆角+细滚动条+选区色），app.js 注入器全站最后加载，11 个主航页面生效
### Fixed
- 命令面板 Enter 失效：纯 NR 工具命中时无 active 项（现首位自动 active + Enter 兜底 items[0]）；verify.sh 增至 13 项

## v2.7.4 - 2026-10-22
### Fixed
- 命令面板页面段缺 .join('') 导致卡片间出现字面逗号（v2.7.0 起）

## v2.7.3 - 2026-10-22
### Fixed
- 命令面板修复：v2.7.0 双重 .join('') 致 renderCommandItems TypeError，浮层永不显示；点击绑定 $ 包裹误用改 $$ 直调（磁盘覆盖秒级上线，断言卫兵拦截一次错误替换）

## v2.7.2 - 2026-10-22
### Added
- NOVA_WEB_DIR 磁盘优先覆盖：设置环境变量时静态资源优先读磁盘（改 JS 免编译免重启刷新即生效），未设置时保持单文件内嵌部署

## v2.7.1 - 2026-10-22
### Fixed
- Mac Chrome 保留 Cmd+K（页面无法拦截）→ 新增 / 键呼出命令面板（输入框内不误触发），Cmd+K 保留兼容 Safari，底部提示更新

## v2.7.0 - 2026-10-22
### Added
- Ctrl+K 命令面板接入 14 个试炼场工具：搜索工具名/id，本页直达自动选中，跨页深链 tools.html?tool=<id>（带下拉填充重试）

## v2.6.2 - 2026-10-22
### Fixed
- 前端接线真正落地（v2.6.1 因补丁断言拦截未改 app.js，实为仅 CHANGELOG 的空版本）：4 新卡入网格、暗道直通、NR_SCHEMA 补 cron count（计数 36→40）

## v2.6.1 - 2026-10-22
### Fixed
- 仅变更日志版本：前端接线补丁被断言正确拦截（锚点错误），app.js 未改动

## v2.6.0 - 2026-10-22
### Added
- P0-2 工具扩容四件套：dev.cron（5 段解析+未来运行时间）、dev.chmod（八进制↔符号双向）、dev.subnet（IPv4 CIDR）、dev.yaml（YAML↔JSON 自动识别，引入 gopkg.in/yaml.v3）
- （仅后端；网格接线与 cron count 参数栏实际落于 v2.6.2）

## v2.5.3 - 2026-10-22
### Added
- 真机试炼场按工具渲染参数栏（NR_SCHEMA）：hmac key/alg、case mode、totp code、lorem paras

## v2.5.2 - 2026-10-22
### Added
- 工具大全网格新增 HMAC/TOTP/JWT/Slug/Lorem 5 卡；字数统计、大小写转换升级后端版；卡片点击直通试炼场（计数 31→36）

## v2.5.1 - 2026-10-22
### Fixed
- e2e 工具数断言跟随 ToolRegistry，新增工具不再改测试

## v2.5.0 - 2026-10-22
### Added
- P0-1 微工具带 7 件（hmac/totp/jwt/case/slug/lorem/count），extraRunners 钩子 tools.go 仅 +3 行；新工具默认开启
