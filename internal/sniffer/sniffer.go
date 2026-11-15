package sniffer

import (
	"context"
	"fmt"
	"io/fs"
	"strings"

	"github.com/Azmus-101/azmus-agent/internal/model"
)

type Sniffer struct {
	plugins *Plugins
}

func New(p *Plugins) *Sniffer {
	return &Sniffer{plugins: p}
}

func (s *Sniffer) Run(ctx context.Context, fsys fs.FS) (model.Snapshot, error) {
	warnings := []string{} // so if there is no warning, not getting null

	// 1. walk through directories to get the files
	files, w, err := walk(ctx, fsys)
	if err != nil {
		return model.Snapshot{}, err
	}

	warnings = append(warnings, w...)

	// 2. fire every locator against the same file list
	var targets []model.Target
	for _, l := range s.plugins.locators {
		found, err := l.Locate(ctx, files)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("locator %s: %v", l.Name(), err))
			continue //if one locator failed, it shouldn't stop the rest
		}
		targets = append(targets, found...)
	}

	var components []model.Component
	for _, t := range targets {
		exc, ok := s.plugins.extractorFor(t.Kind)
		if !ok {
			warnings = append(warnings, fmt.Sprintf("no extractor for kind %q (%s)", t.Kind, strings.Join(t.Files, ", ")))
			continue
		}
		comps, err := exc.Extract(ctx, fsys, t)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("extract %s: %v", strings.Join(t.Files, ", "), err))
			continue
		}
		components = append(components, comps...)
	}
	return model.Snapshot{
		Components: components,
		Warnings:   warnings,
	}, nil
}
