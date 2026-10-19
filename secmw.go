package main

import (
"net/http"
"strings"
"sync"
"time"
)

type rlBucket struct {
tokens float64
last   time.Time
}

type rateLimiter struct {
mu      sync.Mutex
buckets map[string]*rlBucket
max     float64
refill  float64
}

func newRateLimiter(burst float64, perMinute float64) *rateLimiter {
return &rateLimiter{buckets: make(map[string]*rlBucket), max: burst, refill: perMinute / 60.0}
}

func (l *rateLimiter) allow(key string) bool {
l.mu.Lock()
defer l.mu.Unlock()
now := time.Now()
b, ok := l.buckets[key]
if !ok {
b = &rlBucket{tokens: l.max, last: now}
l.buckets[key] = b
}
b.tokens += now.Sub(b.last).Seconds() * l.refill
if b.tokens > l.max {
b.tokens = l.max
}
b.last = now
if b.tokens < 1 {
return false
}
b.tokens--
if len(l.buckets) > 10000 {
for k, v := range l.buckets {
if now.Sub(v.last) > 10*time.Minute {
delete(l.buckets, k)
}
}
}
return true
}

func clientIP(r *http.Request) string {
xf := r.Header.Get("X-Forwarded-For")
if xf != "" {
parts := strings.Split(xf, ",")
return strings.TrimSpace(parts[0])
}
return strings.Split(r.RemoteAddr, ":")[0]
}

func secMiddleware(next http.Handler) http.Handler {
apiRL := newRateLimiter(20, 30)
genRL := newRateLimiter(120, 600)
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
p := r.URL.Path
if strings.HasPrefix(p, "/data/") || strings.HasSuffix(p, ".go") ||
strings.HasSuffix(p, ".bak") || strings.HasSuffix(p, ".sh") ||
strings.HasSuffix(p, "DANGEROUS") {
http.Error(w, "forbidden", http.StatusForbidden)
return
}
if p == "/robots.txt" {
w.Header().Set("Content-Type", "text/plain; charset=utf-8")
_, _ = w.Write([]byte("User-agent: *\nDisallow: /api/\nDisallow: /data/\nCrawl-delay: 5\n"))
return
}
rl := genRL
if strings.HasPrefix(p, "/api/llm/") {
rl = apiRL
}
if !rl.allow(clientIP(r)) {
w.Header().Set("Retry-After", "30")
http.Error(w, "too many requests", http.StatusTooManyRequests)
return
}
h := w.Header()
h.Set("X-Content-Type-Options", "nosniff")
h.Set("X-Frame-Options", "SAMEORIGIN")
h.Set("Referrer-Policy", "no-referrer")
next.ServeHTTP(w, r)
})
}
