package command

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func frame(t *testing.T, cmds []Command, adopted bool, findErr error) (Frame, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	var out, errb bytes.Buffer
	return Frame{
		Version:    "1.2.3",
		WritRunTag: "v0.0.03",
		Commands:   cmds,
		Stdout:     &out,
		Stderr:     &errb,
		Terminal:   &FakeTerminal{},
		FindRepo: func(dir string) (string, bool, error) {
			if findErr != nil {
				return "", false, findErr
			}
			return "/repo", adopted, nil
		},
		Getenv: func(string) string { return "" },
		Getwd:  func() (string, error) { return "/repo/sub", nil },
	}, &out, &errb
}

func TestVersionAnswersAnywhere(t *testing.T) {
	f, out, _ := frame(t, nil, false, errors.New("not inside a git repository"))
	if code := Run(f, []string{"--version"}); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	got := out.String()
	if !strings.Contains(got, "1.2.3") || !strings.Contains(got, "pins WritRun v0.0.03") {
		t.Fatalf("--version = %q; want the client's version and the pinned tag", got)
	}
}

func TestHelpAnswersAnywhere(t *testing.T) {
	cmds := []Command{{Name: "noop", Summary: "does nothing"}}
	for _, args := range [][]string{{"--help"}, {"-h"}, nil} {
		f, out, _ := frame(t, cmds, false, errors.New("not inside a git repository"))
		if code := Run(f, args); code != 0 {
			t.Fatalf("Run(%v) = %d, want 0", args, code)
		}
		got := out.String()
		if !strings.Contains(got, "noop  does nothing") {
			t.Fatalf("help = %q; want one line per command", got)
		}
		if !strings.Contains(got, docsAddress) {
			t.Fatalf("help = %q; want the docs' address", got)
		}
	}
}

func TestUnknownSubcommandAborts(t *testing.T) {
	f, _, errb := frame(t, nil, true, nil)
	if code := Run(f, []string{"bogus"}); code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(errb.String(), `unknown command "bogus"`) || !strings.Contains(errb.String(), "usage:") {
		t.Fatalf("stderr = %q; want the unknown command named and usage", errb.String())
	}
}

func TestUnknownFlagAborts(t *testing.T) {
	f, _, errb := frame(t, nil, true, nil)
	if code := Run(f, []string{"--bogus"}); code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(errb.String(), "unknown flag --bogus") {
		t.Fatalf("stderr = %q; want the flag named", errb.String())
	}
}

func TestNeedEnforcement(t *testing.T) {
	ran := false
	cmd := func(need Need) []Command {
		return []Command{{Name: "c", Need: need, Run: func(*Ctx, []string) error { ran = true; return nil }}}
	}
	cases := []struct {
		name     string
		need     Need
		adopted  bool
		findErr  error
		wantCode int
		wantErr  string
	}{
		{"adopted need met", NeedAdopted, true, nil, 0, ""},
		{"adopted need unmet", NeedAdopted, false, nil, 1, "not an adopted repository"},
		{"adopted outside git", NeedAdopted, false, errors.New("not inside a git repository"), 1, "not inside a git repository"},
		{"absent need met", NeedAbsent, false, nil, 0, ""},
		{"absent need unmet", NeedAbsent, true, nil, 1, "already adopted"},
		{"any runs outside git", NeedAny, false, errors.New("not inside a git repository"), 0, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ran = false
			f, _, errb := frame(t, cmd(tc.need), tc.adopted, tc.findErr)
			if code := Run(f, []string{"c"}); code != tc.wantCode {
				t.Fatalf("exit = %d, want %d (stderr %q)", code, tc.wantCode, errb.String())
			}
			if tc.wantErr != "" {
				if ran {
					t.Fatal("command ran despite the refused need")
				}
				if !strings.Contains(errb.String(), tc.wantErr) {
					t.Fatalf("stderr = %q; want it to name %q", errb.String(), tc.wantErr)
				}
			} else if !ran {
				t.Fatal("command did not run")
			}
		})
	}
}

// scriptVerdict mimics *exec.ExitError: an error carrying the wrapped
// script's own exit code.
type scriptVerdict int

func (v scriptVerdict) Error() string { return "exit status" }
func (v scriptVerdict) ExitCode() int { return int(v) }

