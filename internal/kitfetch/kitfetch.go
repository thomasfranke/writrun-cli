// Package kitfetch fetches a WritRun tag: a shallow clone into a
// directory outside the repository, so a failure has written nothing
// where it matters. init and update both start here, and neither
// reimplements the clone.
package kitfetch

import (
	"fmt"
	"os"
	"path/filepath"
)

// GitRunner executes one git invocation in a directory and returns its
// stdout.
type GitRunner func(dir string, args ...string) (string, error)

// Fetched is a clone on disk: the template directory to read from, and
// the cleanup that removes the whole checkout.
type Fetched struct {
	Template string
	Cleanup  func()
}

// Fetch clones source at tag, shallowly, and verifies it is a WritRun
// repository — a clone carrying no `template/` is something else, and
// saying so here beats a copy loop finding nothing to do.
func Fetch(tag, source string, git GitRunner) (*Fetched, error) {
	tmp, err := os.MkdirTemp("", "writrun-kit-")
	if err != nil {
		return nil, err
	}
	cleanup := func() { os.RemoveAll(tmp) }

	clone := filepath.Join(tmp, "writrun")
	if _, err := git("", "clone", "--depth", "1", "--branch", tag, source, clone); err != nil {
		cleanup()
		return nil, fmt.Errorf("fetching WritRun %s from %s failed — nothing was written: %w", tag, source, err)
	}
	template := filepath.Join(clone, "template")
	if _, err := os.Stat(template); err != nil {
		cleanup()
		return nil, fmt.Errorf("%s carries no template/ at %s — not a WritRun repository", source, tag)
	}
	return &Fetched{Template: template, Cleanup: cleanup}, nil
}
