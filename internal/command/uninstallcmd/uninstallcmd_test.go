package uninstallcmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/hook"

	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

func gitT(t *testing.T, dir string, args ...string) string {
	t.Helper()
	full := append([]string{"-c", "user.name=suite", "-c", "user.email=suite@test", "-c", "commit.gpgsign=false"}, args...)
	out, err := gitx.Run(dir, full...)
	if err != nil {
		t.Fatalf("git %s: %v", strings.Join(args, " "), err)
	}
	return out
}

func write(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, root, rel string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatalf("reading %s: %v", rel, err)
	}
	return string(content)
}

func exists(t *testing.T, root, rel string) bool {
	t.Helper()
	_, err := os.Stat(filepath.Join(root, rel))
	return err == nil
}

const projectAgents = `# A project

Rules the project already had.

## WritRun — working the queue

<!-- writrun:begin
     This section is WritRun's flow. -->

### Picking work

The flow's text.

<!-- writrun:end -->
`

// makeAdopted builds a repository carrying everything init installs,
// plus the project's own record beside it.
func makeAdopted(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	gitT(t, root, "init", "-q")
	write(t, root, "AGENTS.md", projectAgents)
	write(t, root, "WRITRUN.md", "# This project uses WritRun\n")
	write(t, root, "docs/writrun-instructions.md", "# How to work this kit\n")
	write(t, root, ".writrun/VERSION", "v9.9.9\n")
	write(t, root, ".writrun/settings.json", "{\n  \"stage\": 1\n}\n")
	write(t, root, ".writrun/scripts/take.sh", "echo take\n")
	for _, wf := range []string{"approve", "check", "issues", "progress"} {
		write(t, root, ".github/workflows/writrun-"+wf+".yml", "name: writrun "+wf+"\n")
	}
	write(t, root, ".github/workflows/tests.yml", "name: the project's own\n")
	write(t, root, "docs/product/a-chapter.md", "# Our own chapter\n")
	write(t, root, "work/tasks/task-0001-a-task.md", "id: task-0001\n")
	gitT(t, root, "add", ".")
	gitT(t, root, "commit", "-q", "-m", "adopt")
	return root
}

func hookAt(t *testing.T, root string) string {
	t.Helper()
	p, err := hook.Path(root, hook.GitRunner(gitx.Run))
	if err != nil {
		t.Fatalf("hook.Path: %v", err)
	}
	return p
}

func runUninstall(t *testing.T, root string, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	ctx := &command.Ctx{
		Stdout:   &out,
		Stderr:   &out,
		Terminal: &command.FakeTerminal{},
		Root:     root,
		Adopted:  true,
		Yes:      true,
	}
	err := run(ctx, Deps{Git: hook.GitRunner(gitx.Run), Files: vfs.OS{}}, args)
	return out.String(), err
}

func TestRemovesTheKitAndKeepsTheRecord(t *testing.T) {
	root := makeAdopted(t)
	installed := hookAt(t, root)
	if err := hook.Install(vfs.OS{}, installed); err != nil {
		t.Fatal(err)
	}

	out, err := runUninstall(t, root)
	if err != nil {
		t.Fatalf("uninstall: %v\n%s", err, out)
	}

	for _, rel := range []string{".writrun", "WRITRUN.md", "docs/writrun-instructions.md",
		".github/workflows/writrun-approve.yml", ".github/workflows/writrun-check.yml",
		".github/workflows/writrun-issues.yml", ".github/workflows/writrun-progress.yml"} {
		if exists(t, root, rel) {
			t.Errorf("%s survived the removal", rel)
		}
	}
	if _, statErr := os.Stat(installed); statErr == nil {
		t.Error("the commit-msg hook survived the removal")
	}

	if got := read(t, root, "work/tasks/task-0001-a-task.md"); got != "id: task-0001\n" {
		t.Errorf("the queue was touched: %q", got)
	}
	if got := read(t, root, "docs/product/a-chapter.md"); got != "# Our own chapter\n" {
		t.Errorf("the project's docs were touched: %q", got)
	}
	if got := read(t, root, ".github/workflows/tests.yml"); got != "name: the project's own\n" {
		t.Errorf("a workflow that is not the kit's was touched: %q", got)
	}

	agents := read(t, root, "AGENTS.md")
	if strings.Contains(agents, "writrun:begin") {
		t.Error("the fenced section survived")
	}
	if !strings.Contains(agents, "Rules the project already had.") {
		t.Error("what the project wrote outside the fence was lost")
	}
}

