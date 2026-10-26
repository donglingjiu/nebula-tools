package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProfilesMigrateAndCRUD(t *testing.T) {
	dir := t.TempDir()
	d, err := openDBAt(filepath.Join(dir, "nova.db"))
	if err != nil {
		t.Fatal(err)
	}
	saved := novaDB
	novaDB = d
	defer func() { novaDB = saved; d.Close() }()

	jsonPath := filepath.Join(dir, "llm_profiles.json")
	legacy := map[string]interface{}{
		"profiles": []map[string]interface{}{
			{"id": "p2", "name": "B", "baseUrl": "https://b.example.com", "model": "m2", "apiKey": encSecret("sk-bbb"), "enabled": false, "createdAt": int64(2000)},
			{"id": "p1", "name": "A", "baseUrl": "https://a.example.com", "model": "m1", "apiKey": encSecret("sk-aaa"), "enabled": true, "createdAt": int64(1000)},
		},
	}
	jb, _ := json.Marshal(legacy)
	if err := os.WriteFile(jsonPath, jb, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := dbProfilesMigrate(jsonPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(jsonPath); err == nil {
		t.Fatal("legacy json should be renamed")
	}
	fi2, err := os.Stat(jsonPath + ".imported-bak")
	if err != nil {
		t.Fatal("backup missing")
	}
	if fi2.Mode().Perm() != 0o444 {
		t.Fatalf("bak perm=%o want 444", fi2.Mode().Perm())
	}
	var n int
	if err := d.QueryRow("SELECT COUNT(*) FROM profiles").Scan(&n); err != nil || n != 2 {
		t.Fatalf("rows=%d err=%v", n, err)
	}
	var pri1, pri2 int
	if err := d.QueryRow("SELECT priority FROM profiles WHERE id='p1'").Scan(&pri1); err != nil || pri1 != 100 {
		t.Fatalf("pri1=%d err=%v", pri1, err)
	}
	if err := d.QueryRow("SELECT priority FROM profiles WHERE id='p2'").Scan(&pri2); err != nil || pri2 != 101 {
		t.Fatalf("pri2=%d err=%v", pri2, err)
	}
	var k1 string
	if err := d.QueryRow("SELECT api_key FROM profiles WHERE id='p1'").Scan(&k1); err != nil {
		t.Fatal(err)
	}
	if k1 == "sk-aaa" {
		t.Fatal("key stored as plaintext")
	}
	if got := decSecret(k1); got != "sk-aaa" {
		t.Fatalf("roundtrip=%q", got)
	}
	p, ok := dbProfilesGet("p1")
	if !ok || p.APIKey != "sk-aaa" || p.Name != "A" || !p.Enabled {
		t.Fatalf("get=%+v ok=%v", p, ok)
	}
	un := dbProfilesList(false)
	if len(un) != 2 || un[0].ID != "p1" || un[0].APIKey != "sk-aaa" {
		t.Fatalf("listChat=%+v", un)
	}
	mk := dbProfilesList(true)
	if len(mk) != 2 || mk[0].APIKey == "sk-aaa" {
		t.Fatalf("masked=%+v", mk)
	}
	p3 := LLMProfile{Name: "C", BaseURL: "https://c.example.com", Model: "m3", APIKey: "sk-ccc", Enabled: true}
	p3 = dbProfilesUpsert(p3)
	if p3.ID == "" || p3.CreatedAt == 0 {
		t.Fatalf("upsert new=%+v", p3)
	}
	g3, ok3 := dbProfilesGet(p3.ID)
	if !ok3 || g3.APIKey != "sk-ccc" || g3.Name != "C" {
		t.Fatalf("get new=%+v ok=%v", g3, ok3)
	}
	var pri3 int
	if err := d.QueryRow("SELECT priority FROM profiles WHERE id=?", p3.ID).Scan(&pri3); err != nil || pri3 != 102 {
		t.Fatalf("pri3=%d err=%v", pri3, err)
	}
	rawDB, _ := os.ReadFile(filepath.Join(dir, "nova.db"))
	rawWAL, _ := os.ReadFile(filepath.Join(dir, "nova.db-wal"))
	if strings.Contains(string(rawDB), "sk-ccc") || strings.Contains(string(rawWAL), "sk-ccc") || strings.Contains(string(rawDB), "sk-aaa") || strings.Contains(string(rawWAL), "sk-aaa") {
		t.Fatal("at-rest plaintext leak in db files")
	}

	up := LLMProfile{ID: "p1", Name: "A2", BaseURL: "https://a2.example.com", Model: "m1b", APIKey: "", Enabled: false}
	up = dbProfilesUpsert(up)
	g1b, ok1b := dbProfilesGet("p1")
	if !ok1b || g1b.APIKey != "sk-aaa" || g1b.Name != "A2" || g1b.CreatedAt != 1000 || g1b.Enabled {
		t.Fatalf("after update=%+v ok=%v", g1b, ok1b)
	}
	if !dbProfilesDel("p2") {
		t.Fatal("del p2")
	}
	if dbProfilesDel("p2") {
		t.Fatal("del twice")
	}
	if err := dbProfilesMigrate(jsonPath); err != nil {
		t.Fatal(err)
	}
	var n2 int
	if err := d.QueryRow("SELECT COUNT(*) FROM profiles").Scan(&n2); err != nil || n2 != 2 {
		t.Fatalf("count after idempotent migrate=%d err=%v", n2, err)
	}
}
