#!/usr/bin/env sh
set -eu
PORT="${1:-8080}"
echo "Nova 静态原型正在本地提供服务： http://localhost:${PORT}"
exec python3 -m http.server "$PORT" --bind 127.0.0.1
