// Package initcmd is `writrun init`: the adoption of a repository in
// one confirmed act — the kit fetched at the pinned WritRun tag, the
// repository's own conventions extracted, an existing AGENTS.md
// grafted, the commit-message hook installed, the stage chosen and its
// requirements checked on the spot (docs/product/adoption/init.md,
// spec-0002).
package initcmd

import (
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/hook"
	"github.com/thomasfranke/writrun-cli/internal/kitfetch"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// sourceDefault is the WritRun repository the kit is fetched from.
const sourceDefault = "https://github.com/thomasfranke/writrun"

// Deps is the production wiring init needs beyond the frame's Ctx:
// the pinned tag, the source to fetch it from, and the three externals
// behind functions so the tests can fake each.
type Deps struct {
	// Tag is the WritRun tag this release pins.
	Tag string
	// Source is the WritRun repository to clone; empty means the
	// canonical one. The suite points it at a local clone.
	Source string
	// Git runs one git invocation; Gh one gh invocation; LookPath is
	// the PATH probe of the stage-0 checks.
	Git      gitx.Runner
	Gh       func(args ...string) (string, error)
	LookPath func(name string) (string, error)
	// Files is the filesystem this command reads and writes through.
	Files vfs.FS
}

// New returns the init command wired with its dependencies.
func New(d Deps) command.Command {
	if d.Source == "" {
		d.Source = sourceDefault
	}
	return command.Command{
		Name:    "init",
		Summary: "install the WritRun kit into this repository",
		Need:    command.NeedAbsent,
		Run: func(ctx *command.Ctx, args []string) error {
			return run(ctx, d, args)
		},
	}
}

func run(ctx *command.Ctx, d Deps, args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	stageFlag := fs.String("stage", "", "the stage to adopt at: 1, 2 or 3")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}

	// The refusals, before the network is asked for anything: a dirty
	// tree would mix the adoption with unrelated changes, and a foreign
	// hook is another project's to keep (spec-0002).
	if out, err := d.Git(ctx.Root, "status", "--porcelain"); err != nil {
		return fmt.Errorf("reading the working tree: %w", err)
	} else if strings.TrimSpace(out) != "" {
		return fmt.Errorf("the working tree is dirty — commit or stash first (`git stash -u`; untracked files count too), so the adoption is the only change")
	}
	hookAt, err := hook.Path(ctx.Root, d.Git)
	if err != nil {
		return err
	}
	if err := hook.RefuseForeign(d.Files, hookAt); err != nil {
		return err
	}

	// The fetch: a shallow clone of the pinned tag, into a directory
	// outside the repository — a failure here has written nothing
	// (spec-0002, edge cases).
	var kit *kitfetch.Fetched
	if err := ctx.Terminal.Spin("fetching WritRun "+d.Tag, func() error {
		var fetchErr error
		kit, fetchErr = kitfetch.Fetch(d.Files, d.Tag, d.Source, d.Git)
		return fetchErr
	}); err != nil {
		return err
	}
	defer kit.Cleanup()
	template := kit.Template

	stage, err := askStage(ctx, *stageFlag)
	if err != nil {
		return err
	}

	a, err := plan(d.Files, ctx.Root, template, d.Tag, d.Source, stage, hookAt, d.Git)
	if err != nil {
		return err
	}
	a.render(ctx.Stdout)
	if err := ctx.AskConfirm(fmt.Sprintf("Adopt WritRun %s into this repository?", d.Tag)); err != nil {
		return err
	}
	if err := a.apply(); err != nil {
		// The tree was clean before this point — the refusal above saw
		// to it — so everything git now reports is the adoption's, and
		// undoing it is two git commands. The hook lives outside the
		// worktree, where neither of them reaches; left behind, it
		// trips the foreign-hook refusal on the rerun.
		return fmt.Errorf("%w — the adoption is partial; `git checkout -- .` and `git clean -fd` undo what it wrote, `rm -f %s` removes the hook, then rerun writrun init", err, a.hookPath)
	}

	reportGaps(ctx.Stdout, checkStages(ctx.Root, stage, d), stage)
	fmt.Fprintf(ctx.Stdout, "Adopted WritRun %s at stage %d. The queue starts empty — work arrives through the flow.\n", d.Tag, stage)
	return nil
}

// askStage resolves the stage: --stage answers it without asking, a
// terminal arrow-selects it, and anything else aborts naming the flag
// (spec-0002).
func askStage(ctx *command.Ctx, preset string) (int, error) {
	options := []string{"1 — files", "2 — pull requests", "3 — GitHub issues"}
	presetOption := ""
	if preset != "" {
		n, err := strconv.Atoi(preset)
		if err != nil || n < 1 || n > len(options) {
			return 0, fmt.Errorf("--stage must be 1, 2 or 3, not %q", preset)
		}
		presetOption = options[n-1]
	}
	idx, err := ctx.AskSelect("Adopt at which stage?", options, presetOption, "--stage")
	if err != nil {
		return 0, err
	}
	return idx + 1, nil
}

// reportGaps names what the chosen stage's checks found — named, never
// fixed, and never blocking: adoption is not conditioned on the forge
// (product/adoption/init.md).
func reportGaps(w io.Writer, gaps []gap, stage int) {
	if len(gaps) == 0 {
		fmt.Fprintf(w, "Checks for stages 0–%d: all clear.\n", stage)
		return
	}
	fmt.Fprintf(w, "Checks for stages 0–%d found %d gap(s) — named, not fixed:\n", stage, len(gaps))
	for _, g := range gaps {
		fmt.Fprintf(w, "  stage %d: %s\n", g.Stage, g.Text)
	}
}
