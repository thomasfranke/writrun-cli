package command

import (
	"errors"
	"strings"
	"testing"
)

// tty returns a frame whose terminal is interactive at both ends and
// whose screen behaves as the test says.
func tty(t *testing.T, adopted bool, screen func(*Ctx, func(string, string)) error) (Frame, *strTerm) {
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
				f.Screen = func(*Ctx, func(string, string)) error {
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
			if !tc.wantOpened && !strings.Contains(out.String(), "the porcelain for WritRun") {
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
	f.Screen = func(_ *Ctx, run func(string, string)) error {
		run("status", "")
		run("doctor", "")
		run("status", "")
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
	f.Screen = func(_ *Ctx, run func(string, string)) error {
		run("take", "task-0021")
		run("status", "")
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
func chooses(name, arg string) func(*Ctx, func(string, string)) error {
	return func(_ *Ctx, run func(string, string)) error {
		run(name, arg)
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
	f.Screen = func(*Ctx, func(string, string)) error { return nil }
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
	f, _ := tty(t, true, func(*Ctx, func(string, string)) error {
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
	f, _ := tty(t, true, func(ctx *Ctx, _ func(string, string)) error {
		root = ctx.Root
		return nil
	})
	Run(f, nil)
	if root != "/repo" {
		t.Errorf("root = %q, want the resolved /repo", root)
	}
}
