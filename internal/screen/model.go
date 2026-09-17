package screen

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Action is what a key chose: the command to run and the argument it
// carries. A zero Action is nothing chosen.
//
// It is a value rather than a call because a model decides and does not
// act — the session is what runs it, on a terminal it released first,
// so the command owns the keyboard alone while it asks its questions
// (session.go).
type Action struct {
	Command string
	Arg     string
}

// keys are the rule's, and the rule is the whole table
// (docs/product/screens/README.md).
const (
	keyTake   = "enter"
	keyWork   = "w"
	keyStatus = "s"
	keyBack   = "esc"
	keyQuit   = "q"
)

// nothingToSelect is what the movement group says where no row moves.
const nothingToSelect = "nothing to select"

// emptyQueue is the explanation an empty queue carries. It is the
// screen's own sentence because no row is there to explain, and it
// names the key that answers instead of the row
// (docs/product/screens/tasks/list.excalidraw).
const emptyQueue = "Nothing is available means every task is done, in flight, " +
	"or held back — not that there is nothing to do. `s` shows where the " +
	"current branch stands; a held-back row names what it waits on."

// model is the screen. It holds no repository state: the rows arrived
// already read, and nothing here writes.
type model struct {
	chrome chrome
	rows   []Row
	cursor int
	// height is the rows the viewport can show, 0 until the terminal
	// says. The selection is kept in view rather than the list
	// truncated, so a queue longer than the window still reaches its
	// end.
	height int
	width  int
	top    int
	action Action
	// back says `esc` was pressed: the reader came from the entry
	// screen and is going back to it, which is not an action and not a
	// departure.
	back bool
	// left says the reader asked to go. See entryModel.left.
	left bool
}

// queueChrome is the lines the screen keeps around the rows: the two
// header lines, the blank under them, the two blanks over the
// explanation, its two lines, the blank, the footer.
const queueChrome = 9

func newModel(identity string, rows []Row) model {
	return model{
		chrome: chrome{identity: identity, context: queueContext(rows)},
		rows:   rows,
		cursor: firstSelectable(rows),
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		// A height of zero is a terminal that has not said how tall it
		// is, not a terminal with no room: it gets no limit, and the
		// rows all render. A terminal that did say, and said something
		// too short to hold the chrome, still gets a line to read.
		switch {
		case msg.Height == 0:
			m.height = 0
		case msg.Height > queueChrome:
			m.height = msg.Height - queueChrome
		default:
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
		case keyBack:
			m.back = true
			return m, tea.Quit
		case keyQuit, "ctrl+c":
			m.left = true
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
	width := contentWidth(m.width)
	for _, line := range m.chrome.lines(width) {
		b.WriteString(line + "\n")
	}
	b.WriteByte('\n')

	end := window(len(m.rows), m.top, m.height)
	for i := m.top; i < end; i++ {
		b.WriteString(cursorBefore(i == m.cursor) + m.rows[i].Text + "\n")
	}

	// The queue draws no rule over its explanation: its rows are the
	// lister's own, and a rule under them would read as the lister's
	// (docs/product/screens/tasks/list.excalidraw).
	b.WriteString("\n\n")
	for _, line := range wrap(m.explain(), width, " ", "        ") {
		b.WriteString(line + "\n")
	}
	b.WriteByte('\n')
	b.WriteString(m.footer().line() + "\n")
	return b.String()
}

// footer names the keys that can act on the row under the cursor, and
// drops the ones that cannot where there is no row
// (docs/product/screens/tasks/list.excalidraw).
func (m model) footer() footer {
	if m.cursor < 0 {
		return footer{
			movement: nothingToSelect,
			actions:  []string{"s status"},
			way:      backOrQuit,
		}
	}
	return footer{
		movement: "↑↓ move",
		actions:  []string{"enter take", "w work", "s status"},
		way:      backOrQuit,
	}
}

// explain is the selected row named, with the act the next key performs
// on it — never only what the row is (spec-0040).
func (m model) explain() string {
	if m.cursor < 0 || m.cursor >= len(m.rows) {
		return emptyQueue
	}
	r := m.rows[m.cursor]
	return r.ID + " — " + r.State + ". `enter` takes it: the branch is cut " +
		"from a fresh origin/main, given its first commit, pushed, and a " +
		"draft pull request opened. `w` launches the configured agent on " +
		"it instead."
}
