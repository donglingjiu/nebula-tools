package alarm

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"nebula-tools/internal/db"
	"nebula-tools/internal/metrics"
	"nebula-tools/internal/model"
)

// Rule 告警规则
type Rule struct {
	ID        string
	Threshold float64
	Severity  string
	Message   string
	Cooldown  time.Duration
}

// Engine 规则引擎（周期评估 + 即时规则）
type Engine struct {
	hub      *Hub
	cooldown map[string]time.Time
	mu       sync.Mutex
}

// Start 启动周期评估协程
func Start(hub *Hub) *Engine {
	e := &Engine{hub: hub, cooldown: map[string]time.Time{}}
	go e.loop()
	log.Println("[alarm] 规则引擎已启动（30s/轮）")
	return e
}

func (e *Engine) fire(r Rule, val float64) {
	e.mu.Lock()
	if t, ok := e.cooldown[r.ID]; ok && time.Since(t) < r.Cooldown {
		e.mu.Unlock()
		return
	}
	e.cooldown[r.ID] = time.Now()
	e.mu.Unlock()

	db.Write().Create(&model.Alarm{RuleID: r.ID, Severity: r.Severity, Message: r.Message})
	e.hub.Broadcast(Event{RuleID: r.ID, Severity: r.Severity, Message: r.Message,
		At: time.Now().Format(time.RFC3339)})
	log.Printf("[alarm] %s: %s", r.ID, r.Message)
}

// Evaluate 一轮规则评估
func (e *Engine) Evaluate() {
	total, errs, avgMs, p95Ms := metrics.Snapshot()
	if total > 20 {
		rate := float64(errs) / float64(total)
		if rate > 0.05 {
			e.fire(Rule{"err_rate", rate, "critical",
				fmt.Sprintf("HTTP 错误率超过 5%%（当前 %.1f%%）", rate*100),
				5 * time.Minute}, rate)
		}
	}
	if p95Ms > 800 {
		e.fire(Rule{"slow_api", float64(p95Ms), "warning",
			fmt.Sprintf("接口 P95 延迟超 800ms（当前 %dms）", p95Ms),
			5 * time.Minute}, float64(p95Ms))
	}
	if db.Driver() == "sqlite" {
		if fi, err := os.Stat("data/nebula.db"); err == nil && fi.Size() > 1<<30 {
			e.fire(Rule{"db_size", float64(fi.Size()), "warning",
				"SQLite 超过 1GB，建议后台切换 MySQL", time.Hour}, float64(fi.Size()))
		}
	}
}

func (e *Engine) loop() {
	t := time.NewTicker(30 * time.Second)
	for range t.C {
		e.Evaluate()
	}
}

// BruteLogin 登录爆破即时告警（第 5 次失败触发）
func (e *Engine) BruteLogin(ip string, fails int) {
	if fails == 5 {
		e.fire(Rule{"brute_login", float64(fails), "critical",
			"检测到疑似登录爆破: " + ip, 10 * time.Minute}, float64(fails))
	}
}
