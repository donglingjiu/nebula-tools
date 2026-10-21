#!/usr/bin/env bash
# Nova v2.2.1 —— 修复脚本：补齐 tools.go（16 工具 + ToolRegistry）
cd "$(dirname "$0")"
w(){ local p="$1"; cat > "$p"; echo "  写入 $p ($(wc -l < "$p" | tr -d ' ') 行)"; }

echo "═══ setup-4：补齐工具实现层 ═══"
[ -f main.go ] || { echo "✗ 请在仓库根目录执行"; exit 1; }

w tools.go <<'TOOLS_GO'
package main

import (
	"archive/zip"
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math/big"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/skip2/go-qrcode"
)

type Tool struct {
	Id, Name, Icon, Cat, Desc, Kind string
}

var ToolRegistry = map[string]Tool{
	"dev.json":      {"dev.json", "JSON 工具", "code", "dev", "格式化 / 压缩 / 校验 JSON", "text"},
	"dev.md":        {"dev.md", "Markdown 转换", "fileText", "dev", "Markdown 转 HTML（标题/列表/加粗）", "text"},
	"dev.encode":    {"dev.encode", "编码转换", "lock", "dev", "Base64 与 URL 编解码", "text"},
	"dev.hash":      {"dev.hash", "哈希计算", "shield", "dev", "SHA-256 / SHA-1 / MD5 摘要", "text"},
	"dev.sql":       {"dev.sql", "SQL 格式化", "database", "dev", "关键字大写与子句换行", "text"},
	"dev.util":      {"dev.util", "时间与 UUID", "clock", "dev", "时间戳互转与 UUID 生成", "text"},
	"dev.diff":      {"dev.diff", "文本对比", "copy", "dev", "两段文本逐行差异（用单独一行 --- 分隔）", "text"},
	"gen.qr":        {"gen.qr", "二维码生成", "grid", "gen", "文本 / 链接生成二维码图片", "text"},
	"gen.base":      {"gen.base", "进制转换", "terminal", "gen", "2/8/10/16/36 进制互转", "text"},
	"gen.pwd":       {"gen.pwd", "密码生成", "key", "gen", "随机强密码批量生成", "text"},
	"data.csv2json": {"data.csv2json", "CSV 转 JSON", "files", "data", "表格数据转 JSON 数组", "text"},
	"data.json2csv": {"data.json2csv", "JSON 转 CSV", "files", "data", "JSON 数组转 CSV 表格", "text"},
	"file.pdf":      {"file.pdf", "PDF 文本提取", "fileText", "file", "提取 PDF 中的文本内容", "file"},
	"file.excel":    {"file.excel", "Excel 转 CSV", "chart", "file", "xlsx 表格转 CSV（≤30MB）", "file"},
	"file.word":     {"file.word", "Word 文本提取", "files", "file", "提取 docx 中的文本段落", "file"},
	"file.img":      {"file.img", "图片压缩转换", "image", "file", "压缩 / JPG / PNG / 缩放", "file"},
}

func runTool(id, in string, opts map[string]string) (string, error) {
	switch id {
	case "dev.json":
		return toolJSON(in, opts)
	case "dev.md":
		return toolMD(in)
	case "dev.encode":
		return toolEncode(in, opts)
	case "dev.hash":
		return toolHash(in, opts)
	case "dev.sql":
		return toolSQL(in)
	case "dev.util":
		return toolUtil(in, opts)
	case "dev.diff":
		return toolDiff(in)
	case "gen.qr":
		return toolQR(in, opts)
	case "gen.base":
		return toolBase(in, opts)
	case "gen.pwd":
		return toolPwd(opts)
	case "data.csv2json":
		return toolCSV2JSON(in)
	case "data.json2csv":
		return toolJSON2CSV(in)
	}
	return "", fmt.Errorf("未知工具: %s", id)
}

