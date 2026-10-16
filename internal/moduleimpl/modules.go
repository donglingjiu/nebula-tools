package moduleimpl

import (
	"archive/zip"
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"github.com/yuin/goldmark"

	"nebula-tools/internal/module"
)

// RegisterAll 注册全部 12 个模块（幂等）
func RegisterAll() {
	module.Register(&module.Module{ID: "dev.json", Name: "JSON 格式化", Icon: "🧾", Category: "dev", MinRole: 1, PayTier: "free",
		Routes: []module.Route{{"POST", "/api/dev/json/format", jsonFormat}}})
	module.Register(&module.Module{ID: "dev.md", Name: "Markdown 转换", Icon: "📝", Category: "dev", MinRole: 1, PayTier: "free",
		Routes: []module.Route{{"POST", "/api/dev/md/render", mdRender}}})
	module.Register(&module.Module{ID: "dev.encode", Name: "编解码", Icon: "🔐", Category: "dev", MinRole: 1, PayTier: "free",
		Routes: []module.Route{{"POST", "/api/dev/encode", encode}}})
	module.Register(&module.Module{ID: "dev.hash", Name: "哈希计算", Icon: "#️⃣", Category: "dev", MinRole: 1, PayTier: "free",
		Routes: []module.Route{{"POST", "/api/dev/hash", hashCalc}}})
	module.Register(&module.Module{ID: "dev.sql", Name: "SQL 格式化", Icon: "🗄️", Category: "dev", MinRole: 1, PayTier: "free",
		Routes: []module.Route{{"POST", "/api/dev/sql/format", sqlFormat}}})
	module.Register(&module.Module{ID: "file.pdf", Name: "PDF 提取", Icon: "📕", Category: "file", MinRole: 1, PayTier: "free",
		Routes: []module.Route{{"POST", "/api/file/pdf/to-text", pdfText}}})
	module.Register(&module.Module{ID: "file.excel", Name: "Excel 转 CSV", Icon: "📊", Category: "file", MinRole: 1, PayTier: "free",
		Routes: []module.Route{{"POST", "/api/file/excel/to-csv", excelToCsv}}})
	module.Register(&module.Module{ID: "file.word", Name: "Word 提取", Icon: "📘", Category: "file", MinRole: 1, PayTier: "free",
		Routes: []module.Route{{"POST", "/api/file/word/to-text", wordToText}}})
	module.Register(&module.Module{ID: "file.image", Name: "图片压缩", Icon: "🖼️", Category: "file", MinRole: 1, PayTier: "free",
		Routes: []module.Route{{"POST", "/api/file/image/resize", imageResize}}})
	module.Register(&module.Module{ID: "data.chart", Name: "图表生成", Icon: "📈", Category: "data", MinRole: 1, PayTier: "free",
		Routes: []module.Route{{"POST", "/api/data/chart/svg", chartSVG}}})
	module.Register(&module.Module{ID: "data.stat", Name: "词频统计", Icon: "🔢", Category: "data", MinRole: 1, PayTier: "free",
		Routes: []module.Route{{"POST", "/api/data/stat/analyze", statAnalyze}}})
	module.Register(&module.Module{ID: "home.search", Name: "全局搜索", Icon: "🔎", Category: "home", MinRole: 0, PayTier: "free"})
}

func bad(c *gin.Context, msg string) { c.JSON(400, gin.H{"ok": false, "error": msg}) }

// ── dev.json ──
func jsonFormat(c *gin.Context) {
	var in struct {
		JSON   string `json:"json"`
		Indent int    `json:"indent"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		bad(c, "参数错误")
		return
	}
	if in.Indent <= 0 || in.Indent > 8 {
		in.Indent = 2
	}
	var v any
	if err := json.Unmarshal([]byte(in.JSON), &v); err != nil {
		bad(c, "JSON 无效: "+err.Error())
		return
	}
	out, _ := json.MarshalIndent(v, "", strings.Repeat(" ", in.Indent))
	c.JSON(200, gin.H{"ok": true, "result": string(out)})
}

// ── dev.md ──
func mdRender(c *gin.Context) {
	var in struct {
		MD string `json:"md"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		bad(c, "参数错误")
		return
	}
	var buf strings.Builder
	if err := goldmark.Convert([]byte(in.MD), &buf); err != nil {
		bad(c, err.Error())
		return
	}
	c.JSON(200, gin.H{"ok": true, "result": buf.String()})
}

