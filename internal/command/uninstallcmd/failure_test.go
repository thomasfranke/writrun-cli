package uninstallcmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/hook"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// adoptedFake is a repository as the fake holds it: everything init
// installs, plus the project's own record beside it.
func adoptedFake(t *testing.T) (*vfs.Fake, string) {
	t.Helper()
	disk := vfs.NewFake()
	root := "/repo"
	disk.Seed(root+"/AGENTS.md", []byte(projectAgents), 0o644)
	disk.Seed(root+"/WRITRUN.md", []byte("# This project uses WritRun\n"), 0o644)
	disk.Seed(root+"/docs/writrun-instructions.md", []byte("# How to work this kit\n"), 0o644)
	disk.Seed(root+"/.writrun/VERSION", []byte("v9.9.9\n"), 0o644)
	disk.Seed(root+"/.writrun/scripts/take.sh", []byte("echo take\n"), 0o755)
	for _, wf := range []string{"approve", "check", "issues", "progress"} {
		disk.Seed(root+"/.github/workflows/writrun-"+wf+".yml", []byte("name: "+wf+"\n"), 0o644)
	}
	disk.Seed(root+"/docs/product/a-chapter.md", []byte("# Our own chapter\n"), 0o644)
	disk.Seed(root+"/work/tasks/task-0001-a-task.md", []byte("id: task-0001\n"), 0o644)
	return disk, root
}

func TestApplyReportsTheDirectoryItCouldNotRemove(t *testing.T) {
	disk, root := adoptedFake(t)
	boom := errors.New("the kit will not go")
	disk.Fail(root+"/.writrun", boom)

	r := &removal{disk: disk, root: root, dirs: []string{".writrun"}, hookState: hook.Absent}
	err := r.apply()
	if err == nil {
		t.Fatal("a removal that cannot remove succeeded")
	}
	if !errors.Is(err, boom) {
		t.Errorf("the cause did not survive: %v", err)
	}
	if !strings.Contains(err.Error(), ".writrun") {
		t.Errorf("the error does not name what stayed: %v", err)
	}
}

func TestApplyReportsTheFileItCouldNotRemove(t *testing.T) {
	disk, root := adoptedFake(t)
	boom := errors.New("that one stays")
	disk.Fail(root+"/WRITRUN.md", boom)

	r := &removal{disk: disk, root: root, files: []string{"WRITRUN.md"}, hookState: hook.Absent}
	if err := r.apply(); !errors.Is(err, boom) {
		t.Errorf("removing a file that refuses: %v", err)
	}
}

func TestApplyReportsTheDocumentItCouldNotWrite(t *testing.T) {
	disk, root := adoptedFake(t)
	boom := errors.New("AGENTS.md is held open")
	disk.Fail(root+"/AGENTS.md", boom)

	r := &removal{disk: disk, root: root, agents: []byte("# Ours\n"), hookState: hook.Absent}
	if err := r.apply(); !errors.Is(err, boom) {
		t.Errorf("editing a document that refuses: %v", err)
	}

	r = &removal{disk: disk, root: root, agentsWhole: true, hookState: hook.Absent}
	if err := r.apply(); !errors.Is(err, boom) {
		t.Errorf("removing a document that refuses: %v", err)
	}
}

func TestApplyReportsTheHookItCouldNotRemove(t *testing.T) {
	disk, root := adoptedFake(t)
	hookAt := root + "/.git/hooks/commit-msg"
	if err := hook.Install(disk, hookAt); err != nil {
		t.Fatal(err)
	}
	boom := errors.New("the hook will not go")
	disk.Fail(hookAt, boom)

	r := &removal{disk: disk, root: root, hookAt: hookAt, hookState: hook.Ours}
	if err := r.apply(); !errors.Is(err, boom) {
		t.Errorf("removing a hook that refuses: %v", err)
	}
}

func TestApplyToleratesWhatIsAlreadyGone(t *testing.T) {
	disk := vfs.NewFake()
	disk.SeedDir("/repo")
	r := &removal{
		disk:      disk,
		root:      "/repo",
		files:     []string{"WRITRUN.md"},
		hookAt:    "/repo/.git/hooks/commit-msg",
		hookState: hook.Ours,
	}
	// Neither the file nor the hook is there; a removal that finds
	// nothing to remove is done, not broken.
	if err := r.apply(); err != nil {
		t.Fatalf("apply over an already-empty repository: %v", err)
	}
}

func TestAnUnreadableAgentsIsAFault(t *testing.T) {
	disk, root := adoptedFake(t)
	boom := errors.New("AGENTS.md cannot be read")
	disk.Fail(root+"/AGENTS.md", boom)

	if _, err := plan(disk, root, root+"/.git/hooks/commit-msg"); !errors.Is(err, boom) {
		t.Errorf("a document that cannot be read was planned over: %v", err)
	}
}

