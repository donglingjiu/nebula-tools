// Package skills — OPT-029: base skill-pack contracts embedded into the binary.
// V1 scope: machine-readable contracts only (pipeline/routing/contract/meta + gates yaml).
// SKILL.md full-text is V2 via runtime dir (NOVA_SKILLS_DIR) — private repo content never enters git.
package skills

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"sync"
)

//go:embed */*
var files embed.FS

type Pack struct {
	Repo         string          `json:"repo"`
	Meta         json.RawMessage `json:"meta"`
	Contract     json.RawMessage `json:"contract"`
	ContractKind string          `json:"contractKind"` // pipeline | routing | builder-contract
	GatesYAML    string          `json:"gatesYaml,omitempty"`
}

var (
	once   sync.Once
	packs  []Pack
	loadEr error
)

func Load() ([]Pack, error) {
	once.Do(func() {
		specs := []struct{ repo, contractFile, kind string }{
			{"web-delivery-team", "pipeline.json", "pipeline"},
			{"product-expert-team", "routing.json", "routing"},
			{"frontend-builder", "contract.json", "builder-contract"},
			{"AutoDev", "contract.json", "builder-contract"},
		}
		for _, sp := range specs {
			meta, err := fs.ReadFile(files, sp.repo+"/meta-info.json")
			if err != nil { loadEr = fmt.Errorf("%s: meta: %w", sp.repo, err); return }
			contract, err := fs.ReadFile(files, sp.repo+"/"+sp.contractFile)
			if err != nil { loadEr = fmt.Errorf("%s: contract: %w", sp.repo, err); return }
			p := Pack{Repo: sp.repo, Meta: json.RawMessage(meta), Contract: json.RawMessage(contract), ContractKind: sp.kind}
			if sp.repo == "web-delivery-team" {
				g, err := fs.ReadFile(files, sp.repo+"/stage-gates.yaml")
				if err != nil { loadEr = fmt.Errorf("%s: gates: %w", sp.repo, err); return }
				p.GatesYAML = string(g)
			}
			packs = append(packs, p)
		}
	})
	return packs, loadEr
}
