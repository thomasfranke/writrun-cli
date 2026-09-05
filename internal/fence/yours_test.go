package fence

import (
	"strings"
	"testing"
)

func TestGraftAddsTheMissingNewline(t *testing.T) {
	// A file whose last line has no terminator still gets exactly one
	// blank line before the section.
	out := Graft([]byte("# A project\n\nNo trailing newline."), []byte("SECTION"))
	if !strings.HasSuffix(string(out), "No trailing newline.\n\nSECTION\n") {
		t.Errorf("Graft did not close the last line:\n%q", out)
	}

	// An empty document grows no leading newline of its own beyond the
	// separator.
	if got := string(Graft(nil, []byte("SECTION"))); got != "\nSECTION\n" {
		t.Errorf("Graft on an empty document = %q", got)
	}
}

// sectionWith builds a fenced section carrying the given body lines,
// so a case can shape exactly the `yours` arrangement it is about.
func sectionWith(body string) []byte {
	return []byte("## WritRun — working the queue\n\n<!-- writrun:begin\n     flow -->\n\n" +
		body + "\n\n<!-- writrun:end -->")
}

func TestAMarkerGoverningNothingStillPairs(t *testing.T) {
	// The marker sits between two headings: nothing follows it before
	// the next heading, and nothing precedes it after the last one.
	body := "### One\n\n<!-- yours: nothing to govern here. -->\n\n### Two"
	current := sectionWith(body)
	incoming := sectionWith(strings.Replace(body, "### Two", "### Two, reworded", 1))

	out, err := carryYours(current, incoming)
	if err != nil {
		t.Fatalf("carryYours: %v", err)
	}
	if !strings.Contains(string(out), "### Two, reworded") {
		t.Error("the refreshed heading did not land")
	}
	if !strings.Contains(string(out), "<!-- yours: nothing to govern here. -->") {
		t.Error("the marker itself was dropped")
	}
}

func TestASectionWithNoMarkersIsTakenWhole(t *testing.T) {
	current := sectionWith("### One\n\nThe project's prose.")
	incoming := sectionWith("### One\n\nThe kit's prose.")
	out, err := carryYours(current, incoming)
	if err != nil {
		t.Fatalf("carryYours: %v", err)
	}
	if string(out) != string(incoming) {
		t.Errorf("a section with no `yours` marker was not taken whole:\n%s", out)
	}
}

func TestMoreIncomingMarkersThanTheDocumentHolds(t *testing.T) {
	// A tag that adds a marker carries its own block for the new one;
	// the project has no answer to put there yet.
	current := sectionWith("### One\n\n<!-- yours: first. -->\n\nThe project's answer.")
	incoming := sectionWith("### One\n\n<!-- yours: first. -->\n\nThe kit's default.\n\n" +
		"### Two\n\n<!-- yours: second. -->\n\nThe kit's other default.")

	out, err := carryYours(current, incoming)
	if err != nil {
		t.Fatalf("carryYours: %v", err)
	}
	got := string(out)
	if !strings.Contains(got, "The project's answer.") {
		t.Error("the answer the project had did not survive")
	}
	if !strings.Contains(got, "The kit's other default.") {
		t.Error("the block of the marker the project never had did not land")
	}
}

func TestRunBeforeStopsAtAHeading(t *testing.T) {
	// Nothing but a heading precedes the marker and nothing follows it
	// before the fence closes, so it governs no lines at all.
	lines := []string{"### Only a heading", "", "<!-- yours: alone. -->", "", "<!-- writrun:end -->"}
	blocks := yoursBlocks(lines)
	if len(blocks) != 1 {
		t.Fatalf("yoursBlocks found %d markers, want 1", len(blocks))
	}
	if blocks[0].from != blocks[0].to {
		t.Errorf("the marker governs lines %d..%d, want an empty range", blocks[0].from, blocks[0].to)
	}
}

func TestReplaceIsAnErrorWhenTheDocumentHasNoFence(t *testing.T) {
	if _, err := Replace([]byte("# Nothing fenced here.\n"), sectionWith("### One")); err == nil {
		t.Fatal("a document with no fence was accepted")
	}
}

func TestTwoMarkersCannotBothClaimOneBlock(t *testing.T) {
	// The first marker takes the line after it; the second finds the
	// same line before it. The block is carried once, and the pass
	// does not slice backwards over it.
	body := "### One\n\n<!-- yours: first. -->\n\nthe shared line\n<!-- yours: second. -->"
	current := sectionWith(body)
	incoming := sectionWith(strings.Replace(body, "the shared line", "the kit's line", 1))

	out, err := carryYours(current, incoming)
	if err != nil {
		t.Fatalf("carryYours: %v", err)
	}
	got := string(out)
	if !strings.Contains(got, "the shared line") {
		t.Error("the project's line did not survive")
	}
	if strings.Count(got, "the shared line") != 1 {
		t.Errorf("the line was carried %d times:\n%s", strings.Count(got, "the shared line"), got)
	}
}
