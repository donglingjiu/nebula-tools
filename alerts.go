package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// sendAlert pushes a tool-failure notice to the admin-configured webhook.
// Best-effort: 3s timeout, errors logged, never fails the request.
func sendAlert(webhook, tool, user, errMsg string) {
	if webhook == "" {
		return
	}
	payload, _ := json.Marshal(map[string]string{
		"event": "tool_failed", "tool": tool, "user": user,
		"error": errMsg, "time": time.Now().Format(time.RFC3339),
	})
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Post(webhook, "application/json", bytes.NewReader(payload))
	if err != nil {
		log.Printf("[alert] push failed: %v", err)
		return
	}
	resp.Body.Close()
	log.Printf("[alert] pushed tool=%s user=%s status=%d", tool, user, resp.StatusCode)
}

// notifySend posts an alert to one channel (webhook/dingtalk/feishu/email).
// webhook keeps the legacy structured JSON for e2e compat; others wrap text.
func notifySend(url, channel string, ev map[string]string) {
	if url == "" {
		return
	}
	var body []byte
	switch channel {
	case "dingtalk":
		b, _ := json.Marshal(map[string]any{"msgtype": "text", "text": map[string]any{"content": ev["tool"] + " failed: " + ev["error"]}})
		body = b
	case "feishu":
		b, _ := json.Marshal(map[string]any{"msg_type": "text", "content": map[string]any{"text": ev["tool"] + " failed: " + ev["error"]}})
		body = b
	case "email":
		log.Printf("[alert] email占位 to=%s: %s %s", url, ev["tool"], ev["error"])
		return
	default:
		body, _ = json.Marshal(ev)
	}
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("[alert] push failed (%s): %v", channel, err)
		return
	}
	resp.Body.Close()
	log.Printf("[alert] pushed ch=%s tool=%s status=%d", channel, ev["tool"], resp.StatusCode)
}

// alertFanout dispatches a tool-failure alert to all configured channels.
func alertFanout(db *DB, tool, user, errMsg string) {
	ev := map[string]string{
		"event": "tool_failed", "tool": tool, "user": user,
		"error": errMsg, "time": time.Now().Format(time.RFC3339),
	}
	for _, ch := range []struct{ key, name string }{
		{"alert_webhook", "webhook"},
		{"alert_dingtalk", "dingtalk"},
		{"alert_feishu", "feishu"},
		{"alert_email", "email"},
	} {
		go notifySend(db.GetSettingSafe(ch.key), ch.name, ev)
	}
}
