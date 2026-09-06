package command

import (
	"errors"
	"strings"
	"testing"
)

// tty returns a frame whose terminal is interactive at both ends and
// whose screen resolves to what the test names.
func tty(t *testing.T, adopted bool, screen func(*Ctx) (string, string, error)) (Frame, *strTerm) {
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
// is what screen.md prescribes rather than a fallback invented here.
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
				f.Screen = func(*Ctx) (string, string, error) {
					opened = true
					return "", "", nil
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

// A key leaves the screen and runs the command it names, with the
// selected id as its argument — the command's own run, not a copy.
func TestAKeyDispatchesTheCommandItNames(t *testing.T) {
	var got []string
	cmd := Command{
		Name: "take", Summary: "take", Need: NeedAdopted,
		Run: func(_ *Ctx, args []string) error { got = args; return nil },
	}
	f, _, _ := frame(t, []Command{cmd}, true, nil)
	f.Terminal = &strTerm{FakeTerminal{In: true, Out: true}}
	f.Screen = func(*Ctx) (string, string, error) { return "take", "task-0021", nil }
	if code := Run(f, nil); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if strings.Join(got, ",") != "task-0021" {
		t.Errorf("take received %v, want the selected id", got)
	}
}

// The dispatched command's exit code is the process's: the screen adds
// no verdict of its own.
func TestTheDispatchedCommandsCodeIsTheProcesss(t *testing.T) {
	cmd := Command{
		Name: "take", Summary: "take", Need: NeedAdopted,
		Run: func(*Ctx, []string) error { return ErrDeclined },
	}
	f, _, errb := frame(t, []Command{cmd}, true, nil)
	f.Terminal = &strTerm{FakeTerminal{In: true, Out: true}}
	f.Screen = func(*Ctx) (string, string, error) { return "take", "task-0021", nil }
	if code := Run(f, nil); code != 1 {
		t.Errorf("exit = %d, want the declined 1", code)
	}
	if !strings.Contains(errb.String(), "declined") {
		t.Error("the command's own words were not printed")
	}
}

// Leaving with q runs nothing and exits 0.
func TestLeavingTheScreenRunsNothing(t *testing.T) {
	ran := false
	cmd := Command{Name: "take", Summary: "take", Need: NeedAdopted,
		Run: func(*Ctx, []string) error { ran = true; return nil }}
	f, _, _ := frame(t, []Command{cmd}, true, nil)
	f.Terminal = &strTerm{FakeTerminal{In: true, Out: true}}
	f.Screen = func(*Ctx) (string, string, error) { return "", "", nil }
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
	f, _ := tty(t, true, func(*Ctx) (string, string, error) {
		return "", "", errors.New("the lister refused")
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
	f, _ := tty(t, true, func(ctx *Ctx) (string, string, error) {
		root = ctx.Root
		return "", "", nil
	})
	Run(f, nil)
	if root != "/repo" {
		t.Errorf("root = %q, want the resolved /repo", root)
	}
}
