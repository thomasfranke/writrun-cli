package uninstallcmd

import (
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/drawing"
	"github.com/thomasfranke/writrun-cli/internal/hook"
	"github.com/thomasfranke/writrun-cli/internal/term"
)

// The removal a confirmation is about, asserted against the frame that
// states it (docs/product/screens/README.md).
func TestTheRemovalPlanFrame(t *testing.T) {
	drawn, err := drawing.Screen(
		"../../../docs/product/screens/adoption/uninstall.excalidraw",
		"writrun uninstall — the plan, one line selected")
	if err != nil {
		t.Fatalf("the drawing could not be read: %v", err)
	}
	r := &removal{
		root:      "/repo",
		dirs:      []string{".writrun"},
		files:     []string{"WRITRUN.md", ".github/workflows/writrun-check.yml"},
		hookAt:    "/repo/.git/hooks/commit-msg",
		hookState: hook.Ours,
		agents:    []byte("what is left of AGENTS.md\n"),
	}
	// The frame selects the third row, which is the kit's workflow.
	got := term.PlanLines(r.plan(), 71, 2)
	if d := drawing.Compare(got, drawn); d != "" {
		t.Error(d)
	}
}
