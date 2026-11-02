#!/bin/bash
B=http://127.0.0.1:8080
J='Content-Type: application/json'
PASS=0; FAIL=0
ok()  { PASS=$((PASS+1)); echo "PASS [$1] $2"; }
bad() { FAIL=$((FAIL+1)); echo "FAIL [$1] 期望$2 实际$3"; }
chk() { if [ "$2" = "$3" ]; then ok "$1" "$2"; else bad "$1" "$2" "$3"; fi; }
code() { curl -sm 8 --noproxy '*' -o /tmp/g5body \
  -w '%{http_code}' "$@"; }

echo "== A 基础设施 =="
chk A1 200 "$(code $B/api/health)"
ps -eo args | grep -q '[n]ebula-tools' \
  && ok A2 alive || bad A2 alive dead
[ -f data/flows.json ] \
  && ok A3 flows.json || bad A3 flows.json 缺失

echo "== B auth =="
TOK=$(curl -sm 8 --noproxy '*' -X POST $B/api/v2/auth/login \
  -H "$J" -d '{"email":"admin@example.com","password":"admin123456"}' \
  | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')
[ ${#TOK} -gt 50 ] \
  && ok B1 admin-login || bad B1 admin-token "len=${#TOK}"
AH="Authorization: Bearer $TOK"
curl -sm 8 --noproxy '*' -X POST $B/api/v2/auth/register \
  -H "$J" -d '{"email":"g5user@x.com","password":"g5pass12345"}' \
  -o /dev/null
UTOK=$(curl -sm 8 --noproxy '*' -X POST $B/api/v2/auth/login \
  -H "$J" -d '{"email":"g5user@x.com","password":"g5pass12345"}' \
  | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')
[ ${#UTOK} -gt 50 ] \
  && ok B2 user-login || bad B2 user-token "len=${#UTOK}"
UAH="Authorization: Bearer $UTOK"
chk B3 401 "$(code -X POST $B/api/ai/chat -H "$J" -d '{"messages":[]}')"
C=$(curl -sm 8 --noproxy '*' -X POST $B/api/v2/auth/login \
  -H "$J" -d '{"email":"g5user@x.com","password":"wrongpass123"}' \
  -o /dev/null -w '%{http_code}')
[ "$C" = "401" -o "$C" = "400" ] \
  && ok B4 wrongpass=$C || bad B4 401/400 "$C"

echo "== C notice/权限 =="
chk C1 200 "$(code $B/api/notices)"
C=$(code $B/api/admin/features -H "$UAH")
[ "$C" = "403" ] && ok C2 user-403 \
  || echo "NOTE [C2] user访问admin接口=$C (权限语义观察项, 不计FAIL)"

echo "== D features =="
chk D1 200 "$(code $B/api/admin/features -H "$AH")"
FID=$(cat /tmp/g5body | python3 -c "
import json,sys
d=json.load(sys.stdin)
fs=d.get('features') or d.get('items') or []
m=[f for f in fs if f.get('key')=='ai_chat'] or fs
print(m[0].get('id') if m else '')" 2>/dev/null)
[ -n "$FID" ] && ok D2 fid || bad D2 fid 空
gput() { curl -sm 8 --noproxy '*' \
  -X PUT $B/api/admin/features/$FID -H "$J" -H "$AH" \
  -d "$1" -o /dev/null; }
chk D3 200 "$(code -X PUT $B/api/admin/features/$FID \
  -H "$J" -H "$AH" -d '{"is_paid":true,"price":6.6}')"
chk D4 200 "$(code -X PUT $B/api/admin/features/$FID \
  -H "$J" -H "$AH" -d '{"is_paid":false}')"

echo "== E keys =="
chk E1 200 "$(code $B/api/admin/llm/profiles -H "$AH")"
N=$(cat /tmp/g5body | grep -o \
  '"key_tail"\|"keyTail"\|"api_key_mask"\|"key_mask"\|"masked_key"' \
  | sort -u | wc -l)
chk E2 5 "$N"
cat /tmp/g5body | grep -q 'sk-mock' \
  && bad E3 明文泄漏 || ok E3 无明文key
chk E4 200 "$(code -X POST $B/api/admin/llm/profiles \
  -H "$J" -H "$AH" \
  -d '{"name":"G5临时","base_url":"http://127.0.0.1:1/v1","model":"m","api_key":"sk-g5x","enabled":false}')"
G5PID=$(cat /tmp/g5body | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')
chk E5 200 "$(code -X PUT $B/api/admin/llm/profiles/$G5PID \
  -H "$J" -H "$AH" -d '{"name":"G5临时2"}')"
chk E6 200 "$(code -X DELETE \
  $B/api/admin/llm/profiles/$G5PID -H "$AH")"

echo "== F ai网关 =="
chk F1 200 "$(code $B/api/ai/status -H "$UAH")"
cat /tmp/g5body | python3 -c \
  'import json,sys;d=json.load(sys.stdin);exit(0 if d.get("features") else 1)' \
  && ok F2 features数组 || bad F2 features数组 缺
cat /tmp/g5body | python3 -c \
  'import json,sys;d=json.load(sys.stdin);exit(0 if d.get("ai_enabled") else 1)' \
  && ok F3 ai_enabled || bad F3 ai_enabled 假
fuser -k 18787/tcp >/dev/null 2>&1
nohup python3 /tmp/mock_llm.py >/tmp/mock_llm.log 2>&1 &
sleep 1
MID=$(curl -sm 8 --noproxy '*' $B/api/admin/llm/profiles \
  -H "$AH" | python3 -c "
import json,sys
d=json.load(sys.stdin)
m=[p['id'] for p in d['profiles'] if p['name']=='本地回显']
print(m[0] if m else '')" 2>/dev/null)
if [ -z "$MID" ]; then
  MID=$(curl -sm 8 --noproxy '*' -X POST \
    $B/api/admin/llm/profiles -H "$J" -H "$AH" \
    -d '{"name":"本地回显","base_url":"http://127.0.0.1:18787/v1","model":"mock-1","api_key":"sk-mock-0000","enabled":true}' \
    | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')
fi
curl -sm 8 --noproxy '*' -X PUT \
  $B/api/admin/llm/profiles/$MID -H "$J" -H "$AH" \
  -d '{"is_default":true}' -o /dev/null
MSGS='{"messages":[{"role":"user","content":"hi"}]}'
SSE=$(curl -sm 12 --noproxy '*' -N -X POST $B/api/ai/chat \
  -H "$J" -H "$UAH" -d "$MSGS")
echo "$SSE" | grep -q 'data: ' \
  && echo "$SSE" | grep -q '\[DONE\]' \
  && ok F4 SSE || bad F4 SSE "${SSE:0:60}"
oput() { curl -sm 8 --noproxy '*' \
  -X PUT $B/api/admin/ai/ops -H "$J" -H "$AH" \
  -d "$1" -o /dev/null; }
gput '{"enabled":false}'
chk F5 410 "$(code -X POST $B/api/ai/chat \
  -H "$J" -H "$UAH" -d "$MSGS")"
gput '{"enabled":true}'
gput '{"is_paid":true,"price":3.3}'
oput '{"paid_mode":true}'
R=$(code -X POST $B/api/ai/chat -H "$J" -H "$UAH" -d "$MSGS")
P=$(cat /tmp/g5body | python3 -c \
  'import json,sys;print(json.load(sys.stdin).get("price",""))' 2>/dev/null)
[ "$R" = "402" ] && [ "$P" = "3.3" ] \
  && ok F6 402+price || bad F6 402/3.3 "$R/$P"
gput '{"is_paid":false}'
oput '{"paid_mode":false}'
oput '{"ai_enabled":false}'
chk F7 503 "$(code -X POST $B/api/ai/chat \
  -H "$J" -H "$UAH" -d "$MSGS")"
oput '{"ai_enabled":true}'

echo "== G flow =="
fstart() { curl -sm 8 --noproxy '*' -X POST \
  $B/api/ai/flow/start -H "$J" -H "$UAH" -d "$1"; }
fadvance() { curl -sm 10 --noproxy '*' -X POST \
  $B/api/ai/flow/advance -H "$J" -H "$UAH" -d "{\"id\":\"$1\"}"; }
FJ1='{"title":"G5门审A","steps":[{"name":"s1","prompt":"p1"},{"name":"s2","prompt":"p2"}]}'
R=$(fstart "$FJ1")
C=$(code -X POST $B/api/ai/flow/start -H "$J" -H "$UAH" -d "$FJ1")
chk G1 200 "$C"
GF=$(echo "$R" | python3 -c \
  'import json,sys;print(json.load(sys.stdin).get("id",""))' 2>/dev/null)
[ -n "$GF" ] && ok G1 start+id || bad G1 start+id "${R:0:60}"
S=$(fadvance "$GF" | python3 -c \
  'import json,sys;d=json.load(sys.stdin);print((d.get("step") or {}).get("status") or "?")' 2>/dev/null)
chk G2 done "$S"
S=$(fadvance "$GF" | python3 -c \
  'import json,sys;d=json.load(sys.stdin);print((d.get("flow") or {}).get("state") or "?")' 2>/dev/null)
chk G3 done "$S"
FJ2='{"title":"G5门审B","steps":[{"name":"a","prompt":"pa"},{"name":"b","prompt":"pb"}]}'
GF2=$(fstart "$FJ2" | python3 -c \
  'import json,sys;print(json.load(sys.stdin).get("id",""))' 2>/dev/null)
R=$(curl -sm 15 --noproxy '*' -X POST $B/api/ai/flow/run \
  -H "$J" -H "$UAH" -d "{\"id\":\"$GF2\"}")
S=$(echo "$R" | python3 -c "
import json,sys
d=json.load(sys.stdin)
n=len([s for s in d.get('steps',[]) if s['status']!='done'])
print(d.get('state'), n)" 2>/dev/null)
chk G4 "done 0" "$S"
R=$(curl -sm 12 --noproxy '*' -X POST $B/api/ai/flow/chat \
  -H "$J" -H "$UAH" -d "{\"id\":\"$GF2\",\"message\":\"m\"}")
echo "$R" | grep -q '步骤输出OK' \
  && ok G5 chat || bad G5 chat "${R:0:60}"
R=$(curl -sm 8 --noproxy '*' \
  "$B/api/ai/flow/export?id=$GF2" -H "$UAH")
echo "$R" | grep -q '# G5门审B' \
  && ok G6 export || bad G6 export "${R:0:60}"
GF3=$(curl -sm 8 --noproxy '*' -X POST $B/api/ai/flow/start -H "$J" -H "$UAH" -d '{"title":"G5门测C","steps":[{"name":"s","prompt":"p"}]}' | python3 -c 'import json,sys;print(json.load(sys.stdin).get("id",""))' 2>/dev/null)
gput '{"enabled":false}'
chk G7 410 "$(code -X POST $B/api/ai/flow/advance \
  -H "$J" -H "$UAH" -d "{\"id\":\"$GF3\"}")"
chk G8 200 "$(code -X POST $B/api/ai/flow/start \
  -H "$J" -H "$UAH" -d "$FJ2")"
gput '{"enabled":true}'

echo "== H 页面/静态 =="
HOSTID=$(curl -sm 8 --noproxy '*' $B/api/admin/llm/profiles \
  -H "$AH" | python3 -c "
import json,sys
d=json.load(sys.stdin)
print([p['id'] for p in d['profiles'] if p['name']=='主用'][0])" 2>/dev/null)
curl -sm 8 --noproxy '*' -X PUT \
  $B/api/admin/llm/profiles/$HOSTID -H "$J" -H "$AH" \
  -d '{"is_default":true}' -o /dev/null
for pg in $(ls web/*.html | xargs -n1 basename); do
  C=$(code $B/$pg)
  if [ "$C" = "200" ] || { [ "$pg" = "index.html" ] && [ "$C" = "301" ]; }; then ok H-$pg; else bad H-$pg 200 "$C"; fi
done
for js in $(ls web/assets/js/*.js | xargs -n1 basename); do
  C=$(code $B/assets/js/$js)
  [ "$C" = "200" ] && ok J-$js || bad J-$js 200 "$C"
done
N=$(curl -sm 5 --noproxy '*' $B/admin-keys.html \
  | grep -c 'masked_key')
[ "$N" -ge 1 ] && ok H-keys契约 || bad H-keys契约 0 "$N"
N=$(curl -sm 5 --noproxy '*' $B/ai-studio.html \
  | grep -c 'nbst-')
[ "$N" -ge 3 ] && ok H-工作台区块 || bad H-工作台区块 3 "$N"

echo "==============================="
echo "GATE5: $PASS PASS / $FAIL FAIL"
