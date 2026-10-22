# Changelog

## v2.6.0 - 2026-10-22
### Added
- P0-2 工具扩容四件套：dev.cron（5 段解析+未来运行时间）、dev.chmod（八进制↔符号双向）、dev.subnet（IPv4 CIDR）、dev.yaml（YAML↔JSON 自动识别，引入 gopkg.in/yaml.v3）
- 工具大全网格新增 4 卡并接入试炼场暗道（计数 36→40）；NR_SCHEMA 补 cron 的 count 参数

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
