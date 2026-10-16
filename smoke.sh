#!/usr/bin/env bash
# 用法: 先运行 nova-server，再  bash smoke.sh [base_url]
# 注意: 重复运行前 rm -f data/nova.json，否则 register 会因用户已存在而 FAIL
B=${1:-http://localhost:8080}
J=$(mktemp); P=0; F=0
ck(){ if [ "$2" = "1" ]; then echo "PASS $1"; P=$((P+1)); else echo "FAIL $1 → ${3:-}"; F=$((F+1)); fi; }
code(){ curl -s -o /dev/null -w '%{http_code}' "$1"; }

h=$(curl -s $B/api/health);                       ck 'health'        "$(echo "$h" | grep -q '"ok":true' && echo 1)" "$h"
ck 'index页'     "$([ "$(code $B/)" = 200 ] && echo 1)" ''
ck 'app.js'      "$([ "$(code $B/assets/app.js)" = 200 ] && echo 1)" ''
ck 'experiences' "$([ "$(code $B/assets/experiences.css)" = 200 ] && echo 1)" ''
ck '未登录me→401' "$([ "$(code $B/api/auth/me)" = 401 ] && echo 1)" ''

reg=$(curl -s -c $J -X POST -H 'Content-Type: application/json' -d '{"username":"smoke","password":"smoke123"}' $B/api/auth/register)
ck 'register(首个=admin)' "$(echo "$reg" | grep -q '"role":"admin"' && echo 1)" "$reg"

r1=$(curl -s -b $J -X POST -H 'Content-Type: application/json' -d '{"input":"{\"a\":1}"}' $B/api/run/dev.json)
ck 'json格式化' "$(echo "$r1" | grep -q '"a": 1' && echo 1)" "$r1"

r2=$(curl -s -b $J -X POST -H 'Content-Type: application/json' -d '{"input":"https://example.com"}' $B/api/run/gen.qr)
ck '二维码' "$(echo "$r2" | grep -q 'data:image/png' && echo 1)" ''

curl -s -b $J -X POST $B/api/fav/dev.json > /dev/null
fv=$(curl -s -b $J $B/api/favs)
ck '收藏' "$(echo "$fv" | grep -q 'dev.json' && echo 1)" "$fv"

hi=$(curl -s -b $J $B/api/history)
ck '历史' "$(echo "$hi" | grep -q 'dev.json' && echo 1)" "$hi"

ai=$(curl -s -b $J -X POST -H 'Content-Type: application/json' -d '{"messages":[{"role":"user","content":"hi"}]}' $B/api/ai/chat)
ck 'AI演示模式' "$(echo "$ai" | grep -q '演示' && echo 1)" "$ai"

echo "─────────  PASS:$P  FAIL:$F  ─────────"
rm -f $J
