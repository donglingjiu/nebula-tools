package main

// toolkit.go — 传统工具集：CSV 清洗 / 图片处理 / Excel 数据剖析。
// 复用 filetools.go 的 fxUser/fxE/fxSend/fxFiles/fxOpts 助手，全内存处理。

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "image/gif"

	_ "golang.org/x/image/bmp"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"github.com/xuri/excelize/v2"
)

func hCsvClean(db *DB) http.HandlerFunc {
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
		if len(bs) != 1 {
			fxE(w, 400, "请上传一个 CSV 文件")
			return
		}
		opts := fxOpts(r)
		raw := bs[0]
		if opts["charset"] == "gbk" {
			if conv, err := io.ReadAll(transform.NewReader(bytes.NewReader(raw), simplifiedchinese.GBK.NewDecoder())); err == nil {
				raw = conv
			}
		}
		raw = bytes.TrimPrefix(raw, []byte("\xEF\xBB\xBF"))
		rd := csv.NewReader(bytes.NewReader(raw))
		rd.FieldsPerRecord = -1
		rows, err := rd.ReadAll()
		if err != nil {
			fxE(w, 422, "CSV 解析失败: "+err.Error())
			return
		}
		if len(rows) == 0 {
			fxE(w, 422, "CSV 为空")
			return
		}
		header, body := applyCsvExtras(rows[0], rows[1:], opts)
		colIdx := map[string]int{}
		for i, h := range header {
			colIdx[strings.TrimSpace(h)] = i
		}
		before := len(body)
		if opts["dedupe"] == "true" {
			seen := map[string]bool{}
			var kept [][]string
			for _, row := range body {
				k := strings.Join(row, "\x00")
				if !seen[k] {
					seen[k] = true
					kept = append(kept, row)
				}
			}
			body = kept
		}
		if fc := opts["filterCol"]; fc != "" {
			if idx, okc := colIdx[fc]; okc {
				fv := opts["filterVal"]
				mode := opts["filterMode"]
				var kept [][]string
				for _, row := range body {
					cell := ""
					if idx < len(row) {
						cell = row[idx]
					}
					keep := strings.Contains(cell, fv)
					if mode == "eq" {
						keep = cell == fv
					} else if mode == "ne" {
						keep = cell != fv
					}
					if keep {
						kept = append(kept, row)
					}
				}
				body = kept
			} else {
				fxE(w, 400, "筛选列不存在: "+fc)
				return
			}
		}
		if kc := opts["keepCols"]; kc != "" {
			var idxs []int
			for _, name := range strings.Split(kc, ",") {
				if idx, okc := colIdx[strings.TrimSpace(name)]; okc {
					idxs = append(idxs, idx)
				}
			}
			if len(idxs) > 0 {
				nh := make([]string, 0, len(idxs))
				for _, ix := range idxs {
					nh = append(nh, header[ix])
				}
				var kept [][]string
				for _, row := range body {
					nr := make([]string, 0, len(idxs))
					for _, ix := range idxs {
						c := ""
						if ix < len(row) {
							c = row[ix]
						}
						nr = append(nr, c)
					}
					kept = append(kept, nr)
				}
				header, body = nh, kept
			}
		}
		var buf bytes.Buffer
		cw := csv.NewWriter(&buf)
		cw.WriteAll(append([][]string{header}, body...))
		cw.Flush()
		out := buf.Bytes()
		if opts["charset"] == "gbk" {
			if conv, err := io.ReadAll(transform.NewReader(bytes.NewReader(out), simplifiedchinese.GBK.NewEncoder())); err == nil {
				out = conv
			}
		} else {
			out = append([]byte("\xEF\xBB\xBF"), out...)
		}
		db.AddHist(u.Username, "t:csv-clean", names[0], fmt.Sprintf("%d → %d 行", before, len(body)), time.Since(t0).Milliseconds())
		fxSend(w, "cleaned-"+names[0], out, "text/csv; charset=utf-8")
	}
}

func hImgProcess(db *DB) http.HandlerFunc {
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
		if len(bs) != 1 {
			fxE(w, 400, "请上传一张图片")
			return
		}
		src, _, err := image.Decode(bytes.NewReader(bs[0]))
		if err != nil {
			fxE(w, 422, "无法解析图片（支持 jpg/png/gif/webp/bmp）: "+err.Error())
			return
		}
		img := src
		opts := fxOpts(r)
		if ws := opts["width"]; ws != "" {
			if target, e := strconv.Atoi(ws); e == nil && target > 0 && target < src.Bounds().Dx() {
				hh := int(float64(target) * float64(src.Bounds().Dy()) / float64(src.Bounds().Dx()))
				dst := image.NewRGBA(image.Rect(0, 0, target, hh))
				draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
				img = dst
			}
		}
		format := strings.ToLower(opts["format"])
		if format == "" {
			format = "jpeg"
		}
		var out bytes.Buffer
		switch format {
		case "png":
			err = png.Encode(&out, img)
		case "jpg", "jpeg":
			q := 82
			if qs := opts["quality"]; qs != "" {
				if v, e := strconv.Atoi(qs); e == nil && v >= 1 && v <= 100 {
					q = v
				}
			}
			err = jpeg.Encode(&out, img, &jpeg.Options{Quality: q})
		default:
			fxE(w, 400, "输出格式仅支持 jpeg / png")
			return
		}
		if err != nil {
			fxE(w, 500, "图片编码失败: "+err.Error())
			return
		}
		ext := ".jpg"
		if format == "png" {
			ext = ".png"
		}
		base := names[0]
		for _, sfx := range []string{".png", ".jpeg", ".jpg", ".gif", ".webp", ".bmp"} {
			base = strings.TrimSuffix(base, sfx)
		}
		db.AddHist(u.Username, "t:img-process", names[0], fmt.Sprintf("%dx%d → %s %dKB", img.Bounds().Dx(), img.Bounds().Dy(), format, out.Len()/1024), time.Since(t0).Milliseconds())
		fxSend(w, "processed-"+base+ext, out.Bytes(), "image/"+format)
	}
}

