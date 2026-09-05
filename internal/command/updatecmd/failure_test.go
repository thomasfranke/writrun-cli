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

func TestDiffTreeNamesAFileTheTagDropped(t *testing.T) {
	root, template := t.TempDir(), t.TempDir()
	write(t, root, ".writrun/skills/gone/SKILL.md", "# Gone\n")
	write(t, template, ".writrun/skills/kept/SKILL.md", "# Kept\n")

	cs, err := diffTree(root, template, ".writrun/skills")
	if err != nil {
		t.Fatalf("diffTree: %v", err)
	}
	var sawRemoved, sawAdded bool
	for _, c := range cs {
		if c.verb == removed && strings.Contains(c.rel, "gone") {
			sawRemoved = true
		}
		if c.verb == added && strings.Contains(c.rel, "kept") {
			sawAdded = true
		}
	}
	if !sawRemoved {
		t.Errorf("a file only the repository has was not named for removal: %+v", cs)
	}
	if !sawAdded {
		t.Errorf("a file only the tag has was not named as added: %+v", cs)
	}
}

func TestReadTreeReportsAFileItCannotRead(t *testing.T) {
	skipAsRoot(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "unreadable.md")
	if err := os.WriteFile(path, []byte("# x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o644) })

	if _, err := readTree(dir); err == nil {
		t.Error("a file that cannot be read was read")
	}
}

func TestRenderSaysWhenTheSectionAlreadyMatches(t *testing.T) {
	r := &refresh{from: oldTag, to: newTag, changes: []change{
		{rel: ".writrun/skills/select/SKILL.md", verb: changed},
		{rel: ".writrun/VERSION", verb: changed},
	}}
	var out strings.Builder
	r.render(&out)
	if !strings.Contains(out.String(), "already matches — left alone") {
		t.Errorf("an unchanged section was not said to be left alone:\n%s", out.String())
	}
}

func TestApplyReportsADirectoryItCouldNotRemove(t *testing.T) {
	skipAsRoot(t)
	root, template := t.TempDir(), t.TempDir()
	write(t, root, ".writrun/skills/select/SKILL.md", "# Select\n")
	write(t, root, ".writrun/VERSION", oldTag+"\n")
	readOnly(t, filepath.Join(root, ".writrun"))

	// The template ships no skills/ at all, so the removal is what fails.
	r := &refresh{root: root, template: template, to: newTag,
		dirs: []string{".writrun/skills"}, agentsPath: filepath.Join(root, "AGENTS.md")}
	err := r.apply()
	if err == nil {
		t.Fatal("removing a directory under an unwritable parent succeeded")
	}
	if !strings.Contains(err.Error(), "removing .writrun/skills") {
		t.Errorf("the error does not name the act: %v", err)
	}
}

func TestApplyReportsAWorkflowItCouldNotRemove(t *testing.T) {
	skipAsRoot(t)
	root := t.TempDir()
	rel := ".github/workflows/writrun-issues.yml"
	write(t, root, rel, "name: issues\n")
	write(t, root, ".writrun/VERSION", oldTag+"\n")
	readOnly(t, filepath.Join(root, ".github", "workflows"))

	r := &refresh{root: root, template: t.TempDir(), to: newTag,
		changes:    []change{{rel: rel, verb: removed}},
		agentsPath: filepath.Join(root, "AGENTS.md")}
	if err := r.apply(); err == nil {
		t.Fatal("removing a workflow under an unwritable parent succeeded")
	}
}

func TestCopyFileReportsADirectoryItCouldNotCreate(t *testing.T) {
	skipAsRoot(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	if err := os.WriteFile(src, []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	guarded := filepath.Join(dir, "guarded")
	if err := os.Mkdir(guarded, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(guarded, 0o755) })

	if err := copyFile(src, filepath.Join(guarded, "nested", "dst.txt"), 0o644); err == nil {
		t.Error("copying under an uncreatable directory succeeded")
	}
}

func TestAnUnknownFlagIsRefused(t *testing.T) {
	root := makeAdopted(t)
	if _, err := runUpdate(t, root, makeSource(t), "--nope"); err == nil {
		t.Fatal("an unknown flag was accepted")
	}
}

func TestAKitWithNoRecordedTagIsRefused(t *testing.T) {
	root := makeAdopted(t)
	if err := os.Remove(filepath.Join(root, ".writrun", "VERSION")); err != nil {
		t.Fatal(err)
	}
	if _, err := runUpdate(t, root, makeSource(t)); err == nil {
		t.Fatal("a kit recording no tag was refreshed")
	}
}

func TestADeclineRefreshesNothing(t *testing.T) {
	src := makeSource(t)
	root := makeAdopted(t)
	var out strings.Builder
	ctx := &command.Ctx{
		Stdout: &out, Stderr: &out,
		Terminal: &command.FakeTerminal{In: true, ConfirmAnswer: false},
		Root:     root, Adopted: true,
	}
	if err := run(ctx, Deps{Tag: newTag, Source: src, Git: gitx.Run}, nil); err == nil {
		t.Fatal("a decline was not reported")
	}
	if got := read(t, root, ".writrun/VERSION"); strings.TrimSpace(got) != oldTag {
		t.Errorf("a declined refresh moved the tag to %q", got)
	}
}

func TestAPlanThatCannotBeMadeStopsTheRefresh(t *testing.T) {
	src := makeSource(t)
	root := makeAdopted(t)
	// The document's fence is intact, so the run reaches the plan; the
	// template's is not, so the merge inside it cannot be made.
	clone := t.TempDir()
	gitT(t, "", "clone", "-q", "--depth", "1", "--branch", newTag, src, filepath.Join(clone, "src"))
	broken := filepath.Join(clone, "src")
	write(t, broken, "template/AGENTS.md", "# No fence in the template.\n")
	gitT(t, broken, "add", "-A")
	gitT(t, broken, "commit", "-q", "-m", "a template with no fence")
	gitT(t, broken, "tag", "-f", newTag)

	if _, err := runUpdate(t, root, broken); err == nil {
		t.Fatal("a template with no fence was refreshed from")
	}
}

func TestOnlyTheTagMovingWritesOnlyTheTag(t *testing.T) {
	// The kit's files already match the new tag; only VERSION differs,
	// so the run says so and asks nothing.
	src := makeSource(t)
	root := makeAdopted(t)
	// Bring every kit-owned path up to newTag by hand, leaving the
	// recorded tag behind.
	write(t, root, "AGENTS.md", agentsAt("The flow's text, reworded."))
	write(t, root, ".writrun/skills/select/SKILL.md", "# Select, reworded\n")
	write(t, root, ".writrun/templates/spec.md", "# Spec\n")
	write(t, root, ".github/workflows/writrun-check.yml", "name: writrun check\n# reworded\n")
	gitT(t, root, "add", "-A")
	gitT(t, root, "commit", "-q", "-m", "the files, already current")

	out, err := runUpdate(t, root, src)
	if err != nil {
		t.Fatalf("update: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Only the recorded tag differs") {
		t.Errorf("the stand-down was not said:\n%s", out)
	}
	if got := read(t, root, ".writrun/VERSION"); strings.TrimSpace(got) != oldTag {
		t.Errorf("a run that asked nothing still wrote the tag: %q", got)
	}
}
