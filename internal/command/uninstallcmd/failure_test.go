package uninstallcmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/hook"
)

func readOnly(t *testing.T, dir string) {
	t.Helper()
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
}

func skipAsRoot(t *testing.T) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("root removes from a read-only directory all the same")
	}
}

func TestTheCommandNewReturnsDoesTheWork(t *testing.T) {
	root := makeAdopted(t)
	var out strings.Builder
	ctx := &command.Ctx{
		Stdout: &out, Stderr: &out,
		Terminal: &command.FakeTerminal{},
		Root:     root, Adopted: true, Yes: true,
	}
	if err := New(Deps{Git: hook.GitRunner(gitx.Run)}).Run(ctx, nil); err != nil {
		t.Fatalf("uninstall through New: %v\n%s", err, out.String())
	}
	if _, statErr := os.Stat(filepath.Join(root, ".writrun")); statErr == nil {
		t.Error("the kit survived a run through the wired command")
	}
}

func TestRunReportsAGitThatCannotAnswer(t *testing.T) {
	// The hooks directory is git's to name; outside a repository there
	// is no answer, and nothing may be removed on a guess.
	root := t.TempDir()
	var out strings.Builder
	ctx := &command.Ctx{Stdout: &out, Stderr: &out, Terminal: &command.FakeTerminal{}, Root: root, Yes: true}
	err := run(ctx, Deps{Git: hook.GitRunner(gitx.Run)}, nil)
	if err == nil {
		t.Fatal("the hooks directory was resolved outside a repository")
	}
	if !strings.Contains(err.Error(), "resolving the hooks directory") {
		t.Errorf("the error does not name the act: %v", err)
	}
}

func TestADeclineRemovesNothing(t *testing.T) {
	root := makeAdopted(t)
	var out strings.Builder
	ctx := &command.Ctx{
		Stdout: &out, Stderr: &out,
		// A terminal that answers no.
		Terminal: &command.FakeTerminal{In: true, ConfirmAnswer: false},
		Root:     root, Adopted: true,
	}
	err := run(ctx, Deps{Git: hook.GitRunner(gitx.Run)}, nil)
	if err == nil {
		t.Fatal("a decline was not reported")
	}
	if _, statErr := os.Stat(filepath.Join(root, ".writrun")); statErr != nil {
		t.Error("a declined uninstall removed the kit anyway")
	}
}

func TestAnUnreadableAgentsIsAFault(t *testing.T) {
	root := makeAdopted(t)
	// A directory where the document should be: neither fenced, nor
	// the project's prose, nor absent.
	if err := os.Remove(filepath.Join(root, "AGENTS.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "AGENTS.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := plan(root, filepath.Join(root, ".git", "hooks", "commit-msg")); err == nil {
		t.Error("a directory where AGENTS.md should be was read as a document")
	}
}

func TestApplyReportsWhatItCouldNotRemove(t *testing.T) {
	skipAsRoot(t)
	root := makeAdopted(t)
	readOnly(t, root)

	r := &removal{root: root, dirs: []string{".writrun"}, hookState: hook.Absent}
	if err := r.apply(); err == nil {
		t.Fatal("removing under an unwritable root succeeded")
	}

	r = &removal{root: root, files: []string{"WRITRUN.md"}, hookState: hook.Absent}
	if err := r.apply(); err == nil {
		t.Fatal("removing a file under an unwritable root succeeded")
	}

	r = &removal{root: root, agentsWhole: true, hookState: hook.Absent}
	if err := r.apply(); err == nil {
		t.Fatal("removing AGENTS.md under an unwritable root succeeded")
	}

	// The edit rewrites a file that is already there, and a read-only
	// directory does not stop that — the document itself has to be.
	if err := os.Chmod(filepath.Join(root, "AGENTS.md"), 0o444); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(root, "AGENTS.md"), 0o644) })
	r = &removal{root: root, agents: []byte("# Ours\n"), hookState: hook.Absent}
	if err := r.apply(); err == nil {
		t.Fatal("editing an unwritable AGENTS.md succeeded")
	}
}

func TestApplyReportsAHookItCouldNotRemove(t *testing.T) {
	skipAsRoot(t)
	root := t.TempDir()
	hooks := filepath.Join(root, "hooks")
	if err := os.Mkdir(hooks, 0o755); err != nil {
		t.Fatal(err)
	}
	installed := filepath.Join(hooks, "commit-msg")
	if err := hook.Install(installed); err != nil {
		t.Fatal(err)
	}
	readOnly(t, hooks)

	r := &removal{root: root, hookAt: installed, hookState: hook.Ours}
	if err := r.apply(); err == nil {
		t.Fatal("removing a hook under an unwritable directory succeeded")
	}
}
