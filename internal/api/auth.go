package api

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"nebula-tools/internal/alarm"
	"nebula-tools/internal/db"
	"nebula-tools/internal/model"
)

var jwtSecret = []byte("nebula-dev-secret-change-me")
var Hub = alarm.NewHub()
var engine *alarm.Engine

// SetJWTSecret 生产部署时从环境变量注入
func SetJWTSecret(s []byte) { jwtSecret = s }

// StartAlarm 启动告警规则引擎（main 调用一次）
func StartAlarm() *alarm.Engine {
	engine = alarm.Start(Hub)
	return engine
}

func hashPassword(pw string) string {
	h := sha256.Sum256([]byte("nebula-salt-v1:" + pw))
	return hex.EncodeToString(h[:])
}

// BootstrapAdmin 首次启动自动创建管理员
func BootstrapAdmin() {
	var n int64
	db.Write().Model(&model.User{}).Count(&n)
	if n == 0 {
		db.Write().Create(&model.User{Username: "admin",
			PasswordHash: hashPassword("admin123"), Role: 3, PayTier: "enterprise"})
		log.Println("[nebula] 已创建管理员 admin / admin123（请尽快改密）")
	}
}

func issueToken(u *model.User) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": u.ID, "username": u.Username, "role": u.Role,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}).SignedString(jwtSecret)
}

func claims(c *gin.Context) jwt.MapClaims {
	if v, ok := c.Get("claims"); ok {
		cl, _ := v.(jwt.MapClaims)
		return cl
	}
	return nil
}

func roleOf(c *gin.Context) int {
	if cl := claims(c); cl != nil {
		if r, ok := cl["role"].(float64); ok {
			return int(r)
		}
	}
	return 0
}

func uidOf(c *gin.Context) uint {
	if cl := claims(c); cl != nil {
		if v, ok := cl["sub"].(float64); ok {
			return uint(v)
		}
	}
	return 0
}

// Auth JWT 解析中间件（Authorization 头 或 ?token= 均可）
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" {
			if q := c.Query("token"); q != "" {
				h = "Bearer " + q
			}
		}
		if strings.HasPrefix(h, "Bearer ") {
			tok, err := jwt.Parse(strings.TrimPrefix(h, "Bearer "),
				func(*jwt.Token) (any, error) { return jwtSecret, nil })
			if err == nil && tok.Valid {
				if cl, ok := tok.Claims.(jwt.MapClaims); ok {
					c.Set("claims", cl)
				}
			}
		}
		c.Next()
	}
}

var failMu sync.Mutex
var fails = map[string]int{}

func recordFail(ip string) int {
	failMu.Lock()
	defer failMu.Unlock()
	fails[ip]++
	return fails[ip]
}

// Login 登录（失败计数 → 第 5 次触发爆破告警）
func Login(c *gin.Context) {
	var in struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(400, gin.H{"ok": false, "error": "参数错误"})
		return
	}
	var u model.User
	if err := db.Write().Where("username = ?", in.Username).First(&u).Error; err != nil ||
		u.PasswordHash != hashPassword(in.Password) {
		ip := c.ClientIP()
		n := recordFail(ip)
		if engine != nil {
			engine.BruteLogin(ip, n)
		}
		c.JSON(401, gin.H{"ok": false, "error": "用户名或密码错误"})
		return
	}
	now := time.Now()
	db.Write().Model(&u).Update("last_login", &now)
	t, _ := issueToken(&u)
	c.JSON(200, gin.H{"ok": true, "token": t, "role": u.Role})
}
