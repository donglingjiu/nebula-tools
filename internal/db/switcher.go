package db

import (
	"github.com/glebarez/sqlite"
	"fmt"
	"os"
	"sync/atomic"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"nebula-tools/internal/model"
)

var switchState atomic.Bool

// SwitchToMySQL 运行时切换到 MySQL（幂等，切换中阻塞读写）
func SwitchToMySQL(host string, port int, user, pass, dbname string) error {
	if Driver() == "mysql" {
		return fmt.Errorf("当前已是 MySQL 模式")
	}
	switchState.Store(true)
	defer switchState.Store(false)

	// ① 先连接目标库验证四要素
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, port, dbname)
	target, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("连接 MySQL 失败: %w", err)
	}
	// ② 目标库执行 AutoMigrate
	if err := target.AutoMigrate(model.All()...); err != nil {
		return fmt.Errorf("目标库迁移失败: %w", err)
	}
	// ③ 原子替换全局句柄
	mu.Lock()
	old := gdb
	gdb = target
	driver = "mysql"
	mu.Unlock()
	_ = old // 旧连接由 gorm 连接池自行回收

	// ④ 持久化配置，下次启动直接 MySQL
	_ = os.WriteFile("config/db.yaml", []byte(fmt.Sprintf(
		"driver: mysql\nhost: %s\nport: %d\nuser: %s\npassword: %s\ndatabase: %s\n",
		host, port, user, pass, dbname)), 0o644)
	return nil
}

// SwitchToSQLite 回切 SQLite
func SwitchToSQLite() error {
	if Driver() == "sqlite" {
		return fmt.Errorf("当前已是 SQLite 模式")
	}
	switchState.Store(true)
	defer switchState.Store(false)

	path := "data/nebula.db"
	if d := os.Getenv("NEBULA_DATA_DIR"); d != "" {
		path = d + "/nebula.db"
	}
	target, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return err
	}
	if err := target.AutoMigrate(model.All()...); err != nil {
		return err
	}
	mu.Lock()
	gdb = target
	driver = "sqlite"
	mu.Unlock()
	_ = os.WriteFile("config/db.yaml", []byte("driver: sqlite\n"), 0o644)
	return nil
}
