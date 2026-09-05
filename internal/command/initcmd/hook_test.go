package initcmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/hook"
)

// installTestHook adopts just enough of a repository for the hook to
// validate against: the observance vocabulary and the hook itself.
func installTestHook(t *testing.T) (root, hookAt string) {
	t.Helper()
	root = makeTarget(t)
	write(t, root, ".writrun/scripts/stage-2-pull-requests/check_observance.sh", templateObservance)
	hookAt, err := hook.Path(root, hook.GitRunner(gitx.Run))
	if err != nil {
		t.Fatalf("hook.Path = %v", err)
	}
	if err := hook.Install(hookAt); err != nil {
		t.Fatalf("hook.Install = %v", err)
	}
	return root, hookAt
}

// runHook executes the installed hook against one message, the way git
// would, and returns its verdict and output.
func runHook(t *testing.T, root, hookAt, message string) (int, string) {
	t.Helper()
	msg := filepath.Join(t.TempDir(), "COMMIT_EDITMSG")
	if err := os.WriteFile(msg, []byte(message), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("bash", hookAt, msg)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err == nil {
		return 0, string(out)
	}
	if exit, ok := err.(*exec.ExitError); ok {
		return exit.ExitCode(), string(out)
	}
	t.Fatalf("running the hook: %v", err)
	return -1, ""
}

func TestHookVerdicts(t *testing.T) {
	root, hookAt := installTestHook(t)
	cases := []struct {
		name    string
		message string
		code    int
		names   string
	}{
		{"a conventional subject passes", "feat: add the thing\n", 0, ""},
		{"a scoped subject passes", "fix(product): mend it\n", 0, ""},
		{"comments and blanks are skipped to the subject", "\n# a comment\nchore: tidy\n", 0, ""},
		{"a shapeless subject is rejected naming the grammar", "did some stuff\n", 1, "not a Conventional subject"},
		{"a foreign type is rejected naming the vocabulary", "build: compile\n", 1, "the type 'build' is outside the vocabulary"},
		{"a foreign scope is rejected naming the vocabulary", "fix(nowhere): mend\n", 1, "the scope 'nowhere' is outside the vocabulary"},
		{"a merge subject passes untouched", "Merge branch 'x'\n", 0, ""},
		{"a fixup subject passes untouched", "fixup! feat: add the thing\n", 0, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, out := runHook(t, root, hookAt, tc.message)
			if code != tc.code {
				t.Errorf("exit = %d, want %d; output:\n%s", code, tc.code, out)
			}
			if tc.names != "" && !strings.Contains(out, tc.names) {
				t.Errorf("output %q does not name the fault %q", out, tc.names)
			}
		})
	}
}

func TestHookOutlivingTheKitBlocksNothing(t *testing.T) {
	root, hookAt := installTestHook(t)
	if err := os.RemoveAll(filepath.Join(root, ".writrun")); err != nil {
		t.Fatal(err)
	}
	if code, out := runHook(t, root, hookAt, "anything at all\n"); code != 0 {
		t.Errorf("exit = %d with the kit gone, want 0; output:\n%s", code, out)
	}
}

func TestRefuseForeignHookRefusesAndNames(t *testing.T) {
	root := makeTarget(t)
	hookAt, err := hook.Path(root, hook.GitRunner(gitx.Run))
	if err != nil {
		t.Fatalf("hook.Path = %v", err)
	}
	if err := hook.RefuseForeign(hookAt); err != nil {
		t.Fatalf("an absent hook was refused: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(hookAt), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(hookAt, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	err = hook.RefuseForeign(hookAt)
	if err == nil {
		t.Fatal("an existing hook was not refused")
	}
	if !strings.Contains(err.Error(), "commit-msg hook is already installed") {
		t.Errorf("the refusal does not name the hook: %v", err)
	}
}
