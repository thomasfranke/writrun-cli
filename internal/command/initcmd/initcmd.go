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
	// Kit fetches the WritRun kit at a tag — the boundary the tests
	// fake, so an adoption is drivable without a clone.
	Kit kitfetch.Fetcher
}

// New returns the init command wired with its dependencies.
func New(d Deps) command.Command {
	if d.Source == "" {
		d.Source = sourceDefault
	}
	return command.Command{
		Name:    "init",
		Summary: "install WritRun into this repository",
		About: command.About{
			Sentence: []string{
				"install WritRun into this repository",
			},
			Parts: []command.AboutPart{
				{Label: "why you would", Lines: []string{
					"You want this project to use the methodology:",
					"its scripts, its checks, its queue of tasks and",
					"specs.",
				}},
				{Label: "what it does", Lines: []string{
					"Copies the kit at the version this binary pins,",
					"reads your existing conventions rather than",
					"imposing its own, and asks which stage you are",
					"adopting.",
				}},
				{Label: "what it never", Lines: []string{
					"Overwrites what is yours. An existing AGENTS.md",
					"gains one section; your docs are left alone.",
				}},
				{Label: "what comes next", Lines: []string{
					"It runs that stage's checks and names what is",
					"missing. Nothing is fixed for you, and nothing",
					"blocks the adoption.",
				}},
			},
		},
		Need: command.NeedAbsent,
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
		kit, fetchErr = d.Kit.Fetch(d.Tag, d.Source)
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
	options := stageOptions()
	presetOption := ""
	if preset != "" {
		n, err := strconv.Atoi(preset)
		if err != nil || n < 1 || n > len(options) {
			return 0, fmt.Errorf("--stage must be 1, 2 or 3, not %q", preset)
		}
		presetOption = options[n-1].Label
	}
	idx, err := ctx.AskSelect("Which stage?", options, presetOption, "--stage")
	if err != nil {
		return 0, err
	}
	return idx + 1, nil
}

// declaring is the clause every stage's explanation ends with, because
// it is the one thing a reader picking a rung most often has wrong:
// a stage is a declaration, and no check on it blocks the adoption
// (docs/product/adoption/init.md).
const declaring = "Declaring it installs nothing — `doctor` names what is " +
	"missing, and adoption is not conditioned on the forge."

// stageOptions are the three rungs and what each one adds, in the words
// the drawing gives (docs/product/screens/adoption/init.excalidraw).
func stageOptions() []command.Option {
	return []command.Option{
		{
			Label: "1   files",
			Detail: "1 — files. The queue, the checks and the documents are the " +
				"repository's own, and no flow reaches a forge, so this stage " +
				"adds no check beyond the four the scripts need. " + declaring,
		},
		{
			Label: "2   pull requests",
			Detail: "2 — pull requests. The flows open and read pull requests, so " +
				"this stage adds the forge checks: `gh` authenticated, squash " +
				"merging on, and the recording push able to reach main. " + declaring,
		},
		{
			Label: "3   GitHub issues",
			Detail: "3 — GitHub issues. The flows mirror what is recorded here " +
				"into issues, so this stage adds Issues being enabled on this " +
				"repository. " + declaring,
		},
	}
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
