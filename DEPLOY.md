# Nova 部署与运维

## 重启契约（2026-10-28 · 服务模型定谳随件 · 承重条款）

- 产线：/home/work/fanguohui/nova-server-linux-amd64（现役 PID 38593 · mtime 2026-10-27 19:35）
- 服务模型：二进制默认 //go:embed web（烤盘）；携带 NOVA_WEB_DIR 时改 http.Dir 磁盘直服——现役进程 /proc/38593/environ 实测携带
- **承重条款：重启必须携带 NOVA_WEB_DIR=/home/work/fanguohui/nebula-tools/web——漏带即回退烤盘旧资产，web-only 改动「push 即产线」依赖此 env**
- 标准配方：cd /home/work/fanguohui && pkill -f nova-server-linux-amd64; sleep 2; (NOVA_WEB_DIR=/home/work/fanguohui/nebula-tools/web nohup ./nova-server-linux-amd64 > nova8080.log 2>&1 &); sleep 2
- 重启自检：curl -s http://127.0.0.1:8080/admin/ | grep -c data-admin-req 应为 1

## 环境变量清单（7-1 配置外置）

| 变量 | 默认 | 说明 |
|---|---|---|
| PORT | 8080 | 监听端口 |
| NOVA_DATA_DIR | data | 数据目录(users/admin/flows/orders/llm/nova.db), 不存在自动创建 |
| NOVA_WEB_DIR | 烤盘 | web 资产磁盘直服目录, 见顶部承重条款 |

## 产物名关系

| 产物 | 产出者 | 用途 |
|---|---|---|
| nebula-tools | dev.sh build | 本地开发回路 |
| nova-box | build.sh | 产线发布 |
| nova | Dockerfile | 容器内(docker本机不可用, 走静态走查, 真build待有docker环境) |

## 备份与恢复（7-1）

- 备份: `bash scripts/backup.sh [数据目录] [备份目录] [保留份数]`（默认 data/ backups/ 14份）
- 恢复: 停服 → `tar -xzf backups/nova-data-<ts>.tar.gz -C 仓库根` → 覆盖 data/ → 重启
- 已验证: 双实例隔离(9091+NOVA_DATA_DIR=/tmp)数据全落新目录, 8080现役零影响
