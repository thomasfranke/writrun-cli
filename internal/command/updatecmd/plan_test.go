package updatecmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/gitx"

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

func TestDiffFileAnswersEveryVerb(t *testing.T) {
	root, template := t.TempDir(), t.TempDir()
	rel := ".github/workflows/writrun-check.yml"

	// Neither side has it: nothing to say.
	c, err := diffFile(vfs.OS{}, root, template, rel)
	if err != nil || c != nil {
		t.Errorf("absent on both sides: change = %v, err = %v", c, err)
	}

	// Only the template has it: added.
	write(t, template, rel, "name: check\n")
	c, err = diffFile(vfs.OS{}, root, template, rel)
	if err != nil || c == nil || c.verb != added {
		t.Errorf("only in the template: change = %+v, err = %v", c, err)
	}

	// Both, identical: nothing to say.
	write(t, root, rel, "name: check\n")
	c, err = diffFile(vfs.OS{}, root, template, rel)
	if err != nil || c != nil {
		t.Errorf("identical on both sides: change = %v, err = %v", c, err)
	}

	// Both, different: changed.
	write(t, template, rel, "name: check\n# reworded\n")
	c, err = diffFile(vfs.OS{}, root, template, rel)
	if err != nil || c == nil || c.verb != changed {
		t.Errorf("different on the two sides: change = %+v, err = %v", c, err)
	}

	// Only the repository has it: the tag dropped it, so it goes.
	if err := os.Remove(filepath.Join(template, rel)); err != nil {
		t.Fatal(err)
	}
	c, err = diffFile(vfs.OS{}, root, template, rel)
	if err != nil || c == nil || c.verb != removed {
		t.Errorf("only in the repository: change = %+v, err = %v", c, err)
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

func TestMergeAgentsRefusesATemplateWithoutTheFence(t *testing.T) {
	template := t.TempDir()
	if _, err := mergeAgents(vfs.OS{}, template, []byte(agentsAt("x"))); err == nil {
		t.Fatal("a template with no AGENTS.md was accepted")
	}
	write(t, template, "AGENTS.md", "# No fence in here.\n")
	if _, err := mergeAgents(vfs.OS{}, template, []byte(agentsAt("x"))); err == nil {
		t.Fatal("a template AGENTS.md with no fence was accepted")
	}
}

func TestApplyRemovesADirectoryTheTagNoLongerShips(t *testing.T) {
	root, template := t.TempDir(), t.TempDir()
	write(t, root, ".writrun/skills/select/SKILL.md", "# Select\n")
	write(t, root, ".writrun/VERSION", oldTag+"\n")
	// The template ships no skills/ at all.
	write(t, template, ".writrun/scripts/take.sh", "echo take\n")

	r := &refresh{disk: vfs.OS{}, root: root, template: template, to: newTag,
		dirs: []string{".writrun/skills"}, agentsPath: filepath.Join(root, "AGENTS.md")}
	if err := r.apply(); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".writrun/skills")); err == nil {
		t.Error("a directory the tag no longer ships survived")
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
		changes:    []change{{rel: rel, verb: removed}, {rel: ".writrun/VERSION", verb: changed}},
		agentsPath: filepath.Join(root, "AGENTS.md")}
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

func TestRenderSaysWhenOnlyTheTagMoved(t *testing.T) {
	r := &refresh{disk: vfs.OS{}, from: oldTag, to: newTag,
		changes: []change{{rel: ".writrun/VERSION", verb: changed}}}
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
	if err := copyFile(vfs.OS{}, src, dst, 0o755); err != nil {
		t.Fatalf("copyFile: %v", err)
	}
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("the file was not written: %v", err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("the executable bit was lost: %v", info.Mode())
	}
	if err := copyFile(vfs.OS{}, filepath.Join(dir, "not-there"), dst, 0o644); err == nil {
		t.Error("copying a file that is not there was not an error")
	}
}
