#!/usr/bin/env bash
# Nebula Tools 部署器：core 包用 SQLite；full 包自动初始化随包 MySQL（失败自动回退 SQLite）
cd "$(dirname "$0")"
mkdir -p config data
[ -f config/db.yaml ] || echo 'driver: sqlite' > config/db.yaml
chmod +x nebula 2>/dev/null

if [ -x runtime/mysql/bin/mysqld ] && [ ! -d data/mysql/mysql ]; then
  echo "[1/3] 初始化随包 MySQL…"
  mkdir -p data/mysql
  if runtime/mysql/bin/mysqld --no-defaults --initialize-insecure \
       --basedir="$PWD/runtime/mysql" --datadir="$PWD/data/mysql" >> data/mysql.log 2>&1; then
    nohup runtime/mysql/bin/mysqld --no-defaults --basedir="$PWD/runtime/mysql" \
      --datadir="$PWD/data/mysql" --port=3307 --bind-address=127.0.0.1 \
      --pid-file="$PWD/data/mysql.pid" >> data/mysql.log 2>&1 &
    ok=0
    for i in $(seq 1 30); do
      runtime/mysql/bin/mysql --no-defaults -h 127.0.0.1 -P 3307 -u root -e 'SELECT 1' >/dev/null 2>&1 && { ok=1; break; }
      sleep 1
    done
    if [ $ok = 1 ]; then
      runtime/mysql/bin/mysql --no-defaults -h 127.0.0.1 -P 3307 -u root \
        -e 'CREATE DATABASE IF NOT EXISTS nebula CHARACTER SET utf8mb4' >/dev/null 2>&1
      printf 'driver: mysql\nhost: 127.0.0.1\nport: 3307\nuser: root\npassword: ""\ndatabase: nebula\n' > config/db.yaml
      echo "      ✓ MySQL 就绪（127.0.0.1:3307，库 nebula 自动创建）"
    else
      echo "      ✖ MySQL 启动失败，回退 SQLite（详情 data/mysql.log）"
      echo 'driver: sqlite' > config/db.yaml
    fi
  else
    echo "      ✖ 初始化失败，回退 SQLite（详情 data/mysql.log）"
    rm -rf data/mysql
  fi
else
  echo "[1/3] 使用现有数据库配置（$(grep driver: config/db.yaml | head -1)）"
fi

echo "[2/3] 启动 Nebula…"
nohup ./nebula >> nebula.log 2>&1 &
echo "[3/3] 等待就绪…"
for i in $(seq 1 30); do
  curl -sf http://localhost:8080/api/health >/dev/null 2>&1 && {
    echo ""
    echo "✅ Nebula Tools 已启动"
    echo "   前台   http://localhost:8080"
    echo "   管理   右上角 ⚙️（admin / admin123）"
    echo "   日志   nebula.log    停止: bash stop.sh"
    exit 0; }
  sleep 1
done
echo "✖ 30s 未就绪，nebula.log 尾部："
tail -n 15 nebula.log
exit 1
