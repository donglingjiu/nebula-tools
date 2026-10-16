package model

import "time"

// User 用户表（role: 0 guest / 1 user / 2 vip / 3 admin）
type User struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Username     string     `gorm:"uniqueIndex;size:64" json:"username"`
	PasswordHash string     `json:"-"`
	Role         int        `json:"role"`
	PayTier      string     `gorm:"size:16;default:free" json:"pay_tier"`
	LastLogin    *time.Time `json:"last_login"`
	CreatedAt    time.Time  `json:"created_at"`
}

// UsageLog 模块调用日志（画像/热度数据源）
type UsageLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Path      string    `gorm:"size:128;index" json:"path"`
	Status    int       `json:"status"`
	LatencyMs int64     `json:"latency_ms"`
	UID       uint      `gorm:"index" json:"uid"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

// Alarm 告警记录
type Alarm struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	RuleID    string    `gorm:"size:32;index" json:"rule_id"`
	Severity  string    `gorm:"size:16" json:"severity"`
	Message   string    `gorm:"size:256" json:"message"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

// Setting 键值配置（模块开关/告警阈值持久层）
type Setting struct {
	Key   string `gorm:"primaryKey;size:64" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}

func All() []any {
	return []any{&User{}, &UsageLog{}, &Alarm{}, &Setting{}}
}
