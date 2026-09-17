package command

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

// tty returns a frame whose terminal is interactive at both ends and
// whose screen behaves as the test says.
func tty(t *testing.T, adopted bool, screen func(*Ctx, func(string, string, io.Writer)) error) (Frame, *strTerm) {
	t.Helper()
	f, _, _ := frame(t, nil, adopted, nil)
	term := &strTerm{FakeTerminal{In: true, Out: true}}
	f.Terminal = term
	f.Screen = screen
	return f, term
}

type strTerm struct{ FakeTerminal }

// The screen is opened only with a terminal at both ends and inside an
// adopted repository. Every other no-command run prints the help, which
// is what screens/README.md prescribes rather than a fallback invented here.
func TestTheScreenOpensOnlyWhereTheRuleSaysItCan(t *testing.T) {
	for _, tc := range []struct {
		name           string
		in, out        bool
		adopted, wired bool
		wantOpened     bool
	}{
		{"a terminal inside an adoption", true, true, true, true, true},
		{"stdin is not a terminal", false, true, true, true, false},
		{"stdout is not a terminal", true, false, true, true, false},
		{"outside an adoption", true, true, false, true, false},
		{"a binary with no screen", true, true, true, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opened := false
			f, out, _ := frame(t, nil, tc.adopted, nil)
			f.Terminal = &strTerm{FakeTerminal{In: tc.in, Out: tc.out}}
			if tc.wired {
				f.Screen = func(*Ctx, func(string, string, io.Writer)) error {
					opened = true
					return nil
				}
			}
			code := Run(f, nil)
			if code != 0 {
				t.Errorf("exit = %d, want 0", code)
			}
			if opened != tc.wantOpened {
				t.Errorf("screen opened = %v, want %v", opened, tc.wantOpened)
			}
			if !tc.wantOpened && !strings.Contains(out.String(), "run a project by WritRun") {
				t.Error("the help was not printed where the rule asks for it")
			}
		})
	}
}

// A key runs the command it names, with the selected id as its argument
// — the command's own run, not a copy.
func TestAKeyDispatchesTheCommandItNames(t *testing.T) {
	var got []string
	cmd := Command{
		Name: "take", Summary: "take", Need: NeedAdopted,
		Run: func(_ *Ctx, args []string) error { got = args; return nil },
	}
	f, _, _ := frame(t, []Command{cmd}, true, nil)
	f.Terminal = &strTerm{FakeTerminal{In: true, Out: true}}
	f.Screen = chooses("take", "task-0021")
	if code := Run(f, nil); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if strings.Join(got, ",") != "task-0021" {
		t.Errorf("take received %v, want the selected id", got)
	}
}

// The session is what the screen is: more than one command runs in one
// run of `writrun`, in the order they were chosen. Reading two things
// used to be two runs, and that is the whole of what this changes.
func TestASessionRunsEveryCommandChosenInIt(t *testing.T) {
	var ran []string
	note := func(name string) Command {
		return Command{Name: name, Summary: name, Need: NeedAdopted,
			Run: func(*Ctx, []string) error { ran = append(ran, name); return nil }}
	}
	f, _, _ := frame(t, []Command{note("status"), note("doctor")}, true, nil)
	f.Terminal = &strTerm{FakeTerminal{In: true, Out: true}}
	f.Screen = func(_ *Ctx, run func(string, string, io.Writer)) error {
		run("status", "", nil)
		run("doctor", "", nil)
		run("status", "", nil)
		return nil
	}
	if code := Run(f, nil); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if strings.Join(ran, ",") != "status,doctor,status" {
		t.Errorf("ran %v, want each choice in the order it was made", ran)
	}
}

// A command that refuses does not end the session and does not decide
// the run's code. The refusal is the reader's to read and come back
// from; leaving the screen is what ends the run, and leaving is not a
// failure. `writrun take` on its own still answers 1 — that code is for
// a script, and a session has no script reading it.
func TestARefusalInsideTheSessionDoesNotEndIt(t *testing.T) {
	after := false
	cmds := []Command{
		{Name: "take", Summary: "take", Need: NeedAdopted,
			Run: func(*Ctx, []string) error { return ErrDeclined }},
		{Name: "status", Summary: "status", Need: NeedAdopted,
			Run: func(*Ctx, []string) error { after = true; return nil }},
	}
	f, _, errb := frame(t, cmds, true, nil)
	f.Terminal = &strTerm{FakeTerminal{In: true, Out: true}}
	f.Screen = func(_ *Ctx, run func(string, string, io.Writer)) error {
		run("take", "task-0021", nil)
		run("status", "", nil)
		return nil
	}
	if code := Run(f, nil); code != 0 {
		t.Errorf("exit = %d, want 0 — a refusal is not the session's verdict", code)
	}
	if !strings.Contains(errb.String(), "declined") {
		t.Error("the command's own words were not printed")
	}
	if !after {
		t.Error("a refusal ended the session")
	}
}

