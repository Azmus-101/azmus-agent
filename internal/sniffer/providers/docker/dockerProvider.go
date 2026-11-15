package docker

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"strconv"
	"strings"

	"github.com/Azmus-101/azmus-agent/internal/model"
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

const providerName = "docker"

type DockerfileExtractor struct {
}

func (DockerfileExtractor) Kinds() []string {
	return []string{model.KindDockerfile}
}

func (DockerfileExtractor) Extract(ctx context.Context, fsys fs.FS, t model.Target) ([]model.Component, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(t.Files) == 0 {
		return nil, fmt.Errorf("dockerfile target has no files")
	}
	file := t.Files[0]

	stages, err := readStages(fsys, file)
	if err != nil {
		return nil, err
	}

	comp := model.Component{
		ID:       "dockerfile:" + file,
		Name:     "dockerfile",
		Kind:     "image",
		Provider: providerName,
		Source:   file,
		State:    "declared",
		Env:      make(map[string]string),
		Extra:    map[string]any{},
	}

	// Tech and base image come from the FROM line of the final stage.
	base, _ := baseImage(stages, "")
	comp.Extra["base_image"] = base
	comp.Tech = imageName(base)

	// Details of the final stage only (that is what the image runs as).
	last := len(stages) - 1
	var argNames []string
	for _, cmd := range stages[last].Commands {
		switch c := cmd.(type) {

		case *instructions.ExposeCommand:
			for _, e := range c.Ports {
				if port, ok := parseExpose(e); ok {
					comp.Ports = append(comp.Ports, port)
				}
			}

		case *instructions.EnvCommand:
			for _, kv := range c.Env {
				comp.Env[kv.Key] = redact(kv.Key, kv.Value)
			}

		case *instructions.ArgCommand:
			// ARG values often hold secrets too, so only the names are kept.
			for _, kv := range c.Args {
				argNames = append(argNames, kv.Key)
			}

		case *instructions.UserCommand:
			comp.Extra["user"] = c.User

		case *instructions.WorkdirCommand:
			comp.Extra["workdir"] = c.Path

		case *instructions.CmdCommand:
			comp.Extra["cmd"] = c.CmdLine

		case *instructions.EntrypointCommand:
			comp.Extra["entrypoint"] = c.CmdLine

		case *instructions.VolumeCommand:
			comp.Extra["volumes"] = c.Volumes

		case *instructions.LabelCommand:
			labels := map[string]string{}
			for _, kv := range c.Labels {
				labels[kv.Key] = kv.Value
			}
			comp.Extra["labels"] = labels

		case *instructions.HealthCheckCommand:
			if c.Health != nil {
				comp.Extra["healthcheck"] = map[string]any{
					"test": c.Health.Test, "interval": c.Health.Interval.String(),
				}
			}
		}
	}
	if len(argNames) > 0 {
		comp.Extra["build_args"] = argNames
	}

	return []model.Component{comp}, nil
}

// ================ Helpers ================

var secretHints = []string{"PASSWORD", "PASSWD", "SECRET", "TOKEN", "KEY", "CREDENTIAL", "PRIVATE"}

// redact keeps variable names but hides values that look like secrets.
func redact(key, value string) string {
	k := strings.ToUpper(key)
	for _, h := range secretHints {
		if strings.Contains(k, h) {
			return "***"
		}
	}
	return value
}

// readStages reads a Dockerfile from fsys and returns its stages.
func readStages(fsys fs.FS, file string) ([]instructions.Stage, error) {
	data, err := fs.ReadFile(fsys, file)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", file, err)
	}
	res, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", file, err)
	}
	// nil linter is fine: we only want the structure, not lint warnings.
	stages, _, err := instructions.Parse(res.AST, nil)
	if err != nil {
		return nil, fmt.Errorf("parse instructions in %s: %w", file, err)
	}
	if len(stages) == 0 {
		return nil, fmt.Errorf("%s has no FROM instruction", file)
	}
	return stages, nil
}

// baseImage returns the image a stage is based on, following
// "FROM <earlier stage>" links back to a real image reference.
// target names the stage to start from; "" means the last stage (Docker's default).
func baseImage(stages []instructions.Stage, target string) (string, error) {
	i := len(stages) - 1
	if target != "" {
		i = -1
		for j, st := range stages {
			if strings.EqualFold(st.Name, target) {
				i = j
			}
		}
		if i < 0 {
			return "", fmt.Errorf("build target %q not found in Dockerfile", target)
		}
	}
	for {
		base := stages[i].BaseName
		found := false
		for j := 0; j < i; j++ {
			if stages[j].Name != "" && strings.EqualFold(stages[j].Name, base) {
				i, found = j, true
				break
			}
		}
		if !found {
			return base, nil
		}
	}
}

// parseExpose turns "5000" or "9229/udp" into a Port. Ranges and variables are skipped.
func parseExpose(s string) (model.Port, bool) {
	num, proto, _ := strings.Cut(s, "/")
	n, err := strconv.Atoi(num)
	if err != nil {
		return model.Port{}, false
	}
	if proto == "" {
		proto = "tcp"
	}
	return model.Port{Container: n, Protocol: proto}, true
}

// imageName strips the tag, digest and the default registry from an image
// reference: "node:lts-alpine" -> "node", "docker.io/library/postgres:16" -> "postgres".
func imageName(image string) string {
	name := image
	if i := strings.Index(name, "@"); i >= 0 { // digest
		name = name[:i]
	}
	if i := strings.LastIndex(name, ":"); i > strings.LastIndex(name, "/") { // tag
		name = name[:i]
	}
	name = strings.TrimPrefix(name, "docker.io/")
	name = strings.TrimPrefix(name, "library/")
	return name
}
