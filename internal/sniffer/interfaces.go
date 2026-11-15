package sniffer

import (
	"context"
	"io/fs"

	"github.com/Azmus-101/azmus-agent/internal/model"
)

// Locator finds targets among the files the sniffer already walked.
// files holds slash-separated paths relative to the root of fsys, sorted.
type Locator interface {
	Name() string
	Locate(ctx context.Context, files []string) ([]model.Target, error)
}

// Extractor turns one target into components, relations and warnings.
type Extractor interface {
	Kinds() []string // which Target.Kind values this extractor handles
	Extract(ctx context.Context, fsys fs.FS, t model.Target) ([]model.Component, error)
}
