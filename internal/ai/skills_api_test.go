package ai

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestSkillsListHandler(t *testing.T) {
	s := &Service{}
	rr := httptest.NewRecorder()
	s.skillsList(rr, httptest.NewRequest("GET", "/api/ai/skills", nil))
	if rr.Code != 200 {
		t.Fatalf("code=%d body=%s", rr.Code, rr.Body.String())
	}
	var out struct {
		Count int               `json:"count"`
		Packs []json.RawMessage `json:"packs"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Count != 4 || len(out.Packs) != 4 {
		t.Fatalf("count=%d packs=%d", out.Count, len(out.Packs))
	}
}
