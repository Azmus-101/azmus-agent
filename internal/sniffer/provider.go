package sniffer

import (
	"context"

	"github.com/Azmus-101/azmus-agent/internal/model"
)

type Provider interface {
	Name() string
	Locate(ctx context.Context) ([]Target, error) // your ProviderLocator
	Extract(ctx context.Context, t Target) (model.Component, error)
}

// Target = "one thing the locator found", before we understand it.
type Target struct {
	Kind  string // "compose-file", "dockerfile", "process", ...
	Ref   string // path, PID, container ID
	Hints map[string]string
}
