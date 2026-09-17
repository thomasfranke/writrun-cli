package screen

import (
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// FirstRun is what `writrun` opens where `.writrun/` is absent: what
// this is, what the environment answers, and what to do next
// (docs/product/screens/first-run.excalidraw, spec-0038).
//
// It arrives already probed. Nothing here runs a check or knows what a
// mark means — the probe is `internal/requirements`', the same one
// `doctor`'s stage 0 uses, and a copy would be a second authority that
// drifts.
type FirstRun struct {
	// Identity is the line the other screens open with.
	Identity string
	// Context says no kit is here.
	Context string
	Groups  []FirstRunGroup
	// Ready says every environment requirement is met, and so whether
	// `enter` has anything to run. While it is false the footer offers
	// `r` instead, and `init` is out of reach (spec-0038).
	Ready bool
}

// FirstRunGroup is one heading and the rows under it.
type FirstRunGroup struct {
	Name string
	Rows []FirstRunRow
}

// FirstRunRow is one row. A row with a Mark is an environment
// requirement, drawn as `doctor` draws one; a row without is a command
// or a flag, drawn as the entry screen draws one.
type FirstRunRow struct {
	Mark string
	Name string
	// Text is what follows the name: the summary on a command row, the
	// note on a requirement that is unmet.
	Text string
	// Explain is the sentence under the rule when the row is selected.
	Explain string
	// Runs is the command `enter` dispatches on this row, empty where
	// the row runs nothing. A row the cursor stops on and cannot run is
	// still a row worth reading.
	Runs string
	// Selects says the cursor stops here.
	Selects bool
}

// The wordmark, drawn on this screen and on no other. It is the one
// place the product introduces itself, because it is the one screen a
// reader can reach without having adopted anything
// (docs/product/screens/first-run.excalidraw).
const (
	wordmark = "W R I T R U N"
	motto    = "What is written, runs"
)

// OpenFirstRun runs the screen until the reader leaves it, and answers
// the command they chose — empty where they chose none.
//
// load is called at the start and again on every `r`, so the screen
// shows the `PATH` rather than a remembered copy of it: a requirement
// installed while this is open is met the moment it is read again.
func OpenFirstRun(load func() (FirstRun, error), in io.Reader, w io.Writer) (string, error) {
	f, err := load()
	if err != nil {
		return "", err
	}
	p := tea.NewProgram(
		newFirstRun(f, load),
		tea.WithInput(in),
		tea.WithOutput(w),
		tea.WithAltScreen(),
	)
	final, err := p.Run()
	if err != nil {
		return "", err
	}
	return final.(firstRunModel).chosen, nil
}

type firstRunModel struct {
	chrome chrome
	rows   []firstRunRow
	ready  bool
	cursor int
	height int
	width  int
	top    int

	load func() (FirstRun, error)
	err  error
	// chosen is the command `enter` picked, empty until it does.
	chosen string
}

type firstRunRow struct {
	text    string
	explain string
	runs    string
	selects bool
}

// firstRunChrome is the lines kept around the rows: the wordmark's
// three, the header's two, the blank, the rule, the explanation's
// three, the blank, the footer.
const firstRunChrome = 12

func newFirstRun(f FirstRun, load func() (FirstRun, error)) firstRunModel {
	m := firstRunModel{load: load, cursor: -1}
	m.fill(f)
	return m
}

func (m *firstRunModel) fill(f FirstRun) {
	m.chrome = chrome{identity: f.Identity, context: f.Context}
	m.ready = f.Ready

	width := 0
	for _, g := range f.Groups {
		for _, r := range g.Rows {
			if r.Mark == "" && len(r.Name) > width {
				width = len(r.Name)
			}
		}
	}

	var rows []firstRunRow
	for _, g := range f.Groups {
		rows = append(rows, firstRunRow{text: ""}, firstRunRow{text: " " + g.Name})
		for _, r := range g.Rows {
			rows = append(rows, firstRunRow{
				text:    firstRunText(r, width),
				explain: r.Explain,
				runs:    r.Runs,
				selects: r.Selects,
			})
		}
	}
	m.rows = rows
	// The reader's place is kept across a re-read where the row it was
	// on is still a row — a repaired requirement should not move them.
	if m.cursor < 0 || m.cursor >= len(rows) || !rows[m.cursor].selects {
		m.cursor = -1
		for i, r := range rows {
			if r.selects {
				m.cursor = i
				break
			}
		}
	}
	m.scroll()
}

// firstRunText is one row rendered: a marked requirement in the shape
// `doctor` draws one, a command in the shape the entry screen draws
// one.
func firstRunText(r FirstRunRow, width int) string {
	if r.Mark != "" {
		line := "   " + r.Mark + "  " + r.Name
		if r.Text != "" {
			line += " — " + r.Text
		}
		return line
	}
	return "   " + pad(r.Name, width) + "   " + r.Text
}

func (m firstRunModel) Init() tea.Cmd { return nil }

func (m firstRunModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		switch {
		case msg.Height == 0:
			m.height = 0
		case msg.Height > firstRunChrome:
			m.height = msg.Height - firstRunChrome
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
		case keyRerun:
			f, err := m.load()
			m.err = err
			if err == nil {
				m.fill(f)
			}
		case keyTake:
			// A row that runs nothing is a row `enter` does nothing to.
			// `init` is held out of reach the same way: the row says
			// why, and the key it would have answered is not offered
			// (spec-0038).
			if m.ready && m.cursor >= 0 && m.rows[m.cursor].runs != "" {
				m.chosen = m.rows[m.cursor].runs
				return m, tea.Quit
			}
		case keyQuit, "ctrl+c", keyBack:
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *firstRunModel) move(d int) {
	if m.cursor < 0 {
		return
	}
	for i := m.cursor + d; i >= 0 && i < len(m.rows); i += d {
		if m.rows[i].selects {
			m.cursor = i
			m.scroll()
			return
		}
	}
}

func (m *firstRunModel) scroll() {
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

func (m firstRunModel) View() string {
	var b strings.Builder
	width := contentWidth(m.width)
	for _, line := range wordmarkLines(width) {
		b.WriteString(line + "\n")
	}
	for _, line := range m.chrome.lines(width) {
		b.WriteString(line + "\n")
	}

	end := window(len(m.rows), m.top, m.height)
	for i := m.top; i < end; i++ {
		line := m.rows[i].text
		if i == m.cursor {
			line = cursorIn(line)
		}
		b.WriteString(line + "\n")
	}

	b.WriteByte('\n')
	b.WriteString(rule(width) + "\n")
	for _, line := range wrap(m.explain(), width, " ", "        ") {
		b.WriteString(line + "\n")
	}
	b.WriteByte('\n')
	b.WriteString(m.footer().line() + "\n")
	return b.String()
}

// footer offers `enter` only where something can run, and `r` only
// where something has to change before anything can
// (docs/product/screens/first-run.excalidraw).
func (m firstRunModel) footer() footer {
	f := footer{movement: "↑↓ move", way: quitOnly}
	if m.ready {
		f.actions = []string{"enter run"}
	} else {
		f.actions = []string{"r re-check"}
	}
	return f
}

func (m firstRunModel) explain() string {
	if m.err != nil {
		return "the environment could not be read: " + m.err.Error()
	}
	if m.cursor < 0 || m.cursor >= len(m.rows) {
		return ""
	}
	return m.rows[m.cursor].explain
}

// wordmarkLines are the three lines the product introduces itself with.
func wordmarkLines(width int) []string {
	inner := "  " + wordmark + "   ·   " + motto
	gap := width - columns(inner)
	if gap < 0 {
		gap = 0
	}
	return []string{
		" ╭" + strings.Repeat("─", width) + "╮",
		" │" + inner + strings.Repeat(" ", gap) + "│",
		" ╰" + strings.Repeat("─", width) + "╯",
	}
}