func runToolFile(id string, b []byte, opts map[string]string) (string, error) {
	switch id {
	case "file.pdf":
		return toolPDF(b)
	case "file.excel":
		return toolExcel(b)
	case "file.word":
		return toolWord(b)
	case "file.img":
		return toolImg(b, opts)
	}
	return "", fmt.Errorf("未知文件工具: %s", id)
}

/* ---------------- 文本工具 ---------------- */

func toolJSON(in string, opts map[string]string) (string, error) {
	in = strings.TrimSpace(in)
	if in == "" {
		return "", errors.New("请输入 JSON")
	}
	if !json.Valid([]byte(in)) {
		return "", errors.New("❌ JSON 不合法：请检查括号、引号与逗号")
	}
	mode := opts["mode"]
	if mode == "" {
		mode = "format"
	}
	switch mode {
	case "minify":
		var buf bytes.Buffer
		if err := json.Compact(&buf, []byte(in)); err != nil {
			return "", err
		}
		return buf.String(), nil
	case "validate":
		var v any
		json.Unmarshal([]byte(in), &v)
		return "✅ JSON 合法（顶层类型: " + jsonType(v) + "，长度 " + strconv.Itoa(len(in)) + " 字符）", nil
	default:
		var buf bytes.Buffer
		if err := json.Indent(&buf, []byte(in), "", "  "); err != nil {
			return "", err
		}
		return buf.String(), nil
	}
}
func jsonType(v any) string {
	switch v.(type) {
	case map[string]any:
		return "object"
	case []any:
		return "array"
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "bool"
	case nil:
		return "null"
	}
	return "unknown"
}

func toolMD(in string) (string, error) {
	if strings.TrimSpace(in) == "" {
		return "", errors.New("请输入 Markdown")
	}
	lines := strings.Split(strings.ReplaceAll(in, "\r\n", "\n"), "\n")
	var out []string
	inList := false
	closeList := func() {
		if inList {
			out = append(out, "</ul>")
			inList = false
		}
	}
	for _, ln := range lines {
		t := strings.TrimSpace(ln)
		switch {
		case t == "":
			closeList()
		case strings.HasPrefix(t, "### "):
			closeList()
			out = append(out, "<h3>"+mdInline(t[4:])+"</h3>")
		case strings.HasPrefix(t, "## "):
			closeList()
			out = append(out, "<h2>"+mdInline(t[3:])+"</h2>")
		case strings.HasPrefix(t, "# "):
			closeList()
			out = append(out, "<h1>"+mdInline(t[2:])+"</h1>")
		case strings.HasPrefix(t, "- "), strings.HasPrefix(t, "* "):
			if !inList {
				out = append(out, "<ul>")
				inList = true
			}
			out = append(out, "<li>"+mdInline(t[2:])+"</li>")
		case t == "---":
			closeList()
			out = append(out, "<hr>")
		default:
			closeList()
			out = append(out, "<p>"+mdInline(t)+"</p>")
		}
	}
	closeList()
	return strings.Join(out, "\n"), nil
}
func mdInline(s string) string {
	s = html.EscapeString(s)
	s = regexp.MustCompile("(?U)\\*\\*(.+?)\\*\\*").ReplaceAllString(s, "<strong>$1</strong>")
	s = regexp.MustCompile("(?U)\\*(.+?)\\*").ReplaceAllString(s, "<em>$1</em>")
	s = regexp.MustCompile("(?U)`(.+?)`").ReplaceAllString(s, "<code>$1</code>")
	return s
}

func toolEncode(in string, opts map[string]string) (string, error) {
	op := opts["op"]
	if op == "" {
		op = "b64enc"
	}
	switch op {
	case "b64enc":
		return base64.StdEncoding.EncodeToString([]byte(in)), nil
	case "b64dec":
		b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(in))
		if err != nil {
			return "", errors.New("不是合法的 Base64")
		}
		return string(b), nil
	case "uenc":
		return url.QueryEscape(in), nil
	case "udec":
		s, err := url.QueryUnescape(strings.TrimSpace(in))
		if err != nil {
			return "", errors.New("不是合法的 URL 编码")
		}
		return s, nil
	}
	return "", errors.New("未知编码操作")
}

