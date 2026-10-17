package main

// csvopts.go — CSV 清洗增强（对标 OpenRefine 高频集）：trim / 列改名 / 值替换。
// 由 toolkit.go 的 hCsvClean 在解析后、构建列索引前调用。

import (
	"encoding/json"
	"strings"
)

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
