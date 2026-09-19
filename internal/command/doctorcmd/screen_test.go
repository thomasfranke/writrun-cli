package doctorcmd

import (
	"strings"
	"testing"
)

// What the screen is given is what the frame draws: the header line, the
// four group headings with their counts, and every row's glyph and text.
//
// How those lines are painted is internal/screen's own frame case; this
// is the half that decides what they say (spec-0037).
func TestTheScreenIsGivenTheRowsTheFrameDraws(t *testing.T) {
	f := newFixture(t, "1")
	write(t, f.root, ".writrun/VERSION", drawnTag+"\n")
	write(t, f.root, "AGENTS.md", legacyAgents)
	readDefault(f)
	write(t, f.root, workflowsDir+"/record.yml", silentWorkflow)
	f.forge.fails["api repos/{owner}/{repo}/rules/branches/main --jq .[].type"] = errBrokenRead{}

	d := f.deps()
	d.Header = func(string) string { return "writrun-cli v0.0.2 · pins WritRun v0.0.09 · branch main" }
	s := requirementsScreen(f.root, d, 0)

	var got []string
	got = append(got, " "+s.Header, " "+s.Line)
	for _, g := range s.Groups {
		got = append(got, "", " "+g.Name)
		for _, r := range g.Rows {
			got = append(got, "   "+r.Mark+"  "+r.Text)
		}
	}

	drawn := frame(t, doctorDrawing, "the doctor screen — a requirement that is not")
	want := drawn[:ruleAt(t, drawn)]
	for len(want) > 0 && want[len(want)-1] == "" {
		want = want[:len(want)-1]
	}
	// The cursor is the screen's, not the data's: the drawn frame marks
	// the row it has selected and this compares what every row says.
	for i, l := range want {
		want[i] = strings.Replace(l, " › ", "   ", 1)
	}
	sameLines(t, got, want)
}

// Every row the screen is given carries the explanation the document
// states, so the footer never has to compose one.
func TestEveryScreenRowArrivesExplained(t *testing.T) {
	f := newFixture(t, "3")
	s := requirementsScreen(f.root, f.deps(), 0)
	for _, g := range s.Groups {
		for _, r := range g.Rows {
			if strings.TrimSpace(r.Explain) == "" {
				t.Errorf("%q reaches the screen with nothing to say about it", r.Name)
			}
		}
	}
}

func ruleAt(t *testing.T, lines []string) int {
	t.Helper()
	for i, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), "───") {
			return i
		}
	}
	t.Fatalf("the frame draws no rule under its rows")
	return 0
}
