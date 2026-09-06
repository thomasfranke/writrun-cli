package screen

import (
	"io"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// listing is the shape the selection skill's lister writes: headings,
// indented rows, the order note, and a reports section whose triage is
// nobody's to dispatch.
const listing = `Available — any of these may be taken:
  task-0020  medium  Have doctor check the recording push
  task-0022  medium  Survive a signal mid-finish

Order is a suggestion for a person and binding for an agent.

In flight — an open pull request already exists:
  task-0021  #80 by @thomasfranke Navigate the queue

Open reports — waiting to be triaged, never selected:
  report-0016  The template's Derived-work comment
`

func rowsOf(t *testing.T) []Row {
	t.Helper()
	return Parse(listing)
}

func TestOnlyTaskRowsAreSelectable(t *testing.T) {
	rows := rowsOf(t)
	var ids []string
	for _, r := range rows {
		if r.Selectable() {
			ids = append(ids, r.ID)
		}
	}
	want := []string{"task-0020", "task-0022", "task-0021"}
	if strings.Join(ids, ",") != strings.Join(want, ",") {
		t.Errorf("selectable = %v, want %v", ids, want)
	}
	for _, r := range rows {
		if strings.Contains(r.Text, "report-0016") && r.Selectable() {
			t.Error("a report row was selectable; triage is not the screen's to dispatch")
		}
		if strings.HasPrefix(r.Text, "Available") && r.Selectable() {
			t.Error("a heading was selectable")
		}
	}
}

// The lister's text is shown as it arrived: a screen that re-formatted
// it would be a second opinion about a queue that has one.
func TestTheListersLinesAreKeptVerbatim(t *testing.T) {
	rows := rowsOf(t)
	var got []string
	for _, r := range rows {
		got = append(got, r.Text)
	}
	want := strings.Split(strings.TrimRight(listing, "\n"), "\n")
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Error("the lister's lines were not preserved")
	}
}

// An in-flight task is selectable. The screen judges no task: the
// command it dispatches to owns the refusal (spec-0020, Edge cases).
func TestATaskThatCannotBeTakenIsStillSelectable(t *testing.T) {
	m := newModel(rowsOf(t))
	m.move(1)
	m.move(1)
	if got := m.selected(); got != "task-0021" {
		t.Fatalf("selected = %q, want the in-flight task-0021", got)
	}
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if got := out.(model).action; got.Command != "take" || got.Arg != "task-0021" {
		t.Errorf("action = %+v, want take task-0021", got)
	}
}

func TestEachKeyResolvesToTheCommandTheRuleNames(t *testing.T) {
	for _, tc := range []struct {
		key  tea.KeyMsg
		want Action
	}{
		{tea.KeyMsg{Type: tea.KeyEnter}, Action{"take", "task-0020"}},
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}}, Action{"work", "task-0020"}},
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}, Action{"status", ""}},
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}, Action{}},
	} {
		m := newModel(rowsOf(t))
		out, _ := m.Update(tc.key)
		if got := out.(model).action; got != tc.want {
			t.Errorf("%s -> %+v, want %+v", tc.key, got, tc.want)
		}
	}
}

// q leaves and runs nothing; the zero Action is what says so.
func TestQuitDispatchesNothing(t *testing.T) {
	m := newModel(rowsOf(t))
	out, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if out.(model).action != (Action{}) {
		t.Error("q resolved to a command")
	}
	if cmd == nil {
		t.Error("q did not quit")
	}
}

func TestTheSelectionStopsAtTheEndsRatherThanWrapping(t *testing.T) {
	m := newModel(rowsOf(t))
	m.move(-1)
	if got := m.selected(); got != "task-0020" {
		t.Errorf("up from the first row moved to %q", got)
	}
	for i := 0; i < 10; i++ {
		m.move(1)
	}
	if got := m.selected(); got != "task-0021" {
		t.Errorf("down past the last row moved to %q", got)
	}
}

