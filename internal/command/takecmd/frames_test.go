package takecmd

import (
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/drawing"
	"github.com/thomasfranke/writrun-cli/internal/term"
)

// The task question, asserted against the frame that states it. The
// options are the lister's own lines, which is what `selectTask` hands
// over (docs/product/screens/README.md).
func TestTheTaskQuestionFrame(t *testing.T) {
	drawn, err := drawing.Screen(
		"../../../docs/product/screens/tasks/take.excalidraw",
		"writrun take — the available group, a task highlighted")
	if err != nil {
		t.Fatalf("the drawing could not be read: %v", err)
	}
	ids := []string{"task-0001", "task-0002"}
	labels := []string{
		"task-0001   Debounce the mirror updates",
		"task-0002   Another thing entirely",
	}
	got := term.QuestionLines("Which task?", options(ids, labels), 71, 1)
	// The rows and the keys, not the sentence: the frame's explanation
	// names the task's spec and its priority, and the lister's line it
	// is drawn from carries neither. What this command can say about a
	// row is what that row carries.
	if d := drawing.CompareRows(got, drawn); d != "" {
		t.Error(d)
	}
}