func fptr(v float64) *float64 { return &v }

func hXlsxProfile(db *DB) http.HandlerFunc {
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
		if len(bs) != 1 {
			fxE(w, 400, "请上传一个 xlsx 文件")
			return
		}
		f, err := excelize.OpenReader(bytes.NewReader(bs[0]))
		if err != nil {
			fxE(w, 422, "xlsx 解析失败: "+err.Error())
			return
		}
		defer f.Close()
		sm := f.GetSheetMap()
		if len(sm) == 0 {
			fxE(w, 422, "无工作表")
			return
		}
		keys := make([]int, 0, len(sm))
		for k := range sm {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		sheet := sm[keys[0]]
		rows, err := f.GetRows(sheet)
		if err != nil {
			fxE(w, 422, "读取失败: "+err.Error())
			return
		}
		if len(rows) < 2 {
			fxE(w, 422, "数据行不足（需表头 + 至少 1 行数据）")
			return
		}
		if len(rows) > 50000 {
			rows = rows[:50000]
		}
		header, body := rows[0], rows[1:]
		cols := len(header)
		seen, dupSet := map[string]bool{}, map[string]bool{}
		for _, row := range body {
			k := strings.Join(row, "\x00")
			if seen[k] {
				dupSet[k] = true
			}
			seen[k] = true
		}
		type colStat struct {
			Col      string   `json:"col"`
			Type     string   `json:"type"`
			Missing  int      `json:"missing"`
			Distinct int      `json:"distinct"`
			Min      *float64 `json:"min,omitempty"`
			Max      *float64 `json:"max,omitempty"`
			Mean     *float64 `json:"mean,omitempty"`
		}
		stats := make([]colStat, 0, cols)
		for c := 0; c < cols; c++ {
			st := colStat{Col: header[c]}
			nums := []float64{}
			texts := map[string]bool{}
			for _, row := range body {
				cell := ""
				if c < len(row) {
					cell = strings.TrimSpace(row[c])
				}
				if cell == "" {
					st.Missing++
					continue
				}
				texts[cell] = true
				if v, e := strconv.ParseFloat(strings.ReplaceAll(cell, ",", ""), 64); e == nil {
					nums = append(nums, v)
				}
			}
			st.Distinct = len(texts)
			filled := len(body) - st.Missing
			switch {
			case filled > 0 && len(nums) == filled:
				st.Type = "numeric"
				mn, mx, sum := nums[0], nums[0], 0.0
				for _, v := range nums {
					if v < mn {
						mn = v
					}
					if v > mx {
						mx = v
					}
					sum += v
				}
				st.Min, st.Max, st.Mean = fptr(mn), fptr(mx), fptr(sum/float64(len(nums)))
			case filled > 0 && len(texts) <= 20 && len(texts) <= filled/3+1:
				st.Type = "categorical"
			default:
				st.Type = "text"
			}
			stats = append(stats, st)
		}
		opts := fxOpts(r)
		if opts["format"] == "md" {
			var sb strings.Builder
			sb.WriteString("# 数据剖析报告：" + names[0] + "\n\n")
			sb.WriteString(fmt.Sprintf("- 工作表：%s\n- 数据行：%d\n- 列数：%d\n- 重复行种类：%d\n\n", sheet, len(body), cols, len(dupSet)))
			sb.WriteString("| 列 | 类型 | 缺失 | 去重值 | 最小 | 最大 | 均值 |\n|---|---|---|---|---|---|---|\n")
			for _, s := range stats {
				mnS, mxS, meanS := "-", "-", "-"
				if s.Min != nil {
					mnS = fmt.Sprintf("%.4g", *s.Min)
				}
				if s.Max != nil {
					mxS = fmt.Sprintf("%.4g", *s.Max)
				}
				if s.Mean != nil {
					meanS = fmt.Sprintf("%.4g", *s.Mean)
				}
				sb.WriteString(fmt.Sprintf("| %s | %s | %d | %d | %s | %s | %s |\n", s.Col, s.Type, s.Missing, s.Distinct, mnS, mxS, meanS))
			}
			db.AddHist(u.Username, "t:xlsx-profile", names[0], fmt.Sprintf("%d 行 %d 列剖析报告", len(body), cols), time.Since(t0).Milliseconds())
			fxSend(w, strings.TrimSuffix(names[0], ".xlsx")+"-剖析报告.md", []byte(sb.String()), "text/markdown; charset=utf-8")
			return
		}
		db.AddHist(u.Username, "t:xlsx-profile", names[0], fmt.Sprintf("%d 行 %d 列剖析", len(body), cols), time.Since(t0).Milliseconds())
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]interface{}{"file": names[0], "sheet": sheet, "rows": len(body), "cols": cols, "duplicate_rows": len(dupSet), "columns": stats})
	}
}

func registerToolkit(mux *http.ServeMux, db *DB) {
	wrap := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if fxUser(db, r) == nil {
				fxE(w, 401, "请先登录后使用工具")
				return
			}
			h(w, r)
		}
	}
	mux.HandleFunc("POST /api/t/csv-clean", wrap(hCsvClean(db)))
	mux.HandleFunc("POST /api/t/img-process", wrap(hImgProcess(db)))
	mux.HandleFunc("POST /api/t/xlsx-profile", wrap(hXlsxProfile(db)))
}
