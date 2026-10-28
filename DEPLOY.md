# Nova 部署与运维

## 重启契约（2026-10-28 · 服务模型定谳随件 · 承重条款）

- 产线：/home/work/fanguohui/nova-server-linux-amd64（现役 PID 38593 · mtime 2026-10-27 19:35）
- 服务模型：二进制默认 //go:embed web（烤盘）；携带 NOVA_WEB_DIR 时改 http.Dir 磁盘直服——现役进程 /proc/38593/environ 实测携带
- **承重条款：重启必须携带 NOVA_WEB_DIR=/home/work/fanguohui/nebula-tools/web——漏带即回退烤盘旧资产，web-only 改动「push 即产线」依赖此 env**
- 标准配方：cd /home/work/fanguohui && pkill -f nova-server-linux-amd64; sleep 2; (NOVA_WEB_DIR=/home/work/fanguohui/nebula-tools/web nohup ./nova-server-linux-amd64 > nova8080.log 2>&1 &); sleep 2
- 重启自检：curl -s http://127.0.0.1:8080/admin/ | grep -c data-admin-req 应为 1
