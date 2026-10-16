package db

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/glebarez/sqlite" // 纯 Go SQLite，支持 CGO_ENABLED=0 交叉编译
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"nebula-tools/internal/model"
)

var (
	mu     sync.RWMutex
	gdb    *gorm.DB
	driver = "sqlite"
)

// Config 数据库配置（手写 key:value 解析，零 yaml 依赖）
type Config struct {
	Driver   string
	Host     string
	Port     int
	User     string
	Password string
	Database string
	Path     string
}

// Init 初始化数据库并自动迁移
func Init() error {
	cfg := Load()
	var err error
	switch cfg.Driver {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
		gdb, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		driver = "mysql"
	default:
		path := cfg.Path
		if path == "" {
			path = "data/nebula.db"
		}
		_ = os.MkdirAll(filepath.Dir(path), 0o755)
		gdb, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
		driver = "sqlite"
	}
	if err != nil {
		return err
	}
	return gdb.AutoMigrate(model.All()...)
}

// Load 读取 config/db.yaml；支持 NEBULA_DATA_DIR 重定向（容器卷挂载）
func Load() *Config {
	cfg := &Config{Driver: "sqlite"}
	if b, err := os.ReadFile("config/db.yaml"); err == nil {
		for _, line := range strings.Split(string(b), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			kv := strings.SplitN(line, ":", 2)
			if len(kv) != 2 {
				continue
			}
			k, v := strings.TrimSpace(kv[0]), strings.Trim(strings.TrimSpace(kv[1]), `"'`)
			switch k {
			case "driver":
				cfg.Driver = v
			case "host":
				cfg.Host = v
			case "port":
				cfg.Port, _ = strconv.Atoi(v)
			case "user":
				cfg.User = v
			case "password":
				cfg.Password = v
			case "database":
				cfg.Database = v
			case "path":
				cfg.Path = v
			}
		}
	}
	if dir := os.Getenv("NEBULA_DATA_DIR"); dir != "" && cfg.Driver == "sqlite" && cfg.Path == "" {
		_ = os.MkdirAll(dir, 0o755)
		cfg.Path = filepath.Join(dir, "nebula.db")
	}
	return cfg
}

// G 获取全局 gorm 实例（读写锁保护，支持运行时切换）
func G() *gorm.DB { mu.RLock(); defer mu.RUnlock(); return gdb }

// Driver 当前驱动名
func Driver() string { mu.RLock(); defer mu.RUnlock(); return driver }

// IsSwitching 是否正在切换（切换过程阻塞新请求）
func IsSwitching() bool { return switchState.Load() }

// Write 获取写句柄（切换时短暂锁住）
func Write() *gorm.DB { mu.RLock(); defer mu.RUnlock(); return gdb }
