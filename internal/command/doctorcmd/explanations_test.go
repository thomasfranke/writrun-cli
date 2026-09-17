package doctorcmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const doctorDoc = "docs/product/adoption/doctor.md"

// The document is the authority, and this is what keeps the copy beside
// the code from becoming a second one: every line of the table held in
// Go is in the document, in this order. A sentence edited on one side
// and not the other is a red build
// (docs/technical/engineering/coupling.md, rule 5).
func TestTheTableInGoIsTheDocumentsOwn(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", filepath.FromSlash(doctorDoc)))
	if err != nil {
		t.Fatalf("%s could not be read: %v", doctorDoc, err)
	}
	if !strings.Contains(string(raw), requirementsTable) {
		t.Errorf("%s does not carry the table %s holds — one of the two was edited alone",
			doctorDoc, "explanations.go")
	}
}

// Every requirement the binary can report is explained by the document.
// This is the test that keeps the document and the checks in step: a
// check renamed, added or removed shows up here (spec-0037).
func TestEveryRequirementTheBinaryReportsHasARow(t *testing.T) {
	for _, name := range reportable(t) {
		if _, there := explanations[name]; !there {
			t.Errorf("%q has no row in %s — the screen would fall back to the check's own sentence",
				name, doctorDoc)
		}
	}
}

// The document explains nothing the binary never reports: a row for a
// requirement no check makes is a sentence nobody can reach.
func TestTheDocumentExplainsNoRequirementTheBinaryNeverReports(t *testing.T) {
	made := map[string]bool{}
	for _, name := range reportable(t) {
		made[name] = true
	}
	for name := range explanations {
		if !made[name] {
			t.Errorf("%s explains %q, which no check makes", doctorDoc, name)
		}
	}
}

// reportable is every requirement name a run can produce: the whole of
// stages 0 to 3, and the two states the settings file can be in that
// the examination itself reports.
func reportable(t *testing.T) []string {
	t.Helper()
	f := newFixture(t, "3")
	var names []string
	for _, r := range examine(f.root, 3, f.deps()) {
		names = append(names, r.name)
	}

	broken := newFixture(t, "3")
	broken.scripts.verdict[settingsReader] = exitErr(3)
	_, unreadable := declaredStage(broken.root, broken.deps())
	if len(unreadable) == 0 {
		t.Fatal("an unreadable stage reported no requirement")
	}
	for _, r := range unreadable {
		names = append(names, r.name)
	}
	return names
}

// A requirement the table does not name is explained by the check's own
// sentence and nothing else — a check a newer kit adds is reported
// without this binary knowing it (spec-0037, step 5).
func TestARequirementWithNoRowFallsBackToTheChecksOwnSentence(t *testing.T) {
	r := requirement{stage: 1, name: "a check no tag has shipped yet",
		note: "it is not as the kit expects", mark: breaks}
	if got, want := explain(r), "it is not as the kit expects."; got != want {
		t.Errorf("explain = %q, want %q", got, want)
	}
}

// A requirement the table does name is explained in the document's own
// words, then by what this run found.
func TestANamedRequirementIsExplainedInTheDocumentsWords(t *testing.T) {
	r := requirement{stage: 1, name: "AGENTS.md", note: "a writrun:begin/writrun:end section is stale", mark: advises}
	got := explain(r)
	for _, want := range []string{
		explanations["AGENTS.md"].what,
		"Advised: a writrun:begin/writrun:end section is stale.",
		"Clear it by " + explanations["AGENTS.md"].clears,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("explain = %q; misses %q", got, want)
		}
	}
}

// A met requirement is explained too: the reader asks what a row means,
// not only how to clear it (spec-0037).
func TestAMetRequirementIsExplained(t *testing.T) {
	got := explain(requirement{stage: 1, name: "docs/about.md", mark: met})
	if !strings.Contains(got, explanations["docs/about.md"].what) {
		t.Errorf("explain = %q; a met requirement is explained too", got)
	}
	if !strings.Contains(got, "Met.") {
		t.Errorf("explain = %q; it does not say the requirement holds", got)
	}
	if strings.Contains(got, "Clear it by") {
		t.Errorf("explain = %q; a requirement that holds needs no clearing", got)
	}
}
