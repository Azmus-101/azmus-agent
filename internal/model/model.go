package model

// Kind values a locator can put on a Target. Shared here so locators and
// extractors agree on the strings without importing each other.
const (
	KindDockerfile = "dockerfile"
	KindCompose    = "compose"
)

// Target = "one thing the locator found", before we understand it.
type Target struct {
	Kind  string
	Dir   string   // directory relative to the scan root ("." for the root)
	Files []string // slash-separated paths relative to the scan root
	Hints map[string]string
}

// Port: Container is always set; Host is 0 when nothing is published.
type Port struct {
	Container int
	Host      int
	Protocol  string // "tcp" or "udp"
}

// Relation links a component to another one. Condition is optional,
// e.g. "service_healthy" on a compose depends_on.
type Relation struct {
	Type      string // "depends_on", "built_from", ...
	To        string // ID of the other component
	Condition string
}

type Component struct {
	ID        string
	Name      string
	Kind      string // "image", "container", ...
	Provider  string // which plugin found it
	Tech      string
	Source    string // file the component came from
	State     string // offline mode: always "declared"
	Ports     []Port
	Env       map[string]string
	Extra     map[string]any
	Relations []Relation
	Warnings  []string
}

type Snapshot struct {
	Components []Component
	Warnings   []string
}
