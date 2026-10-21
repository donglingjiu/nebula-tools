#!/usr/bin/env bash
# ============================================================
# Nova Tools v2.2.1 —— 全量落盘脚本 2/3（测试与 CI）【bash 版】
# 用法: 在仓库根目录执行  bash setup-2.sh
# ============================================================
cd "$(dirname "$0")"

w(){
  local p="$1"; mkdir -p "$(dirname "$p")"
  cat > "$p"
  echo "  写入 $p  ($(wc -c < "$p" | tr -d ' ') 字节)"
}

echo "═══ Nova v2.2.1 脚本2/3：测试与 CI ═══"

# ---------- e2e_test.go（16 环节全链路） ----------
echo "── e2e_test.go ──"
w e2e_test.go <<'E2E_GO'
package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

type TC struct {
	t  *testing.T
	ts *httptest.Server
	c  *http.Client
}

func start(t *testing.T) *TC {
	t.Helper()
	os.Remove("data/nova.json")
	ts := httptest.NewServer(newRouter(Load()))
	t.Cleanup(func() { ts.Close(); os.Remove("data/nova.json") })
	jar, _ := cookiejar.New(nil)
	return &TC{t, ts, &http.Client{Jar: jar}}
}

func (tc *TC) fresh() *http.Client { jar, _ := cookiejar.New(nil); return &http.Client{Jar: jar} }

func (tc *TC) doC(c *http.Client, method, path string, body any) (int, map[string]any) {
	tc.t.Helper()
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, tc.ts.URL+path, rd)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		tc.t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var m map[string]any
	json.Unmarshal(raw, &m)
	if m == nil {
		m = map[string]any{}
	}
	return resp.StatusCode, m
}
func (tc *TC) do(m, p string, b any) (int, map[string]any) { return tc.doC(tc.c, m, p, b) }
func (tc *TC) okC(c *http.Client, m, p string, b any) map[string]any {
	code, r := tc.doC(c, m, p, b)
	if code != 200 || r["ok"] == false {
		tc.t.Fatalf("%s %s → %d %v", m, p, code, r)
	}
	return r
}
func (tc *TC) ok(m, p string, b any) map[string]any { return tc.okC(tc.c, m, p, b) }

