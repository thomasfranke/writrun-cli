package updatecmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

func TestRecordedTagRefusesAnEmptyFile(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".writrun/VERSION", "   \n")
	if _, err := recordedTag(vfs.OS{}, root); err == nil {
		t.Fatal("an empty VERSION was accepted")
	}
	if _, err := recordedTag(vfs.OS{}, t.TempDir()); err == nil {
		t.Fatal("a missing VERSION was accepted")
	}
}

// TestRefreshMovesTheKitAndLeavesTheProject is update's one case
// against the real fetch: a local WritRun repository, cloned at the
// tag, so the fake is compared with the thing it fakes rather than
// assumed equal to it (spec-0016).
func TestRefreshMovesTheKitAndLeavesTheProject(t *testing.T) {
	src := makeSource(t)
	root := makeAdopted(t)

	// The project answers its own gates, in the file that holds them.
	write(t, root, "writrun/gates.md", "# Human gates\n\n| Transition | Who |\n|---|---|\n| Writing docs | The maintainer reviews before merge. |\n")
	gitT(t, root, "add", "-A")
	gitT(t, root, "commit", "-q", "-m", "our answers")

	out, err := runUpdate(t, root, Deps{Source: src, Kit: realKit()})
	if err != nil {
		t.Fatalf("update: %v\n%s", err, out)
	}

	if got := read(t, root, ".writrun/VERSION"); strings.TrimSpace(got) != newTag {
		t.Errorf("VERSION = %q, want %s", got, newTag)
	}
	if got := read(t, root, ".writrun/skills/select/SKILL.md"); !strings.Contains(got, "reworded") {
		t.Error("the refreshed skill did not land")
	}
	if got := read(t, root, ".writrun/templates/spec.md"); !strings.Contains(got, "# Spec") {
		t.Error("a file the new tag adds was not written")
	}
	if got := read(t, root, ".github/workflows/writrun-check.yml"); !strings.Contains(got, "reworded") {
		t.Error("the reworded workflow was not rewritten")
	}

	// The three files that reach no refresh list, and would have stayed
	// at the tag that installed them under the closed inventory.
	if got := read(t, root, ".writrun/AGENTS.md"); !strings.Contains(got, "reworded") {
		t.Error("the kit's own AGENTS.md was not refreshed")
	}
	if got := read(t, root, ".github/workflows/writrun-intake.yml"); !strings.Contains(got, "intake") {
		t.Error("a workflow the new tag adds was not written")
	}
	if got := read(t, root, ".github/ISSUE_TEMPLATE/writrun-report.yml"); !strings.Contains(got, "report") {
		t.Error("the issue template the new tag adds was not written")
	}

	// What the project owns, byte for byte.
	if got := read(t, root, "writrun/conventions/commits.md"); got != "# Our commits\n" {
		t.Errorf("the conventions were touched: %q", got)
	}
	if got := read(t, root, "writrun/settings.json"); !strings.Contains(got, `"stage": 3`) {
		t.Errorf("the settings were touched: %q", got)
	}
	if got := read(t, root, "docs/product/a-chapter.md"); got != "# Our own chapter\n" {
		t.Errorf("the project's docs were touched: %q", got)
	}
	if got := read(t, root, "work/tasks/task-0001-a-task.md"); got != "id: task-0001\n" {
		t.Errorf("the queue was touched: %q", got)
	}
	if got := read(t, root, ".github/workflows/tests.yml"); got != "name: the project's own\n" {
		t.Errorf("a workflow the project wrote was touched: %q", got)
	}
	if got := read(t, root, "writrun/gates.md"); !strings.Contains(got, "The maintainer reviews before merge.") {
		t.Errorf("the project's gate answers did not survive: %q", got)
	}

	// AGENTS.md is the project's whole from v0.0.04 on.
	if got := read(t, root, "AGENTS.md"); got != agentsDoc {
		t.Errorf("AGENTS.md was rewritten:\n%q", got)
	}
}

