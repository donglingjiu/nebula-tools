package main

// csvopts.go — CSV 清洗增强（对标 OpenRefine 高频集）+ toolkit 专用 opts 解析。
// applyCsvExtras：trim / 列改名 / 值替换，由 toolkit.go 的 hCsvClean 调用。
// tkOpts：自包含 opts 解析，支持嵌套 JSON，不依赖 filetools.go 的 fxOpts 合同。

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// tkOpts — 直读 multipart/URL 中的 opts 字段并宽容解析：
// 标量统一转字符串；嵌套对象序列化回 JSON 字符串存储，
// 供 applyCsvExtras 二次 Unmarshal。解析失败返回空 map，不报错。
func tkOpts(r *http.Request) map[string]string {
	out := map[string]string{}
	raw := r.FormValue("opts")
	if raw == "" {
		return out
	}
	var m map[string]interface{}
	if json.Unmarshal([]byte(raw), &m) != nil {
		return out
	}
	for k, v := range m {
		switch t := v.(type) {
		case string:
			out[k] = t
		case bool:
			if t {
				out[k] = "true"
			} else {
				out[k] = "false"
			}
		case float64:
			out[k] = strconv.FormatFloat(t, 'f', -1, 64)
		default:
			if b, e := json.Marshal(t); e == nil {
				out[k] = string(b)
			}
		}
	}
	return out
}

func applyCsvExtras(header []string, body [][]string, opts map[string]string) ([]string, [][]string) {
	if opts["trim"] == "true" {
		for i := range header {
			header[i] = strings.TrimSpace(header[i])
		}
		for ri := range body {
			for ci := range body[ri] {
				body[ri][ci] = strings.TrimSpace(body[ri][ci])
			}
		}
	}
	if rj := opts["rename"]; rj != "" {
		var m map[string]string
		if json.Unmarshal([]byte(rj), &m) == nil {
			for i, h := range header {
				if nv, ok := m[h]; ok {
					header[i] = nv
				}
			}
		}
	}
	if rj := opts["replace"]; rj != "" {
		var m map[string]map[string]string
		if json.Unmarshal([]byte(rj), &m) == nil {
			idx := map[string]int{}
			for i, h := range header {
				idx[h] = i
			}
			for name, rules := range m {
				if ci, ok := idx[name]; ok {
					for ri := range body {
						if ci < len(body[ri]) {
							for oldV, newV := range rules {
								body[ri][ci] = strings.ReplaceAll(body[ri][ci], oldV, newV)
							}
						}
					}
				}
			}
		}
	}
	return header, body
}