func TestPlanSeparatesWhatGoesFromWhatIsGone(t *testing.T) {
	disk, root := adoptedFake(t)
	// A named file deleted by hand is `gone`; a namespaced one is
	// simply not listed, because the folder is read rather than a list
	// of names checked against it.
	if err := disk.Remove(root + "/WRITRUN.md"); err != nil {
		t.Fatal(err)
	}
	if err := disk.Remove(root + "/.github/workflows/writrun-issues.yml"); err != nil {
		t.Fatal(err)
	}
	r, err := plan(disk, root, root+"/.git/hooks/commit-msg")
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if len(r.dirs) != 1 || r.dirs[0] != ".writrun" {
		t.Errorf("the kit directory is not in the removal set: %v", r.dirs)
	}
	var goneNamed bool
	for _, g := range r.gone {
		if g == "WRITRUN.md" {
			goneNamed = true
		}
	}
	if !goneNamed {
		t.Errorf("a file deleted by hand is not in the gone set: %v", r.gone)
	}
	for _, f := range r.files {
		if f == ".github/workflows/writrun-issues.yml" {
			t.Error("a workflow that is not there is named for removal")
		}
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
	if err := New(Deps{Git: gitx.Run, Files: vfs.OS{}}).Run(ctx, nil); err != nil {
		t.Fatalf("uninstall through New: %v\n%s", err, out.String())
	}
	if exists(t, root, ".writrun") {
		t.Error("the kit survived a run through the wired command")
	}
}

func TestRunReportsAGitThatCannotAnswer(t *testing.T) {
	// The hooks directory is git's to name; outside a repository there
	// is no answer, and nothing may be removed on a guess.
	var out strings.Builder
	ctx := &command.Ctx{Stdout: &out, Stderr: &out, Terminal: &command.FakeTerminal{}, Root: t.TempDir()}
	err := run(ctx, Deps{Git: gitx.Run, Files: vfs.OS{}}, nil)
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
		Terminal: &command.FakeTerminal{In: true, ConfirmAnswer: false},
		Root:     root, Adopted: true,
	}
	if err := run(ctx, Deps{Git: gitx.Run, Files: vfs.OS{}}, nil); err == nil {
		t.Fatal("a decline was not reported")
	}
	if !exists(t, root, ".writrun") {
		t.Error("a declined uninstall removed the kit anyway")
	}
}

func TestARemovalThatFailedHalfwaySaysSo(t *testing.T) {
	root := makeAdopted(t)
	disk := vfs.NewFake()
	// The plan is made against the real tree, then applied against a
	// fake that refuses: what is asserted is the message, not the OS.
	r, err := plan(vfs.OS{}, root, root+"/.git/hooks/commit-msg")
	if err != nil {
		t.Fatal(err)
	}
	r.disk = disk
	disk.SeedDir(root + "/.writrun")
	disk.Fail(root+"/.writrun", errors.New("no"))
	if err := r.apply(); err == nil {
		t.Fatal("a removal that could not write succeeded")
	}
}

func TestAnUnknownFlagIsRefused(t *testing.T) {
	root := makeAdopted(t)
	if _, err := runUninstall(t, root, "--nope"); err == nil {
		t.Fatal("an unknown flag was accepted")
	}
}

func TestAKitDirectoryAlreadyGoneIsNamed(t *testing.T) {
	// A kit somebody deleted by hand: what remains is removed, what was
	// already gone is listed rather than reported as a failure
	// (spec-0005, edge cases).
	disk, root := adoptedFake(t)
	if err := disk.RemoveAll(root + "/.writrun"); err != nil {
		t.Fatal(err)
	}
	r, err := plan(disk, root, root+"/.git/hooks/commit-msg")
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	var named bool
	for _, g := range r.gone {
		if g == ".writrun" {
			named = true
		}
	}
	if !named {
		t.Errorf("a kit directory already removed was not named: %v", r.gone)
	}
	if len(r.dirs) != 0 {
		t.Errorf("it is also in the removal set: %v", r.dirs)
	}
}

func TestAHookThatCannotBeReadStopsThePlan(t *testing.T) {
	disk, root := adoptedFake(t)
	hookAt := root + "/.git/hooks/commit-msg"
	if err := hook.Install(disk, hookAt); err != nil {
		t.Fatal(err)
	}
	boom := errors.New("the hook cannot be read")
	disk.Fail(hookAt, boom)

	if _, err := plan(disk, root, hookAt); !errors.Is(err, boom) {
		t.Errorf("a hook that cannot be read was planned around: %v", err)
	}
}

func TestARemovalThatFailedHalfwayNamesTheState(t *testing.T) {
	// What the user is told when the tree is half-removed: rerun, and
	// the command finishes what it started.
	disk, root := adoptedFake(t)
	// The plan reads it and finds it there; the removal is what fails.
	disk.FailOp("removeall", root+"/.writrun", errors.New("no"))
	var out strings.Builder
	ctx := &command.Ctx{
		Stdout: &out, Stderr: &out,
		Terminal: &command.FakeTerminal{},
		Root:     root, Adopted: true, Yes: true,
	}
	err := run(ctx, Deps{Git: func(string, ...string) (string, error) {
		return root + "/.git/hooks/commit-msg", nil
	}, Files: disk}, nil)
	if err == nil {
		t.Fatalf("a removal that could not write succeeded:\n%s", out.String())
	}
	if !strings.Contains(err.Error(), "the removal is partial") {
		t.Errorf("the error does not say what state the tree is in: %v", err)
	}
	if !strings.Contains(err.Error(), "rerun writrun uninstall") {
		t.Errorf("the error does not say how to finish it: %v", err)
	}
}

func TestAPlanThatCannotBeMadeStopsTheRun(t *testing.T) {
	disk, root := adoptedFake(t)
	disk.Fail(root+"/AGENTS.md", errors.New("unreadable"))
	var out strings.Builder
	ctx := &command.Ctx{
		Stdout: &out, Stderr: &out,
		Terminal: &command.FakeTerminal{},
		Root:     root, Adopted: true, Yes: true,
	}
	err := run(ctx, Deps{Git: func(string, ...string) (string, error) {
		return root + "/.git/hooks/commit-msg", nil
	}, Files: disk}, nil)
	if err == nil {
		t.Fatal("a plan that could not be made was applied")
	}
	if strings.Contains(err.Error(), "the removal is partial") {
		t.Errorf("a failure before any write was reported as a partial removal: %v", err)
	}
}
