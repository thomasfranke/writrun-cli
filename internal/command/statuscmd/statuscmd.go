// Package statuscmd is `writrun status`: where the work stands, read
// from the current branch. It resolves the branch to a task, names the
// task's spec and its status, runs the completion checks read-only and
// names the first that would fail, counts the reports awaiting triage,
// and compares the kit's recorded tag with the tag this client pins. It
// writes nothing, and it makes no judgement the checks have not already
// made (docs/product/queue/status.md, spec-0013).
package statuscmd

import (
	"flag"
	"fmt"
	"io"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/kit"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// Deps is the wiring status needs beyond the frame's Ctx.
type Deps struct {
	// Tag is the WritRun tag this release pins — one half of the
	// comparison step 5 makes, handed in the way init and update
	// receive it rather than read from somewhere of this command's own
	// choosing.
	Tag string
	// Git runs one git invocation; the branch is git's to name.
	Git gitx.Runner
	// Files is the filesystem this command reads through. It never
	// writes: every path here is opened to be read.
	Files vfs.FS
	// Scripts runs the adopted repository's own scripts — the
	// completion checks are `preflight.sh`, and this command wraps it
	// rather than repeating what it checks.
	Scripts kit.Runner
}

// New returns the status command wired with its dependencies.
func New(d Deps) command.Command {
	return command.Command{
		Name:    "status",
		Summary: "say where the work stands: the branch's task, the checks, the reports, the kit",
		Need:    command.NeedAdopted,
		Run: func(ctx *command.Ctx, args []string) error {
			return run(ctx, d, args)
		},
	}
}

// run answers the question and exits 0 having answered it. A failing
// completion check is the answer, not this command's failure — the
// reading `list` makes of "nothing is available" — so only a question
// that could not be asked at all is an error (spec-0013, scope).
func run(ctx *command.Ctx, d Deps, args []string) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}

	branch, err := currentBranch(d.Git, ctx.Root)
	if err != nil {
		return err
	}
	say(ctx.Stdout, "Branch", branchLabel(branch))

	t := resolveTask(d.Files, ctx.Root, branch)
	for _, l := range t.lines() {
		say(ctx.Stdout, l.label, l.text)
	}

	// Step 3 runs only where step 1 found a task: the checks are what
	// `finish` would run on this task, and a branch carrying none has
	// no completion to fall short of (spec-0013, steps).
	if t.found {
		say(ctx.Stdout, "Checks", verdict(preflight(d.Scripts, ctx.Root)))
	}

	say(ctx.Stdout, "Reports", openReports(d.Files, ctx.Root))
	say(ctx.Stdout, "Kit", kitLine(d.Files, ctx.Root, d.Tag))
	return nil
}

// say prints one labelled line; an empty label continues the line above
// it, so a task carrying two specs reads as one entry.
func say(w io.Writer, label, text string) {
	fmt.Fprintf(w, "%-8s %s\n", label, text)
}

// kitLine is step 5: the recorded tag against the pinned one, both
// named, nothing bridged.
func kitLine(files vfs.FS, root, pinned string) string {
	recorded, err := recordedTag(files, root)
	switch {
	case err != nil:
		return fmt.Sprintf("%s — this client pins WritRun %s", err, pinned)
	case sameRelease(recorded, pinned):
		return fmt.Sprintf("WritRun %s — the tag this client pins", recorded)
	default:
		return fmt.Sprintf("WritRun %s recorded, %s pinned by this client — they differ", recorded, pinned)
	}
}
