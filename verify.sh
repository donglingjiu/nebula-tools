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
eq "徽章规范" "$(printf '%s' "$A" | grep -c "info...plan")" 1
eq "内联flex精确清零" "$(printf '%s' "$A" | grep -cF 'style="flex:1"')" 0
eq "内联余量棘轮" "$(printf '%s' "$A" | grep -o 'style="' | wc -l)" 111
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
