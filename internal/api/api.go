package api

import (
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"nebula-tools/internal/buildinfo"
	"nebula-tools/internal/db"
	"nebula-tools/internal/metrics"
	"nebula-tools/internal/model"
	"nebula-tools/internal/module"
	"nebula-tools/internal/moduleimpl"
	"nebula-tools/web"
)

var startTime = time.Now()

// Router 挂载全部路由
func Router(r *gin.Engine) {
	moduleimpl.RegisterAll()

	// 版本标识（前台页脚/后台数据源）
	r.GET("/api/about", func(c *gin.Context) {
		b := buildinfo.Get()
		c.JSON(200, gin.H{"ok": true, "build": gin.H{
			"version": b.Version, "go": b.Go, "node": b.Node,
			"bundled": b.Bundled, "build_at": b.BuildAt}})
	})

	r.GET("/api/health", func(c *gin.Context) {
		b := buildinfo.Get()
		c.JSON(200, gin.H{"ok": true, "version": b.Version,
			"uptime_sec": int(time.Since(startTime).Seconds()),
			"db_driver":  db.Driver(), "switching": db.IsSwitching()})
	})

	r.POST("/api/auth/login", Login)

	// 首页宫格
	r.GET("/api/modules", func(c *gin.Context) {
		role := roleOf(c)
		var out []gin.H
		for _, m := range module.All() {
			if !module.Enabled(m.ID) || role < m.MinRole {
				continue
			}
			out = append(out, gin.H{"id": m.ID, "name": m.Name, "icon": m.Icon,
				"category": m.Category, "pay_tier": m.PayTier})
		}
		c.JSON(200, gin.H{"ok": true, "modules": out})
	})

	// 全局搜索
	r.GET("/api/search", func(c *gin.Context) {
		q := strings.ToLower(c.Query("q"))
		role := roleOf(c)
		var out []gin.H
		for _, m := range module.All() {
			if !module.Enabled(m.ID) || role < m.MinRole {
				continue
			}
			if q != "" && !strings.Contains(strings.ToLower(m.Name+" "+m.ID), q) {
				continue
			}
			out = append(out, gin.H{"id": m.ID, "name": m.Name, "icon": m.Icon,
				"category": m.Category})
		}
		c.JSON(200, gin.H{"ok": true, "modules": out})
	})

	// 模块路由（开关/角色运行时生效）
	for _, m := range module.All() {
		mod := m
		for _, rt := range mod.Routes {
			rt := rt
			r.Handle(rt.Method, rt.Path, func(c *gin.Context) {
				if !module.Enabled(mod.ID) {
					c.JSON(403, gin.H{"ok": false, "error": "模块已下线"})
					return
				}
				if roleOf(c) < mod.MinRole {
					c.JSON(403, gin.H{"ok": false, "error": "权限不足"})
					return
				}
				rt.Handler(c)
			})
		}
	}

	registerAdmin(r)
	serveFrontend(r)
}

// serveFrontend 托管 embed 前端（SPA 回退；API 未知路径返回 JSON 404）
func serveFrontend(r *gin.Engine) {
	sub, err := fs.Sub(web.Dist, "dist")
	if err != nil {
		return
	}
	fileServer := http.FileServer(http.FS(sub))
	r.NoRoute(func(c *gin.Context) {
		p := strings.TrimPrefix(c.Request.URL.Path, "/")
		if p == "" {
			p = "index.html"
		}
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(404, gin.H{"ok": false, "error": "接口不存在"})
			return
		}
		if _, err := fs.Stat(sub, p); err != nil {
			c.Request.URL.Path = "/"
		}
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}

// Usage 用量埋点中间件（滑动窗口指标 + 数据库日志）
func Usage() gin.HandlerFunc {
	return func(c *gin.Context) {
		st := time.Now()
		c.Next()
		lat := time.Since(st).Milliseconds()
		metrics.Hit(c.Writer.Status(), lat)
		uid := uidOf(c)
		go db.Write().Create(&model.UsageLog{Path: c.Request.URL.Path,
			Status: c.Writer.Status(), LatencyMs: lat, UID: uid})
	}
}
