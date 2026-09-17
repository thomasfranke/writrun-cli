// Package screen is the queue as a key-navigated screen — what
// `writrun` with no command opens (docs/product/screens/README.md, spec-0020).
//
// The screen reads and dispatches; it decides nothing. The rows are the
// selection skill's own lister's output, parsed only far enough to know
// which line carries which task, and every key leaves the screen and
// runs an existing command with that command's own checks, questions
// and confirmation.
package screen

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/queue"
)

// taskLine matches a lister row that names a task, whether it writes
// the id whole or as its four digits alone. The number is what the
// screen needs and the whole line is what it shows: re-formatting the
// lister's text would make the screen a second opinion about a queue
// that has one (list.md).
var taskLine = regexp.MustCompile(`^\s+(?:task-)?([0-9]{4})\b`)

// section is the lister's own heading a row sits under. The screen
// reads it to say what a selected row is, and for nothing else — which
// tasks are in which group is the lister's answer (list.md).
type section int

const (
	noSection section = iota
	inProgress
	available
	inFlight
	heldBack
	reports
)

// sections are the lister's headings, matched on their opening phrase
// rather than the whole line: the wording is the skill's, and a release
// that rewords one must not silently lose the group it heads.
var sections = []struct {
	prefix string
	is     section
}{
	{"In progress", inProgress},
	{"Available", available},
	{"Nothing is available", available},
	{"In flight", inFlight},
	{"Held back", heldBack},
	{"Open reports", reports},
}

// words are what a section is called in the sentence that explains a
// row, and in the count the header carries.
var words = map[section]string{
	inProgress: "in progress",
	available:  "available",
	inFlight:   "in flight",
	heldBack:   "held back",
}

// priorities are the queue's own, and the second field of an available
// row is one of them or is the row's title. The vocabulary is the
// queue's, which is the only vocabulary this binary carries
// (docs/technical/engineering/boundaries.md).
var priorities = map[string]bool{"low": true, "medium": true, "high": true}

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
	// State is what the row is, in the lister's own words: the group it
	// sits under, and the priority where the row carries one. It is
	// what the explanation names before it names the act (spec-0040).
	State string
	// at is the section the row sits under, kept so the header can
	// count the groups without reading the text twice.
	at section
}

// Selectable reports whether the arrow keys stop on this row.
func (r Row) Selectable() bool { return r.ID != "" }

// Parse turns the lister's output into rows, in the order it wrote
// them. Trailing blank lines are dropped; nothing else is.
func Parse(out string) []Row {
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	rows := make([]Row, 0, len(lines))
	at := noSection
	for _, line := range lines {
		if s, ok := heading(line); ok {
			at = s
		}
		row := Row{Text: line, at: at}
		if m := taskLine.FindStringSubmatch(line); m != nil && at != reports {
			row.ID = "task-" + m[1]
			row.State = state(at, line)
		}
		rows = append(rows, row)
	}
	return rows
}

// heading reports whether a line opens one of the lister's sections.
func heading(line string) (section, bool) {
	if strings.HasPrefix(line, " ") {
		return noSection, false
	}
	for _, s := range sections {
		if strings.HasPrefix(line, s.prefix) {
			return s.is, true
		}
	}
	return noSection, false
}

// state is what a row is: its group, and the priority where its second
// field is one.
func state(at section, line string) string {
	word, there := words[at]
	if !there {
		word = "in the queue"
	}
	fields := strings.Fields(line)
	if len(fields) > 1 && priorities[fields[1]] {
		return word + ", priority " + fields[1]
	}
	return word
}

// queueContext is the screen's second line: the folder it reads, and
// what the lister's own sections count to.
func queueContext(rows []Row) string {
	counts := map[section]int{}
	for _, r := range rows {
		if r.Selectable() {
			counts[r.at]++
		}
	}
	var parts []string
	for _, s := range []section{available, inProgress, inFlight, heldBack} {
		if n := counts[s]; n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", n, words[s]))
		}
	}
	line := "QUEUE · " + queue.TasksDir + "/ — "
	if len(parts) == 0 {
		return line + "nothing available"
	}
	return line + strings.Join(parts, ", ")
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
