package main

import (
	"path/filepath"
	"testing"
)

func TestConversationsCRUD(t *testing.T) {
	dir := t.TempDir()
	d, err := openDBAt(filepath.Join(dir, "nova.db"))
	if err != nil {
		t.Fatal(err)
	}
	saved := novaDB
	novaDB = d
	defer func() { novaDB = saved; d.Close() }()

	id1, err := dbConvCreate("alice", "conv one")
	if err != nil || id1 <= 0 {
		t.Fatalf("create=%v id=%d", err, id1)
	}
	if _, err := dbMsgAppend(id1, "user", "hello"); err != nil {
		t.Fatal(err)
	}
	if _, err := dbMsgAppend(id1, "assistant", "hi there"); err != nil {
		t.Fatal(err)
	}
	cs := dbConvList("alice")
	if len(cs) != 1 || cs[0].Title != "conv one" || cs[0].ID != id1 {
		t.Fatalf("list=%+v", cs)
	}
	if got := dbConvList("bob"); len(got) != 0 {
		t.Fatalf("isolation broken: %+v", got)
	}
	if !dbConvOwn(id1, "alice") || dbConvOwn(id1, "bob") {
		t.Fatal("own check broken")
	}
	ms := dbMsgList(id1)
	if len(ms) != 2 || ms[0].Role != "user" || ms[1].Role != "assistant" {
		t.Fatalf("msgs=%+v", ms)
	}
	if err := dbConvRename(id1, "alice", "renamed"); err != nil {
		t.Fatal(err)
	}
	if err := dbConvRename(id1, "bob", "hack"); err == nil {
		t.Fatal("rename by stranger should fail")
	}
	if _, err := dbMsgAppend(id1, " hacker", "x"); err == nil {
		t.Fatal("bad role should fail")
	}
	if err := dbConvDelete(id1, "alice"); err != nil {
		t.Fatal(err)
	}
	if got := dbMsgList(id1); len(got) != 0 {
		t.Fatalf("cascade failed: %+v", got)
	}
	if err := dbConvDelete(id1, "alice"); err == nil {
		t.Fatal("double delete should fail")
	}
}