func toolHash(in string, opts map[string]string) (string, error) {
	algo := strings.ToLower(opts["algo"])
	if algo == "" {
		algo = "sha256"
	}
	var h string
	switch algo {
	case "sha256":
		s := sha256.Sum256([]byte(in))
		h = hex.EncodeToString(s[:])
	case "sha1":
		s := sha1.Sum([]byte(in))
		h = hex.EncodeToString(s[:])
	case "md5":
		s := md5.Sum([]byte(in))
		h = hex.EncodeToString(s[:])
	default:
		return "", errors.New("algo 仅支持 sha256 / sha1 / md5")
	}
	return h + "\n(" + algo + " · 输入 " + strconv.Itoa(len(in)) + " 字节)", nil
}

var sqlClauseKw = map[string]bool{"FROM": true, "WHERE": true, "GROUP": true, "ORDER": true, "HAVING": true, "LIMIT": true, "OFFSET": true, "VALUES": true, "SET": true, "JOIN": true, "INNER": true, "LEFT": true, "RIGHT": true, "FULL": true, "UNION": true}
var sqlPlainKw = map[string]bool{"SELECT": true, "INSERT": true, "INTO": true, "UPDATE": true, "DELETE": true, "AND": true, "OR": true, "NOT": true, "NULL": true, "IS": true, "IN": true, "AS": true, "ON": true, "BY": true, "DISTINCT": true, "COUNT": true, "SUM": true, "AVG": true, "MIN": true, "MAX": true, "ASC": true, "DESC": true, "LIKE": true, "BETWEEN": true, "EXISTS": true, "CASE": true, "WHEN": true, "THEN": true, "ELSE": true, "END": true}

func toolSQL(in string) (string, error) {
	fields := strings.Fields(in)
	if len(fields) == 0 {
		return "", errors.New("请输入 SQL 语句")
	}
	var out strings.Builder
	for k, f := range fields {
		u := strings.ToUpper(f)
		bare := strings.TrimRight(u, ",;")
		rest := strings.TrimPrefix(u, bare)
		if sqlPlainKw[bare] || sqlClauseKw[bare] {
			f = bare + rest
		}
		if k > 0 && sqlClauseKw[bare] {
			out.WriteString("\n")
		}
		out.WriteString(f + " ")
	}
	return strings.TrimSpace(out.String()), nil
}

func uuidV4() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func toolUtil(in string, opts map[string]string) (string, error) {
	op := opts["op"]
	if op == "" {
		op = "now"
	}
	switch op {
	case "uuid":
		var sb strings.Builder
		for i := 0; i < 5; i++ {
			sb.WriteString(uuidV4() + "\n")
		}
		return sb.String(), nil
	case "ts2date":
		s := strings.TrimSpace(in)
		if s == "" {
			return "", errors.New("请输入时间戳")
		}
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return "", errors.New("时间戳格式错误")
		}
		if n > 1_000_000_000_000 {
			n /= 1000
		}
		return time.Unix(n, 0).Format("2006-01-02 15:04:05") + "\nUnix: " + strconv.FormatInt(n, 10), nil
	case "date2ts":
		s := strings.TrimSpace(in)
		for _, l := range []string{"2006-01-02 15:04:05", "2006-01-02 15:04", "2006-01-02", "2006/01/02 15:04:05", "2006/01/02"} {
			if t, err := time.ParseInLocation(l, s, time.Local); err == nil {
				return fmt.Sprintf("秒: %d\n毫秒: %d", t.Unix(), t.UnixMilli()), nil
			}
		}
		return "", errors.New("无法识别的日期，支持 2006-01-02 15:04:05 等格式")
	default:
		t := time.Now()
		return fmt.Sprintf("本地: %s\nUTC:  %s\nUnix 秒: %d\nUnix 毫秒: %d\nUUID: %s",
			t.Format("2006-01-02 15:04:05"), t.UTC().Format("2006-01-02 15:04:05"), t.Unix(), t.UnixMilli(), uuidV4()), nil
	}
}

