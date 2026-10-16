package metrics

import (
	"sync"
	"time"
)

// SlidingWindow 滑动窗口指标（错误率/延迟/请求量）
type SlidingWindow struct {
	mu      sync.Mutex
	entries []entry
	window  time.Duration
}

type entry struct {
	at     time.Time
	status int
	ms     int64
}

func New(window time.Duration) *SlidingWindow {
	return &SlidingWindow{window: window}
}

func (w *SlidingWindow) Hit(status int, ms int64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.entries = append(w.entries, entry{time.Now(), status, ms})
	cutoff := time.Now().Add(-w.window)
	i := 0
	for i < len(w.entries) && w.entries[i].at.Before(cutoff) {
		i++
	}
	w.entries = w.entries[i:]
}

func (w *SlidingWindow) Snapshot() (total, errors int, avgMs, p95Ms int64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(w.entries) == 0 {
		return 0, 0, 0, 0
	}
	var sumMs int64
	latencies := make([]int64, 0, len(w.entries))
	for _, e := range w.entries {
		sumMs += e.ms
		if e.status >= 500 {
			errors++
		}
		latencies = append(latencies, e.ms)
	}
	for i := 0; i < len(latencies); i++ {
		for j := i + 1; j < len(latencies); j++ {
			if latencies[j] < latencies[i] {
				latencies[i], latencies[j] = latencies[j], latencies[i]
			}
		}
	}
	p95Idx := int(float64(len(latencies)) * 0.95)
	if p95Idx >= len(latencies) {
		p95Idx = len(latencies) - 1
	}
	return len(w.entries), errors, sumMs / int64(len(w.entries)), latencies[p95Idx]
}

var global = New(5 * time.Minute)

func Hit(status int, ms int64)  { global.Hit(status, ms) }
func Snapshot() (int, int, int64, int64) { return global.Snapshot() }
