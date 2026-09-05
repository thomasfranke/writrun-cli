package updatecmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/gitx"

	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

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
	err := run(ctx, Deps{Tag: newTag, Source: makeSource(t), Git: gitx.Run, Files: vfs.OS{}}, nil)
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

	cs, err := diffTree(vfs.OS{}, root, template, ".writrun/skills")
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

func TestRenderSaysWhenTheSectionAlreadyMatches(t *testing.T) {
	r := &refresh{disk: vfs.OS{}, from: oldTag, to: newTag, changes: []change{
		{rel: ".writrun/skills/select/SKILL.md", verb: changed},
		{rel: ".writrun/VERSION", verb: changed},
	}}
	var out strings.Builder
	r.render(&out)
	if !strings.Contains(out.String(), "already matches — left alone") {
		t.Errorf("an unchanged section was not said to be left alone:\n%s", out.String())
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
	if err := run(ctx, Deps{Tag: newTag, Source: src, Git: gitx.Run, Files: vfs.OS{}}, nil); err == nil {
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

// fakeAt is the refresh's two trees as the fake holds them: the kit the
// repository has, and the template the tag ships.
func fakeAt(t *testing.T) (*vfs.Fake, string, string) {
	t.Helper()
	disk := vfs.NewFake()
	root, template := "/repo", "/kit/template"
	disk.Seed(root+"/.writrun/VERSION", []byte(oldTag+"\n"), 0o644)
	disk.Seed(root+"/.writrun/skills/select/SKILL.md", []byte("# Select\n"), 0o644)
	disk.Seed(root+"/.github/workflows/writrun-check.yml", []byte("name: check\n"), 0o644)
	disk.Seed(template+"/.writrun/skills/select/SKILL.md", []byte("# Select, reworded\n"), 0o644)
	disk.Seed(template+"/.github/workflows/writrun-check.yml", []byte("name: check\n# reworded\n"), 0o644)
	return disk, root, template
}

func TestApplyReportsTheDirectoryItCouldNotReplace(t *testing.T) {
	disk, root, template := fakeAt(t)
	boom := errors.New("the skills folder will not go")
	disk.Fail(root+"/.writrun/skills", boom)

	r := &refresh{disk: disk, root: root, template: template, to: newTag,
		dirs: []string{".writrun/skills"}, agentsPath: root + "/AGENTS.md"}
	err := r.apply()
	if err == nil {
		t.Fatal("a refresh that cannot replace succeeded")
	}
	if !errors.Is(err, boom) {
		t.Errorf("the cause did not survive: %v", err)
	}
	if !strings.Contains(err.Error(), ".writrun/skills") {
		t.Errorf("the error does not name the directory: %v", err)
	}
}

func TestApplyReportsTheDirectoryTheTagDroppedAndCouldNotRemove(t *testing.T) {
	disk, root, _ := fakeAt(t)
	boom := errors.New("it stays")
	disk.Fail(root+"/.writrun/skills", boom)

	// The template ships no skills/ at all, so the removal is what fails.
	r := &refresh{disk: disk, root: root, template: "/kit/empty", to: newTag,
		dirs: []string{".writrun/skills"}, agentsPath: root + "/AGENTS.md"}
	if err := r.apply(); !errors.Is(err, boom) {
		t.Errorf("removing a dropped directory: %v", err)
	}
}

func TestApplyReportsTheWorkflowItCouldNotWrite(t *testing.T) {
	disk, root, template := fakeAt(t)
	rel := ".github/workflows/writrun-check.yml"
	boom := errors.New("the workflow is held open")
	disk.Fail(root+"/"+rel, boom)

	r := &refresh{disk: disk, root: root, template: template, to: newTag,
		changes:    []change{{rel: rel, verb: changed, src: template + "/" + rel, mode: 0o644}},
		agentsPath: root + "/AGENTS.md"}
	if err := r.apply(); !errors.Is(err, boom) {
		t.Errorf("writing a workflow that refuses: %v", err)
	}
}

func TestApplyReportsTheWorkflowItCouldNotRemove(t *testing.T) {
	disk, root, template := fakeAt(t)
	rel := ".github/workflows/writrun-check.yml"
	boom := errors.New("it stays")
	disk.Fail(root+"/"+rel, boom)

	r := &refresh{disk: disk, root: root, template: template, to: newTag,
		changes:    []change{{rel: rel, verb: removed}},
		agentsPath: root + "/AGENTS.md"}
	if err := r.apply(); !errors.Is(err, boom) {
		t.Errorf("removing a workflow that refuses: %v", err)
	}
}

func TestApplyReportsTheTagItCouldNotRecord(t *testing.T) {
	disk, root, template := fakeAt(t)
	boom := errors.New("VERSION is read-only")
	disk.Fail(root+"/.writrun/VERSION", boom)

	r := &refresh{disk: disk, root: root, template: template, to: newTag,
		agentsPath: root + "/AGENTS.md"}
	err := r.apply()
	if !errors.Is(err, boom) {
		t.Fatalf("recording the tag: %v", err)
	}
	if !strings.Contains(err.Error(), "recording the tag") {
		t.Errorf("the error does not name the act: %v", err)
	}
}

func TestApplyReportsTheDocumentItCouldNotRefresh(t *testing.T) {
	disk, root, template := fakeAt(t)
	boom := errors.New("AGENTS.md is held open")
	disk.Fail(root+"/AGENTS.md", boom)

	r := &refresh{disk: disk, root: root, template: template, to: newTag,
		agentsPath: root + "/AGENTS.md", agents: []byte("# refreshed\n")}
	err := r.apply()
	if !errors.Is(err, boom) {
		t.Fatalf("refreshing the document: %v", err)
	}
	if !strings.Contains(err.Error(), "AGENTS.md") {
		t.Errorf("the error does not name the document: %v", err)
	}
}

func TestDiffTreeReportsATreeItCannotWalk(t *testing.T) {
	disk, root, template := fakeAt(t)
	disk.Fail(root+"/.writrun/skills", errors.New("the tree cannot be read"))
	if _, err := diffTree(disk, root, template, ".writrun/skills"); err == nil {
		t.Error("a tree that cannot be walked was diffed all the same")
	}
}

func TestReadTreeReportsAFileItCannotRead(t *testing.T) {
	disk, root, _ := fakeAt(t)
	disk.Fail(root+"/.writrun/skills/select/SKILL.md", errors.New("that file, no"))
	if _, err := readTree(disk, root+"/.writrun/skills"); err == nil {
		t.Error("a file that cannot be read was read")
	}
}

func TestCopyFileReportsTheDirectoryItCouldNotCreate(t *testing.T) {
	disk, _, template := fakeAt(t)
	disk.Fail("/out/nested", errors.New("nowhere to put it"))
	if err := copyFile(disk, template+"/.github/workflows/writrun-check.yml", "/out/nested/x.yml", 0o644); err == nil {
		t.Error("copying under a directory that cannot be made succeeded")
	}
}

func TestCopyTreeReportsWhatItCouldNotRead(t *testing.T) {
	disk := vfs.NewFake()
	if err := copyTree(disk, "/not-there", "/out"); err == nil {
		t.Error("copying a tree that is not there succeeded")
	}
}
