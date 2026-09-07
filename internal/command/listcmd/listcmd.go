// Package listcmd is `writrun list`: the queue as the selection
// skill's own lister prints it. Eligibility, grouping and order are the
// script's; this command runs it and presents what it said
// (docs/product/queue/list.md, spec-0006).
package listcmd

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/kit"
)

// lister is the selection skill's own listing script — the eligibility
// authority this command wraps and never reimplements.
const lister = kit.ListTasks

// Script is the exec port as list uses it: one of the adopted
// repository's own scripts, run from root, its reporting streamed to
// the writers, returning the script's own verdict.
type Script func(root string, stdout, stderr io.Writer, env []string, script string, args ...string) error

// Deps is the wiring list needs beyond the frame's Ctx.
type Deps struct {
	// Script runs one of the adopted repository's own scripts.
	Script Script
}

// New returns the list command wired with its dependencies.
func New(d Deps) command.Command {
	return command.Command{
		Name:    "list",
		Summary: "show the queue: what is available, what is held back, what waits for triage",
		Need:    command.NeedAdopted,
		Run: func(ctx *command.Ctx, args []string) error {
			return run(ctx, d, args)
		},
	}
}

func run(ctx *command.Ctx, d Deps, args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	available := fs.Bool("available", false, "print the available section")
	held := fs.Bool("held", false, "print the held-back section")
	reports := fs.Bool("reports", false, "print the open reports section")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	want := sections{available: *available, held: *held, reports: *reports}

	// Unfiltered, the script writes straight to the user: the sections
	// arrive as they are produced and byte for byte as it wrote them. A
	// filter is the one reason to hold the output back — a section is
	// selectable only once it is whole.
	out := ctx.Stdout
	var buf bytes.Buffer
	if want.filtering() {
		out = &buf
	}
	err := d.Script(ctx.Root, out, ctx.Stderr, nil, lister)
	if want.filtering() {
		want.render(ctx.Stdout, buf.String())
	}
	return verdict(err)
}

// verdict maps the lister's exit onto this command's. Something
// available and nothing available are both answers to the question
// that was asked, so both exit 0 and the script's own message carries
// the difference. Any other exit is the script's refusal, already
// reported on stderr: it travels up with its code rather than being
// restated (spec-0006, steps).
func verdict(err error) error {
	if err == nil {
		return nil
	}
	var exit interface{ ExitCode() int }
	if errors.As(err, &exit) && exit.ExitCode() == 1 {
		return nil
	}
	return err
}