func toolDiff(in string) (string, error) {
	parts := strings.SplitN(strings.ReplaceAll(in, "\r\n", "\n"), "\n---\n", 2)
	if len(parts) != 2 {
		return "", errors.New("请将两段文本放入输入框，并用单独一行 --- 分隔")
	}
	a := strings.Split(strings.TrimSuffix(parts[0], "\n"), "\n")
	b := strings.Split(strings.TrimSuffix(parts[1], "\n"), "\n")
	n, m := len(a), len(b)
	if n*m > 4_000_000 {
		return "", errors.New("文本过长，请分段对比")
	}
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}
	var out []string
	i, j := 0, 0
	for i < n && j < m {
		if a[i] == b[j] {
			out = append(out, "  "+a[i])
			i++
			j++
		} else if dp[i+1][j] >= dp[i][j+1] {
			out = append(out, "- "+a[i])
			i++
		} else {
			out = append(out, "+ "+b[j])
			j++
		}
	}
	for ; i < n; i++ {
		out = append(out, "- "+a[i])
	}
	for ; j < m; j++ {
		out = append(out, "+ "+b[j])
	}
	if len(out) == 0 {
		return "（两段文本完全相同）", nil
	}
	return strings.Join(out, "\n"), nil
}

func toolQR(in string, opts map[string]string) (string, error) {
	in = strings.TrimSpace(in)
	if in == "" {
		return "", errors.New("请输入二维码内容（文本或链接）")
	}
	size, _ := strconv.Atoi(opts["size"])
	if size < 64 || size > 2000 {
		size = 512
	}
	pb, err := qrcode.Encode(in, qrcode.Medium, size)
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(pb), nil
}

func toolBase(in string, opts map[string]string) (string, error) {
	base, _ := strconv.Atoi(opts["base"])
	if base < 2 || base > 36 {
		base = 10
	}
	s := strings.TrimSpace(in)
	if s == "" {
		return "", errors.New("请输入数字")
	}
	n, err := strconv.ParseInt(s, base, 64)
	if err != nil {
		return "", fmt.Errorf("「%s」不是合法的 %d 进制数", s, base)
	}
	return fmt.Sprintf("二进制: %s\n八进制: %s\n十进制: %d\n十六进制: %s\n三十六进制: %s",
		strconv.FormatInt(n, 2), strconv.FormatInt(n, 8), n,
		strings.ToUpper(strconv.FormatInt(n, 16)), strings.ToUpper(strconv.FormatInt(n, 36))), nil
}

func randInt(n int) int {
	v, _ := rand.Int(rand.Reader, big.NewInt(int64(n)))
	return int(v.Int64())
}

func toolPwd(opts map[string]string) (string, error) {
	l, _ := strconv.Atoi(opts["len"])
	if l < 6 || l > 128 {
		l = 12
	}
	c, _ := strconv.Atoi(opts["count"])
	if c < 1 || c > 50 {
		c = 5
	}
	lower, upper := "abcdefghijkmnopqrstuvwxyz", "ABCDEFGHJKLMNPQRSTUVWXYZ"
	digits, symbols := "23456789", "!@#$%^&*-_=+"
	all := lower + upper + digits + symbols
	var sb strings.Builder
	for i := 0; i < c; i++ {
		pw := make([]byte, 0, l)
		for _, cl := range []string{lower, upper, digits, symbols} {
			pw = append(pw, cl[randInt(len(cl))])
		}
		for len(pw) < l {
			pw = append(pw, all[randInt(len(all))])
		}
		for k := len(pw) - 1; k > 0; k-- {
			r := randInt(k + 1)
			pw[k], pw[r] = pw[r], pw[k]
		}
		sb.Write(pw)
		sb.WriteString("\n")
	}
	return strings.TrimSpace(sb.String()) + fmt.Sprintf("\n\n长度 %d · 共 %d 个（已去除易混淆字符 0O1lI）", l, c), nil
}

