package uninstallcmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/hook"

	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

func TestNewNamesTheCommand(t *testing.T) {
	c := New(Deps{Git: hook.GitRunner(gitx.Run), Files: vfs.OS{}})
	if c.Name != "uninstall" {
		t.Errorf("Name = %q", c.Name)
	}
	if c.Need != command.NeedAdopted {
		t.Errorf("Need = %v, want NeedAdopted", c.Need)
	}
	if c.Summary == "" {
		t.Error("the command carries no summary for --help")
	}
	if c.Run == nil {
		t.Fatal("the command carries no work")
	}
}

func TestAnAbsentAgentsIsNamedAsAlreadyGone(t *testing.T) {
	root := makeAdopted(t)
	if err := os.Remove(filepath.Join(root, "AGENTS.md")); err != nil {
		t.Fatal(err)
	}
	out, err := runUninstall(t, root)
	if err != nil {
		t.Fatalf("uninstall: %v\n%s", err, out)
	}
	if !strings.Contains(out, "already gone AGENTS.md") {
		t.Errorf("an absent AGENTS.md was not named:\n%s", out)
	}
}

func TestAnAbsentHookIsNamedAsKept(t *testing.T) {
	root := makeAdopted(t)
	out, err := runUninstall(t, root)
	if err != nil {
		t.Fatalf("uninstall: %v\n%s", err, out)
	}
	if !strings.Contains(out, "no commit-msg hook is installed") {
		t.Errorf("an absent hook was not named:\n%s", out)
	}
}

func TestHookDisplayFallsBackToTheAbsolutePath(t *testing.T) {
	// `core.hooksPath` can put the hook anywhere; a plan naming a path
	// inside the repository while the write lands outside it is consent
	// to something else.
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "shared-hooks", "commit-msg")
	r := &removal{disk: vfs.OS{}, root: root, hookAt: outside}
	if got := r.hookDisplay(); got != outside {
		t.Errorf("hookDisplay = %q, want the absolute %q", got, outside)
	}

	inside := filepath.Join(root, ".git", "hooks", "commit-msg")
	r = &removal{disk: vfs.OS{}, root: root, hookAt: inside}
	if got := r.hookDisplay(); got != ".git/hooks/commit-msg" {
		t.Errorf("hookDisplay = %q, want the path relative to the repository", got)
	}
}

func TestApplyToleratesWhatIsAlreadyGone(t *testing.T) {
	root := t.TempDir()
	r := &removal{
		disk:      vfs.OS{},
		root:      root,
		dirs:      nil,
		files:     []string{"WRITRUN.md"},
		hookAt:    filepath.Join(root, "hooks", "commit-msg"),
		hookState: hook.Ours,
	}
	// Neither the file nor the hook is there; a removal that finds
	// nothing to remove is done, not broken.
	if err := r.apply(); err != nil {
		t.Fatalf("apply over an already-empty repository: %v", err)
	}
}

func TestPlanSeparatesWhatGoesFromWhatIsGone(t *testing.T) {
	root := makeAdopted(t)
	if err := os.Remove(filepath.Join(root, ".github/workflows/writrun-issues.yml")); err != nil {
		t.Fatal(err)
	}
	r, err := plan(vfs.OS{}, root, filepath.Join(root, ".git", "hooks", "commit-msg"))
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if len(r.dirs) != 1 || r.dirs[0] != ".writrun" {
		t.Errorf("the kit directory is not in the removal set: %v", r.dirs)
	}
	var goneNamed bool
	for _, g := range r.gone {
		if g == ".github/workflows/writrun-issues.yml" {
			goneNamed = true
		}
	}
	if !goneNamed {
		t.Errorf("a file deleted by hand is not in the gone set: %v", r.gone)
	}
	for _, f := range r.files {
		if f == ".github/workflows/writrun-issues.yml" {
			t.Error("a file already gone is also in the removal set")
		}
	}
}
