package main

// stage7-3 ③ 27工具可达矩阵: 分发链完整性测试(不测三方库解析能力)
// 双向锁: ToolRegistry每个id必有case(漏测编译产物即红), case必在Registry(防注册名漂移)
// 快照锁: 27 = 12主表文本 + 7 extra + 4 p02 + 4文件(B-1盘点口径)

import (
	"archive/zip"
	"bytes"
	"image"
	"image/png"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func mustOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
}

func has(t *testing.T, out, sub string) {
	t.Helper()
	if !strings.Contains(out, sub) {
		t.Fatalf("输出缺特征串 %q, 实得: %.120s", sub, out)
	}
}

// 分发可达: 错误存在但不得是"未知工具/未知文件工具"(那是分发链断裂, 而非输入问题)
func dispatchOK(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		return
	}
	if strings.Contains(err.Error(), "未知工具") || strings.Contains(err.Error(), "未知文件工具") {
		t.Fatalf("分发链断裂: %v", err)
	}
	t.Logf("非分发错误(输入级, 可接受): %v", err)
}

func runT(t *testing.T, id, in string, opts map[string]string) string {
	t.Helper()
	out, err := runTool(id, in, opts)
	mustOK(t, err)
	return out
}

func fxXLSX() []byte {
	f := excelize.NewFile()
	f.SetCellValue("Sheet1", "A1", "col1")
	f.SetCellValue("Sheet1", "A2", "x")
	b, _ := f.WriteToBuffer()
	return b.Bytes()
}

func fxPNG() []byte {
	buf := &bytes.Buffer{}
	png.Encode(buf, image.NewRGBA(image.Rect(0, 0, 1, 1)))
	return buf.Bytes()
}

func fxDocx() []byte {
	buf := &bytes.Buffer{}
	z := zip.NewWriter(buf)
	w, _ := z.Create("word/document.xml")
	w.Write([]byte(`<?xml version="1.0"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>col1</w:t></w:r></w:p></w:body></w:document>`))
	z.Close()
	return buf.Bytes()
}