// TestARefreshWritesNothingIntoTheProjectsHome covers the rule that
// replaced seeding. A project that never wrote its own gates is
// answered by the kit's default, so there is nothing for a refresh to
// put there — and a file the project did write is left whole.
func TestARefreshWritesNothingIntoTheProjectsHome(t *testing.T) {
	root := makeAdopted(t)
	// makeAdopted predates gates.md, the way a v0.0.03 adoption does.
	out, err := runUpdate(t, root, Deps{})
	if err != nil {
		t.Fatalf("update: %v\n%s", err, out)
	}
	if _, err := os.Stat(filepath.Join(root, "writrun", "gates.md")); !os.IsNotExist(err) {
		t.Fatalf("the refresh wrote into the project's home: %v", err)
	}

	write(t, root, "writrun/gates.md", "# Human gates\n\n| Transition | Who |\n|---|---|\n| Writing docs | Ours. |\n")
	write(t, root, ".writrun/VERSION", oldTag+"\n")
	// One kit file put back a tag, so the second plan has something to
	// render: an empty plan stands down before it names anything.
	write(t, root, ".writrun/skills/select/SKILL.md", "# Select\n")
	gitT(t, root, "add", "-A")
	gitT(t, root, "commit", "-q", "-m", "answered, and back a tag")

	out, err = runUpdate(t, root, Deps{})
	if err != nil {
		t.Fatalf("the second update: %v\n%s", err, out)
	}
	if got := read(t, root, "writrun/gates.md"); !strings.Contains(got, "Ours.") {
		t.Errorf("the second refresh overwrote the answers: %q", got)
	}
}

