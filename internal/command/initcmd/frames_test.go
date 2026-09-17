package initcmd

import (
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/drawing"
	"github.com/thomasfranke/writrun-cli/internal/term"
)

// The stage question, asserted against the frame that states it. The
// drawing is not the thing under test: it is what the test is written
// against (docs/product/screens/README.md).
func TestTheStageQuestionFrame(t *testing.T) {
	drawn, err := drawing.Screen(
		"../../../docs/product/screens/adoption/init.excalidraw",
		"writrun init — the stage question, a stage highlighted")
	if err != nil {
		t.Fatalf("the drawing could not be read: %v", err)
	}
	// The frame highlights the second rung.
	got := term.QuestionLines("Which stage?", stageOptions(), drawnWidth, 1)
	if d := drawing.Compare(got, drawn); d != "" {
		t.Error(d)
	}
}

// drawnWidth is the terminal the frames are drawn at.
const drawnWidth = 71
