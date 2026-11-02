# Nova 部署与运维

## 重启契约（7-2 对账修订 · 进程发现式 · 零运行时断言）

- 产线二进制：/home/work/fanguohui/nova-server-linux-amd64（mtime 2026-10-27 19:35 · 快照值）——**7-2 对账时进程已不在**（PID 38593 不存在，无任何运行实例，停机时点不可考）
- **漂移警示：产线二进制构建于 2026-10-27 19:35，其后 main 新增 62 个提交（含 flow 用户隔离/402 购买闭环/NOVA_DATA_DIR 配置外置）——重启前必须重建，否则产线跑旧代码**
- 服务模型：二进制默认 //go:embed web（烤盘）；携带 NOVA_WEB_DIR 时改 http.Dir 磁盘直服（历史实测携带：2026-10-28 /proc/38593/environ，进程已逝，证据留档）
- **承重条款：产线重启必须携带 NOVA_WEB_DIR=/home/work/fanguohui/nebula-tools/web——漏带即回退烤盘旧资产，web-only 改动「push 即产线」依赖此 env**
- 活配方（进程发现式，不断言 PID）：
  1. pgrep -f 'nova-server-linux-amd64|nova-box' >/dev/null && pkill -f 'nova-server-linux-amd64|nova-box'; sleep 2
  2. cd /home/work/fanguohui/nebula-tools && bash build.sh    # 重建 → /home/work/fanguohui/nova-box（新产物名，取代历史名 nova-server-linux-amd64）
  3. cd /home/work/fanguohui && (NOVA_WEB_DIR=/home/work/fanguohui/nebula-tools/web nohup ./nova-box > nova8080.log 2>&1 &); sleep 2
- 重启自检：curl -s http://127.0.0.1:8080/api/health 应 200；curl -s http://127.0.0.1:8080/admin/ | grep -c data-admin-req 应为 1（本机开发实例同样命中，已实测）
- **本机 8080 现为开发回路实例（./nebula-tools）——上述配方在部署目标机执行；在本机执行步骤 1 会先杀开发实例**
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

## 公开工具箱条款（B-2 拍板②A）

- `POST /api/run/{tool}`、`POST /api/upload/{tool}` 无登录门为**有意设计**（公共工具箱定位）：ToolOn 模块开关是唯一闸门，登录仅影响历史记录（AddHist）与告警归属（sendAlert）
- GET /api/tools 匿名仅见已开启工具，admin 全量可见
- 对外收敛手段：admin 工具开关关停，或前置反代鉴权（应用层不加门）
