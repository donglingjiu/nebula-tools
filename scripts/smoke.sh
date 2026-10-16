#!/usr/bin/env bash
# 用法: bash scripts/smoke.sh [base_url]   默认 http://localhost:8080
set -e
BASE=${1:-http://localhost:8080}
pass=0; fail=0
ck() { if [ "$1" = "$2" ]; then echo "  ✔ $3"; pass=$((pass+1)); else echo "  ✘ $3 (期望 $2, 实得 $1)"; fail=$((fail+1)); fi }
echo "== 冒烟测试 $BASE =="
ck "$(curl -sf $BASE/api/health | grep -o '"ok":true' || echo x)" '"ok":true' "健康检查"
ck "$(curl -sf $BASE/api/about | grep -o '"version"' | head -1)" '"version"' "版本标识"
ck "$(curl -sf $BASE/api/modules | grep -o 'dev.json' | head -1)" 'dev.json' "模块列表"
R=$(curl -sf -X POST $BASE/api/dev/json/format -H 'Content-Type: application/json' -d '{"json":"{\"b\":1,\"a\":2}"}')
ck "$(echo "$R" | grep -o '"a"' | head -1)" '"a"' "JSON 格式化"
T=$(curl -sf -X POST $BASE/api/auth/login -H 'Content-Type: application/json' -d '{"username":"admin","password":"admin123"}' | grep -o '"token"' | head -1)
ck "$T" '"token"' "管理员登录"
ck "$(curl -s -o /dev/null -w '%{http_code}' -X POST $BASE/api/admin/modules -d '{}')" "403" "未登录访问后台=403"
echo "== 结果: $pass 通过 / $fail 失败 =="
[ $fail -eq 0 ]
