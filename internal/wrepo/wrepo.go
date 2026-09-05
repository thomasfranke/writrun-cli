// Package wrepo resolves the repository a command operates on: the git
// toplevel, and whether WritRun is adopted there.
package wrepo

import (
	"fmt"
	"path/filepath"

	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// Find walks up from dir to the git toplevel — the directory holding
// `.git`, a directory or a worktree's file — and reports whether
// `.writrun/` sits beside it. Running from a subdirectory is the same
// answer as running from the root.
func Find(files vfs.FS, dir string) (root string, adopted bool, err error) {
	d, err := filepath.Abs(dir)
	if err != nil {
		return "", false, err
	}
	for {
		if _, statErr := files.Stat(filepath.Join(d, ".git")); statErr == nil {
			info, statErr := files.Stat(filepath.Join(d, ".writrun"))
			return d, statErr == nil && info.IsDir(), nil
		}
		parent := filepath.Dir(d)
		if parent == d {
			return "", false, fmt.Errorf("not inside a git repository")
		}
		d = parent
	}
}
