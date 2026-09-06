package updatecmd

import (
	"strings"
	"testing"

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
	write(t, root, ".writrun/gates.md", "# Human gates\n\n| Transition | Who |\n|---|---|\n| Writing docs | Thomas reviews before merge. |\n")
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
	if got := read(t, root, ".writrun/conventions/commits.md"); got != "# Our commits\n" {
		t.Errorf("the conventions were touched: %q", got)
	}
	if got := read(t, root, ".writrun/settings.json"); !strings.Contains(got, `"stage": 3`) {
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
	if got := read(t, root, ".writrun/gates.md"); !strings.Contains(got, "Thomas reviews before merge.") {
		t.Errorf("the project's gate answers did not survive: %q", got)
	}

	// AGENTS.md is the project's whole from v0.0.04 on.
	if got := read(t, root, "AGENTS.md"); got != agentsDoc {
		t.Errorf("AGENTS.md was rewritten:\n%q", got)
	}
}

// TestTheSeedArrivesOnceAndIsNeverRewritten covers the one file a
// refresh writes only where the repository lacks it.
func TestTheSeedArrivesOnceAndIsNeverRewritten(t *testing.T) {
	root := makeAdopted(t)
	// makeAdopted predates gates.md, the way a v0.0.03 adoption does.
	out, err := runUpdate(t, root, Deps{})
	if err != nil {
		t.Fatalf("update: %v\n%s", err, out)
	}
	if got := read(t, root, ".writrun/gates.md"); !strings.Contains(got, "Transition") {
		t.Fatalf("the seed did not arrive: %q", got)
	}

	write(t, root, ".writrun/gates.md", "# Human gates\n\n| Transition | Who |\n|---|---|\n| Writing docs | Ours. |\n")
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
	if got := read(t, root, ".writrun/gates.md"); !strings.Contains(got, "Ours.") {
		t.Errorf("the second refresh overwrote the answers: %q", got)
	}
	if !strings.Contains(out, "yours; this tag ships one and it is left alone") {
		t.Errorf("the plan does not say the seed is kept:\n%s", out)
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
	for _, want := range []string{oldTag + " → " + newTag, "untouched", ".writrun/conventions", "work", "AGENTS.md"} {
		if !strings.Contains(out, want) {
			t.Errorf("the plan does not name %q:\n%s", want, out)
		}
	}
}
