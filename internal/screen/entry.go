package screen

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Entry is what `writrun` with no command opens: every command that can
// run where the reader is standing, grouped by what it acts on
// (docs/product/screens/entry.excalidraw).
//
// It carries no command's behaviour and no queue. The header's facts are
// cheap reads the caller already made, and the rows are the command
// table's own names and summaries — nothing here is a string written for
// this screen alone, because a second copy of a summary is a second
// answer about one command.
type Entry struct {
	// Header is the first line: the product, its version, the tag it
	// pins, the branch. The caller composes it, because every fact in it
	// is one the caller already holds.
	Header string
	// Stage is the two lines beneath it, or empty where the settings
	// could not be read. A screen that cannot say the stage says nothing
	// about it rather than guessing at one.
	Stage []string
	// Groups are the rows, in the order they are shown.
	Groups []Group
}

// Group is one heading and the commands under it.
type Group struct {
	Name string
	Rows []Command
}

// Command is one row: what the key dispatches, and the one-line summary
// the table already carries.
type Command struct {
	Name    string
	Summary string
}

// listCommand is the one row that opens a screen instead of running a
// command. It is named here because the entry screen is the only place
// that treats a command specially, and the special case is one row.
const listCommand = "list"

// entryModel is the entry screen. Like the queue's, it holds no
// repository state and writes nothing: every key either moves the
// cursor or ends the screen with an action for the caller to perform.
type entryModel struct {
	rows   []entryRow
	cursor int
	height int
	top    int
	action Action
	// queue says the reader chose the row that opens the queue rather
	// than a command to run.
	queue bool
	// left says the reader asked to go, rather than choosing anything.
	// Alone the screen simply ends; inside a session the difference
	// between leaving and choosing nothing is the session's to act on.
	left bool
}

// entryRow is one rendered line. A row with no command is a heading, a
// blank or the separator — shown, and skipped by the selection.
type entryRow struct {
	text    string
	command string
	summary string
}

func newEntry(e Entry) entryModel {
	var rows []entryRow
	push := func(text string) { rows = append(rows, entryRow{text: text}) }

	push(" " + e.Header)
	for _, l := range e.Stage {
		push(" " + l)
	}
	width := 0
	for _, g := range e.Groups {
		for _, c := range g.Rows {
			if len(c.Name) > width {
				width = len(c.Name)
			}
		}
	}
	for _, g := range e.Groups {
		push("")
		push(" " + strings.ToUpper(g.Name))
		for _, c := range g.Rows {
			rows = append(rows, entryRow{
				text:    "   " + pad(c.Name, width) + "  " + c.Summary,
				command: c.Name,
				summary: c.Summary,
			})
		}
	}
	m := entryModel{rows: rows, cursor: firstEntry(rows)}
	return m
}

func pad(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func firstEntry(rows []entryRow) int {
	for i, r := range rows {
		if r.command != "" {
			return i
		}
	}
	return -1
}

func (m entryModel) Init() tea.Cmd { return nil }

func (m entryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Four lines are the separator, the detail's two, and the
		// footer; one more is the blank above them.
		// A height of zero is a terminal that has not said how tall it
		// is, not a terminal with no room: it gets no limit, and the
		// rows all render. A terminal that did say, and said something
		// too short to hold the chrome, still gets a line to read.
		switch {
		case msg.Height == 0:
			m.height = 0
		case msg.Height > 5:
			m.height = msg.Height - 5
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
		case keyTake: // enter
			if name := m.selected(); name != "" {
				if name == listCommand {
					m.queue = true
				} else {
					m.action = Action{Command: name}
				}
				return m, tea.Quit
			}
		case keyQuit, "ctrl+c", "esc":
			m.left = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m entryModel) selected() string {
	if m.cursor < 0 || m.cursor >= len(m.rows) {
		return ""
	}
	return m.rows[m.cursor].command
}

func (m *entryModel) move(d int) {
	if m.cursor < 0 {
		return
	}
	for i := m.cursor + d; i >= 0 && i < len(m.rows); i += d {
		if m.rows[i].command != "" {
			m.cursor = i
			m.scroll()
			return
		}
	}
}

func (m *entryModel) scroll() {
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

func (m entryModel) View() string {
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
		b.WriteString(m.rows[i].text)
		b.WriteByte('\n')
	}
	b.WriteByte('\n')
	// The detail line is the selected command's whole summary — the row
	// shows what fits, and this shows what the command table says.
	if m.cursor >= 0 && m.cursor < len(m.rows) {
		r := m.rows[m.cursor]
		b.WriteString(" " + r.command + " — " + r.summary + "\n")
	}
	b.WriteByte('\n')
	b.WriteString("↑↓ move · enter run · q quit\n")
	return b.String()
}
