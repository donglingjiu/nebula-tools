# Nova 工具箱

单二进制服务：前端已内嵌（go:embed），数据落盘 data/，零外部依赖。

## 部署 A：Docker（推荐）
docker compose up -d
curl -s localhost:8080/api/tools | head -c 200
# 或直接用云端镜像（包需为 public，否则先 docker login ghcr.io）：
docker run -d -p 8080:8080 -v "$PWD/data:/app/data" ghcr.io/donglingjiu/nebula-tools:latest

## 部署 B：裸二进制（Linux / macOS / Windows 通用思路）
# 到 Releases 下载对应平台文件（各带 sha256），然后：
chmod +x nova-server-linux-amd64 && PORT=8080 ./nova-server-linux-amd64

## 发版
git tag vX.Y.Z && git push origin vX.Y.Z
# CI 自动完成：测试 → 三平台二进制 → GHCR 双架构镜像 → Releases 挂载
