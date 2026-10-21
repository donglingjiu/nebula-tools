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
