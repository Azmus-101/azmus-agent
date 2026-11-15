package sniffer

import (
	"context"
	"fmt"
	"io/fs"
)

var skipDir = map[string]bool{
	".git": true, "node_modules": true, "vendor": true,
	".idea": true, ".vscode": true, "dist": true, "build": true,
}

func walk(ctx context.Context, fsys fs.FS) (files, warnings []string, err error) {
	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		if walkErr != nil {
			warnings = append(warnings, fmt.Sprintf("cannot read %s: %v", path, walkErr))
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			if path != "." && skipDir[d.Name()] {
				return fs.SkipDir
			}
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, warnings, err
}