// chooses answers with one command and then leaves.
func chooses(name, arg string) func(*Ctx, func(string, string, io.Writer)) error {
	return func(_ *Ctx, run func(string, string, io.Writer)) error {
		run(name, arg, nil)
		return nil
	}
}

// Leaving with q runs nothing and exits 0.
func TestLeavingTheScreenRunsNothing(t *testing.T) {
	ran := false
	cmd := Command{Name: "take", Summary: "take", Need: NeedAdopted,
		Run: func(*Ctx, []string) error { ran = true; return nil }}
	f, _, _ := frame(t, []Command{cmd}, true, nil)
	f.Terminal = &strTerm{FakeTerminal{In: true, Out: true}}
	f.Screen = func(*Ctx, func(string, string, io.Writer)) error { return nil }
	if code := Run(f, nil); code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	if ran {
		t.Error("leaving the screen ran a command")
	}
}

// A screen that cannot open says so and exits 1 — it does not fall back
// to the help, which would read as "no screen was ever meant to open".
func TestAScreenThatFailsIsReported(t *testing.T) {
	f, _ := tty(t, true, func(*Ctx, func(string, string, io.Writer)) error {
		return errors.New("the lister refused")
	})
	var errb strings.Builder
	f.Stderr = &errb
	if code := Run(f, nil); code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if !strings.Contains(errb.String(), "the lister refused") {
		t.Errorf("stderr = %q; the cause was not named", errb.String())
	}
}

// The screen is handed the repository it will read, already resolved.
func TestTheScreenIsGivenTheAdoptedRoot(t *testing.T) {
	var root string
	f, _ := tty(t, true, func(ctx *Ctx, _ func(string, string, io.Writer)) error {
		root = ctx.Root
		return nil
	})
	Run(f, nil)
	if root != "/repo" {
		t.Errorf("root = %q, want the resolved /repo", root)
	}
}

// A captured command writes where the screen said, not to the frame's
// own streams — that is what lets the screen page an answer instead of
// giving up the terminal for it. Both streams go to the one place: a
// reader reads one account, in the order it was written.
func TestACapturedCommandWritesWhereTheScreenSaid(t *testing.T) {
	cmd := Command{
		Name: "status", Summary: "status", Need: NeedAdopted, AsksNothing: true,
		Run: func(ctx *Ctx, _ []string) error {
			fmt.Fprintln(ctx.Stdout, "the answer")
			fmt.Fprintln(ctx.Stderr, "and a warning")
			return nil
		},
	}
	f, out, errb := frame(t, []Command{cmd}, true, nil)
	f.Terminal = &strTerm{FakeTerminal{In: true, Out: true}}
	var captured strings.Builder
	f.Screen = func(_ *Ctx, run func(string, string, io.Writer)) error {
		run("status", "", &captured)
		return nil
	}
	if code := Run(f, nil); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	for _, want := range []string{"the answer", "and a warning"} {
		if !strings.Contains(captured.String(), want) {
			t.Errorf("%q did not reach the screen; it captured %q", want, captured.String())
		}
	}
	if strings.Contains(out.String(), "the answer") || strings.Contains(errb.String(), "and a warning") {
		t.Error("a captured command also wrote to the terminal the screen is drawing on")
	}
}

