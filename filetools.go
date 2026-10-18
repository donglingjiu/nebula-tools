package main

// filetools.go — 文件工具引擎：PDF 合并/拆分/压缩、PDF→Word、PDF→Excel。
// 全程内存处理，不落盘；通过 main.go 里一行 registerFileTools(mux, db) 挂载。

import (
	"context"
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/xuri/excelize/v2"
)

func fxUser(db *DB, r *http.Request) *User {
	tok := ""
	if c, err := r.Cookie("nova_tok"); err == nil {
		tok = c.Value
	}
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		tok = strings.TrimPrefix(h, "Bearer ")
	}
	return db.SessionUser(tok)
}

func fxE(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func fxSend(w http.ResponseWriter, name string, b []byte, mime string) {
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+name+"\"")
	w.Write(b)
}

func fxFiles(w http.ResponseWriter, r *http.Request) ([][]byte, []string, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<20)
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		fxE(w, 400, "上传解析失败（总大小限 64MB）: "+err.Error())
		return nil, nil, false
	}
	if r.MultipartForm == nil || len(r.MultipartForm.File) == 0 {
		fxE(w, 400, "请选择文件")
		return nil, nil, false
	}
	var bs [][]byte
	var names []string
	for _, fhs := range r.MultipartForm.File {
		for _, fh := range fhs {
			f, err := fh.Open()
			if err != nil {
				fxE(w, 400, "读取文件失败")
				return nil, nil, false
			}
			b, err := io.ReadAll(f)
			f.Close()
			if err != nil {
				fxE(w, 400, "读取文件失败")
				return nil, nil, false
			}
			bs = append(bs, b)
			names = append(names, fh.Filename)
		}
	}
	return bs, names, true
}

func fxOpts(r *http.Request) map[string]string {
	opts := map[string]string{}
	if sv := r.FormValue("opts"); sv != "" {
		json.Unmarshal([]byte(sv), &opts)
	}
	return opts
}

func fxPDF(b []byte) bool { return len(b) > 4 && string(b[:4]) == "%PDF" }

func fxPdfText(b []byte) (string, error) {
	pr, err := pdf.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return "", err
	}
	rd, err := pr.GetPlainText()
	if err != nil {
		return "", err
	}
	sb := &strings.Builder{}
	io.Copy(sb, rd)
	return sb.String(), nil
}

func hPdfMerge(db *DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := fxUser(db, r)
		if u == nil {
			fxE(w, 401, "请先登录")
			return
		}
		t0 := time.Now()
		bs, names, ok := fxFiles(w, r)
		if !ok {
			return
		}
		if len(bs) < 2 {
			fxE(w, 400, "合并至少需要两个 PDF 文件")
			return
		}
		for i, b := range bs {
			if !fxPDF(b) {
				fxE(w, 400, "第 "+strconv.Itoa(i+1)+" 个文件不是 PDF: "+names[i])
				return
			}
		}
		rs := make([]io.ReadSeeker, 0, len(bs))
		for _, b := range bs {
			rs = append(rs, bytes.NewReader(b))
		}
		var out bytes.Buffer
		if err := api.MergeRaw(context.Background(), rs, &out, false, nil); err != nil {
			fxE(w, 422, "合并失败: "+err.Error())
			return
		}
		data := out.Bytes()
		if fxOpts(r)["compress"] == "true" {
			var c bytes.Buffer
			if err := api.Optimize(bytes.NewReader(data), &c, nil); err == nil && c.Len() > 0 && c.Len() < len(data) {
				data = c.Bytes()
			}
		}
		db.AddHist(u.Username, "fx:pdf-merge", fmt.Sprintf("合并 %d 个文件", len(bs)), fmt.Sprintf("输出 %d KB", len(data)/1024), time.Since(t0).Milliseconds())
		fxSend(w, "merged.pdf", data, "application/pdf")
	}
}

func hPdfSplit(db *DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := fxUser(db, r)
		if u == nil {
			fxE(w, 401, "请先登录")
			return
		}
		t0 := time.Now()
		bs, names, ok := fxFiles(w, r)
		if !ok {
			return
		}
		if len(bs) != 1 || !fxPDF(bs[0]) {
			fxE(w, 400, "请上传一个 PDF 文件")
			return
		}
		n, err := api.PageCount(bytes.NewReader(bs[0]), nil)
		if err != nil {
			fxE(w, 422, "读取页数失败: "+err.Error())
			return
		}
		if n < 2 {
			fxE(w, 400, "只有 1 页，无需拆分")
			return
		}
		base := strings.TrimSuffix(names[0], ".pdf")
		var zipBuf bytes.Buffer
		zw := zip.NewWriter(&zipBuf)
		for p := 1; p <= n; p++ {
			var pb bytes.Buffer
			if err := api.Collect(bytes.NewReader(bs[0]), &pb, []string{strconv.Itoa(p)}, nil); err != nil {
				zw.Close()
				fxE(w, 422, fmt.Sprintf("拆分第 %d 页失败: %v", p, err))
				return
			}
			fw, _ := zw.Create(fmt.Sprintf("%s-第%d页.pdf", base, p))
			fw.Write(pb.Bytes())
		}
		zw.Close()
		db.AddHist(u.Username, "fx:pdf-split", names[0], fmt.Sprintf("%d 页 → %d 个 PDF", n, n), time.Since(t0).Milliseconds())
		fxSend(w, base+"-拆分.zip", zipBuf.Bytes(), "application/zip")
	}
}

