package amendcmd

import (
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/drawing"
	"github.com/thomasfranke/writrun-cli/internal/term"
)

// The amendment a confirmation is about, asserted against the frame
// that states it (docs/product/screens/README.md).
func TestTheAmendmentFrame(t *testing.T) {
	drawn, err := drawing.Screen(
		"../../../docs/product/screens/authoring/amend.excalidraw",
		"writrun amend — the amendment, one line selected")
	if err != nil {
		t.Fatalf("the drawing could not be read: %v", err)
	}
	p := plan{
		specID:  "spec-0011",
		relPath: "work/specs/spec-0011-amend-command.md",
		branch:  "docs/amend-command",
		subject: "docs(specs): return spec-0011 to draft",
		title:   "[Docs][Specs] Reopen the amendment gate",
		body:    "the body\n",
	}
	susp := []suspension{{task: "task-0012", number: 42}}
	// The frame selects the second row, which is the suspension.
	got := term.PlanLines(amendPlan(p, []string{"task-0012"}, susp, true), 71, 1)
	if d := drawing.Compare(got, drawn); d != "" {
		t.Error(d)
	}
}