func textCases() map[string]func(*testing.T) {
	return map[string]func(*testing.T){
		"dev.json": func(t *testing.T) { has(t, runT(t, "dev.json", `{"a":1}`, nil), `"a"`) },
		"dev.md":   func(t *testing.T) { has(t, runT(t, "dev.md", "# 标题", nil), "<h1>") },
		"dev.encode": func(t *testing.T) {
			if runT(t, "dev.encode", "hello", nil) == "" {
				t.Fatal("空输出")
			}
		},
		"dev.hash": func(t *testing.T) {
			if len(runT(t, "dev.hash", "abc", nil)) < 16 {
				t.Fatal("摘要过短")
			}
		},
		"dev.sql": func(t *testing.T) { has(t, runT(t, "dev.sql", "select a from t", nil), "SELECT") },
		"dev.util": func(t *testing.T) {
			if runT(t, "dev.util", "", nil) == "" {
				t.Fatal("空输出")
			}
		},
		"dev.diff": func(t *testing.T) {
			if runT(t, "dev.diff", "a\n---\nb", nil) == "" {
				t.Fatal("空输出")
			}
		},
		"gen.qr": func(t *testing.T) { has(t, runT(t, "gen.qr", "https://example.com", nil), "data:image/png;base64,") },
		"gen.base": func(t *testing.T) {
			if runT(t, "gen.base", "255", nil) == "" {
				t.Fatal("空输出")
			}
		},
		"gen.pwd": func(t *testing.T) {
			if len(runT(t, "gen.pwd", "", map[string]string{"len": "16", "count": "3"})) < 8 {
				t.Fatal("输出过短")
			}
		},
		"data.csv2json": func(t *testing.T) { o := runT(t, "data.csv2json", "a,b\n1,2", nil); has(t, o, "["); has(t, o, "1") },
		"data.json2csv": func(t *testing.T) { has(t, runT(t, "data.json2csv", `[{"a":1,"b":2}]`, nil), "a") },
		"dev.hmac": func(t *testing.T) {
			if len(runT(t, "dev.hmac", "msg", map[string]string{"alg": "sha256", "key": "k"})) < 16 {
				t.Fatal("HMAC过短")
			}
		},
		"dev.totp": func(t *testing.T) {
			if len(runT(t, "dev.totp", "JBSWY3DPEHPK3PXP", nil)) < 6 {
				t.Fatal("口令过短")
			}
		},
		"dev.jwt": func(t *testing.T) {
			if len(runT(t, "dev.jwt", "eyJhbGciOiJIUzI1NiJ9.eyJhIjoxfQ.sig", nil)) < 10 {
				t.Fatal("解析输出过短")
			}
		},
		"dev.case": func(t *testing.T) {
			o := runT(t, "dev.case", "FooBar", map[string]string{"mode": "snake"})
			has(t, o, "foo")
			has(t, o, "bar")
		},
		"dev.slug": func(t *testing.T) {
			o := runT(t, "dev.slug", "Hello World 123", nil)
			for _, r := range o {
				if r > 127 {
					t.Fatalf("slug含非ASCII: %q", r)
				}
			}
		},
		"dev.lorem": func(t *testing.T) {
			if len(runT(t, "dev.lorem", "", map[string]string{"paras": "2"})) < 20 {
				t.Fatal("假文过短")
			}
		},
		"dev.count": func(t *testing.T) {
			if len(runT(t, "dev.count", "hello world 你好", nil)) < 3 {
				t.Fatal("统计过短")
			}
		},
		"dev.cron": func(t *testing.T) {
			if len(runT(t, "dev.cron", "*/5 * * * *", map[string]string{"count": "2"})) < 10 {
				t.Fatal("解析过短")
			}
		},
		"dev.chmod":  func(t *testing.T) { has(t, runT(t, "dev.chmod", "754", nil), "rwxr-xr--") },
		"dev.subnet": func(t *testing.T) { has(t, runT(t, "dev.subnet", "192.168.1.0/24", nil), "255.255.255.0") },
		"dev.yaml":   func(t *testing.T) { has(t, runT(t, "dev.yaml", "a: 1", nil), `"a"`) },
	}
}

func fileCases() map[string]func(*testing.T) {
	return map[string]func(*testing.T){
		"file.pdf": func(t *testing.T) {
			_, err := runToolFile("file.pdf", []byte("%PDF-1.4\nfake-xref"), nil)
			dispatchOK(t, err)
		},
		"file.excel": func(t *testing.T) {
			o, err := runToolFile("file.excel", fxXLSX(), nil)
			mustOK(t, err)
			has(t, o, "col1")
		},
		"file.word": func(t *testing.T) {
			o, err := runToolFile("file.word", fxDocx(), nil)
			dispatchOK(t, err)
			if err == nil && !strings.Contains(o, "col1") {
				t.Logf("docx解析成功但文本缺失: %.80s", o)
			}
		},
		"file.img": func(t *testing.T) { _, err := runToolFile("file.img", fxPNG(), nil); mustOK(t, err) },
	}
}

func TestRegistrySnapshot(t *testing.T) {
	if n := len(ToolRegistry); n != 27 {
		t.Errorf("注册表漂移: %d≠27 — 更新B-1报告与本矩阵case表", n)
	}
	tc, fc := textCases(), fileCases()
	for id, tool := range ToolRegistry {
		want := fc
		if tool.Kind == "text" {
			want = tc
		}
		if _, ok := want[id]; !ok {
			t.Errorf("工具 %s(kind=%s) 无矩阵case — 新增工具必须带测试", id, tool.Kind)
		}
	}
	for id := range tc {
		if _, ok := ToolRegistry[id]; !ok {
			t.Errorf("case表含未注册工具: %s", id)
		}
	}
	for id := range fc {
		if _, ok := ToolRegistry[id]; !ok {
			t.Errorf("case表含未注册文件工具: %s", id)
		}
	}
}

func TestTextToolMatrix(t *testing.T) {
	for id, fn := range textCases() {
		t.Run(id, fn)
	}
}

func TestFileToolMatrix(t *testing.T) {
	for id, fn := range fileCases() {
		t.Run(id, fn)
	}
}
