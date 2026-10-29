#!/usr/bin/env bash
# nebula-tools 本地开发回路 —— 仅依赖 go + python3 + curl/wget，不依赖物理机特有环境
set -e
PORT="${2:-8080}"
GET() { if command -v curl >/dev/null 2>&1; then curl -fsS "$1"; else wget -qO- "$1"; fi; }

case "${1:-help}" in
  web)   # 前端极速回路：静态服务 web/，改完刷新即生效（当前阶段主力，0编译）
     BUSY() { (exec 3<>"/dev/tcp/127.0.0.1/$1") 2>/dev/null && { exec 3>&-; return 0; } || return 1; }
     TRY="${PORT}"
     while BUSY "${TRY}"; do echo "[dev] 端口 ${TRY} 被占用(可能是你的Go服务), 换 $((TRY+1))"; TRY=$((TRY+1)); [ "${TRY}" -gt $((PORT+20)) ] && echo "[dev] 无空闲端口" && exit 1; done
     [ "${TRY}" != "${PORT}" ] && echo "[dev] 实际使用端口 ${TRY}"
     echo "[dev] 打开 http://127.0.0.1:${TRY}/ai-studio.html 或 http://服务器IP:${TRY}/ai-studio.html  (Ctrl+C 停止)"
     cd web && exec python3 -m http.server "${TRY}" --bind 0.0.0.0 ;;
  app)   # 全量应用：go run（含 go:embed 与后端），Go 代码改动需 Ctrl+C 重启
     echo "[dev] go run 启动中（首次编译数秒）"
     CGO_ENABLED=0 go run . ;;
  test)  # Go 单测: 新包 + 存量
     CGO_ENABLED=0 go test ./internal/... && echo TESTS_NEW_OK
     CGO_ENABLED=0 go test . 2>&1 | tail -5 ;;
  build) # 标准产物（等价 CI 构建参数）
     CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o nebula-tools . && echo BUILD_OK ;;
  smoke) # 冒烟：./dev.sh smoke http://127.0.0.1:8080
     B="${2:-http://127.0.0.1:8080}"; fail=0
     for u in / /ai-studio.html /login.html /register.html /profile.html /ai-chat.html /admin.html /admin-keys.html /admin-aiops.html /admin-features.html /assets/css/tokens.css /assets/css/admin.css /assets/js/core.js /assets/js/ai-studio.js /assets/js/ai-chat.js /assets/js/admin.js; do
       GET "${B}${u}" >/dev/null 2>&1 && echo "PASS 200 ${u}" || { echo "FAIL ${u}"; fail=1; }
     done
     GET "${B}/ai-studio.html" | grep -q "AI 工作台" && echo "PASS mark ai-studio" || { echo "FAIL mark ai-studio"; fail=1; }
     GET "${B}/assets/css/tokens.css" | grep -q "ai-gradient" && echo "PASS mark tokens" || { echo "FAIL mark tokens"; fail=1; }
     GET "${B}/assets/js/core.js" | grep -q "NB.stream" && echo "PASS mark core" || { echo "FAIL mark core"; fail=1; }
     GET "${B}/assets/js/ai-studio.js" | grep -q "toFixed(1)" && echo "PASS fresh ai-studio.js" || { echo "FAIL fresh: 该服务内嵌旧前端(重启Go服务即可)"; fail=1; }
     [ "$fail" = 0 ] && echo SMOKE_OK || echo SMOKE_FAIL
     exit "$fail" ;;
  stop)  pkill -f "python3 -m http.server" && echo stopped || echo "无进程" ;;
  check) # 每批必跑：CRLF归一 + 第72行类损伤自修 + 双编码/FFFD乱码扫描 + 严格UTF8 + 哨兵
     python3 - <<'PYCHK'
