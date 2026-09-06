// Package workcmd is `writrun work`: the task the selection algorithm
// offers — or the one named — handed to the agent the adopter
// configured, with the brief the methodology's own script assembled.
// It selects and it launches; it reasons about no task, writes nothing
// to the queue, and opens nothing on the forge
// (docs/product/queue/work.md, spec-0007).
package workcmd

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/kit"
)

// The authorities this command wraps: the selection algorithm's own
// lister, and the script that assembles step 7's brief. Both are the
// adopted repository's own, and neither is reimplemented here.
const (
	listScript  = ".writrun/skills/writrun-select-next-task/list_tasks.sh"
	briefScript = ".writrun/skills/writrun-select-next-task/brief.sh"
)

// agentsFile is the project's instructions — the second thing the
// launched agent is pointed at, after the brief.
const agentsFile = "AGENTS.md"

// Deps is the wiring work needs beyond the frame's Ctx.
type Deps struct {
	// Git reads the configured agent command. `git config` layers
	// repository over user for free, which is the whole reason the key
	// lives there (decision 0003).
	Git gitx.Runner
	// Scripts runs the adopted repository's own scripts — the lister
	// and the brief.
	Scripts kit.Runner
	// Launch starts the configured agent.
	Launch Launcher
}

// New returns the work command wired with its dependencies.
func New(d Deps) command.Command {
	return command.Command{
		Name:    "work",
		Summary: "launch the configured agent on the next available task, or on the one named",
		Need:    command.NeedAdopted,
		Run: func(ctx *command.Ctx, args []string) error {
			return run(ctx, d, args)
		},
	}
}

// run is the spec's four steps in order. The agent is read first: with
// none configured there is nothing to select for, and a run that had
// already picked a task would read as though something had begun
// (spec-0007, steps).
func run(ctx *command.Ctx, d Deps, args []string) error {
	fs := flag.NewFlagSet("work", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 1 {
		return fmt.Errorf("two task ids given (%q and %q) — work takes one", fs.Arg(0), fs.Arg(1))
	}

	agent, err := configured(d.Git, ctx.Root)
	if err != nil {
		return err
	}
	name, argv, err := words(agent)
	if err != nil {
		return err
	}

	id, err := selectTask(ctx, d, fs.Arg(0))
	if err != nil {
		return err
	}
	brief, err := assemble(ctx, d, id)
	if err != nil {
		return err
	}

	fmt.Fprintf(ctx.Stdout, "%s — launching %s\n", id, agent)
	return passthrough(name, d.Launch(ctx.Root, name, append(argv, prompt(id, brief))...))
}

// assemble is step 3: the brief, in the script's own output. Nothing
// here reads it — an incomplete brief is refused on the script's exit
// code, never on a judgement about what the text says (spec-0007, edge
// cases).
func assemble(ctx *command.Ctx, d Deps, id string) (string, error) {
	var out, errb bytes.Buffer
	err := d.Scripts(ctx.Root, &out, &errb, nil, briefScript, id)
	code := exitCode(err)
	if code == 0 {
		return out.String(), nil
	}

	// Whatever the script said, it said in its own words and on its own
	// stream; a partial brief printed what it could resolve, and that
	// half is shown too rather than swallowed.
	if code == 2 {
		fmt.Fprint(ctx.Stdout, out.String())
	}
	fmt.Fprint(ctx.Stderr, errb.String())
	switch code {
	case 1:
		return "", fmt.Errorf("%s resolved no task for %s — nothing was launched", briefScript, id)
	case 2:
		return "", fmt.Errorf("the brief for %s is incomplete — an incomplete brief is not a working brief, and nothing was launched", id)
	default:
		return "", fmt.Errorf("running %s: %w", briefScript, err)
	}
}

// prompt is step 4's pointing: the task, the project's instructions and
// the brief, in one argument. The brief travels unedited — this command
// frames it and adds nothing to it, because everything an agent needs
// to decide with is already inside it (spec-0007, acceptance criteria).
func prompt(id, brief string) string {
	return fmt.Sprintf(`Work %s in this repository.

Read %s at the repository root first: it is this project's instructions,
and every gate in it binds you exactly as it binds a human. Take the
task the way it says to.

Below is the brief %s assembled — the task, its specs, and the
documentation each one anchors.

%s`, id, agentsFile, briefScript, brief)
}

// passthrough hands the launched command's verdict up unedited: the
// frame turns an error carrying an exit code into that exit code, and
// what the agent printed it printed on the terminal it inherited. Only
// a command that never started is this command's failure to report.
func passthrough(name string, err error) error {
	if err == nil {
		return nil
	}
	if exitCode(err) < 0 {
		return fmt.Errorf("launching %s: %w", name, err)
	}
	return err
}

// exitCode reads a child's own verdict off the error a port returned;
// -1 says it never ran, which is not a verdict to map.
func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var verdict interface{ ExitCode() int }
	if errors.As(err, &verdict) && verdict.ExitCode() > 0 {
		return verdict.ExitCode()
	}
	return -1
}
