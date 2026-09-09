// Package kitfetch is the kit-fetch port: a WritRun tag fetched as a
// shallow clone into a directory outside the repository, so a failure
// has written nothing where it matters. init and update both start
// here, and neither reimplements the clone.
//
// boundaries.md puts everything leaving the process behind a small
// interface with a fake beside it. Script execution, `gh`, the
// terminal and the filesystem were; this is the fifth, and the one
// whose absence made a partial adoption untestable without a clone.
package kitfetch

import (
	"fmt"
	"path/filepath"

	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// Fetched is a kit on disk: the template directory to read from, and
// the cleanup that removes the whole checkout. The cleanup is part of
// the contract — one handing back a directory and no cleanup leaves
// the checkout standing.
type Fetched struct {
	Template string
	Cleanup  func()
}

// Fetcher is the fetch as its consumers name it: one call taking a tag
// and a source, handing back a template directory and its cleanup.
// Clone is the production implementation; Fake, beside it, is the one
// the tests inject.
type Fetcher interface {
	Fetch(tag, source string) (*Fetched, error)
}

// Clone is the production fetcher: the shallow clone, through the
// filesystem and the git runner it was wired with.
type Clone struct {
	Files vfs.FS
	Git   gitx.Runner
}

// Fetch clones source at tag through the wiring Clone holds.
func (c Clone) Fetch(tag, source string) (*Fetched, error) {
	return Fetch(c.Files, tag, source, c.Git)
}

// Fetch clones source at tag, shallowly, and verifies it is a WritRun
// repository — a clone carrying no `template/` is something else, and
// saying so here beats a copy loop finding nothing to do.
func Fetch(files vfs.FS, tag, source string, git gitx.Runner) (*Fetched, error) {
	tmp, err := files.MkdirTemp("", "writrun-kit-")
	if err != nil {
		return nil, err
	}
	cleanup := func() { _ = files.RemoveAll(tmp) }

	clone := filepath.Join(tmp, "writrun")
	if _, err := git("", "clone", "--depth", "1", "--branch", tag, source, clone); err != nil {
		cleanup()
		return nil, errClone(tag, source, err)
	}
	// The folder the kit ships from takes the kit's name. It was
	// `template/` before WritRun v0.0.07 renamed it; the binary pins one
	// tag at a time, so it looks where the tag it pins puts it and
	// refuses anything else (docs/about.md — pins, never tracks).
	shipped := filepath.Join(clone, shippedDir)
	if _, err := files.Stat(shipped); err != nil {
		cleanup()
		return nil, errNoTemplate(tag, source)
	}
	return &Fetched{Template: shipped, Cleanup: cleanup}, nil
}

// errClone is the refusal a clone that could not run produces: it
// names the tag and the source and says the tree is untouched. The
// fake raises this one too, so a test reads the message the user gets.
func errClone(tag, source string, cause error) error {
	return fmt.Errorf("fetching WritRun %s from %s failed — nothing was written: %w", tag, source, cause)
}

// shippedDir is the folder a WritRun tag ships its adopter tree from.
const shippedDir = "kit"

// errNoTemplate is the refusal a clone carrying no shipped folder
// produces: a repository, but not a WritRun one at the pinned tag.
func errNoTemplate(tag, source string) error {
	return fmt.Errorf("%s carries no %s/ at %s — not a WritRun repository", source, shippedDir, tag)
}
