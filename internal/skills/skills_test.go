package skills

import (
	"encoding/json"
	"testing"
)

func TestLoad(t *testing.T) {
	packs, err := Load()
	if err != nil { t.Fatalf("load: %v", err) }
	if len(packs) != 4 { t.Fatalf("want 4 packs, got %d", len(packs)) }
	seen := map[string]bool{}
	for _, p := range packs {
		seen[p.Repo] = true
		var m map[string]any
		if err := json.Unmarshal(p.Meta, &m); err != nil { t.Fatalf("%s meta: %v", p.Repo, err) }
		var c map[string]any
		if err := json.Unmarshal(p.Contract, &c); err != nil { t.Fatalf("%s contract: %v", p.Repo, err) }
		if p.ContractKind == "" { t.Fatalf("%s: empty contractKind", p.Repo) }
	}
	for _, want := range []string{"web-delivery-team", "product-expert-team", "frontend-builder", "AutoDev"} {
		if !seen[want] { t.Fatalf("missing pack: %s", want) }
	}
	wdt := packs[0]
	if wdt.Repo != "web-delivery-team" { t.Fatal("packs[0] not web-delivery-team") }
	if wdt.GatesYAML == "" { t.Fatal("gates yaml empty") }
	if _, err := json.Marshal(packs); err != nil { t.Fatalf("marshal (yaml must not break encoding): %v", err) }
}