// A command that asks runs as its own process.
//
// This is the whole of spec-0044 in one assertion: the command chosen on
// a screen is not called inside this process, it is handed to the spawn
// port. What it buys is not visible from here — a reader that cannot
// outlive a process it is not in — so this is the only place the rule
// can be held at all (report-0040).
func TestACommandThatAsksIsSpawned(t *testing.T) {
	ran := false
	cmd := Command{
		Name: "report", Summary: "report", Need: NeedAdopted,
		Run: func(*Ctx, []string) error { ran = true; return nil },
	}
	f, _, _ := frame(t, []Command{cmd}, true, nil)
	f.Terminal = &strTerm{FakeTerminal{In: true, Out: true}}
	f.Screen = chooses("report", "")

	var got []string
	f.Again = func(args []string) error { got = args; return nil }

	if code := Run(f, nil); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if strings.Join(got, ",") != "report" {
		t.Errorf("spawned %v, want the command alone", got)
	}
	if ran {
		t.Error("the command ran in this process as well as in its own")
	}
}

// The frame's flags travel to the process the question runs in.
//
// A session running under `--yes` that spawned a bare command would ask
// a question its parent had already answered, and the reader would have
// to answer it twice — or worse, be asked one they had declined.
func TestTheSpawnedCommandCarriesTheSessionsFlags(t *testing.T) {
	cmd := Command{Name: "take", Summary: "take", Need: NeedAdopted,
		Run: func(*Ctx, []string) error { return nil }}
	f, _, _ := frame(t, []Command{cmd}, true, nil)
	f.Terminal = &strTerm{FakeTerminal{In: true, Out: true}}
	f.Screen = chooses("take", "task-0021")

	var got []string
	f.Again = func(args []string) error { got = args; return nil }

	if code := Run(f, []string{"--yes", "--no-color"}); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	want := "--no-color,--yes,take,task-0021"
	if strings.Join(got, ",") != want {
		t.Errorf("spawned %v, want %q — flags first, then the command and its argument", got, want)
	}
}

// A command that asks nothing is captured here, not spawned.
//
// It opens no terminal program, so it has no reader that could outlive
// it — and the screen keeps the terminal for it precisely so its output
// can be paged. Spawning it would throw that away.
func TestACapturedCommandIsNotSpawned(t *testing.T) {
	f, _, _ := frame(t, []Command{{
		Name: "status", Summary: "status", Need: NeedAdopted,
		Run: func(ctx *Ctx, _ []string) error { fmt.Fprint(ctx.Stdout, "read"); return nil },
	}}, true, nil)
	f.Terminal = &strTerm{FakeTerminal{In: true, Out: true}}

	var page strings.Builder
	f.Screen = func(_ *Ctx, run func(string, string, io.Writer)) error {
		run("status", "", &page)
		return nil
	}
	spawned := false
	f.Again = func([]string) error { spawned = true; return nil }

	if code := Run(f, nil); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if spawned {
		t.Error("a captured command was spawned")
	}
	if page.String() != "read" {
		t.Errorf("captured %q, want the command's output", page.String())
	}
}

// A binary that cannot spawn itself still opens its questions.
//
// The port answers an error only when the process could not be started,
// and the answer to that is the behaviour every command had before the
// port existed. A screen that refuses to open a question is worse than
// one that may swallow a key, and a reader cannot tell the two apart
// from a question that never appeared.
func TestAFailedSpawnFallsBackToThisProcess(t *testing.T) {
	ran := false
	f, _, _ := frame(t, []Command{{
		Name: "report", Summary: "report", Need: NeedAdopted,
		Run: func(*Ctx, []string) error { ran = true; return nil },
	}}, true, nil)
	f.Terminal = &strTerm{FakeTerminal{In: true, Out: true}}
	f.Screen = chooses("report", "")
	f.Again = func([]string) error { return errors.New("no such binary") }

	if code := Run(f, nil); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if !ran {
		t.Error("the command did not run at all after the spawn failed")
	}
}

// A frame with no spawn port runs the command where it always ran.
//
// The port is production wiring, and a binary built without it — or a
// suite driving the routing without one — must still dispatch. This is
// the same fallback as a failed spawn, reached a different way.
func TestNoSpawnPortRunsTheCommandHere(t *testing.T) {
	ran := false
	f, _, _ := frame(t, []Command{{
		Name: "report", Summary: "report", Need: NeedAdopted,
		Run: func(*Ctx, []string) error { ran = true; return nil },
	}}, true, nil)
	f.Terminal = &strTerm{FakeTerminal{In: true, Out: true}}
	f.Screen = chooses("report", "")

	if code := Run(f, nil); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if !ran {
		t.Error("the command did not run with no spawn port wired")
	}
}
