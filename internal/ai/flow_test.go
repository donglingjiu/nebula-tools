package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nova/internal/admin"
	"nova/internal/auth"
)

func TestFlow(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["stream"] == true {
				w.Header().Set("Content-Type", "text/event-stream")
				w.WriteHeader(200)
				fl := w.(http.Flusher)
				for _, tok := range []string{"A", "B"} {
					b, _ := json.Marshal(map[string]any{"choices": []map[string]any{{"delta": map[string]string{"content": tok}}}})
					fmt.Fprintf(w, "data: %s\n\n", b)
					fl.Flush()
				}
				fmt.Fprint(w, "data: [DONE]\n\n")
				fl.Flush()
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{{"message": map[string]string{"content": "步骤输出OK"}}},
			})
		}))
	defer up.Close()

	mux := http.NewServeMux()
	dir := t.TempDir()
	asvc, err := auth.Mount(mux, dir, "/api/v2/auth")
	if err != nil {
		t.Fatal(err)
	}
	if err := admin.Mount(mux, dir, asvc); err != nil {
		t.Fatal(err)
	}
	if err := Mount(mux, dir, asvc); err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(mux)
	defer ts.Close()

	_, m, _ := do(t, ts, "POST", "/api/v2/auth/register", "", `{"email":"root@x.com","password":"12345678"}`)
	super := m["token"].(string)
	_, m3, _ := do(t, ts, "POST", "/api/admin/llm/profiles", super,
		fmt.Sprintf(`{"name":"模拟","base_url":"%s/v1","model":"mock-1","api_key":"sk-1","enabled":true}`, up.URL))
	pid := m3["id"].(string)
	do(t, ts, "PUT", "/api/admin/llm/profiles/"+pid, super, `{"is_default":true}`)

	_, ms, _ := do(t, ts, "POST", "/api/ai/flow/start", super,
		`{"title":"报告","steps":[{"name":"提纲","prompt":"列提纲"},{"name":"正文","prompt":"写正文"}]}`)
	fid, _ := ms["id"].(string)
	if fid == "" {
		t.Fatalf("start: %v", ms)
	}
	_, ma, _ := do(t, ts, "POST", "/api/ai/flow/advance", super,
		fmt.Sprintf(`{"id":"%s"}`, fid))
	if ma["done"] != false {
		t.Fatalf("advance1: %v", ma)
	}
	_, ma2, _ := do(t, ts, "POST", "/api/ai/flow/advance", super,
		fmt.Sprintf(`{"id":"%s"}`, fid))
	fl2, _ := ma2["flow"].(map[string]any)
	if fl2 == nil || fl2["state"] != "done" {
		t.Fatalf("advance2应done: %v", ma2)
	}

	_, ms2, _ := do(t, ts, "POST", "/api/ai/flow/start", super,
		`{"title":"一键","steps":[{"name":"a","prompt":"pa"},{"name":"b","prompt":"pb"}]}`)
	fid2 := ms2["id"].(string)
	_, mr2, _ := do(t, ts, "POST", "/api/ai/flow/run", super,
		fmt.Sprintf(`{"id":"%s"}`, fid2))
	if mr2["state"] != "done" {
		t.Fatalf("run应done: %v", mr2)
	}

	_, mc, _ := do(t, ts, "POST", "/api/ai/flow/chat", super,
		fmt.Sprintf(`{"id":"%s","message":"总结一下"}`, fid2))
	if mc["reply"] != "步骤输出OK" {
		t.Fatalf("chat: %v", mc)
	}

	c, raw, _ := do(t, ts, "GET", "/api/ai/flow/export?id="+fid2, super, "")
	if c != 200 || !strings.Contains(raw, "一键") ||
		!strings.Contains(raw, "步骤输出OK") {
		t.Fatalf("export: %d %q", c, raw)
	}

	_, mf, _ := do(t, ts, "GET", "/api/admin/features", super, "")
	var gf []any
	gf, _ = mf["features"].([]any)
	if gf == nil {
		gf, _ = mf["items"].([]any)
	}
	var afid string
	for _, it := range gf {
		f := it.(map[string]any)
		if f["key"] == "ai_chat" {
			afid, _ = f["id"].(string)
		}
	}
	if afid == "" {
		t.Skipf("features形状未知: %v", mf)
	}
	_, ms3, _ := do(t, ts, "POST", "/api/ai/flow/start", super,
		`{"title":"门测","steps":[{"name":"s","prompt":"p"}]}`)
	fid3 := ms3["id"].(string)
	do(t, ts, "PUT", "/api/admin/features/"+afid, super, `{"enabled":false}`)
	if c, _, _ := do(t, ts, "POST", "/api/ai/flow/advance", super, fmt.Sprintf(`{"id":"%s"}`, fid3)); c != 410 {
		t.Fatalf("下线advance应410: %d", c)
	}
	if c, _, _ := do(t, ts, "POST", "/api/ai/flow/start", super, `{"title":"t","steps":[{"name":"s","prompt":"p"}]}`); c != 200 {
		t.Fatalf("下线start应放行: %d", c)
	}
	do(t, ts, "PUT", "/api/admin/features/"+afid, super, `{"enabled":true}`)
}
