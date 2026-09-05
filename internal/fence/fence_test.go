package fence

import (
	"errors"
	"strings"
	"testing"
)

// kitSection is the shape the kit ships: two `yours` markers, one
// before the block it governs and one after it.
const kitSection = `## WritRun — working the queue

<!-- writrun:begin
     This section is WritRun's flow. -->

### Picking work

The flow's text.

### Human gates

<!-- yours: this table is the project's own answers; it survives updates. -->

| Transition | Who |
|---|---|
| Writing docs | <!-- TODO — default: human reviews --> |
| Everything else | Agent, autonomously. |

**The forge row is not optional.** Settings live outside the repository.

### Deriving work

Present the derived tasks in the session before opening the PR.
<!-- yours: keep, invert, or drop this default — it is the project's. -->

<!-- writrun:end -->`

func templateDoc() []byte {
	return []byte("# AGENTS.md — entry point\n\n<!-- TODO: one paragraph. -->\n\n" + kitSection + "\n")
}

// adopted is the kit section as a project has answered it: the gates
// table filled in, the deriving default inverted.
func adopted() []byte {
	s := strings.Replace(kitSection,
		"| Writing docs | <!-- TODO — default: human reviews --> |",
		"| Writing docs | Thomas writes or reviews before merge. |", 1)
	s = strings.Replace(s,
		"Present the derived tasks in the session before opening the PR.",
		"Open the derived pull request directly; no session review.", 1)
	return []byte("# A project\n\nIts own prose.\n\n" + s + "\n")
}

func TestSectionCutsHeadingThroughClosingMarker(t *testing.T) {
	got, err := Section(templateDoc())
	if err != nil {
		t.Fatalf("Section: %v", err)
	}
	if string(got) != kitSection {
		t.Errorf("Section returned:\n%s\n\nwant:\n%s", got, kitSection)
	}
}

func TestSectionWithoutFenceIsAnError(t *testing.T) {
	if _, err := Section([]byte("# AGENTS.md\n\nNo fence here.\n")); err == nil {
		t.Fatal("Section accepted a document with no fence")
	}
}

func TestGraftThenRemoveIsARoundTrip(t *testing.T) {
	original := []byte("# A project\n\nIts own prose.\n")
	section, err := Section(templateDoc())
	if err != nil {
		t.Fatalf("Section: %v", err)
	}
	grafted := Graft(original, section)
	back, only, err := Remove(grafted)
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if only {
		t.Error("Remove called a document with prose of its own a bare skeleton")
	}
	if string(back) != string(original) {
		t.Errorf("round trip changed the project's bytes:\n%q\nwant:\n%q", back, original)
	}
}

func TestRemoveNamesADocumentThatIsOnlyTheSection(t *testing.T) {
	_, only, err := Remove([]byte(kitSection + "\n"))
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if !only {
		t.Error("a document holding nothing but the section was not named as one")
	}
}

func TestRemoveKeepsEveryByteOutsideTheFence(t *testing.T) {
	doc := []byte("# A project\n\nBefore.\n\n" + kitSection + "\n\n## After\n\nTrailing prose.\n")
	back, _, err := Remove(doc)
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
	for _, want := range []string{"# A project", "Before.", "## After", "Trailing prose."} {
		if !strings.Contains(string(back), want) {
			t.Errorf("Remove dropped %q", want)
		}
	}
	if strings.Contains(string(back), "writrun:begin") {
		t.Error("Remove left the fence behind")
	}
}

func TestReplaceCarriesTheProjectsAnswersAcross(t *testing.T) {
	// The refreshed section reworded the kit's prose and re-emptied the
	// gates table — exactly what a new tag does.
	next := strings.Replace(kitSection, "The flow's text.", "The flow's text, reworded.", 1)
	out, err := Replace(adopted(), []byte(next))
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}
	got := string(out)
	if !strings.Contains(got, "| Writing docs | Thomas writes or reviews before merge. |") {
		t.Error("the project's gates table did not survive the refresh")
	}
	if strings.Contains(got, "<!-- TODO — default: human reviews -->") {
		t.Error("the kit's empty gates row overwrote the project's answer")
	}
	if !strings.Contains(got, "Open the derived pull request directly; no session review.") {
		t.Error("the project's deriving default did not survive the refresh")
	}
	if !strings.Contains(got, "The flow's text, reworded.") {
		t.Error("the refreshed prose did not land")
	}
	if !strings.Contains(got, "# A project") || !strings.Contains(got, "Its own prose.") {
		t.Error("Replace touched bytes outside the fence")
	}
}

func TestReplaceRefusesToDropAYoursBlock(t *testing.T) {
	// A tag that deleted the markers would silently take the gates
	// table with it.
	stripped := strings.Replace(kitSection,
		"<!-- yours: this table is the project's own answers; it survives updates. -->\n\n", "", 1)
	stripped = strings.Replace(stripped,
		"<!-- yours: keep, invert, or drop this default — it is the project's. -->\n", "", 1)
	_, err := Replace(adopted(), []byte(stripped))
	if err == nil {
		t.Fatal("Replace accepted a section that would drop the project's answers")
	}
	if !strings.Contains(err.Error(), "yours") {
		t.Errorf("the refusal does not name what was at stake: %v", err)
	}
}

func TestDamagedMarkersStopEverything(t *testing.T) {
	damaged := []byte("# A project\n\n## WritRun — working the queue\n\n<!-- writrun:begin -->\n\nNo closing marker.\n")
	if _, err := Replace(damaged, []byte(kitSection)); !errors.Is(err, ErrNoFence) {
		t.Errorf("Replace on a damaged fence: got %v, want ErrNoFence", err)
	}
	if _, _, err := Remove(damaged); !errors.Is(err, ErrNoFence) {
		t.Errorf("Remove on a damaged fence: got %v, want ErrNoFence", err)
	}
}
