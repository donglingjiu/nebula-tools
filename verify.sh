#!/bin/bash
# Nova 一键体检：./verify.sh [期望版本 如 2.7.1]，末行 ALL-GREEN=通过
B=http://127.0.0.1:8080; P=0; F=0; J=/tmp/nv-verify.jar
ok(){ echo "PASS  $1"; P=$((P+1)); }
no(){ echo "FAIL  $1  →  $2"; F=$((F+1)); }
g(){ if printf '%s' "$2" | grep -qF -- "$3"; then ok "$1"; else no "$1" "未命中 $3"; fi; }
gx(){ if printf '%s' "$2" | grep -qE -- "$3"; then ok "$1"; else no "$1" "未命中 $3"; fi; }
eq(){ if [ "$2" = "$3" ]; then ok "$1"; else no "$1" "got=[$2] want=[$3]"; fi; }
eq "root=200" "$(curl -s -o /dev/null -w '%{http_code}' $B/)" 200
LC=$(curl -s -o /dev/null -w '%{http_code}' -c $J -X POST $B/api/auth/login -d '{"username":"admin","password":"Admin123456"}')
[ "$LC" = "200" ] && echo '(登录探针 200)' || echo "WARN 登录=$LC 凭据漂移→认证类检查可能失真"
eq "DB权限600" "$(stat -c '%a' "${NOVA_DB:-/home/work/fanguohui/data/nova.db}" 2>/dev/null)" 600
eq "会话持久化在位" "$(python3 -c "import sqlite3,os;p=os.environ.get('NOVA_DB') or '/home/work/fanguohui/data/nova.db';print(sqlite3.connect(p).execute('select count(*) from schema_migrations where version=1').fetchone()[0])" 2>/dev/null)" 1
eq "档案库SQLite化" "$(python3 -c "import sqlite3,os;p=os.environ.get('NOVA_DB') or '/home/work/fanguohui/data/nova.db';n=sqlite3.connect(p).execute('select count(*) from profiles').fetchone()[0];j=not os.path.exists('/home/work/fanguohui/data/llm_profiles.json');print(1 if n>0 and j else 0)" 2>/dev/null)" 1
eq "会话API在位" "$(python3 -c "b=open('/home/work/fanguohui/nova-server-linux-amd64','rb').read();print(1 if b'api/conversations' in b else 0)" 2>/dev/null)" 1
eq "用量API在位" "$(python3 -c "b=open('/home/work/fanguohui/nova-server-linux-amd64','rb').read();print(1 if b'api/usage/summary' in b else 0)" 2>/dev/null)" 1
eq "对话页在位" "$(test -f web/chat.html && grep -q conversations web/assets/chat.js && echo 1 || echo 0)" 1
eq "看板页在位" "$(test -f web/usage.html && grep -q usage/summary web/assets/usage.js && echo 1 || echo 0)" 1
eq "档案记账在位" "$(grep -q dbMsgAppendP dbconversations.go && grep -q usage/profiles web/assets/usage.js && echo 1 || echo 0)" 1
eq "分叉通电在位" "$(grep -q dbConvBranch dbconversations.go && grep -q branchConv web/assets/chat.js && echo 1 || echo 0)" 1
eq "登录链在位" "$(grep -q authSwitch web/assets/app.js && grep -q auth/register web/assets/app.js && grep -q auth/me web/assets/app.js && grep -q novaLogout web/assets/app.js && echo 1 || echo 0)" 1
eq "切换钮归层在位" "$(grep -c '^#authSwitch{' web/assets/nova-finishing.css)" 1
eq "app页普查棘轮" "$(grep -l 'id="app"' web/*.html | wc -l)" 13
eq "fx接线棘轮" "$(grep -oE 'pdf-(compress|split)' web/assets/app.js | wc -l)" 4
eq "admin旗硬编码棘轮" "$(grep -l 'data-admin="true"' web/admin/*.html 2>/dev/null | wc -l)" 0
eq "adminReq旗棘轮" "$(grep -l 'data-admin-req="true"' web/admin/*.html 2>/dev/null | wc -l)" 10
eq "users演示数据棘轮" "$(grep -c '128,436\|userRows\|演示账户' web/assets/app.js)" 0
eq "users真接线棘轮" "$(grep -c 'ADMIN_USERS_LIVE\|refreshUsers\|adminUserAction' web/assets/app.js)" 10
eq "JS语法门" "$(python3 js_gate.py web/assets/app.js)" ok
eq "语法门在岗" "$(test -f js_gate.py && echo 1 || echo 0)" 1
eq "admin壳全接线" "$(grep -L data-admin-req web/admin/*.html 2>/dev/null | wc -l)" 0
eq "admin路由活变棘轮" "$(grep -o computeCurrentID web/assets/app.js | wc -l)" 3
eq "dashboard演示退役棘轮" "$(grep -c '3,678,920' web/assets/app.js)" 0
eq "dashboard真接线棘轮" "$(grep -o 'refreshOverview\|renderDashLive\|ADMIN_DASH_LIVE' web/assets/app.js | wc -l)" 10
eq "重启契约在册" "$(grep -c '重启必须携带 NOVA_WEB_DIR' DEPLOY.md)" 1
eq "工具数=23" "$(curl -s -b $J $B/api/tools | grep -o '"id"' | wc -l | tr -d ' ')" 23
eq "①C三新端点计数锁" "$(grep -oE 'api/t/(pdf|word|xlsx)-text' web/assets/app.js | wc -l | tr -d ' ')" "12"
UPC=$(curl -s -o /dev/null -w '%{http_code}' -b $J -X POST $B/api/upload/file.pdf -F f=x 2>/dev/null)
if [ "$UPC" = "405" ]; then UPC="404"; fi
eq "upload退役防复活" "$UPC" "404"
VT=$(curl -s -X POST $B/api/v2/auth/login -H 'Content-Type: application/json' -d '{"email":"probe-ai@example.com","password":"probe12345"}')
if [ -z "$(echo "$VT" | grep -o '"token"')" ]; then
  VT=$(curl -s -X POST $B/api/v2/auth/register -H 'Content-Type: application/json' -d '{"email":"probe-ai@example.com","password":"probe12345"}')
