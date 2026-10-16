#!/usr/bin/env bash
# >>> 部署前停掉占用 8080 的老进程（CentOS6 兼容：netstat 优先 / ss 全路径兜底）
OLD=""
if command -v netstat >/dev/null 2>&1; then
  OLD=$(netstat -ltnp 2>/dev/null | grep ':8080 ' | awk '{print $NF}' | grep -oE '^[0-9]+' | sort -u)
fi
if [ -z "$OLD" ]; then
  for S in ss /usr/sbin/ss; do
    command -v $S >/dev/null 2>&1 && OLD=$($S -ltnp 2>/dev/null | grep ':8080 ' | grep -oE 'pid=[0-9]+' | cut -d= -f2 | sort -u) && break
  done
fi
if [ -n "$OLD" ]; then
  echo "停止老进程: $OLD"
  kill $OLD 2>/dev/null || true
  sleep 1
  STILL=""
  if command -v netstat >/dev/null 2>&1; then
    STILL=$(netstat -ltnp 2>/dev/null | grep ':8080 ' | awk '{print $NF}' | grep -oE '^[0-9]+' | sort -u)
  fi
  [ -n "$STILL" ] && { echo "强制结束: $STILL"; kill -9 $STILL 2>/dev/null || true; sleep 1; }
fi
# <<<
set -e
BIN=nova-server-linux-amd64; PORT=8080
[ -f "$BIN" ] || BIN=dist/$BIN
[ -f "$BIN" ] || { echo "✗ 未找到二进制"; exit 1; }
chmod +x "$BIN"

# ① 端口占用检查：是自家旧实例就停，是别人的服务就报错退出
OWNER=$(ss -tlnp 2>/dev/null | grep ":$PORT " | grep -oP 'pid=\K[0-9]+' | head -1)
[ -z "$OWNER" ] && OWNER=$(fuser $PORT/tcp 2>/dev/null | tr -s ' ' '\n' | head -1)
if [ -n "$OWNER" ]; then
  echo "⚠ 端口 $PORT 被 PID=$OWNER 占用："
  ps -p "$OWNER" -o pid,user,cmd --no-headers || true
  case "$(ps -p "$OWNER" -o cmd= 2>/dev/null)" in
    *nova*|*nebula*) kill "$OWNER" 2>/dev/null || kill -9 "$OWNER"; echo "→ 已停止旧实例"; sleep 1;;
    *) echo "✗ 占用者不是 nova/nebula，请人工确认后手动 kill $OWNER"; exit 1;;
  esac
fi

# ② 启动 + 立即验证进程没死（bind 失败会在这里被抓到）
mkdir -p data
nohup "./$BIN" > nova.log 2>&1 & PID=$!
sleep 1
kill -0 "$PID" 2>/dev/null || { echo "✗ 进程启动即退出："; tail -20 nova.log; exit 1; }
echo "已启动 PID=$PID"

# ③ 健康检查 + 展示进程信息供人工核对
for i in $(seq 1 10); do
  H=$(curl -s "http://localhost:$PORT/api/health" || true)
  if printf '%s' "$H" | grep -q '"ok":true'; then
    echo "── health ──"; printf '%s\n' "$H"
    echo "── 进程 ──"; ps -p "$PID" -o pid,cmd --no-headers
    echo "═══ 部署成功 ═══"; exit 0
  fi; sleep 1
done
echo "✗ 未就绪："; tail -20 nova.log; exit 1
