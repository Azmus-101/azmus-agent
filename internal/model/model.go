// Package model defines the shared vocabulary of the Azmus agent.
// It must not import any other internal package.
package model

type Component struct {
	ID       string
	Name     string
	Kind     string // container, database, runtime, etc...
	Provider string // which provider found it (DockerProvider for example)
	Tech     string // "node", "postgres", "python", "unknown"
	Source   string // file path or API(like docker/client) it came from
	State    string // "running", "stopped", ...
	// Ports    []Port
	Env map[string]string
	// Mounts   []Mount
	// Health   *Health
	Extra map[string]any // provider-specific details (compose profiles, build stages...)
}
