package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var novaDB *sql.DB

func dbPath() string {
	if p := os.Getenv("NOVA_DB"); p != "" {
		return p
	}
	return filepath.Join("data", "nova.db")
}

func openDBAt(path string) (*sql.DB, error) {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	dsn := path + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)"
	d, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	d.SetMaxOpenConns(1)
	if err := dbMigrate(d); err != nil {
		d.Close()
		return nil, err
	}
	for _, f := range []string{path, path + "-wal", path + "-shm"} {
		_ = os.Chmod(f, 0o600)
	}
	return d, nil
}

func openDB() {
	d, err := openDBAt(dbPath())
	if err != nil {
		log.Printf("[db] init soft-fail: %v", err)
		return
	}
	novaDB = d
	log.Printf("[db] ready %s", dbPath())
}

var dbMigrateV1 = []string{
	"CREATE TABLE IF NOT EXISTS schema_migrations(version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL DEFAULT (datetime('now')))",
	"CREATE TABLE IF NOT EXISTS settings(key TEXT PRIMARY KEY, value TEXT)",
	"CREATE TABLE IF NOT EXISTS profiles(id TEXT PRIMARY KEY, name TEXT NOT NULL, base_url TEXT NOT NULL, model TEXT NOT NULL, api_key TEXT NOT NULL, enabled INTEGER NOT NULL DEFAULT 1, priority INTEGER NOT NULL DEFAULT 100, created_at INTEGER NOT NULL DEFAULT 0)",
	"CREATE TABLE IF NOT EXISTS conversations(id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT NOT NULL DEFAULT '新对话', profile_id TEXT, user_id TEXT, created_at TEXT NOT NULL DEFAULT (datetime('now')), updated_at TEXT NOT NULL DEFAULT (datetime('now')))",
	"CREATE TABLE IF NOT EXISTS messages(id INTEGER PRIMARY KEY AUTOINCREMENT, conversation_id INTEGER NOT NULL REFERENCES conversations(id) ON DELETE CASCADE, role TEXT NOT NULL CHECK(role IN ('user','assistant','system')), content TEXT NOT NULL, tokens INTEGER, branch_of INTEGER, created_at TEXT NOT NULL DEFAULT (datetime('now')))",
	"CREATE INDEX IF NOT EXISTS idx_messages_conv ON messages(conversation_id, id)",
	"CREATE INDEX IF NOT EXISTS idx_profiles_fallback ON profiles(enabled, priority)",
	"INSERT OR IGNORE INTO schema_migrations(version) VALUES (1)",
}

func dbMigrate(d *sql.DB) error {
	for i, s := range dbMigrateV1 {
		if _, err := d.Exec(s); err != nil {
			log.Printf("[db] step %d: %v", i, err)
			return err
		}
	}
	return nil
}