import glob, pathlib
files = ["web/assets/js/ai-studio.js", "web/assets/js/core.js", "web/assets/js/ai-chat.js", "web/assets/js/admin.js",
         "web/ai-studio.html", "web/login.html", "web/register.html", "web/profile.html",
         "web/ai-chat.html", "web/admin.html", "web/admin-keys.html", "web/admin-aiops.html",
         "web/admin-features.html", "web/index.html",
         "web/assets/css/tokens.css", "web/assets/css/admin.css", "todo.md"]
ok = True
for f in files:
    p = pathlib.Path(f)
    b = p.read_bytes().replace(b"\r\n", b"\n").replace(b"\r", b"\n")
    if f.endswith("ai-studio.js"):
        lines = b.split(b"\n")
        for i, ln in enumerate(lines):
            if b"stTime" in ln:
                j = i + 1
                if j < len(lines) and lines[j].strip() != b"}":
                    print("autofix line", j + 1, "-> '  }'")
                    lines[j] = b"  }"
                    b = b"\n".join(lines)
                break
    p.write_bytes(b)
    try:
        b.decode("utf-8"); u = "OK"
    except UnicodeDecodeError as e:
        u, ok = "FAIL " + str(e), False
    mo = (b"\xef\xbf\xbd" in b) or (b"\xc3\xaf\xc2\xbf\xc2\xbd" in b)
    if mo: ok = False
    print(f, "utf8:" + u, "mojibake:" + ("YES" if mo else "no"))
s = pathlib.Path("web/assets/js/ai-studio.js").read_text(encoding="utf-8")
ls = s.splitlines()
idx = next((i for i, l in enumerate(ls) if "stTime" in l), -1)
if idx < 0:
    ok = False; print("sentinel stTime: MISSING")
else:
    nxt = ls[idx + 1]
    print("sentinel stTime.next:", repr(nxt))
    if nxt.strip() != "}": ok = False
h = pathlib.Path("web/ai-studio.html").read_text(encoding="utf-8")
if '/ai-chat.html">AI' not in h:
    ok = False; print("sentinel nav: FAIL")
c = pathlib.Path("web/assets/js/core.js").read_text(encoding="utf-8")
for mark in ("NB.stream", "String.fromCharCode(96, 96, 96)", "showPaywall"):
    if mark not in c:
        ok = False; print("sentinel core missing:", mark)
for f, marks in [("web/login.html", ["NB.auth.login", "insEls"]),
                 ("web/register.html", ["NB.auth.register", "max-width: 400px", "margin-bottom: 14px"]),
                 ("web/profile.html", ["btnLogout", "device-width", "billing"]),
                 ("web/assets/js/ai-chat.js", ["NB.stream"]),
                 ("web/admin.html", ["admin-keys.html", "NB.adminGate"]),
                 ("web/admin-keys.html", ["masked_key", "f_key", "llm/profiles"]),
                 ("web/assets/js/admin.js", ["adminGate"]),
                 ("web/admin-aiops.html", ["sw_ai", "/ai/ops').then(function(r){", "sel_model"]),
                 ("web/admin-features.html", ["admin/features", "nb-switch", "calls_7d", "(f.price || 0)"]),
                 ("web/index.html", ["nb-hero-ai", '<a href="/ai-studio.html"', "立即体验", "font-weight:700;padding:9px 18px"])]:
    t = pathlib.Path(f).read_text(encoding="utf-8")
    for m in marks:
        if m not in t:
            ok = False; print("sentinel missing:", f, "::", m[:40])
for hf in sorted(glob.glob("web/*.html")):
    ht = pathlib.Path(hf).read_text(encoding="utf-8")
    if "</html>" not in ht:
        ok = False; print("sentinel missing:", hf, ":: </html>")
    if "<header" in ht and "</header>" not in ht:
        ok = False; print("sentinel missing:", hf, ":: </header>")
print("CHECK_OK" if ok else "CHECK_FAIL")
raise SystemExit(0 if ok else 1)
PYCHK
     ;;
  accept) # R1 前端验收（阶段门4自动证据）
     python3 - <<'PYACC'
