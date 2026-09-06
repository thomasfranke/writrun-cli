package updatecmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/kittag"

	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

func TestNewNamesTheCommand(t *testing.T) {
	c := New(Deps{Tag: newTag})
	if c.Name != "update" {
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

func TestNewDefaultsTheSource(t *testing.T) {
	// An empty Source means the canonical repository, resolved once at
	// wiring time rather than at every call.
	root := t.TempDir()
	write(t, root, ".writrun/VERSION", newTag+"\n")
	var out strings.Builder
	ctx := &command.Ctx{Stdout: &out, Stderr: &out, Terminal: &command.FakeTerminal{}, Root: root, Yes: true}
	if err := New(Deps{Tag: newTag, Git: gitx.Run, Files: vfs.OS{}}).Run(ctx, nil); err != nil {
		t.Fatalf("the stand-down path needs no source: %v", err)
	}
	if !strings.Contains(out.String(), "Already at") {
		t.Errorf("unexpected output: %s", out.String())
	}
}

// TestThePlanAnswersEveryVerb replaces TestDiffFileAnswersEveryVerb:
// diffFile is gone with the closed list it served, and the same four
// answers are now the plan's over one walk of the template.
func TestThePlanAnswersEveryVerb(t *testing.T) {
	root, template := t.TempDir(), t.TempDir()
	write(t, root, ".writrun/VERSION", oldTag+"\n")
	write(t, template, ".writrun/VERSION", newTag+"\n")

	// Only the template has it: added.
	write(t, template, ".writrun/skills/select/SKILL.md", "# Select\n")
	// Both, different: changed.
	write(t, root, ".writrun/scripts/take.sh", "echo take\n")
	write(t, template, ".writrun/scripts/take.sh", "echo take, reworded\n")
	// Both, identical: nothing to say.
	write(t, root, ".writrun/templates/task.md", "# Task\n")
	write(t, template, ".writrun/templates/task.md", "# Task\n")
	// Only the repository has it: the tag dropped it, so it goes.
	write(t, root, ".writrun/scripts/gone.sh", "echo gone\n")

	r, err := plan(vfs.OS{}, root, template, oldTag, newTag)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	want := map[string]verb{
		".writrun/skills/select/SKILL.md": added,
		".writrun/scripts/take.sh":        changed,
		".writrun/scripts/gone.sh":        removed,
		kittag.Rel:                        changed,
	}
	got := map[string]verb{}
	for _, c := range r.changes {
		got[c.rel] = c.verb
	}
	for rel, v := range want {
		if got[rel] != v {
			t.Errorf("%s = %q, want %q", rel, got[rel], v)
		}
	}
	if _, there := got[".writrun/templates/task.md"]; there {
		t.Error("an identical file was named as a change")
	}
}

// TestAFileNoListNamesIsRefreshed is the rule's own test: a tag that
// adds a file needs no Go change for update to write it
// (docs/technical/engineering/coupling.md).
func TestAFileNoListNamesIsRefreshed(t *testing.T) {
	root, template := t.TempDir(), t.TempDir()
	write(t, root, ".writrun/VERSION", oldTag+"\n")
	write(t, template, ".writrun/VERSION", newTag+"\n")
	write(t, template, ".writrun/a-folder-nobody-anticipated/thing.md", "# New\n")
	write(t, template, ".github/workflows/writrun-invented.yml", "name: invented\n")

	r, err := plan(vfs.OS{}, root, template, oldTag, newTag)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if err := r.apply(); err != nil {
		t.Fatalf("apply: %v", err)
	}
	for _, rel := range []string{
		".writrun/a-folder-nobody-anticipated/thing.md",
		".github/workflows/writrun-invented.yml",
	} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			t.Errorf("%s was planned and never written: %v", rel, err)
		}
	}
}

func TestReadTreeOnAnAbsentDirectoryIsEmpty(t *testing.T) {
	// A tag may add a folder the adopted kit never had; that is not a
	// failure, it is every file in it being new.
	got, err := readTree(vfs.OS{}, filepath.Join(t.TempDir(), "never-existed"))
	if err != nil {
		t.Fatalf("readTree: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("readTree = %v, want empty", got)
	}
}

func TestApplyRemovesAFileTheTagNoLongerShips(t *testing.T) {
	root, template := t.TempDir(), t.TempDir()
	write(t, root, ".writrun/skills/select/SKILL.md", "# Select\n")
	write(t, root, ".writrun/VERSION", oldTag+"\n")
	// The template ships no skills/ at all.
	write(t, template, ".writrun/VERSION", newTag+"\n")
	write(t, template, ".writrun/scripts/take.sh", "echo take\n")

	r, err := plan(vfs.OS{}, root, template, oldTag, newTag)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if err := r.apply(); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".writrun/skills/select/SKILL.md")); err == nil {
		t.Error("a file the tag no longer ships survived")
	}
	if got := read(t, root, ".writrun/VERSION"); strings.TrimSpace(got) != newTag {
		t.Errorf("VERSION = %q", got)
	}
}

func TestApplyRemovesAWorkflowTheTagDropped(t *testing.T) {
	root, template := t.TempDir(), t.TempDir()
	rel := ".github/workflows/writrun-issues.yml"
	write(t, root, rel, "name: issues\n")
	write(t, root, ".writrun/VERSION", oldTag+"\n")

	r := &refresh{disk: vfs.OS{}, root: root, template: template, to: newTag,
		changes: []change{{rel: rel, verb: removed}, {rel: kittag.Rel, verb: changed}}}
	if err := r.apply(); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, rel)); err == nil {
		t.Error("a workflow the tag dropped survived")
	}
	// Removing what is already gone is not a failure.
	if err := r.apply(); err != nil {
		t.Fatalf("a second apply failed on an already-removed file: %v", err)
	}
}

func TestAWorkflowTheProjectWroteIsNeverRemoved(t *testing.T) {
	root, template := t.TempDir(), t.TempDir()
	write(t, root, ".writrun/VERSION", oldTag+"\n")
	write(t, root, ".github/workflows/tests.yml", "name: the project's own\n")
	write(t, template, ".writrun/VERSION", newTag+"\n")

	r, err := plan(vfs.OS{}, root, template, oldTag, newTag)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	for _, c := range r.changes {
		if c.rel == ".github/workflows/tests.yml" {
			t.Fatalf("the project's own workflow is planned to %s", c.verb)
		}
	}
}

func TestRenderSaysWhenOnlyTheTagMoved(t *testing.T) {
	r := &refresh{disk: vfs.OS{}, from: oldTag, to: newTag,
		changes: []change{{rel: kittag.Rel, verb: changed}}}
	if !r.empty() {
		t.Fatal("a refresh with nothing but the tag is not empty()")
	}
	var out strings.Builder
	r.render(&out)
	if !strings.Contains(out.String(), "Only the recorded tag differs") {
		t.Errorf("the stand-down is not said:\n%s", out.String())
	}
}

func TestCopyFileCarriesTheModeBit(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "take.sh")
	if err := os.WriteFile(src, []byte("echo take\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "nested", "deeper", "take.sh")
	if err := copyFile(vfs.OS{}, src, dst); err != nil {
		t.Fatalf("copyFile: %v", err)
	}
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("the file was not written: %v", err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("the executable bit was lost: %v", info.Mode())
	}
	if err := copyFile(vfs.OS{}, filepath.Join(dir, "not-there"), dst); err == nil {
		t.Error("copying a file that is not there was not an error")
	}
}