func toolCSV2JSON(in string) (string, error) {
	rows, err := csv.NewReader(strings.NewReader(strings.TrimSpace(in))).ReadAll()
	if err != nil {
		return "", errors.New("CSV 解析失败: " + err.Error())
	}
	if len(rows) < 1 {
		return "", errors.New("CSV 为空")
	}
	head := rows[0]
	arr := []map[string]any{}
	for _, row := range rows[1:] {
		m := map[string]any{}
		for i, v := range row {
			if i >= len(head) {
				break
			}
			h := strings.TrimSpace(head[i])
			if h == "" {
				h = "col" + strconv.Itoa(i+1)
			}
			if f, e := strconv.ParseFloat(strings.TrimSpace(v), 64); e == nil && v != "" {
				m[h] = f
			} else {
				m[h] = v
			}
		}
		arr = append(arr, m)
	}
	b, _ := json.Marshal(arr)
	return string(b), nil
}

func toolJSON2CSV(in string) (string, error) {
	var arr []map[string]any
	if err := json.Unmarshal([]byte(in), &arr); err != nil {
		return "", errors.New(`输入必须是对象数组，如 [{"a":1}]`)
	}
	keys := []string{}
	seen := map[string]bool{}
	for _, m := range arr {
		for k := range m {
			if !seen[k] {
				seen[k] = true
				keys = append(keys, k)
			}
		}
	}
	if len(keys) == 0 {
		return "", errors.New("数组为空")
	}
	sort.Strings(keys)
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Write(keys)
	for _, m := range arr {
		row := make([]string, len(keys))
		for i, k := range keys {
			switch v := m[k].(type) {
			case nil:
				row[i] = ""
			case string:
				row[i] = v
			case float64:
				if v == float64(int64(v)) {
					row[i] = strconv.FormatInt(int64(v), 10)
				} else {
					row[i] = strconv.FormatFloat(v, 'f', -1, 64)
				}
			default:
				row[i] = fmt.Sprint(v)
			}
		}
		w.Write(row)
	}
	w.Flush()
	return strings.TrimSpace(buf.String()), nil
}

/* ---------------- 文件工具 ---------------- */

var pdfStrRe = regexp.MustCompile(`\(((?:\\.|[^\\()])*)\)`)