// ── dev.encode ──
func encode(c *gin.Context) {
	var in struct {
		Text string `json:"text"`
		Op   string `json:"op"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		bad(c, "参数错误")
		return
	}
	var out string
	switch in.Op {
	case "base64enc":
		out = base64.StdEncoding.EncodeToString([]byte(in.Text))
	case "base64dec":
		b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(in.Text))
		if err != nil {
			bad(c, "Base64 无效: "+err.Error())
			return
		}
		out = string(b)
	case "urlenc":
		out = url.QueryEscape(in.Text)
	case "urldec":
		s, err := url.QueryUnescape(in.Text)
		if err != nil {
			bad(c, "URL 解码失败: "+err.Error())
			return
		}
		out = s
	case "hexenc":
		out = hex.EncodeToString([]byte(in.Text))
	case "hexdec":
		b, err := hex.DecodeString(strings.TrimSpace(in.Text))
		if err != nil {
			bad(c, "Hex 无效")
			return
		}
		out = string(b)
	default:
		bad(c, "op 须为 base64enc/base64dec/urlenc/urldec/hexenc/hexdec")
		return
	}
	c.JSON(200, gin.H{"ok": true, "result": out})
}

// ── dev.hash ──
func hashCalc(c *gin.Context) {
	var in struct {
		Text string `json:"text"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		bad(c, "参数错误")
		return
	}
	m1 := md5.Sum([]byte(in.Text))
	s1 := sha1.Sum([]byte(in.Text))
	s2 := sha256.Sum256([]byte(in.Text))
	c.JSON(200, gin.H{"ok": true, "result": gin.H{
		"md5":    hex.EncodeToString(m1[:]),
		"sha1":   hex.EncodeToString(s1[:]),
		"sha256": hex.EncodeToString(s2[:]),
	}})
}

// ── dev.sql ──
var sqlMajor = map[string]bool{"from": true, "where": true, "left": true, "right": true,
	"inner": true, "outer": true, "full": true, "cross": true, "join": true, "on": true,
	"order": true, "group": true, "having": true, "limit": true, "offset": true,
	"union": true, "values": true, "set": true, "into": true}
var sqlKw = map[string]bool{"select": true, "insert": true, "update": true, "delete": true,
	"and": true, "or": true, "not": true, "null": true, "is": true, "in": true, "as": true,
	"asc": true, "desc": true, "by": true, "distinct": true, "case": true, "when": true,
	"then": true, "else": true, "end": true, "between": true, "like": true, "exists": true,
	"from": true, "where": true, "join": true, "on": true, "order": true, "group": true,
	"having": true, "limit": true, "offset": true, "union": true, "values": true, "set": true,
	"into": true, "left": true, "right": true, "inner": true, "outer": true}

func sqlFormat(c *gin.Context) {
	var in struct {
		SQL string `json:"sql"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		bad(c, "参数错误")
		return
	}
	toks := regexp.MustCompile(`'[^']*'|\w+|[(),*;=<>!+\-/.%]|\S`).FindAllString(in.SQL, -1)
	var b strings.Builder
	depth := 0
	for i, t := range toks {
		low := strings.ToLower(t)
		switch t {
		case "(":
			depth++
			b.WriteString("(")
		case ")":
			if depth > 0 {
				depth--
			}
			b.WriteString(")")
		case ",":
			b.WriteString(",")
			if depth == 0 {
				b.WriteString("\n   ")
			}
		case ";":
			b.WriteString(";")
			if i < len(toks)-1 {
				b.WriteString("\n")
			}
		default:
			if depth == 0 && i > 0 && sqlMajor[low] && toks[i-1] != "(" && toks[i-1] != "." {
				b.WriteString("\n  ")
			}
			if sqlKw[low] {
				t = strings.ToUpper(t)
			}
			b.WriteString(t)
		}
		if t != "," && t != ";" && t != "(" {
			b.WriteString(" ")
		}
	}
	out := regexp.MustCompile(`\n\s*\n`).ReplaceAllString(b.String(), "\n")
	c.JSON(200, gin.H{"ok": true, "result": strings.TrimSpace(out)})
}

// ── file.pdf（零依赖极简提取：支持未压缩/Flate 流；扫描件与加密件不支持）──
func pdfText(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		bad(c, "请上传文件字段 file")
		return
	}
	f, _ := fh.Open()
	defer f.Close()
	raw, _ := io.ReadAll(f)
	out := extractPDFText(raw)
	if strings.TrimSpace(out) == "" {
		bad(c, "未能提取到文本（可能为扫描件/加密/复杂编码 PDF）")
		return
	}
	c.JSON(200, gin.H{"ok": true, "result": out})
}