func TestAForeignHookIsLeftStanding(t *testing.T) {
	root := makeAdopted(t)
	installed := hookAt(t, root)
	if err := os.MkdirAll(filepath.Dir(installed), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(installed, []byte("#!/bin/sh\n# somebody else wrote this\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	out, err := runUninstall(t, root)
	if err != nil {
		t.Fatalf("uninstall: %v\n%s", err, out)
	}
	if !strings.Contains(out, "not the one init writes") {
		t.Errorf("the plan does not say the hook is left:\n%s", out)
	}
	content, readErr := os.ReadFile(installed)
	if readErr != nil {
		t.Fatalf("the foreign hook was removed: %v", readErr)
	}
	if !strings.Contains(string(content), "somebody else wrote this") {
		t.Error("the foreign hook was overwritten")
	}
}

func TestABareSkeletonAgentsGoesWhole(t *testing.T) {
	root := makeAdopted(t)
	// Everything before the section removed: the file is the kit's alone.
	agents := read(t, root, "AGENTS.md")
	write(t, root, "AGENTS.md", agents[strings.Index(agents, "## WritRun"):])

	out, err := runUninstall(t, root)
	if err != nil {
		t.Fatalf("uninstall: %v\n%s", err, out)
	}
	if !strings.Contains(out, "nothing in it but the kit") {
		t.Errorf("the plan does not name it for removal:\n%s", out)
	}
	if exists(t, root, "AGENTS.md") {
		t.Error("a bare skeleton AGENTS.md survived")
	}
}

func TestAnAgentsWithNoFenceIsLeftAlone(t *testing.T) {
	root := makeAdopted(t)
	write(t, root, "AGENTS.md", "# Ours alone\n\nNo fence here.\n")

	out, err := runUninstall(t, root)
	if err != nil {
		t.Fatalf("uninstall: %v\n%s", err, out)
	}
	if !strings.Contains(out, "no fenced WritRun section found") {
		t.Errorf("the plan does not say it is kept:\n%s", out)
	}
	if got := read(t, root, "AGENTS.md"); got != "# Ours alone\n\nNo fence here.\n" {
		t.Errorf("a file with no fence was edited: %q", got)
	}
}

func TestWhatIsAlreadyGoneIsNamed(t *testing.T) {
	root := makeAdopted(t)
	if err := os.Remove(filepath.Join(root, "WRITRUN.md")); err != nil {
		t.Fatal(err)
	}

	out, err := runUninstall(t, root)
	if err != nil {
		t.Fatalf("uninstall: %v\n%s", err, out)
	}
	if !strings.Contains(out, "already gone WRITRUN.md") {
		t.Errorf("a file deleted by hand was not named:\n%s", out)
	}
}

func TestBothSetsAreShown(t *testing.T) {
	root := makeAdopted(t)
	out, err := runUninstall(t, root)
	if err != nil {
		t.Fatalf("uninstall: %v\n%s", err, out)
	}
	for _, want := range []string{"remove       .writrun/", "stays        work/", "stays        docs/"} {
		if !strings.Contains(out, want) {
			t.Errorf("the plan does not show %q:\n%s", want, out)
		}
	}
}

func TestAnUnexpectedArgumentIsRefused(t *testing.T) {
	root := makeAdopted(t)
	if _, err := runUninstall(t, root, "everything"); err == nil {
		t.Fatal("an argument was accepted")
	}
}
