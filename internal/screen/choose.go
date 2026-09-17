package screen

import (
	"errors"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// A question and a plan are one screen, and this is it: rows, a cursor,
// and a footer naming the act the next key performs on the row under it
// (docs/product/screens/tasks/take.excalidraw,
// docs/product/screens/adoption/uninstall.excalidraw, spec-0040).
//
// There is one component because there was nearly a second: `take` and
// `init` asked through a form, and five commands printed a plan and
// asked a yes/no beside it, so the rows a reader was deciding about
// carried no cursor and no explanation at all. What a row means is the
// hardest thing to recover once a plan has run.
//
// It carries no header and no `q`: it is not a screen a reader arrived
// at, and `q` over a list of options would be a value (spec-0042).

// Choice is what the reader is choosing between.
type Choice struct {
	// Title is the question. A plan has none — its own first lines say
	// what it is.
	Title string
	// Rows are every line, in the order the caller composed them.
	Rows []Line
	// Verb is what `enter` does to the selected row: `choose` on a
	// question, `refresh`, `remove`, `mark ready` on a plan.
	Verb string
}

// Line is one line of a question or a plan: what it says, whether the
// cursor stops on it, and what the explanation says about it.
type Line struct {
	// Text carries its own indent, because the caller composed it — the
	// cursor takes the column in front of it and moves nothing.
	Text string
	// Detail is the sentence under the rule when this line is selected.
	// A line with none shows its own text instead, rather than an empty
	// pane.
	Detail  string
	Selects bool
}

// ErrCancelled is `esc`: the reader chose nothing and nothing was done.
var ErrCancelled = errors.New("cancelled")

// Choose runs the list until the reader answers it, and returns the
// index of the row they chose. ErrCancelled is `esc`.
func Choose(c Choice, in io.Reader, out io.Writer) (int, error) {
	m := newChoice(c)
	if len(m.selectable()) == 0 {
		return -1, ErrCancelled
	}
	p := tea.NewProgram(m, tea.WithInput(in), tea.WithOutput(out), tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return -1, err
	}
	done := final.(choiceModel)
	if done.chosen < 0 {
		return -1, ErrCancelled
	}
	return done.index(done.chosen), nil
}

// choiceModel is the list. It decides nothing about what a row means:
// the rows arrived composed, and the explanation arrived with them.
type choiceModel struct {
	title  string
	rows   []Line
	verb   string
	cursor int
	height int
	width  int
	top    int
	// chosen is the row `enter` took, -1 until it does.
	chosen int
}

func newChoice(c Choice) choiceModel {
	m := choiceModel{title: c.Title, rows: c.Rows, verb: c.Verb, chosen: -1}
	m.cursor = -1
	for i, r := range c.Rows {
		if r.Selects {
			m.cursor = i
			break
		}
	}
	return m
}

// selectable is the rows the cursor stops on.
func (m choiceModel) selectable() []int {
	var out []int
	for i, r := range m.rows {
		if r.Selects {
			out = append(out, i)
		}
	}
	return out
}

// index is the caller's own numbering: how many selectable rows come
// before this one. A caller composed its rows and its options in one
// order, and this is what makes the answer that order's.
func (m choiceModel) index(row int) int {
	n := 0
	for i, r := range m.rows {
		if i == row {
			return n
		}
		if r.Selects {
			n++
		}
	}
	return -1
}

// chooseChrome is the lines kept below the rows: the two blanks, the
// explanation's two lines, the blank, the footer. The title and the
// blank under it are counted where there is one.
const chooseChrome = 6

func (m choiceModel) Init() tea.Cmd { return nil }

func (m choiceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		above := chooseChrome
		if m.title != "" {
			above += 2
		}
		switch {
		case msg.Height == 0:
			m.height = 0
		case msg.Height > above:
			m.height = msg.Height - above
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
			if m.cursor >= 0 {
				m.chosen = m.cursor
				return m, tea.Quit
			}
		case keyBack, "ctrl+c":
			m.chosen = -1
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *choiceModel) move(d int) {
	if m.cursor < 0 {
		return
	}
	for i := m.cursor + d; i >= 0 && i < len(m.rows); i += d {
		if m.rows[i].Selects {
			m.cursor = i
			m.scroll()
			return
		}
	}
}

func (m *choiceModel) scroll() {
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

func (m choiceModel) View() string {
	var b strings.Builder
	width := contentWidth(m.width)
	if m.title != "" {
		b.WriteString(" " + m.title + "\n\n")
	}

	end := window(len(m.rows), m.top, m.height)
	for i := m.top; i < end; i++ {
		b.WriteString(cursorBefore(i == m.cursor) + m.rows[i].Text + "\n")
	}

	b.WriteString("\n\n")
	for _, line := range wrap(m.explain(), width, " ", "        ") {
		b.WriteString(line + "\n")
	}
	b.WriteByte('\n')
	b.WriteString(footer{
		movement: "↑↓ move",
		actions:  []string{"enter " + m.verb},
		way:      cancels,
	}.line() + "\n")
	return b.String()
}

// explain is the selected row's own sentence, or the row itself where
// the caller wrote none — a pane with nothing in it says less than the
// row already did (spec-0040, edge cases).
func (m choiceModel) explain() string {
	if m.cursor < 0 || m.cursor >= len(m.rows) {
		return ""
	}
	r := m.rows[m.cursor]
	if r.Detail != "" {
		return r.Detail
	}
	return strings.TrimSpace(r.Text)
}

// Draw is the lines a question or a plan renders at a width, with the
// cursor on the nth row that can take one.
//
// It runs no program. The view is what a drawing states, and the suite
// holds a command's composition to its frame by reading it here rather
// than by driving a terminal that does not exist
// (docs/product/screens/README.md).
func Draw(c Choice, width, on int) []string {
	m := newChoice(c)
	next, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: 0})
	m = next.(choiceModel)
	if rows := m.selectable(); on >= 0 && on < len(rows) {
		m.cursor = rows[on]
	}
	return trimmed(m.View())
}

// trimmed is rendered output as a frame is read: trailing spaces gone,
// and the blank lines at either end with them.
func trimmed(out string) []string {
	var lines []string
	for _, l := range strings.Split(out, "\n") {
		lines = append(lines, strings.TrimRight(l, " "))
	}
	for len(lines) > 0 && lines[0] == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