import glob, pathlib, re
ok = True
def chk(name, cond, detail=''):
    global ok
    print(('PASS' if cond else 'FAIL'), name, ('| ' + detail) if detail else '')
    if not cond: ok = False

need = ['index.html','ai-studio.html','login.html','register.html','profile.html',
        'ai-chat.html','admin.html','admin-keys.html','admin-aiops.html','admin-features.html',
        'assets/css/tokens.css','assets/css/admin.css',
        'assets/js/core.js','assets/js/ai-studio.js','assets/js/ai-chat.js','assets/js/admin.js']
miss = [f for f in need if not pathlib.Path('web', f).exists()]
chk('A.inventory', not miss, '16/16' if not miss else 'missing: ' + ','.join(miss))

hrefs = set()
for f in glob.glob('web/*.html'):
    hrefs |= set(re.findall(r'href="(/[a-z0-9-]+\.html)"', pathlib.Path(f).read_text(encoding='utf-8')))
bad = [h for h in sorted(hrefs) if not pathlib.Path('web' + h).exists()]
chk('B.nav-closure', not bad, str(len(hrefs)) + ' links ok' if not bad else 'broken: ' + ','.join(bad))

core = pathlib.Path('web/assets/js/core.js').read_text(encoding='utf-8')
chk('C.error-codes', all(c in core for c in ['401', '402', '410', '429']))
chk('C.paywall-modal', 'showPaywall' in core)
admin_pages = glob.glob('web/admin-*.html') + ['web/admin.html']
chk('C.admin-gate', all('adminGate' in pathlib.Path(f).read_text(encoding='utf-8') for f in admin_pages), str(len(admin_pages)) + ' pages')

stu = pathlib.Path('web/assets/js/ai-studio.js').read_text(encoding='utf-8')
chk('D.export-fallback', 'clientMd' in stu and 'format=zip' in stu)
studio_html = pathlib.Path('web/ai-studio.html').read_text(encoding='utf-8')
chat_js = pathlib.Path('web/assets/js/ai-chat.js').read_text(encoding='utf-8')
chk('D.login-gates', 'loginGate' in studio_html and 'login.html?next=/ai-chat.html' in chat_js)

tok = pathlib.Path('web/assets/css/tokens.css').read_text(encoding='utf-8')
defined = set(re.findall(r'(--[a-z0-9-]+)\s*:', tok))
used = set()
for f in glob.glob('web/*.html') + glob.glob('web/assets/js/*.js') + ['web/assets/css/admin.css']:
    used |= set(re.findall(r'var\((--[a-z0-9-]+)\)', pathlib.Path(f).read_text(encoding='utf-8')))
undef = sorted(used - defined)
chk('E.tokens-coverage', not undef, str(len(used)) + ' vars defined' if not undef else 'undefined: ' + ','.join(undef))
chk('E.focus-visible', 'focus-visible' in tok)

alljs = '\n'.join(pathlib.Path(f).read_text(encoding='utf-8') for f in glob.glob('web/assets/js/*.js'))
chk('F.storage-keys', all(k in alljs for k in ['nb_token', 'nb_flow', 'nb_chat']))

apis = set()
for f in glob.glob('web/assets/js/*.js') + glob.glob('web/*.html'):
    for x in re.findall(r'/api/[a-z0-9_/{}\-]+', pathlib.Path(f).read_text(encoding='utf-8')):
        apis.add(x.rstrip('/'))
print('INFO api-endpoints (' + str(len(apis)) + '):')
for a in sorted(apis):
    print('   ', a)
chk('G.contract-surface', len(apis) >= 8)

print('ACCEPT_OK' if ok else 'ACCEPT_FAIL')
raise SystemExit(0 if ok else 1)
PYACC
     ;;  *) echo "用法: ./dev.sh web|app|build|test|smoke|stop|check|accept [port|base]" ;;

esac