// A queue with nothing to select still opens, says so, and takes s and q
// (spec-0020, acceptance criteria).
func TestAQueueWithNothingSelectableStillOpens(t *testing.T) {
	m := newModel(Parse("Nothing is available.\n\nOrder is a suggestion.\n"))
	if m.selected() != "" {
		t.Fatal("something was selected in an empty queue")
	}
	if !strings.Contains(m.View(), "nothing to select") {
		t.Error("the screen did not say the queue offers nothing")
	}
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if out.(model).action != (Action{}) {
		t.Error("enter dispatched with no selection")
	}
	out, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if got := out.(model).action; got.Command != "status" {
		t.Errorf("s -> %+v, want status", got)
	}
}

// A terminal too short scrolls the selection into view rather than
// truncating the list silently.
func TestAShortTerminalScrollsTheSelectionIntoView(t *testing.T) {
	m := newModel(rowsOf(t))
	out, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 5})
	m = out.(model)
	for i := 0; i < 5; i++ {
		m.move(1)
	}
	view := m.View()
	if !strings.Contains(view, "task-0021") {
		t.Errorf("the selected row is outside the window:\n%s", view)
	}
	if strings.Count(view, "\n") > 6 {
		t.Errorf("the window rendered more lines than it has:\n%s", view)
	}
}

func TestParseDropsNoLineButTrailingBlanks(t *testing.T) {
	rows := Parse("a\n\nb\n\n\n")
	if len(rows) != 3 {
		t.Fatalf("rows = %d, want 3 (a, blank, b)", len(rows))
	}
	if rows[0].Text != "a" || rows[1].Text != "" || rows[2].Text != "b" {
		t.Errorf("rows = %+v", rows)
	}
}

// Open drives the real program. The suite has no terminal, so the input
// is a reader and the output a buffer — the same seam the term port
// uses so a guarded flow stays exercisable end to end.
func TestOpenRunsTheProgramAndReturnsWhatWasChosen(t *testing.T) {
	act, err := Open(listing, strings.NewReader("q"), io.Discard)
	if err != nil {
		t.Fatalf("Open = %v", err)
	}
	if act != (Action{}) {
		t.Errorf("q resolved to %+v, want the zero action", act)
	}
}

func TestOpenCarriesTheChosenCommandOut(t *testing.T) {
	act, err := Open(listing, strings.NewReader("w"), io.Discard)
	if err != nil {
		t.Fatalf("Open = %v", err)
	}
	if act.Command != "work" || act.Arg != "task-0020" {
		t.Errorf("Open = %+v, want work on the first selectable row", act)
	}
}

// Init asks for nothing: the rows arrived already read, so there is no
// first command to run.
func TestInitAsksForNothing(t *testing.T) {
	if cmd := newModel(rowsOf(t)).Init(); cmd != nil {
		t.Error("the screen asked for work on start; the rows are already read")
	}
}

// Scrolling back up brings the window with it.
func TestTheWindowFollowsTheSelectionUpward(t *testing.T) {
	m := newModel(rowsOf(t))
	out, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 4})
	m = out.(model)
	for i := 0; i < 5; i++ {
		m.move(1)
	}
	for i := 0; i < 5; i++ {
		m.move(-1)
	}
	// The window follows the selection, not the top of the list: the
	// first selectable row is the second line, so a window starting
	// there is the selection in view.
	if m.top > m.cursor {
		t.Errorf("top = %d is past the cursor %d", m.top, m.cursor)
	}
	if !strings.Contains(m.View(), "task-0020") {
		t.Error("the selected row is outside the window after scrolling back")
	}
}

// A window of one line still renders, rather than dividing by nothing.
func TestAWindowTooShortForAnythingStillRenders(t *testing.T) {
	m := newModel(rowsOf(t))
	out, _ := m.Update(tea.WindowSizeMsg{Width: 20, Height: 1})
	m = out.(model)
	if got := m.height; got != 1 {
		t.Errorf("height = %d, want the floor of 1", got)
	}
	if m.View() == "" {
		t.Error("a one-line window rendered nothing")
	}
}

// An unknown key changes nothing.
func TestAnUnknownKeyIsIgnored(t *testing.T) {
	m := newModel(rowsOf(t))
	out, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	if out.(model).action != (Action{}) || cmd != nil {
		t.Error("an unknown key did something")
	}
	if out.(model).selected() != "task-0020" {
		t.Error("an unknown key moved the selection")
	}
}
