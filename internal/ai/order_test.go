package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"nova/internal/admin"
	"nova/internal/auth"
)

func TestOrderPaywall(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(200)
			fl := w.(http.Flusher)
			if body["stream"] == true {
				for _, tok := range []string{"购", "OK"} {
					b, _ := json.Marshal(map[string]any{
						"choices": []map[string]any{
							{"delta": map[string]string{"content": tok}}}})
					fmt.Fprintf(w, "data: %s\n\n", b)
					fl.Flush()
				}
				fmt.Fprint(w, "data: [DONE]\n\n")
				fl.Flush()
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{
					{"message": map[string]string{"content": "购买后输出OK"}}}})
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
	do(t, ts, "PUT", "/api/admin/llm/profiles/"+m3["id"].(string), super, `{"is_default":true}`)
	do(t, ts, "POST", "/api/v2/auth/register", "", `{"email":"ua@x.com","password":"12345678"}`)
	do(t, ts, "POST", "/api/v2/auth/register", "", `{"email":"ub@x.com","password":"12345678"}`)
	_, ma, _ := do(t, ts, "POST", "/api/v2/auth/login", "", `{"email":"ua@x.com","password":"12345678"}`)
	ua := ma["token"].(string)
	_, mb, _ := do(t, ts, "POST", "/api/v2/auth/login", "", `{"email":"ub@x.com","password":"12345678"}`)
	ub := mb["token"].(string)

	_, mf, _ := do(t, ts, "GET", "/api/admin/features", super, "")
	var afid string
	for _, it := range mf["features"].([]any) {
		f := it.(map[string]any)
		if f["key"] == "ai_chat" {
			afid, _ = f["id"].(string)
		}
	}
	do(t, ts, "PUT", "/api/admin/features/"+afid, super, `{"is_paid":true,"price":7.7}`)
	do(t, ts, "PUT", "/api/admin/ai/ops", super, `{"paid_mode":true}`)

	MSGS := `{"messages":[{"role":"user","content":"hi"}]}`
	if c, _, _ := do(t, ts, "POST", "/api/ai/chat", ua, MSGS); c != 402 {
		t.Fatalf("未购应402: %d", c)
	}
	_, mo, _ := do(t, ts, "POST", "/api/ai/order", ua, `{"feature_key":"ai_chat"}`)
	oid, _ := mo["id"].(string)
	if oid == "" || mo["price"].(float64) != 7.7 {
		t.Fatalf("order: %v", mo)
	}
	if c, _, _ := do(t, ts, "POST", "/api/ai/order/pay", ub,
		fmt.Sprintf(`{"id":"%s"}`, oid)); c != 404 {
		t.Fatalf("B付A单应404: %d", c)
	}
	if c, mp, _ := do(t, ts, "POST", "/api/ai/order/pay", ua,
		fmt.Sprintf(`{"id":"%s"}`, oid)); c != 200 || mp["status"] != "paid" {
		t.Fatalf("pay: %d %v", c, mp)
	}
	if c, _, _ := do(t, ts, "POST", "/api/ai/chat", ua, MSGS); c != 200 {
		t.Fatalf("已购应放行200: %d", c)
	}
	if c, _, _ := do(t, ts, "POST", "/api/ai/chat", ub, MSGS); c != 402 {
		t.Fatalf("UB未购仍应402: %d", c)
	}
	_, ml, _ := do(t, ts, "GET", "/api/ai/orders", ua, "")
	n := len(ml["orders"].([]any))
	if n != 1 {
		t.Fatalf("UA订单应1: %d", n)
	}
	do(t, ts, "PUT", "/api/admin/features/"+afid, super, `{"is_paid":false}`)
	do(t, ts, "PUT", "/api/admin/ai/ops", super, `{"paid_mode":false}`)
}
