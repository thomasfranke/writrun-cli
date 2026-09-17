package screen

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// A plan is rows, a cursor, and a footer naming the act the next key
// performs on the row under it (spec-0040).
func samplePlan() Choice {
	return Choice{
		Rows: []Line{
			{Text: "writrun uninstall — the plan:"},
			{Text: ""},
			{Text: "  remove       .writrun/", Detail: ".writrun/ — the kit, whole.", Selects: true},
			{Text: "  stays        work/", Detail: "work/ — the project's.", Selects: true},
		},
		Verb: "remove",
	}
}

func TestThePlanPutsACursorOnItsRowsAndSkipsTheRest(t *testing.T) {
	m := newChoice(samplePlan())
	if m.cursor != 2 {
		t.Errorf("the cursor opens at row %d, want the first row that selects", m.cursor)
	}
	view := m.View()
	if !strings.HasPrefix(strings.Split(view, "\n")[2], "›") {
		t.Errorf("the cursor is not on the first selectable row:\n%s", view)
	}
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if got := out.(choiceModel).cursor; got != 2 {
		t.Errorf("up moved to %d — a list read top to bottom stops at its ends", got)
	}
}

// The footer names the selected row's act, and ends `esc cancels`.
func TestThePlansFooterNamesTheAct(t *testing.T) {
	last := keyLine(Draw(samplePlan(), drawnWidth, 0))
	if last != "↑↓ move · enter remove · esc cancels" {
		t.Errorf("footer = %q", last)
	}
}

// The explanation follows the cursor and says what the row is.
func TestTheExplanationFollowsTheCursor(t *testing.T) {
	first := strings.Join(Draw(samplePlan(), drawnWidth, 0), "\n")
	if !strings.Contains(first, ".writrun/ — the kit, whole.") {
		t.Errorf("the pane does not explain the selected row:\n%s", first)
	}
	second := strings.Join(Draw(samplePlan(), drawnWidth, 1), "\n")
	if !strings.Contains(second, "work/ — the project's.") {
		t.Errorf("the pane did not follow the cursor:\n%s", second)
	}
}

// A row with no sentence shows its own text rather than an empty pane.
func TestARowWithNoSentenceShowsItself(t *testing.T) {
	c := Choice{Rows: []Line{{Text: "  a row nobody explained", Selects: true}}, Verb: "choose"}
	if !strings.Contains(strings.Join(Draw(c, drawnWidth, 0), "\n"), "a row nobody explained") {
		t.Error("a row with no sentence rendered an empty pane")
	}
}

// A plan with one row: the cursor rests on it and the footer explains
// it (spec-0040, edge cases).
func TestAPlanWithOneRow(t *testing.T) {
	c := Choice{Rows: []Line{{Text: "  the only row", Detail: "the only row — and this is it.", Selects: true}}, Verb: "refresh"}
	lines := Draw(c, drawnWidth, 0)
	if !strings.HasPrefix(lines[0], "›") {
		t.Errorf("the cursor does not rest on the one row: %q", lines[0])
	}
	if !strings.Contains(strings.Join(lines, "\n"), "and this is it.") {
		t.Error("the one row was not explained")
	}
}

// `enter` answers with the caller's own numbering: how many selectable
// rows came before the one chosen, so a plan's prose lines never shift
// the answer.
func TestTheAnswerIsTheCallersOwnNumbering(t *testing.T) {
	m := newChoice(samplePlan())
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = out.(choiceModel)
	if got := m.index(m.cursor); got != 1 {
		t.Errorf("the second selectable row answered %d, want 1", got)
	}
}

// `esc` is the decline, and it chooses nothing.
func TestEscChoosesNothing(t *testing.T) {
	m := newChoice(samplePlan())
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if got := out.(choiceModel).chosen; got != -1 {
		t.Errorf("esc chose row %d", got)
	}
}
