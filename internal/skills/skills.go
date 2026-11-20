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

// Contract returns the raw contract bytes for a repo (OPT-030b: flow engine consumption).
// Reuses Load() (sync.Once cached) — Pack already carries Contract bytes; no second spec table,
// no Load() modification (nt100/100b lesson: reuse accessors, never reach into unseen internals).
func Contract(repo string) ([]byte, error) {
	ps, err := Load()
	if err != nil {
		return nil, err
	}
	for _, p := range ps {
		if p.Repo == repo {
			return []byte(p.Contract), nil
		}
	}
	return nil, fmt.Errorf("unknown contract repo: %s", repo)
}

// DisplayName returns the localized display name for a pack repo (V2 polish).
// Meta is json.RawMessage — runtime parse; tags from meta-info.json实录 (skillName).
func DisplayName(repo string) string {
	ps, err := Load()
	if err != nil {
		return ""
	}
	for _, p := range ps {
		if p.Repo == repo && len(p.Meta) > 0 {
			var m struct {
		SkillName string `json:"skillName"`
			}
			if json.Unmarshal(p.Meta, &m) == nil {
				for _, v := range []string{m.SkillName} {
					if v != "" {
						return v
					}
				}
			}
		}
	}
	return ""
}
