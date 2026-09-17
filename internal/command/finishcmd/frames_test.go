package finishcmd

import (
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/drawing"
	"github.com/thomasfranke/writrun-cli/internal/term"
)

// The summary a confirmation is about, asserted against the frame that
// states it (docs/product/screens/README.md).
func TestTheReadyForReviewFrame(t *testing.T) {
	drawn, err := drawing.Screen(
		"../../../docs/product/screens/tasks/finish.excalidraw",
		"writrun finish — the summary, one line selected")
	if err != nil {
		t.Fatalf("the drawing could not be read: %v", err)
	}
	// The frame opens with preflight's own last line, printed before
	// this plan is composed, and the blank under it.
	drawn = drawn[2:]
	p := readyPlan("task-0001", "spec-0001",
		pullRequest{Number: 7, Title: "[TASK-0001] A thing to do"})
	// The frame selects the third row, which is the pull request.
	got := term.PlanLines(p, 71, 2)
	if d := drawing.Compare(got, drawn); d != "" {
		t.Error(d)
	}
}
