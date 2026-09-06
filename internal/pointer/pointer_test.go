package pointer

import (
	"errors"
	"strings"
	"testing"
)

// kitSection is the shape v0.0.04 ships: a heading, and under it the
// link that is the whole of how the section is recognised.
const kitSection = "## WritRun\n" +
	"\n" +
	"This project tracks its work with WritRun. Before touching `work/`,\n" +
	"`docs/`, or any task, spec, or report, read and follow\n" +
	"[`.writrun/AGENTS.md`](.writrun/AGENTS.md)."

// templateDoc is the skeleton AGENTS.md the kit ships: the project's
// own TODO paragraph above, WritRun's section below.
func templateDoc() []byte {
	return []byte("# AGENTS.md — entry point\n\n<!-- TODO: one paragraph. -->\n\n" + kitSection + "\n")
}

// legacyDoc is an AGENTS.md a v0.0.03 adoption left: the fenced
// section, markers and all, with no pointer anywhere.
func legacyDoc() []byte {
	return []byte("# AGENTS.md\n\nThe project's own paragraph.\n\n" +
		"## WritRun — working the queue\n\n" +
		"<!-- writrun:begin\n     This section is WritRun's flow. -->\n\n" +
		"### Picking work\n\nThe flow's text.\n\n" +
		"<!-- writrun:end -->\n")
}

func TestSectionIsTheHeadingThatIntroducesTheLink(t *testing.T) {
	got, err := Section(templateDoc())
	if err != nil {
		t.Fatalf("Section: %v", err)
	}
	if string(got) != kitSection {
		t.Errorf("Section =\n%q\nwant\n%q", got, kitSection)
	}
}

func TestSectionStopsAtTheNextHeadingOfTheSameLevel(t *testing.T) {
	doc := []byte("# Top\n\n" + kitSection + "\n\n## Something else\n\nNot WritRun's.\n")
	got, err := Section(doc)
	if err != nil {
		t.Fatalf("Section: %v", err)
	}
	if string(got) != kitSection {
		t.Errorf("Section carried past its heading:\n%q", got)
	}
}

func TestSectionStopsAtAHigherHeading(t *testing.T) {
	sub := strings.Replace(kitSection, "## WritRun", "### WritRun", 1)
	doc := []byte("# Top\n\n" + sub + "\n\n## Another chapter\n\nNot WritRun's.\n")
	got, err := Section(doc)
	if err != nil {
		t.Fatalf("Section: %v", err)
	}
	if string(got) != sub {
		t.Errorf("Section =\n%q\nwant\n%q", got, sub)
	}
}

func TestSectionRefusesADocumentWithNoLink(t *testing.T) {
	_, err := Section([]byte("# AGENTS.md\n\nNothing of WritRun's here.\n"))
	if !errors.Is(err, ErrNoSection) {
		t.Fatalf("err = %v, want ErrNoSection", err)
	}
}

func TestSectionRefusesTheLegacyFence(t *testing.T) {
	// The fenced section links nothing, so there is no pointer to cut —
	// which is the refusal `update` met when the pin first moved.
	_, err := Section(legacyDoc())
	if !errors.Is(err, ErrNoSection) {
		t.Fatalf("err = %v, want ErrNoSection", err)
	}
}

func TestHasReadsBothAnswers(t *testing.T) {
	if !Has(templateDoc()) {
		t.Error("Has(template) = false, want true")
	}
	if Has(legacyDoc()) {
		t.Error("Has(legacy) = true — the fence carries no pointer")
	}
}

func TestLegacyReadsTheMarkersAndNothingElse(t *testing.T) {
	if !Legacy(legacyDoc()) {
		t.Error("Legacy(legacy) = false, want true")
	}
	if Legacy(templateDoc()) {
		t.Error("Legacy(template) = true — v0.0.04 ships no markers")
	}
	if Legacy([]byte("<!-- writrun:end -->\n<!-- writrun:begin -->\n")) {
		t.Error("Legacy = true on markers out of order")
	}
}

func TestGraftAppendsAndChangesNoByteAbove(t *testing.T) {
	existing := []byte("# AGENTS.md\n\nThe project's own rules.\n")
	section, err := Section(templateDoc())
	if err != nil {
		t.Fatalf("Section: %v", err)
	}
	got := string(Graft(existing, section))
	if !strings.HasPrefix(got, string(existing)) {
		t.Errorf("Graft rewrote what was there:\n%q", got)
	}
	if !strings.Contains(got, Target) {
		t.Errorf("Graft dropped the pointer:\n%q", got)
	}
	if !strings.Contains(got, "rules.\n\n## WritRun") {
		t.Errorf("Graft did not separate with one blank line:\n%q", got)
	}
}

func TestGraftThenRemoveIsARoundTrip(t *testing.T) {
	existing := []byte("# AGENTS.md\n\nThe project's own rules.\n")
	section, err := Section(templateDoc())
	if err != nil {
		t.Fatalf("Section: %v", err)
	}
	out, only, err := Remove(Graft(existing, section))
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if only {
		t.Fatal("onlySection = true — the project's own paragraph was there")
	}
	if string(out) != string(existing) {
		t.Errorf("round trip =\n%q\nwant\n%q", out, existing)
	}
}

func TestRemoveReportsADocumentThatIsNothingButTheSection(t *testing.T) {
	out, only, err := Remove([]byte(kitSection + "\n"))
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if !only {
		t.Fatalf("onlySection = false, out = %q", out)
	}
}

func TestRemoveCutsTheFenceWhereBothShapesAreThere(t *testing.T) {
	// A repository adopted at v0.0.03 and re-inited at v0.0.04 carries
	// both. The fence holds the kit's own prose, so it is the one cut.
	doc := append(legacyDoc(), []byte("\n"+kitSection+"\n")...)
	out, only, err := Remove(doc)
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if only {
		t.Fatal("onlySection = true — the project's paragraph was there")
	}
	if strings.Contains(string(out), "writrun:begin") {
		t.Errorf("the fence survived:\n%q", out)
	}
	if !strings.Contains(string(out), Target) {
		t.Errorf("the pointer was cut instead of the fence:\n%q", out)
	}
}

func TestRemoveRefusesADocumentCarryingNeitherShape(t *testing.T) {
	_, _, err := Remove([]byte("# AGENTS.md\n\nThe project's alone.\n"))
	if !errors.Is(err, ErrNoSection) {
		t.Fatalf("err = %v, want ErrNoSection", err)
	}
}
