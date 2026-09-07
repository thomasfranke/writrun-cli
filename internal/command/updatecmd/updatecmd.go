// Package updatecmd is `writrun update`: the adopted kit refreshed to
// the tag this binary pins. It rewrites what the methodology declares
// refreshable and nothing else — the conventions, the settings, the
// project's docs and the queue are never read for writing
// (docs/product/adoption/update.md, spec-0003).
package updatecmd

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/kitfetch"
	"github.com/thomasfranke/writrun-cli/internal/kittag"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// sourceDefault is the WritRun repository the kit is fetched from.
const sourceDefault = "https://github.com/thomasfranke/writrun"

// Deps is the wiring update needs beyond the frame's Ctx.
type Deps struct {
	// Tag is the WritRun tag this release pins — the target of every
	// refresh, so what a binary installs and what it refreshes to can
	// never disagree.
	Tag string
	// Source is the WritRun repository to clone; empty means the
	// canonical one.
	Source string
	// Git runs one git invocation.
	Git gitx.Runner
	// Files is the filesystem this command reads and writes through.
	Files vfs.FS
	// Kit fetches the WritRun kit at a tag — the boundary the tests
	// fake, so a refresh is drivable without a clone.
	Kit kitfetch.Fetcher
}

// New returns the update command wired with its dependencies.
func New(d Deps) command.Command {
	if d.Source == "" {
		d.Source = sourceDefault
	}
	return command.Command{
		Name:    "update",
		Summary: "refresh the adopted kit to the tag this binary pins",
		Need:    command.NeedAdopted,
		Run: func(ctx *command.Ctx, args []string) error {
			return run(ctx, d, args)
		},
	}
}

func run(ctx *command.Ctx, d Deps, args []string) error {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}

	current, err := recordedTag(d.Files, ctx.Root)
	if err != nil {
		return err
	}
	switch cmp := kittag.Compare(d.Tag, current); {
	case cmp == 0:
		fmt.Fprintf(ctx.Stdout, "Already at WritRun %s — nothing to refresh.\n", current)
		return nil
	case cmp < 0:
		return fmt.Errorf("this binary pins WritRun %s and the kit records %s — a downgrade is a deliberate act update does not offer", d.Tag, current)
	}

	// A dirty tree would mix the refresh with unrelated changes, and
	// unlike an adoption a refresh *overwrites*: an uncommitted edit
	// inside a kit-owned folder would be gone with nothing to restore
	// it from.
	if out, gitErr := d.Git(ctx.Root, "status", "--porcelain"); gitErr != nil {
		return fmt.Errorf("reading the working tree: %w", gitErr)
	} else if strings.TrimSpace(out) != "" {
		return fmt.Errorf("the working tree is dirty — commit or stash first, so the refresh is the only change")
	}

	var kit *kitfetch.Fetched
	if err := ctx.Terminal.Spin("fetching WritRun "+d.Tag, func() error {
		var fetchErr error
		kit, fetchErr = d.Kit.Fetch(d.Tag, d.Source)
		return fetchErr
	}); err != nil {
		return err
	}
	defer kit.Cleanup()

	r, err := plan(d.Files, ctx.Root, kit.Template, current, d.Tag)
	if err != nil {
		return err
	}
	r.render(ctx.Stdout)
	if r.empty() {
		return nil
	}
	if err := ctx.AskConfirm(fmt.Sprintf("Refresh the kit to WritRun %s?", d.Tag)); err != nil {
		return err
	}
	if err := r.apply(); err != nil {
		return fmt.Errorf("%w — the refresh is partial; `git checkout -- .` and `git clean -fd` undo what it wrote, then rerun writrun update", err)
	}
	fmt.Fprintf(ctx.Stdout, "Refreshed to WritRun %s.\n", d.Tag)
	return nil
}

// recordedTag is the tag the kit records — the refresh's starting
// point, read from the file the adoption wrote. The file and its
// parsing are kittag's; what an unrecorded tag means to a refresh is
// this command's.
func recordedTag(files vfs.FS, root string) (string, error) {
	tag, err := kittag.Read(files, root)
	if err != nil {
		return "", fmt.Errorf("reading .writrun/VERSION: %w", err)
	}
	if tag == "" {
		return "", fmt.Errorf(".writrun/VERSION records no tag — the kit's version is what a refresh refreshes from")
	}
	return tag, nil
}
