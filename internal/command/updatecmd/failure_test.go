package updatecmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/gitx"
)

// readOnly makes dir unwritable for the rest of the test, so a refresh
// writing into it fails the way a checkout somebody else owns would.
func readOnly(t *testing.T, dir string) {
	t.Helper()
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
}

// readOnlyFile makes one file unwritable. A read-only *directory*
// still lets an existing file be overwritten, so a case about a failed
// write has to reach the file itself.
func readOnlyFile(t *testing.T, path string) {
	t.Helper()
	if err := os.Chmod(path, 0o444); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o644) })
}

func skipAsRoot(t *testing.T) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("root writes into a read-only directory all the same")
	}
}

func TestApplyReportsWhatItCouldNotWrite(t *testing.T) {
	skipAsRoot(t)
	root, template := t.TempDir(), t.TempDir()
	write(t, template, ".writrun/skills/select/SKILL.md", "# Select\n")
	write(t, root, ".writrun/skills/select/SKILL.md", "# Select\n")
	readOnly(t, filepath.Join(root, ".writrun"))

	r := &refresh{root: root, template: template, to: newTag,
		dirs: []string{".writrun/skills"}, agentsPath: filepath.Join(root, "AGENTS.md")}
	err := r.apply()
	if err == nil {
		t.Fatal("replacing a directory under an unwritable parent succeeded")
	}
	if !strings.Contains(err.Error(), ".writrun/skills") {
		t.Errorf("the error does not name the directory: %v", err)
	}
}

func TestApplyReportsAWorkflowItCouldNotWrite(t *testing.T) {
	skipAsRoot(t)
	root, template := t.TempDir(), t.TempDir()
	rel := ".github/workflows/writrun-check.yml"
	write(t, template, rel, "name: check\n")
	write(t, root, rel, "name: check, older\n")
	readOnly(t, filepath.Join(root, ".github", "workflows"))

	r := &refresh{root: root, template: template, to: newTag,
		changes:    []change{{rel: rel, verb: changed, src: filepath.Join(template, rel), mode: 0o644}},
		agentsPath: filepath.Join(root, "AGENTS.md")}
	if err := r.apply(); err == nil {
		t.Fatal("writing a workflow under an unwritable parent succeeded")
	}
}

func TestApplyReportsATagItCouldNotRecord(t *testing.T) {
	skipAsRoot(t)
	root := t.TempDir()
	write(t, root, ".writrun/VERSION", oldTag+"\n")
	// The file itself, not its directory: a read-only directory still
	// lets an existing file be overwritten.
	readOnlyFile(t, filepath.Join(root, ".writrun", "VERSION"))

	r := &refresh{root: root, template: t.TempDir(), to: newTag,
		agentsPath: filepath.Join(root, "AGENTS.md")}
	err := r.apply()
	if err == nil {
		t.Fatal("recording the tag into an unwritable directory succeeded")
	}
	if !strings.Contains(err.Error(), "recording the tag") {
		t.Errorf("the error does not name the act: %v", err)
	}
}

func TestApplyReportsAnAgentsItCouldNotRefresh(t *testing.T) {
	skipAsRoot(t)
	root := t.TempDir()
	write(t, root, ".writrun/VERSION", oldTag+"\n")
	write(t, root, "AGENTS.md", "# Ours\n")
	readOnlyFile(t, filepath.Join(root, "AGENTS.md"))

	r := &refresh{root: root, template: t.TempDir(), to: newTag,
		agentsPath: filepath.Join(root, "AGENTS.md"),
		agents:     []byte("# refreshed\n")}
	// The tag lands first and lands fine; what fails here is the
	// document.
	err := r.apply()
	if err == nil {
		t.Fatal("refreshing AGENTS.md under an unwritable root succeeded")
	}
	if !strings.Contains(err.Error(), "AGENTS.md") {
		t.Errorf("the error does not name the document: %v", err)
	}
}

func TestCopyTreeReportsWhatItCouldNotRead(t *testing.T) {
	if err := copyTree(filepath.Join(t.TempDir(), "not-there"), t.TempDir()); err == nil {
		t.Error("copying a tree that is not there succeeded")
	}
}

func TestDiffTreeReportsAnUnreadableTree(t *testing.T) {
	skipAsRoot(t)
	root, template := t.TempDir(), t.TempDir()
	write(t, root, ".writrun/skills/select/SKILL.md", "# Select\n")
	if err := os.Chmod(filepath.Join(root, ".writrun", "skills"), 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(root, ".writrun", "skills"), 0o755) })

	if _, err := diffTree(root, template, ".writrun/skills"); err == nil {
		t.Error("a tree that cannot be walked was diffed all the same")
	}
}

func TestRunReportsAnApplyThatFailedHalfway(t *testing.T) {
	skipAsRoot(t)
	src := makeSource(t)
	root := makeAdopted(t)
	readOnly(t, filepath.Join(root, ".writrun"))

	out, err := runUpdate(t, root, src)
	if err == nil {
		t.Fatalf("a refresh that could not write succeeded:\n%s", out)
	}
	if !strings.Contains(err.Error(), "the refresh is partial") {
		t.Errorf("the error does not say what state the tree is in: %v", err)
	}
}

func TestRunRefusesAnUnreadableAgents(t *testing.T) {
	src := makeSource(t)
	root := makeAdopted(t)
	if err := os.Remove(filepath.Join(root, "AGENTS.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := runUpdate(t, root, src); err == nil {
		t.Fatal("a repository with no AGENTS.md was refreshed")
	}
}

func TestRunReportsAGitThatCannotAnswer(t *testing.T) {
	// The tree is read through git; outside a repository there is no
	// answer, and a refresh may not proceed on a guess.
	root := t.TempDir()
	write(t, root, ".writrun/VERSION", oldTag+"\n")
	write(t, root, "AGENTS.md", agentsAt("The flow's text."))
	var out strings.Builder
	ctx := &command.Ctx{Stdout: &out, Stderr: &out, Terminal: &command.FakeTerminal{}, Root: root, Yes: true}
	err := run(ctx, Deps{Tag: newTag, Source: makeSource(t), Git: gitx.Run}, nil)
	if err == nil {
		t.Fatal("the working tree was read outside a repository")
	}
	if !strings.Contains(err.Error(), "reading the working tree") {
		t.Errorf("the error does not name the act: %v", err)
	}
}

func TestRunReportsAnUnreachableSource(t *testing.T) {
	root := makeAdopted(t)
	out, err := runUpdate(t, root, filepath.Join(t.TempDir(), "not-a-repository"))
	if err == nil {
		t.Fatalf("an unreachable source was accepted:\n%s", out)
	}
	if !strings.Contains(err.Error(), "nothing was written") {
		t.Errorf("the error does not say the tree is untouched: %v", err)
	}
}
