#!/bin/bash
# Nova 一键体检：./verify.sh [期望版本 如 2.7.1]，末行 ALL-GREEN=通过
B=http://127.0.0.1:8080; P=0; F=0; J=/tmp/nv-verify.jar
ok(){ echo "PASS  $1"; P=$((P+1)); }
no(){ echo "FAIL  $1  →  $2"; F=$((F+1)); }
g(){ if printf '%s' "$2" | grep -qF -- "$3"; then ok "$1"; else no "$1" "未命中 $3"; fi; }
gx(){ if printf '%s' "$2" | grep -qE -- "$3"; then ok "$1"; else no "$1" "未命中 $3"; fi; }
eq(){ if [ "$2" = "$3" ]; then ok "$1"; else no "$1" "got=[$2] want=[$3]"; fi; }
eq "root=200" "$(curl -s -o /dev/null -w '%{http_code}' $B/)" 200
curl -s -c $J -X POST $B/api/auth/login -d '{"username":"admin","password":"Admin123456"}' >/dev/null
eq "工具数=27" "$(curl -s -b $J $B/api/tools | grep -o '"id"' | wc -l | tr -d ' ')" 27
g "chmod 754" "$(curl -s -b $J -X POST $B/api/run/dev.chmod -d '{"input":"754","opts":{}}')" 'rwxr-xr--'
g "subnet /26=62台" "$(curl -s -b $J -X POST $B/api/run/dev.subnet -d '{"input":"192.168.1.130/26","opts":{}}')" '62 台'
g "yaml→json" "$(curl -s -b $J -X POST $B/api/run/dev.yaml -d '{"input":"a: 1","opts":{}}' | python3 -c 'import sys,json;print(json.load(sys.stdin).get("output",""))')" '"a": 1'
gx "cron 跳周末" "$(curl -s -b $J -X POST $B/api/run/dev.cron -d '{"input":"0 9 * * 1-5","opts":{}}')" 'Mon|Tue|Wed|Thu|Fri'
A=$(curl -s $B/assets/app.js)
eq "面板双join雷区" "$(printf '%s' "$A" | grep -c "join('').join('')")" 0
eq "面板pages join" "$(printf '%s' "$A" | grep -c "join('')+nrt.map(")" 1
eq "面板Enter兜底" "$(printf '%s' "$A" | grep -cF '||items[0];t&&t.click()')" 1
eq "独立页tokens" "$(curl -s $B/mini.html | grep -c 'nova-tokens.css')" 1
eq "word2pdf tokens" "$(curl -s $B/word2pdf.html | grep -c 'nova-tokens.css')" 1
eq "徽章规范" "$(printf '%s' "$A" | grep -c "info..:.plan")" 1
g "面板融合 NR_TOOLS" "$A" 'NR_TOOLS'
g "跨页深链" "$A" 'tools.html?tool='
g "/ 键呼出" "$A" '呼出'
[ -n "$1" ] && g "banner v$1" "$(grep -o 'Nova Tools v[0-9.]* 已启动' /home/work/fanguohui/nova8080.log | tail -1)" "Nova Tools v$1 已启动"
echo "────────"; echo "体检: $P PASS / $F FAIL"; [ "$F" = 0 ] && echo ALL-GREEN || exit 1
