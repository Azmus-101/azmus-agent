package locate

import (
	"context"
	"path"
	"strings"

	"github.com/Azmus-101/azmus-agent/internal/model"
)

type DockerfileLocator struct{}

func (DockerfileLocator) Name() string { return "dockerfileLocator" }

func (DockerfileLocator) Locate(ctx context.Context, files []string) ([]model.Target, error) {
	var out []model.Target
	for _, f := range files {
		if !isDockerfileName(path.Base(f)) {
			continue
		}
		out = append(out, model.Target{
			Kind:  model.KindDockerfile,
			Dir:   path.Dir(f),
			Files: []string{f},
		})
	}
	return out, nil
}

// Matches: Dockerfile, Dockerfile.dev, api.Dockerfile, Containerfile (any case).
func isDockerfileName(base string) bool {
	b := strings.ToLower(base)
	return b == "dockerfile" || b == "containerfile" ||
		strings.HasPrefix(b, "dockerfile.") ||
		strings.HasSuffix(b, ".dockerfile")
}
