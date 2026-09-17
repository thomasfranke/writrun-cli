package screen

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func press(m requirementsModel, key string) requirementsModel {
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return next.(requirementsModel)
}

func arrow(m requirementsModel, t tea.KeyType) requirementsModel {
	next, _ := m.Update(tea.KeyMsg{Type: t})
	return next.(requirementsModel)
}

func newDrawn() requirementsModel {
	r := drawnRows()
	return newRequirements(r, func() (Requirements, error) { return r, nil })
}

// Every requirement row is selectable, the met ones included, and the
// headings, the counts and the blank lines are shown and skipped — as
// the queue screen's rows already are (spec-0037, step 1).
func TestTheSelectionStopsOnEveryRequirementAndNothingElse(t *testing.T) {
	m := newDrawn()
	if m.cursor < 0 {
		t.Fatal("the screen opened on no row")
	}
	seen := 0
	for {
		if !m.rows[m.cursor].selectable() {
			t.Fatalf("the cursor stopped on %q, which is not a requirement", m.rows[m.cursor].text)
		}
		seen++
		before := m.cursor
		m = arrow(m, tea.KeyDown)
		if m.cursor == before {
			break
		}
	}
	want := 0
	for _, g := range drawnRows().Groups {
		want += len(g.Rows)
	}
	if seen != want {
		t.Errorf("the cursor stopped on %d rows, want every one of the %d requirements", seen, want)
	}
}

// The first row a heading or a blank would be is never the cursor's, and
// the ends do not wrap.
func TestTheCursorStopsAtTheEnds(t *testing.T) {
	m := newDrawn()
	first := m.cursor
	m = arrow(m, tea.KeyUp)
	if m.cursor != first {
		t.Errorf("up from the first row moved to %d, want %d", m.cursor, first)
	}
}

func TestTheFooterNamesTheSelectedRequirementAndExplainsIt(t *testing.T) {
	cases := []struct{ name, row, want string }{
		{"a requirement that is met", "writrun/gates.md", "who operates each gate the methodology names."},
		{"a requirement that is not", "the recording push can write to main", "Not met: .github/workflows/record.yml"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := screenAt(t, c.row)
			view := m.View()
			if !strings.Contains(view, " "+c.row+" — ") {
				t.Errorf("the footer does not name %q:\n%s", c.row, view)
			}
			if !strings.Contains(unwrapped(view), c.want) {
				t.Errorf("the footer misses %q:\n%s", c.want, view)
			}
		})
	}
}

// A requirement the caller could explain in no other words is shown with
// what it was given and nothing invented (spec-0037, step 5).
func TestARequirementWithNoExplanationShowsItsNameAlone(t *testing.T) {
	r := Requirements{Header: "h", Line: "l", Groups: []RequirementGroup{
		{Name: "GROUP", Rows: []Requirement{{Mark: "✗", Text: "a check no tag has shipped yet",
			Name: "a check no tag has shipped yet", Explain: "it is not as the kit expects."}}},
	}}
	m := newRequirements(r, func() (Requirements, error) { return r, nil })
	if want := " a check no tag has shipped yet — it is not as the kit expects."; !strings.Contains(unwrapped(m.View()), want) {
		t.Errorf("the footer misses %q:\n%s", want, m.View())
	}
}

// `r` runs the checks again and keeps the cursor where it was: a reader
// who fixed one thing is looking at that row (spec-0037, step 6).
func TestRerunKeepsTheCursor(t *testing.T) {
	runs := 0
	r := drawnRows()
	m := newRequirements(r, func() (Requirements, error) { runs++; return r, nil })
	m = arrow(m, tea.KeyDown)
	m = arrow(m, tea.KeyDown)
	was := m.cursor
	m = press(m, "r")
	if runs != 1 {
		t.Errorf("the checks ran %d times on one `r`, want 1", runs)
	}
	if m.cursor != was {
		t.Errorf("cursor = %d after a re-run, want %d", m.cursor, was)
	}
}

// A re-run that cannot answer says so and leaves the screen open: a run
// that could not be made is not a reason to close the door on somebody.
func TestARerunThatCannotAnswerKeepsTheScreen(t *testing.T) {
	r := drawnRows()
	fail := false
	m := newRequirements(r, func() (Requirements, error) {
		if fail {
			return Requirements{}, errors.New("the settings could not be read")
		}
		return r, nil
	})
	fail = true
	m = press(m, "r")
	if !strings.Contains(m.View(), "the requirements could not be read") {
		t.Errorf("the screen does not say why:\n%s", m.View())
	}
	if !strings.Contains(m.View(), "q quit") {
		t.Errorf("the screen offers no way out:\n%s", m.View())
	}
}

// A terminal too short for both scrolls the list and keeps the footer:
// the selected row is explained whatever the height (spec-0037).
func TestAShortTerminalScrollsTheListAndKeepsTheFooter(t *testing.T) {
	m := newDrawn()
	next, _ := m.Update(tea.WindowSizeMsg{Width: 65, Height: 12})
	m = next.(requirementsModel)
	for i := 0; i < 20; i++ {
		m = arrow(m, tea.KeyDown)
	}
	view := m.View()
	if m.top == 0 {
		t.Error("the list never scrolled")
	}
	if !strings.Contains(view, "↑↓ move · r re-run · esc back · q quit") {
		t.Errorf("the footer went with the rows:\n%s", view)
	}
	if !strings.Contains(view, " "+m.rows[m.cursor].name+" —") {
		t.Errorf("the footer does not name the selected requirement:\n%s", view)
	}
}

// A sentence longer than the pane wraps under the hanging indent; it is
// never cut to fit (spec-0037, edge cases).
func TestALongExplanationWrapsAndIsNeverTruncated(t *testing.T) {
	long := strings.Repeat("a sentence that keeps going ", 12)
	r := Requirements{Header: "h", Line: "l", Groups: []RequirementGroup{
		{Name: "GROUP", Rows: []Requirement{{Mark: "✓", Text: "x", Name: "x", Explain: long}}},
	}}
	m := newRequirements(r, func() (Requirements, error) { return r, nil })
	next, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 0})
	m = next.(requirementsModel)
	view := m.View()
	for _, word := range strings.Fields(long) {
		if !strings.Contains(view, word) {
			t.Fatalf("the explanation was cut — %q is missing:\n%s", word, view)
		}
	}
	for _, line := range strings.Split(view, "\n") {
		if n := len([]rune(line)); n > 40 {
			t.Errorf("a line runs to %d columns in a 40-column terminal: %q", n, line)
		}
	}
}

// unwrapped is a view with its wrapping undone, so a case asserts a
// sentence rather than where a terminal broke it.
func unwrapped(view string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(view, "\n", " ")), " ")
}
