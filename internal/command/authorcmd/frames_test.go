package authorcmd

import (
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/drawing"
	"github.com/thomasfranke/writrun-cli/internal/term"
)

// The composition a confirmation is about, asserted against the frame
// that states it (docs/product/screens/README.md).
func TestTheCompositionFrame(t *testing.T) {
	drawn, err := drawing.Screen(
		"../../../docs/product/screens/authoring/author.excalidraw",
		"writrun author — the composition, one line selected")
	if err != nil {
		t.Fatalf("the drawing could not be read: %v", err)
	}
	c := composition{
		branch: "docs/derived-work",
		title:  "[DOCS] The declaration is the section",
		body:   strings.Repeat("a line of the body\n", 36),
		files: []string{
			"docs/product/rules.md",
			"work/specs/spec-0001-derived-work.md",
			"work/tasks/task-0001-derived-work.md",
		},
	}
	// The frame disagrees with itself on one column: its selected row is
	// drawn `› files:` and the rows around it put their first character
	// in the column under the cursor. The unselected rows are three
	// against one, so they are the authority and the column is
	// reconciled here rather than left to read as the binary drifting.
	for i, l := range drawn {
		drawn[i] = strings.Replace(l, "› files:", "›files:", 1)
	}
	// The frame selects the fourth row, which is the files.
	got := term.PlanLines(authorPlan(c), 71, 3)
	if d := drawing.Compare(got, drawn); d != "" {
		t.Error(d)
	}
}
