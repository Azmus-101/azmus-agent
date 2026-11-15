package locate

import (
	"context"
	"path"
	"strings"

	"github.com/Azmus-101/azmus-agent/internal/model"
)

var composeNames = map[string]bool{
	"docker-compose.yml": true, "docker-compose.yaml": true,
	"compose.yml": true, "compose.yaml": true,
}

// Compose finds compose files by name. A sibling "<name>.override.<ext>"
// is added to the same target, because compose merges the two.
type Compose struct{}

func (Compose) Name() string { return "compose" }

func (Compose) Locate(ctx context.Context, files []string) ([]model.Target, error) {
	have := make(map[string]bool, len(files))
	for _, f := range files {
		have[f] = true
	}

	var out []model.Target
	for _, f := range files {
		if !composeNames[strings.ToLower(path.Base(f))] {
			continue
		}
		t := model.Target{Kind: model.KindCompose, Dir: path.Dir(f), Files: []string{f}}
		ext := path.Ext(f)
		if override := strings.TrimSuffix(f, ext) + ".override" + ext; have[override] {
			t.Files = append(t.Files, override)
		}
		out = append(out, t)
	}
	return out, nil
}
