package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"nebula-tools/internal/buildinfo"
	"nebula-tools/internal/db"
	"nebula-tools/internal/metrics"
	"nebula-tools/internal/model"
	"nebula-tools/internal/module"
)

func registerAdmin(r *gin.Engine) {
	g := r.Group("/api/admin", func(c *gin.Context) {
		if roleOf(c) < 3 {
			c.AbortWithStatusJSON(403, gin.H{"ok": false, "error": "需要管理员权限"})
			return
		}
		c.Next()
	})

	// 模块池管理
	g.GET("/modules", func(c *gin.Context) {
		var out []gin.H
		for _, m := range module.All() {
			out = append(out, gin.H{"id": m.ID, "name": m.Name, "icon": m.Icon,
				"category": m.Category, "enabled": module.Enabled(m.ID), "min_role": m.MinRole})
		}
		c.JSON(200, gin.H{"ok": true, "modules": out})
	})
	g.PATCH("/modules/:id", func(c *gin.Context) {
		var in struct {
			Enabled *bool `json:"enabled"`
		}
		if err := c.ShouldBindJSON(&in); err != nil || in.Enabled == nil {
			c.JSON(400, gin.H{"ok": false, "error": "需要 enabled 字段"})
			return
		}
		if !module.SetEnabled(c.Param("id"), *in.Enabled) {
			c.JSON(404, gin.H{"ok": false, "error": "模块不存在"})
			return
		}
		c.JSON(200, gin.H{"ok": true})
	})

	// 数据画像
	g.GET("/insight/overview", func(c *gin.Context) {
		var users, unread, reqs, errs int64
		db.Write().Model(&model.User{}).Count(&users)
		db.Write().Model(&model.Alarm{}).Where("read = ?", false).Count(&unread)
		today := time.Now().Format("2006-01-02")
		db.Write().Model(&model.UsageLog{}).Where("date(created_at) = ?", today).Count(&reqs)
		db.Write().Model(&model.UsageLog{}).Where("date(created_at) = ? AND status >= 500", today).Count(&errs)
		_, _, avgMs, p95 := metrics.Snapshot()
		c.JSON(200, gin.H{"ok": true, "data": gin.H{
			"users": users, "requests_today": reqs, "error_today": errs,
			"avg_latency_ms": avgMs, "p95_latency_ms": p95, "unread_alarms": unread}})
	})
	g.GET("/insight/modules", func(c *gin.Context) {
		type row struct {
			Module string `json:"module"`
			Count  int64  `json:"count"`
		}
		var rows []row
		db.Write().Model(&model.UsageLog{}).
			Select("path as module, count(*) as count").
			Where("created_at > ?", time.Now().AddDate(0, 0, -7)).
			Group("path").Order("count desc").Limit(10).Scan(&rows)
		c.JSON(200, gin.H{"ok": true, "heat": rows})
	})

	// 告警中心
	g.GET("/alarms", func(c *gin.Context) {
		var list []model.Alarm
		db.Write().Order("id desc").Limit(50).Find(&list)
		c.JSON(200, gin.H{"ok": true, "alarms": list})
	})
	g.POST("/alarms/read", func(c *gin.Context) {
		db.Write().Model(&model.Alarm{}).Where("read = ?", false).Update("read", true)
		c.JSON(200, gin.H{"ok": true})
	})

	// 版本/依赖标识
	g.GET("/deps", func(c *gin.Context) {
		b := buildinfo.Get()
		c.JSON(200, gin.H{"ok": true, "pkgs": []gin.H{
			{"name": "Go（构建）", "current": b.Go},
			{"name": "Node（前端构建）", "current": b.Node},
			{"name": "MySQL（随包）", "current": b.Bundled["mysql"]},
			{"name": "Python（随包）", "current": b.Bundled["python"]},
		}})
	})

	// 一键切换数据库（SQLite ↔ MySQL）
	g.POST("/db/switch", func(c *gin.Context) {
		var in struct {
			Driver   string `json:"driver"`
			Host     string `json:"host"`
			Port     int    `json:"port"`
			User     string `json:"user"`
			Password string `json:"password"`
			Database string `json:"database"`
		}
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(400, gin.H{"ok": false, "error": "参数错误"})
			return
		}
		switch in.Driver {
		case "mysql":
			if in.Port == 0 {
				in.Port = 3306
			}
			if err := db.SwitchToMySQL(in.Host, in.Port, in.User, in.Password, in.Database); err != nil {
				c.JSON(400, gin.H{"ok": false, "error": err.Error()})
				return
			}
		case "sqlite":
			if err := db.SwitchToSQLite(); err != nil {
				c.JSON(400, gin.H{"ok": false, "error": err.Error()})
				return
			}
		default:
			c.JSON(400, gin.H{"ok": false, "error": "driver 须为 mysql/sqlite"})
			return
		}
		c.JSON(200, gin.H{"ok": true, "driver": db.Driver()})
	})

	// SSE 实时告警推送
	g.GET("/events", func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		flusher, ok := c.Writer.(http.Flusher)
		if !ok {
			c.Status(500)
			return
		}
		ch := Hub.Subscribe()
		defer Hub.Unsubscribe(ch)
		for {
			select {
			case ev := <-ch:
				b, _ := json.Marshal(ev)
				fmt.Fprintf(c.Writer, "data: %s\n\n", b)
				flusher.Flush()
			case <-c.Request.Context().Done():
				return
			}
		}
	})
}
