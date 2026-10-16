#!/usr/bin/env bash
cd "$(dirname "$0")"
pkill -f '/nebula' 2>/dev/null && echo "✓ nebula 已停止" || echo "· nebula 未在运行"
if [ -f data/mysql.pid ]; then
  kill "$(cat data/mysql.pid)" 2>/dev/null && echo "✓ 随包 MySQL 已停止"
fi