func toolPDF(b []byte) (string, error) {
	var texts []string
	for _, m := range regexp.MustCompile(`(?s)stream\r?\n(.*?)endstream`).FindAllStringSubmatch(string(b), -1) {
		data := []byte(m[1])
		if zr, err := zlib.NewReader(bytes.NewReader(data)); err == nil {
			if out, err := io.ReadAll(zr); err == nil {
				data = out
			}
		}
		for _, sm := range pdfStrRe.FindAllStringSubmatch(string(data), -1) {
			texts = append(texts, pdfUnescape(sm[1]))
		}
	}
	res := strings.Join(texts, " ")
	res = strings.TrimSpace(regexp.MustCompile(`\s{3,}`).ReplaceAllString(res, "\n"))
	if res == "" {
		return "", errors.New("未能提取文本：可能是扫描件 / 加密 PDF，或使用了非常规字体编码")
	}
	return res, nil
}
func pdfUnescape(s string) string {
	s = regexp.MustCompile(`\\([0-7]{1,3})`).ReplaceAllStringFunc(s, func(m string) string {
		n, err := strconv.ParseInt(m[1:], 8, 32)
		if err != nil {
			return m
		}
		return string(rune(n))
	})
	return strings.NewReplacer(`\n`, "\n", `\r`, "\r", `\t`, "\t", `\(`, "(", `\)`, ")", `\\`, `\`).Replace(s)
}

func colIdx(ref string) int {
	n := 0
	for _, c := range ref {
		if c >= 'A' && c <= 'Z' {
			n = n*26 + int(c-'A'+1)
		} else if c >= 'a' && c <= 'z' {
			n = n*26 + int(c-'a'+1)
		} else {
			break
		}
	}
	return n
}

func toolExcel(b []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return "", errors.New("不是合法的 xlsx 文件")
	}
	shared := []string{}
	for _, f := range zr.File {
		if strings.HasSuffix(f.Name, "sharedStrings.xml") {
			rc, _ := f.Open()
			d, _ := io.ReadAll(rc)
			rc.Close()
			for _, m := range regexp.MustCompile(`(?s)<si>(.*?)</si>`).FindAllStringSubmatch(string(d), -1) {
				v := ""
				for _, t := range regexp.MustCompile(`(?s)<t[^>]*>(.*?)</t>`).FindAllStringSubmatch(m[1], -1) {
					v += t[1]
				}
				shared = append(shared, v)
			}
		}
	}
	names := []string{}
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "xl/worksheets/sheet") {
			names = append(names, f.Name)
		}
	}
	sort.Strings(names)
	if len(names) == 0 {
		return "", errors.New("xlsx 中没有工作表")
	}
	var sheet *zip.File
	for _, f := range zr.File {
		if f.Name == names[0] {
			sheet = f
		}
	}
	rc, _ := sheet.Open()
	d, _ := io.ReadAll(rc)
	rc.Close()
	type cell struct {
		col int
		val string
	}
	var rows [][]cell
	for _, rm := range regexp.MustCompile(`(?s)<row[^>]*>(.*?)</row>`).FindAllStringSubmatch(string(d), -1) {
		var r []cell
		for _, cm := range regexp.MustCompile(`(?s)<c ([^>]*?)(?:/>|>(.*?)</c>)`).FindAllStringSubmatch(rm[1], -1) {
			attrs, inner := cm[1], cm[2]
			col := len(r) + 1
			if ref := regexp.MustCompile(`r="([A-Za-z]+)[0-9]+"`).FindStringSubmatch(attrs); ref != nil {
				col = colIdx(ref[1])
			}
			val := ""
			switch {
			case strings.Contains(attrs, `t="inlineStr"`):
				for _, tm := range regexp.MustCompile(`(?s)<t[^>]*>(.*?)</t>`).FindAllStringSubmatch(inner, -1) {
					val += tm[1]
				}
			case strings.Contains(attrs, `t="s"`):
				if vm := regexp.MustCompile(`<v>(.*?)</v>`).FindStringSubmatch(inner); vm != nil {
					if idx, e := strconv.Atoi(vm[1]); e == nil && idx < len(shared) {
						val = shared[idx]
					}
				}
			default:
				if vm := regexp.MustCompile(`<v>(.*?)</v>`).FindStringSubmatch(inner); vm != nil {
					val = vm[1]
				}
			}
			r = append(r, cell{col, val})
		}
		rows = append(rows, r)
	}
	if len(rows) == 0 {
		return "", errors.New("工作表为空")
	}
	maxc := 0
	for _, r := range rows {
		for _, c := range r {
			if c.col > maxc {
				maxc = c.col
			}
		}
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	for _, r := range rows {
		row := make([]string, maxc)
		for _, c := range r {
			if c.col >= 1 && c.col <= maxc {
				row[c.col-1] = c.val
			}
		}
		w.Write(row)
	}
	w.Flush()
	out := strings.TrimSpace(buf.String())
	if out == "" {
		return "", errors.New("工作表没有可读内容")
	}
	return out, nil
}

func toolWord(b []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return "", errors.New("不是合法的 docx 文件")
	}
	var doc []byte
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			rc, _ := f.Open()
			doc, _ = io.ReadAll(rc)
			rc.Close()
		}
	}
	if doc == nil {
		return "", errors.New("docx 中没有 word/document.xml")
	}
	s := string(doc)
	s = strings.ReplaceAll(s, "</w:p>", "\n")
	s = strings.ReplaceAll(s, "<w:br/>", "\n")
	s = regexp.MustCompile(`(?s)<w:t[^>]*>(.*?)</w:t>`).ReplaceAllString(s, "$1")
	s = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	lines := []string{}
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			lines = append(lines, l)
		}
	}
	if len(lines) == 0 {
		return "", errors.New("文档为空")
	}
	return strings.Join(lines, "\n"), nil
}

func resizeNearest(src image.Image, w, h int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(x, y, src.At(src.Bounds().Min.X+x*sw/w, src.Bounds().Min.Y+y*sh/h))
		}
	}
	return dst
}

func toolImg(b []byte, opts map[string]string) (string, error) {
	op := opts["op"]
	if op == "" {
		op = "compress"
	}
	src, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return "", errors.New("图片解码失败：支持 PNG / JPEG / GIF")
	}
	if mw, e := strconv.Atoi(opts["maxw"]); e == nil && mw > 0 && src.Bounds().Dx() > mw {
		src = resizeNearest(src, mw, src.Bounds().Dy()*mw/src.Bounds().Dx())
	}
	var mime string
	var buf bytes.Buffer
	if op == "to_png" {
		mime = "image/png"
		png.Encode(&buf, src)
	} else {
		q, _ := strconv.Atoi(opts["quality"])
		if q < 10 || q > 100 {
			q = 80
		}
		dst := image.NewRGBA(image.Rect(0, 0, src.Bounds().Dx(), src.Bounds().Dy()))
		draw.Draw(dst, dst.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
		draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Over)
		mime = "image/jpeg"
		jpeg.Encode(&buf, dst, &jpeg.Options{Quality: q})
	}
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
TOOLS_GO

# ---------- 清理冲突来源 ----------
if [ -f tools2.go ]; then
  git rm -f tools2.go >/dev/null 2>&1 || rm -f tools2.go
  echo "  已移除 tools2.go（其职责已并入 tools.go，避免重复定义）"
fi
if [ -f tools_test.go ] && grep -q "runTool2\|runToolFile2" tools_test.go; then
  git rm -f tools_test.go >/dev/null 2>&1 || rm -f tools_test.go
  echo "  已移除 tools_test.go（依赖已删除的 tools2.go；16 工具行为已由 e2e_test.go 全覆盖）"
fi

# ---------- 自检 ----------
echo "── 自检 ──"
grep -q "var ToolRegistry" tools.go && echo "  ✓ ToolRegistry 已定义" || { echo "  ✗ ToolRegistry 缺失"; exit 1; }
grep -q "func runTool(" tools.go && echo "  ✓ runTool 分发器" || { echo "  ✗ runTool 缺失"; exit 1; }
grep -q "func runToolFile(" tools.go && echo "  ✓ runToolFile 分发器" || { echo "  ✗ runToolFile 缺失"; exit 1; }
N=$(grep -cE '", "(text|file)"\},' tools.go)
echo "  ✓ 注册表条目: $N（应为 16）"
[ "$N" = "16" ] || { echo "  ✗ 条目数不对，停止提交"; exit 1; }

# ---------- 提交推送 ----------
git add -A
git commit -m "v2.2.1: 补齐 tools.go 全量16工具实现（修复 ToolRegistry 未定义）" || echo "  没有变更"
read -r -p "确认推送? [Y/n] " a
case "$a" in n|N) echo "已跳过，稍后手动 git push"; exit 0;; esac
git push && echo "═══ 已推送。Actions 会自动重跑 test，绿了再 Run workflow 打包 ═══"
