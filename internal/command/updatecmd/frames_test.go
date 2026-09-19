package updatecmd

import (
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/drawing"
	"github.com/thomasfranke/writrun-cli/internal/kittag"
	"github.com/thomasfranke/writrun-cli/internal/term"
)

// The refresh a confirmation is about, asserted against the frame that
// states it (docs/product/screens/README.md).
func TestTheRefreshPlanFrame(t *testing.T) {
	drawn, err := drawing.Screen(
		"../../../docs/product/screens/adoption/update.excalidraw",
		"writrun update — the plan, one path selected")
	if err != nil {
		t.Fatalf("the drawing could not be read: %v", err)
	}
	r := &refresh{
		from: "v0.0.00",
		to:   "v0.0.09",
		changes: []change{
			{rel: ".github/ISSUE_TEMPLATE/writrun-report.yml", verb: added},
			{rel: ".github/workflows/writrun-intake.yml", verb: added},
			{rel: ".github/workflows/writrun-check.yml", verb: changed},
			{rel: kittag.Rel, verb: changed},
			{rel: ".writrun/skills/writrun-check-front-matter/check_front_matter.sh", verb: changed},
			{rel: ".writrun/templates/pull_request_template.md", verb: added},
		},
	}
	// The frame selects the second row, which is the workflows.
	got := term.PlanLines(r.plan("v0.0.09"), 71, 1)
	if d := drawing.Compare(got, drawn); d != "" {
		t.Error(d)
	}
}