func extractPDFText(raw []byte) string {
	var out strings.Builder
	reStream := regexp.MustCompile(`(?s)stream\r?\n(.*?)endstream`)
	reLiteral := regexp.MustCompile(`\((?:\\.|[^\\()])*\)`)
	for _, m := range reStream.FindAllSubmatch(raw, -1) {
		data := m[1]
		if dec, err := zlib.NewReader(bytes.NewReader(data)); err == nil {
			if d, err := io.ReadAll(dec); err == nil {
				data = d
			}
		}
		for _, sm := range reLiteral.FindAllSubmatch(data, -1) {
			s := string(sm[0])
			s = strings.TrimSuffix(strings.TrimPrefix(s, "("), ")")
			s = strings.ReplaceAll(s, `\(`, "(")
			s = strings.ReplaceAll(s, `\)`, ")")
			s = strings.ReplaceAll(s, `\\`, `\`)
			out.WriteString(s)
		}
		out.WriteString("\n")
	}
	return out.String()
}

// ── file.excel ──
func excelToCsv(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		bad(c, "请上传文件字段 file")
		return
	}
	f, err := fh.Open()
	if err != nil {
		bad(c, err.Error())
		return
	}
	defer f.Close()
	xf, err := excelize.OpenReader(f)
	if err != nil {
		bad(c, "Excel 解析失败: "+err.Error())
		return
	}
	defer xf.Close()
	sheets := xf.GetSheetList()
	if len(sheets) == 0 {
		bad(c, "无工作表")
		return
	}
	rows, err := xf.GetRows(sheets[0])
	if err != nil {
		bad(c, err.Error())
		return
	}
	var buf bytes.Buffer
	buf.WriteString("\xEF\xBB\xBF") // UTF-8 BOM，Excel 打开不乱码
	w := csv.NewWriter(&buf)
	_ = w.WriteAll(rows)
	c.Header("Content-Disposition", "attachment; filename=converted.csv")
	c.Data(200, "text/csv; charset=utf-8", buf.Bytes())
}

// ── file.word（docx = zip + word/document.xml，零依赖）──
func wordToText(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		bad(c, "请上传文件字段 file")
		return
	}
	f, _ := fh.Open()
	defer f.Close()
	zr, err := zip.NewReader(f.(io.ReaderAt), fh.Size)
	if err != nil {
		bad(c, "不是有效的 docx 文件（老版 .doc 请先另存为 .docx）")
		return
	}
	var doc []byte
	for _, zf := range zr.File {
		if zf.Name == "word/document.xml" {
			rc, _ := zf.Open()
			doc, _ = io.ReadAll(rc)
			rc.Close()
			break
		}
	}
	if doc == nil {
		bad(c, "docx 中缺少 word/document.xml")
		return
	}
	s := string(doc)
	s = regexp.MustCompile(`</w:p>`).ReplaceAllString(s, "\n")
	s = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	c.JSON(200, gin.H{"ok": true, "result": s})
}

// ── file.image（最近邻缩放 + JPEG 质量压缩，零依赖）──
func imageResize(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		bad(c, "请上传文件字段 file")
		return
	}
	width, _ := strconv.Atoi(c.PostForm("width"))
	quality, _ := strconv.Atoi(c.PostForm("quality"))
	if width <= 0 {
		width = 800
	}
	if width > 4096 {
		width = 4096
	}
	if quality <= 0 || quality > 100 {
		quality = 80
	}
	f, _ := fh.Open()
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		bad(c, "图片解码失败（支持 jpg/png/gif）")
		return
	}
	src := img.Bounds()
	dstImg := img
	dx, dy := src.Dx(), src.Dy()
	if dx > width { // 不放大，只缩/压
		dy = src.Dy() * width / dx
		dst := image.NewRGBA(image.Rect(0, 0, width, dy))
		for y := 0; y < dy; y++ {
			sy := src.Min.Y + y*src.Dy()/dy
			for x := 0; x < width; x++ {
				sx := src.Min.X + x*src.Dx()/width
				dst.Set(x, y, img.At(sx, sy))
			}
		}
		dstImg = dst
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dstImg, &jpeg.Options{Quality: quality}); err != nil {
		bad(c, err.Error())
		return
	}
	c.Header("Content-Disposition", "attachment; filename=compressed.jpg")
	c.Data(200, "image/jpeg", buf.Bytes())
}

// ── data.chart（SVG，零依赖）──
func chartSVG(c *gin.Context) {
	var in struct {
		Type   string    `json:"type"`
		Title  string    `json:"title"`
		Labels []string  `json:"labels"`
		Values []float64 `json:"values"`
	}
	if err := c.ShouldBindJSON(&in); err != nil || len(in.Values) == 0 {
		bad(c, "需要 values 数组")
		return
	}
	if in.Type != "line" {
		in.Type = "bar"
	}
	mx := in.Values[0]
	for _, v := range in.Values {
		if v > mx {
			mx = v
		}
	}
	if mx <= 0 {
		mx = 1
	}
	W, H, pad := 640.0, 320.0, 44.0
	n := len(in.Values)
	div := 1.0
	if n > 1 {
		div = float64(n - 1)
	}
	var b strings.Builder
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="640" height="320" font-family="sans-serif">`)
	if in.Title != "" {
		b.WriteString(`<text x="320" y="24" font-size="14" font-weight="bold" text-anchor="middle" fill="#334155">` +
			html.EscapeString(in.Title) + `</text>`)
	}
	b.WriteString(fmt.Sprintf(`<line x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="#cbd5e1"/>`,
		pad, H-pad, W-pad, H-pad))
	if in.Type == "line" {
		pts := ""
		for i, v := range in.Values {
			x := pad + float64(i)*(W-2*pad)/div
			y := H - pad - v/mx*(H-2*pad)
			pts += fmt.Sprintf("%.1f,%.1f ", x, y)
			b.WriteString(fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="3" fill="#4f6ef7"/>`, x, y))
		}
		b.WriteString(fmt.Sprintf(`<polyline points="%s" fill="none" stroke="#4f6ef7" stroke-width="2"/>`,
			strings.TrimSpace(pts)))
	} else {
		bw := (W - 2*pad) / float64(n)
		for i, v := range in.Values {
			hh := v / mx * (H - 2 * pad)
			b.WriteString(fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="3" fill="#4f6ef7"/>`,
				pad+float64(i)*bw+bw*0.15, H-pad-hh, bw*0.7, hh))
			b.WriteString(fmt.Sprintf(`<text x="%.1f" y="%.1f" font-size="10" text-anchor="middle" fill="#64748b">%.0f</text>`,
				pad+float64(i)*bw+bw*0.5, H-pad-hh-4, v))
		}
	}
	for i := 0; i < len(in.Labels) && i < n; i++ {
		lb := in.Labels[i]
		if r := []rune(lb); len(r) > 12 {
			lb = string(r[:12])
		}
		var x float64
		if in.Type == "line" {
			x = pad + float64(i)*(W-2*pad)/div
		} else {
			bw := (W - 2*pad) / float64(n)
			x = pad + float64(i)*bw + bw*0.5
		}
		b.WriteString(fmt.Sprintf(`<text x="%.1f" y="%.0f" font-size="10" text-anchor="middle" fill="#64748b">%s</text>`,
			x, H-pad+14, html.EscapeString(lb)))
	}
	b.WriteString(`</svg>`)
	c.JSON(200, gin.H{"ok": true, "result": b.String()})
}

// ── data.stat（词频：中文 2 字组合 + 英文单词）──
func statAnalyze(c *gin.Context) {
	var in struct {
		Text string `json:"text"`
		TopN int    `json:"topn"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		bad(c, "参数错误")
		return
	}
	if in.TopN <= 0 {
		in.TopN = 20
	}
	if in.TopN > 100 {
		in.TopN = 100
	}
	toks := regexp.MustCompile(`[\p{Han}]{2}|[a-z][a-z0-9_]*|[0-9]+`).
		FindAllString(strings.ToLower(in.Text), -1)
	freq := map[string]int{}
	for _, t := range toks {
		freq[t]++
	}
	type kv struct {
		Word  string `json:"word"`
		Count int    `json:"count"`
	}
	arr := make([]kv, 0, len(freq))
	for k, v := range freq {
		arr = append(arr, kv{k, v})
	}
	sort.Slice(arr, func(i, j int) bool { return arr[i].Count > arr[j].Count })
	if len(arr) > in.TopN {
		arr = arr[:in.TopN]
	}
	c.JSON(200, gin.H{"ok": true, "total_tokens": len(toks), "result": arr})
}
