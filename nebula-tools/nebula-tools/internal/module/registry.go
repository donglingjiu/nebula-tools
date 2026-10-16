package module

import (
	"sync"

	"github.com/gin-gonic/gin"
)

// Route 模块自带的 HTTP 路由
type Route struct {
	Method  string
	Path    string
	Handler gin.HandlerFunc
}

// Module 模块定义
type Module struct {
	ID          string
	Name        string
	Icon        string
	Category    string // dev / file / data / home
	MinRole     int    // 0 guest / 1 user / 3 admin
	PayTier     string // free / pro
	AISupported bool
	Routes      []Route
}

var (
	mu      sync.RWMutex
	all     []*Module
	flags   = map[string]bool{}
	loaded  bool
)

// Register 注册模块（幂等）
func Register(m *Module) {
	mu.Lock()
	defer mu.Unlock()
	for _, e := range all {
		if e.ID == m.ID {
			return
		}
	}
	all = append(all, m)
	flags[m.ID] = true
}

// All 返回全部模块快照
func All() []*Module {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]*Module, len(all))
	copy(out, all)
	return out
}

// Enabled 模块是否启用
func Enabled(id string) bool {
	mu.RLock()
	defer mu.RUnlock()
	return flags[id]
}

// SetEnabled 运行时开关（返回 false 表示模块不存在）
func SetEnabled(id string, on bool) bool {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := flags[id]; !ok {
		return false
	}
	flags[id] = on
	return true
}

// Visible 按角色+开关过滤可见模块
func Visible(role int) []*Module {
	var out []*Module
	for _, m := range All() {
		if flags[m.ID] && role >= m.MinRole {
			out = append(out, m)
		}
	}
	return out
}