func TestTheSameTagChangesNothing(t *testing.T) {
	root := makeAdopted(t)
	write(t, root, ".writrun/VERSION", newTag+"\n")
	gitT(t, root, "add", "-A")
	gitT(t, root, "commit", "-q", "-m", "already current")

	out, err := runUpdate(t, root, Deps{})
	if err != nil {
		t.Fatalf("update: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Already at WritRun "+newTag) {
		t.Errorf("the stand-down was not said:\n%s", out)
	}
	if diff := gitT(t, root, "status", "--porcelain"); strings.TrimSpace(diff) != "" {
		t.Errorf("something was written:\n%s", diff)
	}
}

func TestADowngradeIsRefused(t *testing.T) {
	root := makeAdopted(t)
	write(t, root, ".writrun/VERSION", "v99.0.0\n")
	gitT(t, root, "add", "-A")
	gitT(t, root, "commit", "-q", "-m", "a kit from the future")

	out, err := runUpdate(t, root, Deps{})
	if err == nil {
		t.Fatalf("the downgrade was accepted:\n%s", out)
	}
	if !strings.Contains(err.Error(), "downgrade") {
		t.Errorf("the refusal does not name it: %v", err)
	}
}

// TestALegacyFenceIsNamedAndNotTouched inverts
// TestADamagedFenceStopsEverything. The fence was what a refresh
// rewrote, so a damaged one stopped everything; from v0.0.04 the whole
// of AGENTS.md is the project's, so a leftover section is named and
// left exactly as it is.
func TestALegacyFenceIsNamedAndNotTouched(t *testing.T) {
	root := makeAdopted(t)
	write(t, root, "AGENTS.md", legacyAgents)
	gitT(t, root, "add", "-A")
	gitT(t, root, "commit", "-q", "-m", "still on the fenced shape")

	out, err := runUpdate(t, root, Deps{})
	if err != nil {
		t.Fatalf("update: %v\n%s", err, out)
	}
	if !strings.Contains(out, "still carries a writrun:begin/writrun:end section") {
		t.Errorf("the plan does not name the stale section:\n%s", out)
	}
	if got := read(t, root, "AGENTS.md"); got != legacyAgents {
		t.Errorf("AGENTS.md was rewritten:\n%q", got)
	}
	if got := read(t, root, ".writrun/VERSION"); strings.TrimSpace(got) != newTag {
		t.Errorf("the refresh did not proceed: VERSION = %q", got)
	}
}

// TestAnAbsentAgentsFileDoesNotStopTheRefresh: it is the project's
// file, and a refresh has no opinion about one that is not there.
func TestAnAbsentAgentsFileDoesNotStopTheRefresh(t *testing.T) {
	root := makeAdopted(t)
	gitT(t, root, "rm", "-q", "AGENTS.md")
	gitT(t, root, "commit", "-q", "-m", "no entry point")

	out, err := runUpdate(t, root, Deps{})
	if err != nil {
		t.Fatalf("update: %v\n%s", err, out)
	}
	if got := read(t, root, ".writrun/VERSION"); strings.TrimSpace(got) != newTag {
		t.Errorf("the refresh did not proceed: VERSION = %q", got)
	}
}

func TestADirtyTreeIsRefused(t *testing.T) {
	root := makeAdopted(t)
	write(t, root, ".writrun/scripts/take.sh", "echo edited by hand\n")

	out, err := runUpdate(t, root, Deps{})
	if err == nil {
		t.Fatalf("a dirty tree was accepted:\n%s", out)
	}
	if !strings.Contains(err.Error(), "dirty") {
		t.Errorf("the refusal does not name the tree: %v", err)
	}
	if got := read(t, root, ".writrun/scripts/take.sh"); !strings.Contains(got, "edited by hand") {
		t.Error("the uncommitted edit was overwritten by a run that refused")
	}
}

func TestAnUnexpectedArgumentIsRefused(t *testing.T) {
	root := makeAdopted(t)
	if _, err := runUpdate(t, root, Deps{}, "v1.2.3"); err == nil {
		t.Fatal("an argument was accepted")
	}
}

func TestRenderNamesWhatItWillNotTouch(t *testing.T) {
	root := makeAdopted(t)
	out, err := runUpdate(t, root, Deps{})
	if err != nil {
		t.Fatalf("update: %v\n%s", err, out)
	}
	for _, want := range []string{oldTag + " → " + newTag, "untouched", "writrun", "work", "AGENTS.md"} {
		if !strings.Contains(out, want) {
			t.Errorf("the plan does not name %q:\n%s", want, out)
		}
	}
}

// makeLegacyAdopted is a repository adopted before WritRun v0.0.05:
// its answers still sit inside the kit's home, where the kit's own
// scripts no longer look.
func makeLegacyAdopted(t *testing.T) string {
	t.Helper()
	root := makeAdopted(t)
	for _, rel := range []string{"writrun/settings.json", "writrun/conventions"} {
		if err := os.RemoveAll(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			t.Fatal(err)
		}
	}
	write(t, root, ".writrun/settings.json", "{\n  \"stage\": 3\n}\n")
	write(t, root, ".writrun/gates.md", "# Human gates\n\n| Transition | Who |\n|---|---|\n| Writing docs | Ours. |\n")
	write(t, root, ".writrun/conventions/commits.md", "# Our commits\n")
	gitT(t, root, "add", "-A")
	gitT(t, root, "commit", "-q", "-m", "adopted before the two homes")
	return root
}

func TestTheMigrationCarriesTheAnswersAcrossOnce(t *testing.T) {
	root := makeLegacyAdopted(t)
	out, err := runUpdate(t, root, Deps{})
	if err != nil {
		t.Fatalf("update: %v\n%s", err, out)
	}
	for rel, want := range map[string]string{
		"writrun/settings.json":          `"stage": 3`,
		"writrun/gates.md":               "Ours.",
		"writrun/conventions/commits.md": "# Our commits",
	} {
		if got := read(t, root, rel); !strings.Contains(got, want) {
			t.Errorf("%s does not carry the answer: %q", rel, got)
		}
	}
	for _, rel := range []string{".writrun/settings.json", ".writrun/gates.md", ".writrun/conventions"} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); !os.IsNotExist(err) {
			t.Errorf("%s survived the migration: %v", rel, err)
		}
	}
	if !strings.Contains(out, "move") {
		t.Errorf("the plan did not name the migration:\n%s", out)
	}
}

