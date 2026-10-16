package buildinfo

import (
	_ "embed"
	"encoding/json"
	"runtime"
)

//go:embed runtime.json
var raw []byte

type Info struct {
	Version string            `json:"version"`
	Go      string            `json:"go"`
	Node    string            `json:"node_build"`
	Bundled map[string]string `json:"bundled"`
	BuildAt string            `json:"build_at"`
}

func Get() Info {
	var i Info
	if err := json.Unmarshal(raw, &i); err != nil {
		return Info{Version: "dev", Go: runtime.Version(), Bundled: map[string]string{}}
	}
	if i.Bundled == nil {
		i.Bundled = map[string]string{}
	}
	return i
}