func TestScriptExitCodePassesThroughUnrestated(t *testing.T) {
	cmds := []Command{{Name: "c", Need: NeedAny, Run: func(*Ctx, []string) error {
		return scriptVerdict(3)
	}}}
	f, _, errb := frame(t, cmds, true, nil)
	if code := Run(f, []string{"c"}); code != 3 {
		t.Fatalf("exit = %d, want the script's own 3", code)
	}
	if errb.Len() != 0 {
		t.Fatalf("stderr = %q; the script already reported — want nothing restated", errb.String())
	}
}

func TestSignaledScriptStillExitsOneWithTheCause(t *testing.T) {
	cmds := []Command{{Name: "c", Need: NeedAny, Run: func(*Ctx, []string) error {
		return scriptVerdict(-1) // killed, no exit code to pass through
	}}}
	f, _, errb := frame(t, cmds, true, nil)
	if code := Run(f, []string{"c"}); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(errb.String(), "writrun c:") {
		t.Fatalf("stderr = %q; want the failure named", errb.String())
	}
}

func TestDeclinedExitsNonZeroNamingIt(t *testing.T) {
	cmds := []Command{{Name: "c", Need: NeedAny, Run: func(c *Ctx, _ []string) error {
		return c.AskConfirm("push the branch?")
	}}}
	f, _, errb := frame(t, cmds, true, nil)
	f.Terminal = &FakeTerminal{In: true, ConfirmAnswer: false}
	if code := Run(f, []string{"c"}); code != 1 {
		t.Fatalf("exit = %d, want 1 on decline", code)
	}
	if !strings.Contains(errb.String(), "declined — nothing changed") {
		t.Fatalf("stderr = %q; want the decline named", errb.String())
	}
}

func TestDoubleDashHandsReservedFlagsToTheCommand(t *testing.T) {
	var gotArgs []string
	cmds := []Command{{Name: "c", Need: NeedAny, Run: func(_ *Ctx, a []string) error {
		gotArgs = a
		return nil
	}}}
	f, out, _ := frame(t, cmds, true, nil)
	if code := Run(f, []string{"c", "--", "--help", "--yes"}); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if len(gotArgs) != 2 || gotArgs[0] != "--help" || gotArgs[1] != "--yes" {
		t.Fatalf("args = %v, want [--help --yes] verbatim after --", gotArgs)
	}
	if strings.Contains(out.String(), "Docs:") {
		t.Fatalf("stdout = %q; the global help answered a command's own --help", out.String())
	}
}

func TestDoubleDashBeforeTheNameStillDispatches(t *testing.T) {
	ran := false
	cmds := []Command{{Name: "c", Need: NeedAny, Run: func(_ *Ctx, a []string) error {
		ran = true
		if len(a) != 1 || a[0] != "--no-color" {
			t.Errorf("args = %v, want [--no-color]", a)
		}
		return nil
	}}}
	f, _, _ := frame(t, cmds, true, nil)
	if code := Run(f, []string{"--", "c", "--no-color"}); code != 0 || !ran {
		t.Fatalf("exit = %d ran = %v, want the command dispatched", code, ran)
	}
}

func TestCommandErrorNamesCommandAndExitsNonZero(t *testing.T) {
	cmds := []Command{{Name: "c", Need: NeedAny, Run: func(*Ctx, []string) error {
		return errors.New("the check failed")
	}}}
	f, _, errb := frame(t, cmds, true, nil)
	if code := Run(f, []string{"c"}); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(errb.String(), "writrun c: the check failed") {
		t.Fatalf("stderr = %q; want the failure named", errb.String())
	}
}

func TestGlobalFlagsReachCtxAndArgsReachCommand(t *testing.T) {
	var got *Ctx
	var gotArgs []string
	cmds := []Command{{Name: "c", Need: NeedAny, Run: func(c *Ctx, a []string) error {
		got, gotArgs = c, a
		return nil
	}}}
	f, _, _ := frame(t, cmds, true, nil)
	if code := Run(f, []string{"--yes", "c", "--no-color", "extra"}); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if !got.Yes {
		t.Fatal("--yes did not reach the command")
	}
	if got.Color {
		t.Fatal("--no-color did not disable color")
	}
	if len(gotArgs) != 1 || gotArgs[0] != "extra" {
		t.Fatalf("args = %v, want [extra]", gotArgs)
	}
	if got.Root != "/repo" || !got.Adopted {
		t.Fatalf("repo = %q adopted=%v, want /repo true", got.Root, got.Adopted)
	}
}
