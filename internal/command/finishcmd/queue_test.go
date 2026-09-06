package finishcmd

import (
	"testing"
)

// How the front matter, the id and the file are read is
// internal/queue's, and its tests hold where each disputed reading came
// from — including the refusal of an id that declares the other kind,
// which `finish spec-0012` used to carry to the write stage
// (report-0020). What is held here is the one reading that is this
// command's own.

func TestOutcomeFilled(t *testing.T) {
	cases := map[string]bool{
		"What was built, and what diverged.": true,
		"":                                   false,
		"_(fill after execution)_":           false,
		"TODO":                               false,
		"\n\n":                               false,
	}
	for outcome, want := range cases {
		if got := outcomeFilled([]byte(specFixture("approved", outcome))); got != want {
			t.Errorf("outcomeFilled(%q) = %v, want %v", outcome, got, want)
		}
	}
	// A section after Outcome does not leak into it.
	trailing := "---\nid: spec-0010\n---\n\n## Outcome\n\n## Notes\n\nSomething else.\n"
	if outcomeFilled([]byte(trailing)) {
		t.Error("a later section was read as the Outcome")
	}
	// A spec with no Outcome heading at all has none filled.
	if outcomeFilled([]byte("---\nid: spec-0010\n---\n\n# spec\n\nprose\n")) {
		t.Error("a spec carrying no Outcome heading reported one")
	}
}
