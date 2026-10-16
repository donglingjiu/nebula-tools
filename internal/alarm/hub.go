package alarm

import "sync"

// Event SSE 推送的告警事件
type Event struct {
	RuleID   string `json:"rule_id"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	At       string `json:"created_at"`
}

// Hub 告警广播中心（SSE 订阅）
type Hub struct {
	mu   sync.Mutex
	subs map[chan Event]bool
}

func NewHub() *Hub { return &Hub{subs: map[chan Event]bool{}} }

func (h *Hub) Subscribe() chan Event {
	ch := make(chan Event, 8)
	h.mu.Lock()
	h.subs[ch] = true
	h.mu.Unlock()
	return ch
}

func (h *Hub) Unsubscribe(ch chan Event) {
	h.mu.Lock()
	delete(h.subs, ch)
	h.mu.Unlock()
}

func (h *Hub) Broadcast(e Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subs {
		select {
		case ch <- e:
		default: // 订阅者满了就丢弃，避免阻塞
		}
	}
}