func hPdfCompress(db *DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := fxUser(db, r)
		if u == nil {
			fxE(w, 401, "请先登录")
			return
		}
		t0 := time.Now()
		bs, names, ok := fxFiles(w, r)
		if !ok {
			return
		}
		if len(bs) != 1 || !fxPDF(bs[0]) {
			fxE(w, 400, "请上传一个 PDF 文件")
			return
		}
		var out bytes.Buffer
		if err := api.Optimize(bytes.NewReader(bs[0]), &out, nil); err != nil {
			fxE(w, 422, "优化失败: "+err.Error())
			return
		}
		if out.Len() == 0 || out.Len() >= len(bs[0]) {
			db.AddHist(u.Username, "fx:pdf-compress", names[0], "已是最优体积，返回原文件", time.Since(t0).Milliseconds())
			fxSend(w, names[0], bs[0], "application/pdf")
			return
		}
		db.AddHist(u.Username, "fx:pdf-compress", names[0], fmt.Sprintf("%d KB → %d KB", len(bs[0])/1024, out.Len()/1024), time.Since(t0).Milliseconds())
		fxSend(w, "compressed-"+names[0], out.Bytes(), "application/pdf")
	}
}

func fxBuildDocx(paras []string) []byte {
	esc := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	var bb strings.Builder
	bb.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	bb.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`)
	for _, p := range paras {
		bb.WriteString(`<w:p><w:r><w:t xml:space="preserve">` + esc.Replace(p) + `</w:t></w:r></w:p>`)
	}
	bb.WriteString(`<w:sectPr><w:pgSz w:w="11906" w:h="16838"/></w:sectPr></w:body></w:document>`)
	ct := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>`
	rels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	f1, _ := zw.Create("[Content_Types].xml")
	f1.Write([]byte(ct))
	f2, _ := zw.Create("_rels/.rels")
	f2.Write([]byte(rels))
	f3, _ := zw.Create("word/document.xml")
	f3.Write([]byte(bb.String()))
	zw.Close()
	return buf.Bytes()
}

func hPdfToDocx(db *DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := fxUser(db, r)
		if u == nil {
			fxE(w, 401, "请先登录")
			return
		}
		t0 := time.Now()
		bs, names, ok := fxFiles(w, r)
		if !ok {
			return
		}
		if len(bs) != 1 || !fxPDF(bs[0]) {
			fxE(w, 400, "请上传一个 PDF 文件")
			return
		}
		text, err := fxPdfText(bs[0])
		if err != nil || strings.TrimSpace(text) == "" {
			fxE(w, 422, "未提取到文本（可能是扫描件/图片型 PDF，文字层转换只支持文本型 PDF）")
			return
		}
		paras := []string{}
		for _, line := range strings.Split(text, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				paras = append(paras, line)
			}
			if len(paras) >= 5000 {
				break
			}
		}
		docx := fxBuildDocx(paras)
		base := strings.TrimSuffix(names[0], ".pdf")
		db.AddHist(u.Username, "fx:pdf-to-docx", names[0], fmt.Sprintf("提取 %d 段 → docx", len(paras)), time.Since(t0).Milliseconds())
		fxSend(w, base+".docx", docx, "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	}
}

func hPdfToXlsx(db *DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := fxUser(db, r)
		if u == nil {
			fxE(w, 401, "请先登录")
			return
		}
		t0 := time.Now()
		bs, names, ok := fxFiles(w, r)
		if !ok {
			return
		}
		if len(bs) != 1 || !fxPDF(bs[0]) {
			fxE(w, 400, "请上传一个 PDF 文件")
			return
		}
		text, err := fxPdfText(bs[0])
		if err != nil || strings.TrimSpace(text) == "" {
			fxE(w, 422, "未提取到文本（可能是扫描件/图片型 PDF）")
			return
		}
		f := excelize.NewFile()
		defer f.Close()
		row := 1
		for _, line := range strings.Split(text, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			cells := strings.Split(line, "\t")
			if len(cells) == 1 {
				cells = strings.Split(line, ",")
			}
			for c, cv := range cells {
				cell, _ := excelize.CoordinatesToCellName(c+1, row)
				f.SetCellValue("Sheet1", cell, cv)
			}
			row++
			if row > 20000 {
				break
			}
		}
		buf, err := f.WriteToBuffer()
		if err != nil {
			fxE(w, 500, "生成 xlsx 失败: "+err.Error())
			return
		}
		base := strings.TrimSuffix(names[0], ".pdf")
		db.AddHist(u.Username, "fx:pdf-to-xlsx", names[0], fmt.Sprintf("%d 行 → xlsx", row-1), time.Since(t0).Milliseconds())
		fxSend(w, base+".xlsx", buf.Bytes(), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	}
}

func registerFileTools(mux *http.ServeMux, db *DB) {
	wrap := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if fxUser(db, r) == nil {
				fxE(w, 401, "请先登录后使用文件工具")
				return
			}
			h(w, r)
		}
	}
	mux.HandleFunc("POST /api/fx/pdf-merge", wrap(hPdfMerge(db)))
	mux.HandleFunc("POST /api/fx/pdf-split", wrap(hPdfSplit(db)))
	mux.HandleFunc("POST /api/fx/pdf-compress", wrap(hPdfCompress(db)))
	mux.HandleFunc("POST /api/fx/pdf-to-docx", wrap(hPdfToDocx(db)))
	mux.HandleFunc("POST /api/fx/pdf-to-xlsx", wrap(hPdfToXlsx(db)))
}