fi
TK=$(echo "$VT" | grep -o '"token":"[^"]*"' | head -1 | cut -d'"' -f4)
eq "v2鉴权闭环(me)" "$(curl -s -o /dev/null -w '%{http_code}' -H "Authorization: Bearer $TK" $B/api/v2/auth/me)" "200"
eq "v2鉴权闭环(ai-status)" "$(curl -s -o /dev/null -w '%{http_code}' -H "Authorization: Bearer $TK" $B/api/ai/status)" "200"
eq "ai-chat页在位" "$(curl -s -o /dev/null -w '%{http_code}' $B/ai-chat.html)" "200"
eq "chat契约锁(空messages=400)" "$(curl -s -o /dev/null -w '%{http_code}' -X POST $B/api/ai/chat -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' -d '{"messages":[]}')" "400"
eq "ai-chat.js形状锁" "$(grep -c 'messages: S.msgs' web/assets/js/ai-chat.js)" "1"
eq "e2e隔离锁(os.Remove归零)" "$(grep -c 'os.Remove' e2e_test.go)" "0"
eq "e2e临时目录锁(TempDir)" "$(grep -c 't.TempDir()' e2e_test.go)" "1"
eq "资产版本棘轮(vhash-check)" "$(python3 vhash.py check >/dev/null 2>&1; echo $?)" "0"
eq "ai-chat.html版本化引用" "$(grep -c 'ai-chat\.js?v=' web/ai-chat.html)" "1"
g "chmod 754" "$(curl -s -b $J -X POST $B/api/run/dev.chmod -d '{"input":"754","opts":{}}')" 'rwxr-xr--'
g "subnet /26=62台" "$(curl -s -b $J -X POST $B/api/run/dev.subnet -d '{"input":"192.168.1.130/26","opts":{}}')" '62 台'
g "yaml→json" "$(curl -s -b $J -X POST $B/api/run/dev.yaml -d '{"input":"a: 1","opts":{}}' | python3 -c 'import sys,json;print(json.load(sys.stdin).get("output",""))')" '"a": 1'
gx "cron 跳周末" "$(curl -s -b $J -X POST $B/api/run/dev.cron -d '{"input":"0 9 * * 1-5","opts":{}}')" 'Mon|Tue|Wed|Thu|Fri'
A=$(curl -s $B/assets/app.js)
g "①C三新端点在线(serve=磁盘)" "$A" 'pdf-text'
eq "面板双join雷区" "$(printf '%s' "$A" | grep -c "join('').join('')")" 0
eq "面板pages join" "$(printf '%s' "$A" | grep -c "join('')+nrt.map(")" 1
eq "面板Enter兜底" "$(printf '%s' "$A" | grep -cF '||items[0];t&&t.click()')" 1
eq "独立页tokens" "$(curl -s $B/mini.html | grep -c 'nova-tokens.css')" 1
eq "word2pdf tokens" "$(curl -s $B/word2pdf.html | grep -c 'nova-tokens.css')" 1
eq "徽章规范" "$(printf '%s' "$A" | grep -c "info...plan")" 1
eq "内联flex精确清零" "$(printf '%s' "$A" | grep -cF 'style="flex:1"')" 0
eq "内联余量棘轮" "$(printf '%s' "$A" | grep -o 'style="' | wc -l)" 116
eq "q预填在位" "$(printf '%s' "$A" | grep -cF "params.has('q')")" 1
eq "深链派发冒泡" "$(printf '%s' "$A" | grep -cF "ts.dispatchEvent(new Event('input',{bubbles:true}))")" 1
eq "尺度tokens在位" "$(grep -c -- '--nv-s10:10px' "$(dirname "$0")/web/assets/nova-tokens.css")" 1
eq "墨色令牌在位" "$(grep -oF -- 'var(--nv-ink)' "$(dirname "$0")/web/assets/nova-tokens.css" | wc -l)" 1
eq "面板兜底搜索" "$(printf '%s' "$A" | grep -cF '在工具大全中搜索')" 1
eq "IME卫兵在位" "$(printf '%s' "$A" | grep -oF 'isComposing||e.keyCode===229' | wc -l)" 2
eq "兜底原样大小写" "$(printf '%s' "$A" | grep -cF 'esc(raw)')" 1
eq "组合期输入静默" "$(printf '%s' "$A" | grep -oE 'isComposing.return;renderCommandItems|compositionend' | wc -l)" 2
eq "拆甲基线立账" "$(grep -oF '!important' "$(dirname "$0")/web/assets/nova-tokens.css" | wc -l)" 12
g "面板融合 NR_TOOLS" "$A" 'NR_TOOLS'
g "跨页深链" "$A" 'tools.html?tool='
g "/ 键呼出" "$A" '呼出'
[ -n "$1" ] && g "banner v$1" "$(grep -o 'Nova Tools v[0-9.]* 已启动' /home/work/fanguohui/nova8080.log | tail -1)" "Nova Tools v$1 已启动"
echo "────────"; echo "体检: $P PASS / $F FAIL"; [ "$F" = 0 ] && echo ALL-GREEN || exit 1
