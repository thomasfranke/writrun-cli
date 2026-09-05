package reportcmd

import (
	"strings"
	"testing"
)

func TestSubstitutePutsTheObservationWhereThePlaceholderWas(t *testing.T) {
	got := substitute(generated, "One paragraph of what was seen.")
	want := `---
id: report-0009
status: open
task_ref: []
doc_ref: null
created: 2026-09-05T10:00:00Z
triaged: null
---

# Something that was noticed

One paragraph of what was seen.
`
	if got != want {
		t.Errorf("substitute =\n%s\nwant\n%s", got, want)
	}
}

// The references line the generator writes sits above the placeholder,
// and everything around the paragraph survives untouched.
func TestSubstituteKeepsEverythingAroundThePlaceholder(t *testing.T) {
	file := `---
id: report-0009
---

# A title

**References:** [technical/testing/suites.md](../../docs/technical/testing/suites.md)

TODO: what was observed, and the evidence at hand. What should be done
about it is triage's output, never this file's content.
`
	got := substitute(file, "The observation.")
	for _, want := range []string{"id: report-0009", "# A title", "**References:**", "The observation."} {
		if !strings.Contains(got, want) {
			t.Errorf("substitute dropped %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "TODO") {
		t.Errorf("the placeholder survived:\n%s", got)
	}
}

// An adopter's template may write a body of its own below the
// placeholder; only the placeholder's own paragraph is replaced.
func TestSubstituteReplacesOneParagraphOnly(t *testing.T) {
	file := "---\nid: report-0009\n---\n\n# A title\n\nTODO: the placeholder.\n\nA paragraph the template added.\n"
	got := substitute(file, "The observation.")
	if !strings.Contains(got, "A paragraph the template added.") {
		t.Errorf("substitute reached past the placeholder's paragraph:\n%s", got)
	}
	if strings.Contains(got, "TODO") {
		t.Errorf("the placeholder survived:\n%s", got)
	}
}

// A template with no placeholder is still a report, and the
// reporter's words are never the part that goes missing.
func TestSubstituteAppendsWhereThereIsNoPlaceholder(t *testing.T) {
	file := "---\nid: report-0009\n---\n\n# A title\n"
	got := substitute(file, "The observation.")
	want := "---\nid: report-0009\n---\n\n# A title\n\nThe observation.\n"
	if got != want {
		t.Errorf("substitute =\n%s\nwant\n%s", got, want)
	}
}

// Front matter is skipped, never searched: a field whose value reads
// like the placeholder is a field, not a body.
func TestSubstituteNeverEditsTheFrontMatter(t *testing.T) {
	file := "---\nid: report-0009\ndoc_ref: TODO\n---\n\n# A title\n\nTODO: the placeholder.\n"
	got := substitute(file, "The observation.")
	if !strings.Contains(got, "doc_ref: TODO") {
		t.Errorf("the front matter was edited:\n%s", got)
	}
	if strings.Contains(got, "TODO: the placeholder.") {
		t.Errorf("the placeholder survived:\n%s", got)
	}
}

// The observation is taken as given: its own line breaks stand, and
// the file ends in exactly one newline whatever the reporter typed.
func TestSubstituteTakesTheObservationAsGiven(t *testing.T) {
	got := substitute(generated, "First line.\nSecond line.\n\n\n")
	if !strings.Contains(got, "First line.\nSecond line.\n") {
		t.Errorf("the observation was reflowed:\n%s", got)
	}
	if !strings.HasSuffix(got, "Second line.\n") {
		t.Errorf("the file does not end in one newline:\n%q", got)
	}
}

func TestFillLeavesTheFileAloneForABlankObservation(t *testing.T) {
	h := newHarness(t)
	h.files.Seed("/repo/"+file, []byte(generated), 0o644)

	if err := fill(h.files, "/repo/"+file, file, "   \n\t "); err != nil {
		t.Fatalf("fill = %v", err)
	}
	if got := h.read(t, file); got != generated {
		t.Errorf("a blank observation edited the file:\n%s", got)
	}
}
