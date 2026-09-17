package doctorcmd

import (
	"strings"
	"testing"
)

// The two report frames, asserted line for line against what the binary
// renders. The drawing is not the thing under test: it is what the test
// is written against, and where the two disagree the binary is what
// changes (docs/product/screens/README.md).
//
// Each fixture is the repository the frame was drawn from: stage 1
// declared, the kit tag the frame shows, and the forge answering what
// the frame's six rows say it answered.

// drawnTag is the kit tag the frames show in their `.writrun/VERSION`
// row.
const drawnTag = "v0.0.08"

func TestTheFrameForAStageWithinReach(t *testing.T) {
	f := newFixture(t, "1")
	write(t, f.root, ".writrun/VERSION", drawnTag+"\n")

	out, err := runDoctor(t, f)
	if err != nil {
		t.Fatalf("run = %v (exit %d), want 0 — the frame is captioned exit 0", err, exitCode(err))
	}
	want := drop(frame(t, doctorDrawing, "writrun doctor — stage 2 within reach, exit 0"), "$ writrun doctor")
	sameLines(t, rendered(out), want)
}

func TestTheFrameForAStageOutOfReach(t *testing.T) {
	f := newFixture(t, "1")
	write(t, f.root, ".writrun/VERSION", drawnTag+"\n")
	// The one stage-1 row the frame marks `!`.
	write(t, f.root, "AGENTS.md", legacyAgents)
	// The three stage-2 rows the frame does not mark `✓`: a workflow
	// that pushes to main and raises nothing, under a repository default
	// of read, and the rules on main unreadable.
	readDefault(f)
	write(t, f.root, workflowsDir+"/record.yml", silentWorkflow)
	f.forge.fails["api repos/{owner}/{repo}/rules/branches/main --jq .[].type"] = errBrokenRead{}

	out, err := runDoctor(t, f)
	if err != nil {
		t.Fatalf("run = %v (exit %d), want 0 — a previewed row never reaches the exit status", err, exitCode(err))
	}
	want := drop(frame(t, doctorDrawing, "writrun doctor — stage 2 out of reach, exit 0"), "$ writrun doctor")
	sameLines(t, unwrapDetails(rendered(out)), unwrapDetails(want))
}

// errBrokenRead is a forge read that refuses. What it says is not the
// frame's subject — the row is `?` whatever the reason — so the case
// keeps it to one line.
type errBrokenRead struct{}

func (errBrokenRead) Error() string { return "gh api: HTTP 403" }

// unwrapDetails folds a detail line back into the one above it.
//
// The drawing wraps a long detail to fit its canvas, and the binary
// writes it whole: a terminal is as wide as it is, and this command
// prints plain text that a pager or the terminal wraps
// (docs/product/rules.md). So the two are compared unwrapped, which is
// the only difference between them that is the canvas's and not the
// binary's.
func unwrapDetails(lines []string) []string {
	var out []string
	for _, l := range lines {
		if strings.HasPrefix(l, detailIndent) && len(out) > 0 && strings.HasPrefix(out[len(out)-1], detailIndent) {
			out[len(out)-1] += " " + strings.TrimSpace(l)
			continue
		}
		out = append(out, l)
	}
	return out
}
