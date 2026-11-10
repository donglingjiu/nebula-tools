package ai

import (
	"strings"
	"testing"
)

func mkChats(pairs, size int) []map[string]string {
	out := make([]map[string]string, 0, pairs*2)
	for i := 0; i < pairs; i++ {
		s := strings.Repeat("字", size)
		out = append(out,
			map[string]string{"role": "user", "content": s},
			map[string]string{"role": "assistant", "content": s})
	}
	return out
}

func TestCompressChatsUnderBudget(t *testing.T) {
	cs := mkChats(3, 10)
	got, trunc := compressChats(cs, 1000)
	if trunc || len(got) != 6 {
		t.Fatalf("want 6 msgs no-trunc, got %d/%v", len(got), trunc)
	}
}

func TestCompressChatsOverBudget(t *testing.T) {
	cs := mkChats(5, 100)
	got, trunc := compressChats(cs, 300)
	if !trunc {
		t.Fatal("want truncated")
	}
	if len(got) == 0 || len(got) >= len(cs) {
		t.Fatalf("want partial keep, got %d/%d", len(got), len(cs))
	}
	if got[0]["role"] != "user" {
		t.Fatalf("window must start with user, got %s", got[0]["role"])
	}
	if got[len(got)-1]["content"] != cs[len(cs)-1]["content"] {
		t.Fatal("newest message lost")
	}
}

func TestCompressChatsHugeSingle(t *testing.T) {
	cs := []map[string]string{{"role": "user", "content": strings.Repeat("x", 500)}}
	got, trunc := compressChats(cs, 100)
	if len(got) != 1 {
		t.Fatalf("want 1, got %d", len(got))
	}
	if trunc {
		t.Fatal("single kept message must not flag window truncation")
	}
	want := truncRunes(cs[0]["content"], 100)
	if got[0]["content"] != want {
		t.Fatalf("want truncRunes form (%d runes), got %d runes",
			len([]rune(want)), len([]rune(got[0]["content"])))
	}
	if !strings.HasPrefix(want, strings.Repeat("x", 90)) {
		t.Fatal("truncated form lost content prefix")
	}
}