// The content is the adopter's, and a migration that reformatted it
// would be editing an answer it was only carrying.
func TestTheMigrationPreservesTheContentByteForByte(t *testing.T) {
	root := makeLegacyAdopted(t)
	before := read(t, root, ".writrun/settings.json")
	if out, err := runUpdate(t, root, Deps{}); err != nil {
		t.Fatalf("update: %v\n%s", err, out)
	}
	if got := read(t, root, "writrun/settings.json"); got != before {
		t.Errorf("the migration rewrote the answer:\ngot  %q\nwant %q", got, before)
	}
}

// Two answers about one setting are the adopter's to reconcile. A merge
// would pick for them, and a silent overwrite would pick worse.
func TestBothAddressesKeepsTheNewOneAndNamesTheOld(t *testing.T) {
	root := makeLegacyAdopted(t)
	write(t, root, "writrun/settings.json", "{\n  \"stage\": 1\n}\n")
	gitT(t, root, "add", "-A")
	gitT(t, root, "commit", "-q", "-m", "both addresses answer")

	out, err := runUpdate(t, root, Deps{})
	if err != nil {
		t.Fatalf("update: %v\n%s", err, out)
	}
	if got := read(t, root, "writrun/settings.json"); !strings.Contains(got, `"stage": 1`) {
		t.Errorf("the new address did not win: %q", got)
	}
	if got := read(t, root, ".writrun/settings.json"); !strings.Contains(got, `"stage": 3`) {
		t.Errorf("the old address was touched: %q", got)
	}
	if !strings.Contains(out, "left") || !strings.Contains(out, "nothing is merged") {
		t.Errorf("the plan did not name what it left behind:\n%s", out)
	}
}

func TestAnAlreadyMigratedLayoutMovesNothing(t *testing.T) {
	root := makeAdopted(t)
	out, err := runUpdate(t, root, Deps{})
	if err != nil {
		t.Fatalf("update: %v\n%s", err, out)
	}
	if strings.Contains(out, "move ") {
		t.Errorf("the plan named a migration a migrated repository does not owe:\n%s", out)
	}
}

// The migration is a change to the repository, so it waits for the same
// yes every other change does.
func TestADeclinedRefreshMovesNothing(t *testing.T) {
	root := makeLegacyAdopted(t)
	var out bytes.Buffer
	ctx := &command.Ctx{
		Stdout:   &out,
		Stderr:   &out,
		Terminal: &command.FakeTerminal{In: true, Out: true, ConfirmAnswer: false},
		Root:     root,
		Adopted:  true,
	}
	if err := run(ctx, Deps{Tag: newTag, Source: sourceDefault, Git: gitx.Run, Files: vfs.OS{}, Kit: fakeKit(t)}, nil); err == nil {
		t.Fatal("the declined run reported no refusal")
	}
	if _, err := os.Stat(filepath.Join(root, ".writrun", "settings.json")); err != nil {
		t.Errorf("a declined refresh moved the answer anyway: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "writrun", "settings.json")); !os.IsNotExist(err) {
		t.Errorf("a declined refresh wrote the new address: %v", err)
	}
}

// A half-finished move leaves the folder without its files. Reading
// that as the project having answered would strand the answers the
// migration exists to deliver.
func TestAnEmptyFolderAtTheNewAddressIsNoAnswer(t *testing.T) {
	root := makeLegacyAdopted(t)
	if err := os.MkdirAll(filepath.Join(root, "writrun", "conventions"), 0o755); err != nil {
		t.Fatal(err)
	}
	out, err := runUpdate(t, root, Deps{})
	if err != nil {
		t.Fatalf("update: %v\n%s", err, out)
	}
	if got := read(t, root, "writrun/conventions/commits.md"); !strings.Contains(got, "# Our commits") {
		t.Errorf("the migration was blocked by an empty folder: %q", got)
	}
	if strings.Contains(out, "left         writrun/conventions") {
		t.Errorf("an empty folder was reported as an answer:\n%s", out)
	}
}
