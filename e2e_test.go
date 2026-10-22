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

// doArrC：响应体是 JSON 数组时专用（tools/favs/history/workflows/notices 等）
func (tc *TC) doArrC(c *http.Client, method, path string, body any) (int, []any) {
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
	var arr []any
	json.Unmarshal(raw, &arr)
	if arr == nil {
		arr = []any{}
	}
	return resp.StatusCode, arr
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

func (tc *TC) okArr(m, p string) []any {
	code, arr := tc.doArrC(tc.c, m, p, nil)
	if code != 200 {
		tc.t.Fatalf("%s %s → %d", m, p, code)
	}
	return arr
}

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
	// 04 工具列表 16 个（数组端点 → okArr）
	tools := tc.okArr("GET", "/api/tools")
	if len(tools) != len(ToolRegistry) {
		t.Fatalf("应 len(ToolRegistry) 个工具: %d", len(tools))
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
	// 08 收藏 → 列表回显（数组）
	tc.ok("POST", "/api/fav/dev.json", nil)
	favs := tc.okArr("GET", "/api/favs")
	if !strings.Contains(fmt.Sprint(favs), "dev.json") {
		t.Fatalf("收藏: %v", favs)
	}
	// 09 历史已记录（数组）
	hist := tc.okArr("GET", "/api/history")
	if !strings.Contains(fmt.Sprint(hist), "dev.json") {
		t.Fatalf("历史: %v", hist)
	}
	// 10 工作流：保存 → 运行 → 两步轨迹（数组）
	tc.ok("POST", "/api/workflows", map[string]any{"name": "清洗", "steps": []string{"dev.json", "dev.encode"}})
	wfs := tc.okArr("GET", "/api/workflows")
	if len(wfs) == 0 {
		t.Fatal("工作流未保存")
	}
	wf := wfs[0].(map[string]any)
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
	notices := tc.okArr("GET", "/api/notices")
	if !strings.Contains(fmt.Sprint(notices), "上线") {
		t.Fatalf("公告未发布: %v", notices)
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
