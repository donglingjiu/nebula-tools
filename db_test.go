package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenDBAtMigrate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nova.db")
	d, err := openDBAt(path)
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := d.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('schema_migrations','settings','profiles','conversations','messages')").Scan(&n); err != nil || n != 5 {
		t.Fatalf("tables=%d err=%v", n, err)
	}
	var idx int
	if err := d.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name IN ('idx_messages_conv','idx_profiles_fallback')").Scan(&idx); err != nil || idx != 2 {
		t.Fatalf("indexes=%d err=%v", idx, err)
	}
	var ver int
	if err := d.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version=1").Scan(&ver); err != nil || ver != 1 {
		t.Fatalf("version=%d err=%v", ver, err)
	}
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}
	d2, err := openDBAt(path)
	if err != nil {
		t.Fatal("reopen:", err)
	}
	defer d2.Close()
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Fatalf("perm=%o want 600", fi.Mode().Perm())
	}
}
