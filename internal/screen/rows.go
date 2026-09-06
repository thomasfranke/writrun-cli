// Package screen is the queue as a key-navigated screen — what
// `writrun` with no command opens (docs/product/screen.md, spec-0020).
//
// The screen reads and dispatches; it decides nothing. The rows are the
// selection skill's own lister's output, parsed only far enough to know
// which line carries which task, and every key leaves the screen and
// runs an existing command with that command's own checks, questions
// and confirmation.
package screen

import (
	"regexp"
	"strings"
)

// taskLine matches a lister row that names a task. The id is what the
// screen needs and the whole line is what it shows: re-formatting the
// lister's text would make the screen a second opinion about a queue
// that has one (list.md).
var taskLine = regexp.MustCompile(`^\s+(task-[0-9]{4})\b`)

// Row is one line of the lister's output. Task rows carry an id and can
// be selected; everything else — headings, blank lines, the order note,
// the reports whose triage is nobody's to dispatch — is shown as it
// arrived and skipped by the selection.
//
// **Every task row is selectable, including one that cannot be taken.**
// The screen judges no task: a held-back or in-flight row dispatches
// like any other and the command's own refusal is the answer. spec-0020
// said this twice and disagreed with itself once — its step 3 called a
// held-back entry unselectable while its edge cases gave `take`'s
// refusal as the answer for a task that is not ready. The second is the
// reading kept, because "the screen offers no action a command does not
// already provide" cuts both ways: it must not withhold one either.
type Row struct {
	Text string
	ID   string
}

// Selectable reports whether the arrow keys stop on this row.
func (r Row) Selectable() bool { return r.ID != "" }

// Parse turns the lister's output into rows, in the order it wrote
// them. Trailing blank lines are dropped; nothing else is.
func Parse(out string) []Row {
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	rows := make([]Row, 0, len(lines))
	for _, line := range lines {
		row := Row{Text: line}
		if m := taskLine.FindStringSubmatch(line); m != nil {
			row.ID = m[1]
		}
		rows = append(rows, row)
	}
	return rows
}

// firstSelectable is the row the screen opens on, or -1 when the queue
// offers none — an empty queue is a screen that opens and says so.
func firstSelectable(rows []Row) int {
	for i, r := range rows {
		if r.Selectable() {
			return i
		}
	}
	return -1
}
