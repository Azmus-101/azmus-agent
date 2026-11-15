package docker

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/compose-spec/compose-go/v2/loader"
	"github.com/compose-spec/compose-go/v2/types"

	"github.com/Azmus-101/azmus-agent/internal/model"
)

// ComposeExtractor turns one compose project (a compose file plus its
// override file, if any) into one "container" component per service.
type ComposeExtractor struct{}

func (ComposeExtractor) Kinds() []string { return []string{model.KindCompose} }

func (ComposeExtractor) Extract(ctx context.Context, fsys fs.FS, t model.Target) ([]model.Component, error) {
	if len(t.Files) == 0 {
		return nil, fmt.Errorf("compose target has no files")
	}
	file := t.Files[0]

	// Read the files ourselves, so the library never touches the disk.
	var configFiles []types.ConfigFile
	for _, f := range t.Files {
		data, err := fs.ReadFile(fsys, f)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", f, err)
		}
		configFiles = append(configFiles, types.ConfigFile{Filename: f, Content: data})
	}

	project, err := loader.LoadWithContext(ctx,
		types.ConfigDetails{WorkingDir: ".", ConfigFiles: configFiles},
		func(o *loader.Options) {
			o.SetProjectName("project", true)
			o.SkipInterpolation = true      // we want what is written, not what the env would give
			o.SkipNormalization = true      // keep paths relative to the scan root
			o.SkipValidation = true         // be lenient with real-world files
			o.SkipConsistencyCheck = true   // a missing dependency is not fatal
			o.SkipResolveEnvironment = true // never read .env / env_file from disk
			o.SkipInclude = true            // include / extends need disk access
			o.SkipExtends = true
			o.Profiles = []string{"*"} // otherwise services behind a profile (e.g. "prod") are dropped
		})
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", file, err)
	}

	names := make([]string, 0, len(project.Services))
	for name := range project.Services {
		names = append(names, name)
	}
	sort.Strings(names)

	id := func(service string) string { return "compose:" + file + "#" + service }

	var comps []model.Component
	for _, name := range names {
		svc := project.Services[name]
		c := model.Component{
			ID:       id(name),
			Name:     name,
			Kind:     "container",
			Provider: providerName,
			Source:   file,
			State:    "declared",
			Env:      map[string]string{},
			Extra:    map[string]any{},
		}
		// ports
		for _, p := range svc.Ports {
			host, _ := strconv.Atoi(p.Published) // ranges and unset values become 0
			c.Ports = append(c.Ports, model.Port{Container: int(p.Target), Host: host, Protocol: p.Protocol})
		}
		// environment: names kept, secret-looking values hidden
		for k, v := range svc.Environment {
			if v == nil {
				c.Env[k] = ""
			} else {
				c.Env[k] = redact(k, *v)
			}
		}
		// depends_on -> relations (sorted so the output is stable)
		deps := make([]string, 0, len(svc.DependsOn))
		for dep := range svc.DependsOn {
			deps = append(deps, dep)
		}
		sort.Strings(deps)
		for _, dep := range deps {
			c.Relations = append(c.Relations, model.Relation{
				Type: "depends_on", To: id(dep), Condition: svc.DependsOn[dep].Condition,
			})
		}
		// image service: the tech is the image name. Built service: look at its Dockerfile.
		switch {
		case svc.Build != nil:
			dockerfile, ok := dockerfilePath(t.Dir, svc.Build)
			if !ok {
				c.Warnings = append(c.Warnings, "build context is not a local directory; Dockerfile not read")
				break
			}
			c.Relations = append(c.Relations, model.Relation{Type: "built_from", To: "dockerfile:" + dockerfile})
			c.Extra["dockerfile"] = dockerfile
			if svc.Build.Target != "" {
				c.Extra["build_target"] = svc.Build.Target
			}
			stages, err := readStages(fsys, dockerfile)
			if err == nil {
				var base string
				if base, err = baseImage(stages, svc.Build.Target); err == nil {
					c.Tech = imageName(base)
					c.Extra["base_image"] = base
				}

			}
			if err != nil {
				c.Warnings = append(c.Warnings, "tech unknown: "+err.Error())
			}
		case svc.Image != "":
			c.Extra["image"] = svc.Image
			c.Tech = imageName(svc.Image)
		}
		// the rest goes in Extra
		if svc.ContainerName != "" {
			c.Extra["container_name"] = svc.ContainerName
		}
		if len(svc.Command) > 0 {
			c.Extra["command"] = []string(svc.Command)
		}
		if svc.Restart != "" {
			c.Extra["restart"] = svc.Restart
		}
		if len(svc.Profiles) > 0 {
			c.Extra["profiles"] = svc.Profiles
		}
		if len(svc.Volumes) > 0 {
			var vols []string
			for _, v := range svc.Volumes {
				s := v.Source + ":" + v.Target
				if v.ReadOnly {
					s += ":ro"
				}
				vols = append(vols, s)
			}
			c.Extra["volumes"] = vols
		}
		if len(svc.EnvFiles) > 0 { // paths only: the files themselves are never read
			var paths []string
			for _, ef := range svc.EnvFiles {
				paths = append(paths, ef.Path)
			}
			c.Extra["env_files"] = paths
		}
		if hc := svc.HealthCheck; hc != nil && !hc.Disable {
			c.Extra["healthcheck"] = []string(hc.Test)
		}

		comps = append(comps, c)

	}
	return comps, nil
}

// dockerfilePath works out where a service's Dockerfile is, relative to the
// scan root. ok is false when the context is a URL or points outside the root.
func dockerfilePath(composeDir string, b *types.BuildConfig) (string, bool) {
	if strings.Contains(b.Context, "://") || strings.HasPrefix(b.Context, "git@") {
		return "", false
	}
	name := b.Dockerfile
	if name == "" {
		name = "Dockerfile"
	}
	p := path.Join(composeDir, b.Context, name)
	if !fs.ValidPath(p) {
		return "", false
	}
	return p, true
}
