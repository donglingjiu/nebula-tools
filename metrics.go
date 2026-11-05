package main

// NT-4: lightweight metrics per AutoDev llmops-metrics.md.
// stdlib only. MetricsWrap is outermost (counts everything incl. auth rejects).
// SSE paths are counted but excluded from latency stats (long-lived).

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

const metP = "/api/metrics"

type mPathStat struct {
	Count   int   `json:"count"`
	Errs    int   `json:"errs"`
	TotalMs int64 `json:"-"`
}

type mRecent struct {
	Path   string `json:"path"`
	Status int    `json:"status"`
	Ms     int64  `json:"ms"`
	TS     string `json:"ts"`
}

type mStore struct {
	mu       sync.Mutex
	start    time.Time
	reqTotal int
	reqErrs  int
	paths    map[string]*mPathStat
	recent   []mRecent
	win      []int64 // unix-ms of requests, sliding 60s for rps
	llmCalls int
	llmErrs  int
	llmMs    []int64 // cap 500 for p95
}

var met = newMStore()

func newMStore() *mStore {
	return &mStore{start: time.Now(), paths: map[string]*mPathStat{}, recent: make([]mRecent, 0, 64)}
}

func (s *mStore) record(path string, status int, ms int64) {
	sse := strings.HasSuffix(path, "/stream")
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	s.reqTotal++
	if status >= 400 {
		s.reqErrs++
	}
	st := s.paths[path]
	if st == nil {
		st = &mPathStat{}
		s.paths[path] = st
	}
	st.Count++
	if status >= 400 {
		st.Errs++
	}
	if !sse {
		st.TotalMs += ms
	}
	s.win = append(s.win, now.UnixNano()/int64(time.Millisecond))
	if len(s.win) > 2000 {
		s.win = s.win[len(s.win)-2000:]
	}
	if path == "/api/llm/chat" {
		s.llmCalls++
		if status >= 400 {
			s.llmErrs++
		}
		if !sse {
			s.llmMs = append(s.llmMs, ms)
			if len(s.llmMs) > 500 {
				s.llmMs = s.llmMs[len(s.llmMs)-500:]
			}
		}
	}
	s.recent = append(s.recent, mRecent{Path: path, Status: status, Ms: ms, TS: now.Format("15:04:05")})
	if len(s.recent) > 60 {
		s.recent = s.recent[len(s.recent)-60:]
	}
}

func (s *mStore) snapshot() map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	nowMs := time.Now().UnixNano() / int64(time.Millisecond)
	cut := nowMs - 60000
	rpm := 0
	for _, t := range s.win {
		if t >= cut {
			rpm++
		}
	}
	errRate := 0.0
	if s.reqTotal > 0 {
		errRate = float64(s.reqErrs) / float64(s.reqTotal) * 100
	}
	type topItem struct {
		Path  string  `json:"path"`
		Count int     `json:"count"`
		AvgMs float64 `json:"avgMs"`
		Errs  int     `json:"errs"`
	}
	tops := make([]topItem, 0, len(s.paths))
	for p, st := range s.paths {
		avg := 0.0
		if st.Count > 0 && !strings.HasSuffix(p, "/stream") {
			avg = float64(st.TotalMs) / float64(st.Count)
		}
		tops = append(tops, topItem{p, st.Count, avg, st.Errs})
	}
	sort.Slice(tops, func(i, j int) bool { return tops[i].Count > tops[j].Count })
	if len(tops) > 8 {
		tops = tops[:8]
	}
	llmAvg, llmP95 := 0.0, int64(0)
	if n := len(s.llmMs); n > 0 {
		var sum int64
		for _, v := range s.llmMs {
			sum += v
		}
		llmAvg = float64(sum) / float64(n)
		sorted := append([]int64(nil), s.llmMs...)
		sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
		llmP95 = sorted[n*95/100]
	}
	recent := append([]mRecent(nil), s.recent...)
	for i, j := 0, len(recent)-1; i < j; i, j = i+1, j-1 {
		recent[i], recent[j] = recent[j], recent[i]
	}
	return map[string]interface{}{
		"ok": true, "uptime_sec": int(time.Since(s.start).Seconds()),
		"req_total": s.reqTotal, "req_errors": s.reqErrs,
		"error_rate": errRate, "rpm": rpm,
		"llm": map[string]interface{}{"calls": s.llmCalls, "errors": s.llmErrs, "avg_ms": llmAvg, "p95_ms": llmP95},
		"top": tops, "recent": recent,
	}
}

type mWriter struct {
	http.ResponseWriter
	status int
}

func (w *mWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// MetricsWrap is the outermost wrapper: counts every request incl. auth rejects.
func MetricsWrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, metP) {
			metHJSON(w, r)
			return
		}
		mw := &mWriter{ResponseWriter: w, status: 200}
		t0 := time.Now()
		next.ServeHTTP(mw, r)
		met.record(r.URL.Path, mw.status, time.Since(t0).Milliseconds())
	})
}

func metHJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(met.snapshot())
}

// init registers on DefaultServeMux as fallback.
func init() {
	http.HandleFunc(metP, metHJSON)
}