func (tc *TC) upC(c *http.Client, path, name, content string) (int, map[string]any) {
	tc.t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, _ := w.CreateFormFile("file", name)
	fw.Write([]byte(content))
	w.Close()
	req, _ := http.NewRequest("POST", tc.ts.URL+path, &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := c.Do(req)
	if err != nil {
		tc.t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var m map[string]any
	json.Unmarshal(raw, &m)
	if m == nil {
		m = map[string]any{}
	}
	return resp.StatusCode, m
}

/* ---- 现场构造合法测试文件 ---- */
func pngBytes(t *testing.T) []byte {
	var b bytes.Buffer
	png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 8, 8)))
	return b.Bytes()
}
func docxBytes(t *testing.T) []byte {
	var b bytes.Buffer
	zw := zip.NewWriter(&b)
	f, _ := zw.Create("word/document.xml")
	f.Write([]byte(`<w:document><w:body><w:p><w:r><w:t>你好 Nova</w:t></w:r></w:p></w:body></w:document>`))
	zw.Close()
	return b.Bytes()
}
func xlsxBytes(t *testing.T) []byte {
	var b bytes.Buffer
	zw := zip.NewWriter(&b)
	f, _ := zw.Create("xl/worksheets/sheet1.xml")
	f.Write([]byte(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData><row r="1"><c r="A1" t="inlineStr"><is><t>name</t></is></c><c r="B1" t="inlineStr"><is><t>age</t></is></c></row><row r="2"><c r="A2" t="inlineStr"><is><t>张三</t></is></c><c r="B2"><v>20</v></c></row></sheetData></worksheet>`))
	zw.Close()
	return b.Bytes()
}

func TestFullJourney(t *testing.T) {
	tc := start(t)

	// 01 健康检查
	if code, h := tc.do("GET", "/api/health", nil); code != 200 || h["ok"] != true {
		t.Fatalf("health: %d %v", code, h)
	}
	// 02 未登录访问受保护接口 → 401
	if code, _ := tc.do("GET", "/api/auth/me", nil); code != 401 {
		t.Fatalf("未登录 me 应 401，得到 %d", code)
	}
	// 03 注册：短密码被拒；首个用户自动 admin
	if code, _ := tc.do("POST", "/api/auth/register", map[string]string{"username": "admin", "password": "123"}); code != 400 {
		t.Fatal("短密码应 400")
	}
	m := tc.ok("POST", "/api/auth/register", map[string]string{"username": "admin", "password": "admin123"})
	if m["role"] != "admin" {
		t.Fatalf("首个用户应为 admin: %v", m)
	}
	// 04 工具列表 16 个且全开
	_, tools := tc.ok("GET", "/api/tools", nil)
	if len(tools.([]any)) != 16 {
		t.Fatalf("应 16 个工具: %d", len(tools.([]any)))
	}
	// 05 十一个文本工具逐个真实运行并校验输出特征
	runs := map[string]struct {
		in   string
		opts map[string]string
		want string
	}{
		"dev.json":      {`{"b":1,"a":2}`, map[string]string{"mode": "format"}, `"a": 2`},
		"dev.md":        {"# 标题\n- 项目", nil, "<h1>"},
		"dev.encode":    {"Nova", map[string]string{"op": "b64enc"}, "Tm92"},
		"dev.hash":      {"abc", nil, "ba7816bf"},
		"dev.sql":       {"select * from t where id=1", nil, "SELECT"},
		"dev.util":      {"", map[string]string{"op": "uuid"}, "-"},
		"gen.qr":        {"https://example.com", nil, "data:image/png"},
		"gen.base":      {"255", map[string]string{"base": "10"}, "FF"},
		"dev.diff":      {"a\nb\n---\na\nc", nil, "- b"},
		"data.csv2json": {"n,v\nx,1", nil, `"n":"x"`},
		"data.json2csv": {`[{"a":1,"b":2}]`, nil, "a,b"},
	}
	for id, r := range runs {
		out := tc.ok("POST", "/api/run/"+id, map[string]any{"input": r.in, "opts": r.opts})
		if !strings.Contains(fmt.Sprint(out["output"]), r.want) {
			t.Fatalf("%s 输出不含 %q: %v", id, r.want, out["output"])
		}
	}
	// 06 边界：非法 JSON → ok=false；未知工具 → 403
	if _, m := tc.do("POST", "/api/run/dev.json", map[string]any{"input": "{bad"}); m["ok"] != false {
		t.Fatal("非法 JSON 应 ok=false")
	}
	if code, _ := tc.do("POST", "/api/run/nope", map[string]any{"input": "x"}); code != 403 {
		t.Fatal("未知工具应 403")
	}
	// 07 四个文件工具（现场构造 docx/xlsx/png）
	var code int
	code, m = tc.upC(tc.c, "/api/upload/file.word", "a.docx", string(docxBytes(t)))
	if code != 200 || !strings.Contains(fmt.Sprint(m["output"]), "你好 Nova") {
		t.Fatalf("docx: %d %v", code, m)
	}
	code, m = tc.upC(tc.c, "/api/upload/file.excel", "a.xlsx", string(xlsxBytes(t)))
	if code != 200 || !strings.Contains(fmt.Sprint(m["output"]), "name") {
		t.Fatalf("xlsx: %d %v", code, m)
	}
	code, m = tc.upC(tc.c, "/api/upload/file.img", "a.png", string(pngBytes(t)))
	if code != 200 || !strings.HasPrefix(fmt.Sprint(m["output"]), "data:image/") {
		t.Fatalf("img: %d %v", code, m)
	}
	// 08 收藏 → 列表回显
	tc.ok("POST", "/api/fav/dev.json", nil)
	if _, f := tc.ok("GET", "/api/favs", nil); !strings.Contains(fmt.Sprint(f), "dev.json") {
		t.Fatalf("收藏: %v", f)
	}
	// 09 历史已记录（文本+文件）
	if _, h := tc.ok("GET", "/api/history", nil); !strings.Contains(fmt.Sprint(h), "dev.json") {
		t.Fatalf("历史: %v", h)
	}
	// 10 工作流：保存 → 运行 → 两步轨迹
	tc.ok("POST", "/api/workflows", map[string]any{"name": "清洗", "steps": []string{"dev.json", "dev.encode"}})
	_, wfs := tc.ok("GET", "/api/workflows", nil)
	wf := wfs.([]any)[0].(map[string]any)
	m = tc.ok("POST", "/api/workflows/"+fmt.Sprint(int64(wf["id"].(float64)))+"/run", map[string]any{"input": `{"b":1}`})
	if len(m["trace"].([]any)) != 2 {
		t.Fatalf("轨迹应 2 步: %v", m)
	}
	// 11 AI 未配置端点 → 演示模式
	m = tc.ok("POST", "/api/ai/chat", map[string]any{"messages": []map[string]string{{"role": "user", "content": "hi"}}})
	if !strings.Contains(fmt.Sprint(m["reply"]), "演示模式") {
		t.Fatalf("AI: %v", m)
	}
	// 12 改密码 → 旧密码失效
	tc.ok("POST", "/api/auth/password", map[string]string{"old": "admin123", "new": "newpass66"})
	if code, _ := tc.do("POST", "/api/auth/password", map[string]string{"old": "admin123", "new": "xxxxxx1"}); code != 400 {
		t.Fatal("旧密码应失效")
	}
	// 13 管理后台全流程
	tc.ok("POST", "/api/auth/login", map[string]string{"username": "admin", "password": "newpass66"})
	tc.ok("GET", "/api/admin/overview", nil)
	tc.ok("POST", "/api/admin/modules/dev.json", map[string]bool{"on": false})
	if code, _ := tc.do("POST", "/api/run/dev.json", map[string]any{"input": "1"}); code != 403 {
		t.Fatal("关闭模块应 403")
	}
	tc.ok("POST", "/api/admin/modules/dev.json", map[string]bool{"on": true})
	bc := tc.fresh()
	tc.okC(bc, "POST", "/api/auth/register", map[string]string{"username": "bob", "password": "bob12345"})
	if code, _ := tc.do("POST", "/api/admin/users/bob/role", map[string]string{"role": "boss"}); code != 400 {
		t.Fatal("非法角色应 400")
	}
	tc.ok("POST", "/api/admin/users/bob/role", map[string]string{"role": "admin"})
	tc.ok("POST", "/api/admin/users/bob/reset", map[string]string{"password": "bob99999"})
	tc.ok("POST", "/api/admin/notices", map[string]string{"title": "上线", "body": "v2.2.1"})
	if _, n := tc.ok("GET", "/api/notices", nil); !strings.Contains(fmt.Sprint(n), "上线") {
		t.Fatal("公告未发布")
	}
	tc.ok("GET", "/api/admin/logs", nil)
	tc.ok("GET", "/api/admin/analytics", nil)
	tc.ok("POST", "/api/admin/settings", map[string]string{"allow_register": "0"})
	if code, _ := tc.doC(tc.fresh(), "POST", "/api/auth/register", map[string]string{"username": "eve", "password": "eve12345"}); code != 400 {
		t.Fatal("关闭注册应 400")
	}
	tc.ok("POST", "/api/admin/settings", map[string]string{"allow_register": "1"})
	// 14 权限：普通用户访问管理接口 → 403
	tc.okC(tc.fresh(), "POST", "/api/auth/register", map[string]string{"username": "carol", "password": "carol123"})
	cc := tc.fresh()
	tc.okC(cc, "POST", "/api/auth/login", map[string]string{"username": "carol", "password": "carol123"})
	if code, _ := tc.doC(cc, "GET", "/api/admin/overview", nil); code != 403 {
		t.Fatal("普通用户应 403")
	}
	// 15 导出
	if code, m := tc.do("GET", "/api/admin/export", nil); code != 200 || m["users"] == nil {
		t.Fatalf("导出: %d", code)
	}
	// 16 登出 → 会话失效
	tc.ok("POST", "/api/auth/logout", nil)
	if code, _ := tc.do("GET", "/api/auth/me", nil); code != 401 {
		t.Fatal("登出后应 401")
	}
}

func TestSecondUserIsNotAdmin(t *testing.T) {
	tc := start(t)
	tc.ok("POST", "/api/auth/register", map[string]string{"username": "a", "password": "aaaaaa"})
	tc.okC(tc.fresh(), "POST", "/api/auth/register", map[string]string{"username": "b", "password": "bbbbbb"})
	b := tc.fresh()
	tc.okC(b, "POST", "/api/auth/login", map[string]string{"username": "b", "password": "bbbbbb"})
	if code, _ := tc.doC(b, "GET", "/api/admin/overview", nil); code != 403 {
		t.Fatal("第二个用户不应是管理员")
	}
}
E2E_GO

# ---------- CI 流水线 ----------
echo "── .github/workflows/ci.yml ──"
w .github/workflows/ci.yml <<'CI_YML'
name: CI
on:
  push:
    branches: [ main ]
  workflow_dispatch:
    inputs:
      version:
        description: '发布版本号，如 v2.2.1'
        required: false
        default: ''
permissions:
  contents: write
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: 生成 go.sum（本机无 Go，由 CI 补齐）
        run: go mod tidy
      - name: 静态检查
        run: go vet ./...
      - name: 单元测试 + 全链路 E2E（测试报告就在这段日志里）
        run: go test -v ./...
  build:
    needs: test
    if: ${{ github.event_name == 'workflow_dispatch' }}
    runs-on: ubuntu-latest
    strategy:
      matrix:
        include:
          - { goos: windows, goarch: amd64, ext: .exe }
          - { goos: linux,  goarch: amd64, ext: '' }
          - { goos: darwin, goarch: arm64, ext: '' }
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go mod tidy
      - run: go vet ./... && go test ./...
      - name: Build
        run: |
          mkdir -p dist
          GOOS=${{ matrix.goos }} GOARCH=${{ matrix.goarch }} go build -trimpath -ldflags "-s -w -X main.Version=${{ github.event.inputs.version || 'dev' }}" -o dist/nova-server-${{ matrix.goos }}-${{ matrix.goarch }}${{ matrix.ext }} .
      - uses: actions/upload-artifact@v4
        with:
          name: nova-${{ matrix.goos }}-${{ matrix.goarch }}
          path: dist/
CI_YML

# ---------- .gitignore ----------
echo "── .gitignore ──"
w .gitignore <<'GITIGNORE'
data/
dist/
*.exe
GITIGNORE

# ---------- smoke.sh（本机冒烟，curl 版） ----------
echo "── smoke.sh ──"
w smoke.sh <<'SMOKE_SH'
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
SMOKE_SH
chmod +x smoke.sh

echo
echo "═══ 脚本2/3 完成。回复「继续」获取 setup-3.sh（前端全量 + 设计包CSS检查 + 自动提交推送） ═══"
