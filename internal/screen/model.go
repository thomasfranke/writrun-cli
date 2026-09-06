package screen

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Action is what the screen resolved to when it closed: the command to
// run and the argument it carries. A zero Action is `q` — the screen
// left and nothing runs.
//
// The screen closes *before* the command runs, so the command owns the
// terminal it asks its questions on. That is why this is a value handed
// back rather than a call made from inside the model: a huh form
// rendering underneath a live Bubble Tea program is two programs
// holding one terminal.
type Action struct {
	Command string
	Arg     string
}

// keys are the rule's, and the rule is the whole table
// (docs/product/screen.md).
const (
	keyTake   = "enter"
	keyWork   = "w"
	keyStatus = "s"
	keyQuit   = "q"
)

// model is the screen. It holds no repository state: the rows arrived
// already read, and nothing here writes.
type model struct {
	rows   []Row
	cursor int
	// height is the rows the viewport can show, 0 until the terminal
	// says. The selection is kept in view rather than the list
	// truncated, so a queue longer than the window still reaches its
	// end.
	height int
	top    int
	action Action
}

func newModel(rows []Row) model {
	return model{rows: rows, cursor: firstSelectable(rows)}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Two lines are the footer and the blank above it.
		m.height = msg.Height - 2
		if m.height < 1 {
			m.height = 1
		}
		m.scroll()
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.move(-1)
		case "down", "j":
			m.move(1)
		case keyTake:
			if id := m.selected(); id != "" {
				m.action = Action{Command: "take", Arg: id}
				return m, tea.Quit
			}
		case keyWork:
			if id := m.selected(); id != "" {
				m.action = Action{Command: "work", Arg: id}
				return m, tea.Quit
			}
		case keyStatus:
			m.action = Action{Command: "status"}
			return m, tea.Quit
		case keyQuit, "ctrl+c", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

// selected is the id under the cursor, empty when the queue offers no
// selectable row.
func (m model) selected() string {
	if m.cursor < 0 || m.cursor >= len(m.rows) {
		return ""
	}
	return m.rows[m.cursor].ID
}

// move steps to the next selectable row in a direction, and stops at
// the ends rather than wrapping — a queue read top to bottom should not
// silently start again.
func (m *model) move(d int) {
	if m.cursor < 0 {
		return
	}
	for i := m.cursor + d; i >= 0 && i < len(m.rows); i += d {
		if m.rows[i].Selectable() {
			m.cursor = i
			m.scroll()
			return
		}
	}
}

// scroll brings the cursor into the window.
func (m *model) scroll() {
	if m.height <= 0 || m.cursor < 0 {
		return
	}
	if m.cursor < m.top {
		m.top = m.cursor
	}
	if m.cursor >= m.top+m.height {
		m.top = m.cursor - m.height + 1
	}
	if m.top < 0 {
		m.top = 0
	}
}

func (m model) View() string {
	var b strings.Builder
	end := len(m.rows)
	if m.height > 0 && m.top+m.height < end {
		end = m.top + m.height
	}
	for i := m.top; i < end; i++ {
		if i == m.cursor {
			b.WriteString("›")
		} else {
			b.WriteByte(' ')
		}
		b.WriteString(m.rows[i].Text)
		b.WriteByte('\n')
	}
	b.WriteByte('\n')
	if m.cursor < 0 {
		b.WriteString("nothing to select · s status · q quit\n")
	} else {
		b.WriteString("↑↓ move · enter take · w work · s status · q quit\n")
	}
	return b.String()
}
