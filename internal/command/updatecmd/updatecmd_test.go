package updatecmd

import (
	"strings"
	"testing"
)

func TestCompareTags(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"v0.0.03", "v0.0.03", 0},
		{"v0.0.3", "v0.0.03", 0}, // the padding is not identity
		{"v0.0.10", "v0.0.9", 1}, // which string order gets wrong
		{"v0.0.9", "v0.0.10", -1},
		{"v0.1.0", "v0.0.99", 1},
		{"v1.0", "v1.0.0", 0}, // a missing component is zero
		{"v1.0.1", "v1.0", 1},
		{"not-a-tag", "v0.0.1", 1}, // unreadable is never a downgrade
	}
	for _, tc := range cases {
		if got := compareTags(tc.a, tc.b); got != tc.want {
			t.Errorf("compareTags(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestRecordedTagRefusesAnEmptyFile(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".writrun/VERSION", "   \n")
	if _, err := recordedTag(root); err == nil {
		t.Fatal("an empty VERSION was accepted")
	}
	if _, err := recordedTag(t.TempDir()); err == nil {
		t.Fatal("a missing VERSION was accepted")
	}
}

func TestRefreshMovesTheKitAndLeavesTheProject(t *testing.T) {
	src := makeSource(t)
	root := makeAdopted(t)

	// The project answers its gate and inverts the deriving default.
	agents := read(t, root, "AGENTS.md")
	agents = strings.Replace(agents, "<!-- TODO — default: human reviews -->", "Thomas reviews before merge.", 1)
	agents = strings.Replace(agents, "Present the derived tasks in the session before opening the PR.", "Open the derived pull request directly.", 1)
	write(t, root, "AGENTS.md", agents)
	gitT(t, root, "add", "-A")
	gitT(t, root, "commit", "-q", "-m", "our answers")

	out, err := runUpdate(t, root, src)
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

	// What the project owns, byte for byte.
	if got := read(t, root, ".writrun/conventions/commits.md"); got != "# Our commits\n" {
		t.Errorf("the conventions were touched: %q", got)
	}
	if got := read(t, root, "docs/product/a-chapter.md"); got != "# Our own chapter\n" {
		t.Errorf("the project's docs were touched: %q", got)
	}
	if got := read(t, root, "work/tasks/task-0001-a-task.md"); got != "id: task-0001\n" {
		t.Errorf("the queue was touched: %q", got)
	}

	refreshed := read(t, root, "AGENTS.md")
	if !strings.Contains(refreshed, "Thomas reviews before merge.") {
		t.Error("the project's gates answer did not survive")
	}
	if !strings.Contains(refreshed, "Open the derived pull request directly.") {
		t.Error("the project's deriving default did not survive")
	}
	if strings.Contains(refreshed, "TODO — default: human reviews") {
		t.Error("the kit's empty gates row came back")
	}
	if !strings.Contains(refreshed, "The flow's text, reworded.") {
		t.Error("the refreshed prose did not land")
	}
	if !strings.Contains(refreshed, "Prose the project wrote.") {
		t.Error("bytes outside the fence were touched")
	}
}

func TestTheSameTagChangesNothing(t *testing.T) {
	src := makeSource(t)
	root := makeAdopted(t)
	write(t, root, ".writrun/VERSION", newTag+"\n")
	gitT(t, root, "add", "-A")
	gitT(t, root, "commit", "-q", "-m", "already current")

	out, err := runUpdate(t, root, src)
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
	src := makeSource(t)
	root := makeAdopted(t)
	write(t, root, ".writrun/VERSION", "v99.0.0\n")
	gitT(t, root, "add", "-A")
	gitT(t, root, "commit", "-q", "-m", "a kit from the future")

	out, err := runUpdate(t, root, src)
	if err == nil {
		t.Fatalf("the downgrade was accepted:\n%s", out)
	}
	if !strings.Contains(err.Error(), "downgrade") {
		t.Errorf("the refusal does not name it: %v", err)
	}
}

func TestADamagedFenceStopsEverything(t *testing.T) {
	src := makeSource(t)
	root := makeAdopted(t)
	agents := read(t, root, "AGENTS.md")
	write(t, root, "AGENTS.md", strings.Replace(agents, "<!-- writrun:end -->", "", 1))
	gitT(t, root, "add", "-A")
	gitT(t, root, "commit", "-q", "-m", "a damaged fence")

	out, err := runUpdate(t, root, src)
	if err == nil {
		t.Fatalf("a damaged fence was accepted:\n%s", out)
	}
	if !strings.Contains(err.Error(), "fenced section") {
		t.Errorf("the refusal does not name the fence: %v", err)
	}
	if got := read(t, root, ".writrun/VERSION"); strings.TrimSpace(got) != oldTag {
		t.Errorf("the tag moved to %q despite the refusal", got)
	}
	if diff := gitT(t, root, "status", "--porcelain"); strings.TrimSpace(diff) != "" {
		t.Errorf("something was written:\n%s", diff)
	}
}

func TestADirtyTreeIsRefused(t *testing.T) {
	src := makeSource(t)
	root := makeAdopted(t)
	write(t, root, ".writrun/scripts/take.sh", "echo edited by hand\n")

	out, err := runUpdate(t, root, src)
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
	src := makeSource(t)
	root := makeAdopted(t)
	if _, err := runUpdate(t, root, src, "v1.2.3"); err == nil {
		t.Fatal("an argument was accepted")
	}
}

func TestRenderNamesWhatItWillNotTouch(t *testing.T) {
	src := makeSource(t)
	root := makeAdopted(t)
	out, err := runUpdate(t, root, src)
	if err != nil {
		t.Fatalf("update: %v\n%s", err, out)
	}
	for _, want := range []string{oldTag + " → " + newTag, "untouched", ".writrun/conventions", "work", "AGENTS.md"} {
		if !strings.Contains(out, want) {
			t.Errorf("the plan does not name %q:\n%s", want, out)
		}
	}
}
