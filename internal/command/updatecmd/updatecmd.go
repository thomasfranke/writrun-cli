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
	"os"
	"path/filepath"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/fence"
	"github.com/thomasfranke/writrun-cli/internal/kitfetch"
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
	Git kitfetch.GitRunner
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

	current, err := recordedTag(ctx.Root)
	if err != nil {
		return err
	}
	switch cmp := compareTags(d.Tag, current); {
	case cmp == 0:
		fmt.Fprintf(ctx.Stdout, "Already at WritRun %s — nothing to refresh.\n", current)
		return nil
	case cmp < 0:
		return fmt.Errorf("this binary pins WritRun %s and the kit records %s — a downgrade is a deliberate act update does not offer", d.Tag, current)
	}

	// The fence is read before the network is asked for anything: a
	// damaged one stops the whole refresh, and stopping after a clone
	// would have spent the fetch to reach the same refusal (spec-0003).
	agentsPath := filepath.Join(ctx.Root, "AGENTS.md")
	agents, err := os.ReadFile(agentsPath)
	if err != nil {
		return fmt.Errorf("reading AGENTS.md: %w", err)
	}
	if _, _, err := fence.Remove(agents); err != nil {
		return fmt.Errorf("AGENTS.md: %w — the fenced section is what a refresh rewrites, so nothing was changed", err)
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
		kit, fetchErr = kitfetch.Fetch(d.Tag, d.Source, d.Git)
		return fetchErr
	}); err != nil {
		return err
	}
	defer kit.Cleanup()

	r, err := plan(ctx.Root, kit.Template, current, d.Tag, agents)
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
// point, read from the file the adoption wrote.
func recordedTag(root string) (string, error) {
	raw, err := os.ReadFile(filepath.Join(root, ".writrun", "VERSION"))
	if err != nil {
		return "", fmt.Errorf("reading .writrun/VERSION: %w", err)
	}
	tag := strings.TrimSpace(string(raw))
	if tag == "" {
		return "", fmt.Errorf(".writrun/VERSION records no tag — the kit's version is what a refresh refreshes from")
	}
	return tag, nil
}
