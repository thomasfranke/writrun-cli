package listcmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
)

// listing is one run of the lister with every section present, in the
// order it prints them.
const listing = `In progress — resume before selecting anything new:
  task-0004  Take a task without memorising the flow

Available — any of these may be taken:
  task-0006  high    Show the queue with writrun list
  task-0008  low     Something smaller

Order is a suggestion for a person and binding for an agent.

In flight — an open pull request already exists:
  task-0003  #12 by @someone  Package the kit

Held back:
  task-0009  spec-0009 is draft

Open reports — waiting to be triaged, never selected:
  report-0002  A thing that was noticed

Note: could not reach GitHub, so nothing above accounts for work
already in flight, nor for an amendment suspending a task from an
open pull request — only a spec already returned to draft on this
checkout was seen. Check open pull requests before starting.
`

// call is what the exec port was asked for.
type call struct {
	root   string
	script string
	args   []string
}

// exitErr is a script's own verdict as the port hands it up: an error
// carrying the exit code, which is all this command reads of it.
type exitErr int

func (e exitErr) Error() string { return fmt.Sprintf("exit status %d", int(e)) }
func (e exitErr) ExitCode() int { return int(e) }

// script is the exec port faked: canned output, a canned verdict, and
// a record of what it was asked to run.
func script(out string, err error, rec *call) Script {
	return func(root string, stdout, stderr io.Writer, name string, args ...string) error {
		if rec != nil {
			*rec = call{root: root, script: name, args: args}
		}
		fmt.Fprint(stdout, out)
		return err
	}
}

func runList(t *testing.T, out string, verdict error, args ...string) (string, call, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	var rec call
	ctx := &command.Ctx{Stdout: &stdout, Stderr: &stderr, Root: "/repo"}
	err := run(ctx, Deps{Script: script(out, verdict, &rec)}, args)
	return stdout.String(), rec, err
}

// exitCode is the code an error carries, or -1 for an error carrying
// none — the reading the frame makes of a wrapped script's verdict.
func exitCode(err error) int {
	var exit interface{ ExitCode() int }
	if errors.As(err, &exit) {
		return exit.ExitCode()
	}
	return -1
}

func TestRunsTheSelectionSkillsListerFromTheRoot(t *testing.T) {
	_, rec, err := runList(t, listing, nil)
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	if rec.root != "/repo" {
		t.Errorf("root = %q; want the repository root", rec.root)
	}
	if rec.script != ".writrun/skills/writrun-select-next-task/list_tasks.sh" {
		t.Errorf("script = %q; want the selection skill's lister", rec.script)
	}
	if len(rec.args) != 0 {
		t.Errorf("args = %v; want the script's own defaults", rec.args)
	}
}

func TestUnfilteredOutputIsTheListersOwn(t *testing.T) {
	stdout, _, err := runList(t, listing, nil)
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	if stdout != listing {
		t.Errorf("output was edited:\n%s", stdout)
	}
}

func TestNothingAvailableIsAnAnswerNotAFailure(t *testing.T) {
	stdout, _, err := runList(t, "Nothing is available.\n", exitErr(1))
	if err != nil {
		t.Fatalf("run = %v; want nothing available to exit 0", err)
	}
	if !strings.Contains(stdout, "Nothing is available.") {
		t.Errorf("output = %q; want the script's own message", stdout)
	}
}

func TestTheScriptsRefusalTravelsUpWithItsCode(t *testing.T) {
	_, _, err := runList(t, "", exitErr(3))
	if err == nil {
		t.Fatal("a missing queue directory exited 0")
	}
	if got := exitCode(err); got != 3 {
		t.Errorf("exit = %d; want the script's own 3", got)
	}
}

func TestAFilterStillReportsTheScriptsRefusal(t *testing.T) {
	stdout, _, err := runList(t, "", exitErr(3), "--available")
	if err == nil {
		t.Fatal("a missing queue directory exited 0 under a filter")
	}
	if stdout != "" {
		t.Errorf("output = %q; want nothing where the script printed nothing", stdout)
	}
}

func TestAnErrorCarryingNoExitCodeTravelsUp(t *testing.T) {
	boom := errors.New("bash is not installed")
	_, _, err := runList(t, "", boom)
	if !errors.Is(err, boom) {
		t.Errorf("err = %v; want the cause preserved", err)
	}
}

func TestAnUnexpectedArgumentIsRefused(t *testing.T) {
	_, _, err := runList(t, listing, nil, "task-0006")
	if err == nil || !strings.Contains(err.Error(), "task-0006") {
		t.Errorf("err = %v; want the argument named", err)
	}
}

func TestAnUnknownFlagIsRefused(t *testing.T) {
	if _, _, err := runList(t, listing, nil, "--everything"); err == nil {
		t.Fatal("an unknown flag was accepted")
	}
}

func TestNewDeclaresTheCommand(t *testing.T) {
	c := New(Deps{Script: script(listing, nil, nil)})
	if c.Name != "list" {
		t.Errorf("name = %q", c.Name)
	}
	if c.Need != command.NeedAdopted {
		t.Errorf("need = %v; want an adopted repository", c.Need)
	}
	if c.Summary == "" {
		t.Error("no summary for --help")
	}
	var stdout bytes.Buffer
	if err := c.Run(&command.Ctx{Stdout: &stdout, Stderr: io.Discard, Root: "/repo"}, nil); err != nil {
		t.Errorf("Run = %v", err)
	}
	if stdout.String() != listing {
		t.Error("the wired command did not print the lister's output")
	}
}
