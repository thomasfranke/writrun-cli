package screen

import (
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Requirements is what the doctor screen shows: the groups a report
// already counted, and under each the rows it already marked.
//
// It arrives already examined. Nothing here runs a check, knows what a
// mark means, or writes anything — those are `doctor`'s, and a copy
// would be a second authority that drifts
// (docs/product/screens/adoption/doctor.excalidraw).
type Requirements struct {
	// Header is the first line: the product, its version, the tag it
	// pins, the branch. The caller composes it from facts it holds.
	Header string
	// Line is the second: the declaration, the range examined and the
	// rung previewed.
	Line   string
	Groups []RequirementGroup
}

// RequirementGroup is one stage: its heading, already counted, and the
// rows under it.
type RequirementGroup struct {
	Name string
	Rows []Requirement
}

// Requirement is one row: the glyph, what the row says, the stable name
// the document explains it under, and that explanation.
//
// Explain arrives filled because looking a sentence up is the caller's
// act: the document states it, and a screen holding its own copy would
// be the drift the lookup exists to stop.
type Requirement struct {
	Mark    string
	Text    string
	Name    string
	Explain string
}

// keyRerun runs every check again. The cursor stays where it was: a
// reader who fixed one thing is looking at that row.
const keyRerun = "r"

// requirementsChrome is the lines the screen keeps below the rows: the
// blank, the rule, two for the explanation, the blank, the footer.
const requirementsChrome = 6

// OpenRequirements runs the doctor screen until the reader leaves it.
//
// load is called at the start and again on every `r`, so the screen
// shows the repository rather than a remembered copy of it: whether a
// requirement holds is the repository's answer.
func OpenRequirements(load func() (Requirements, error), in io.Reader, w io.Writer) error {
	r, err := load()
	if err != nil {
		return err
	}
	p := tea.NewProgram(
		newRequirements(r, load),
		tea.WithInput(in),
		tea.WithOutput(w),
		tea.WithAltScreen(),
	)
	_, err = p.Run()
	return err
}

type requirementsModel struct {
	rows   []requirementRow
	cursor int
	height int
	width  int
	top    int

	load func() (Requirements, error)
	// err is what went wrong examining the repository, shown in place of
	// the rows: a run that could not be made is not a reason to close
	// the screen on somebody.
	err error
}

// requirementRow is one rendered line. A row with no name is a heading,
// a blank or the first two lines — shown, and skipped by the selection,
// as the queue screen's rows already are (Row.Selectable).
type requirementRow struct {
	text    string
	name    string
	explain string
}

func (r requirementRow) selectable() bool { return r.name != "" }

func newRequirements(r Requirements, load func() (Requirements, error)) requirementsModel {
	m := requirementsModel{load: load, cursor: -1}
	m.fill(r)
	return m
}

func (m *requirementsModel) fill(r Requirements) {
	var rows []requirementRow
	push := func(text string) { rows = append(rows, requirementRow{text: text}) }

	for _, line := range (chrome{identity: r.Header, context: r.Line}).lines(contentWidth(m.width)) {
		push(line)
	}
	for _, g := range r.Groups {
		push("")
		push(" " + g.Name)
		for _, row := range g.Rows {
			rows = append(rows, requirementRow{
				text:    "   " + row.Mark + "  " + row.Text,
				name:    row.Name,
				explain: row.Explain,
			})
		}
	}
	m.rows = rows
	// The reader's place is kept across a re-run where the row it was on
	// is still a row — a repaired requirement should not move them.
	if m.cursor < 0 || m.cursor >= len(rows) || !rows[m.cursor].selectable() {
		m.cursor = firstRequirement(rows)
	}
	m.scroll()
}

func firstRequirement(rows []requirementRow) int {
	for i, r := range rows {
		if r.selectable() {
			return i
		}
	}
	return -1
}

func (m requirementsModel) Init() tea.Cmd { return nil }

func (m requirementsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		switch {
		case msg.Height == 0:
			m.height = 0
		case msg.Height > requirementsChrome:
			m.height = msg.Height - requirementsChrome
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
			r, err := m.load()
			m.err = err
			if err == nil {
				m.fill(r)
			}
		case keyQuit, "ctrl+c", keyBack:
			return m, tea.Quit
		}
	}
	return m, nil
}

// move steps to the next selectable row and stops at the ends rather
// than wrapping — a list read top to bottom should not silently start
// again.
func (m *requirementsModel) move(d int) {
	if m.cursor < 0 {
		return
	}
	for i := m.cursor + d; i >= 0 && i < len(m.rows); i += d {
		if m.rows[i].selectable() {
			m.cursor = i
			m.scroll()
			return
		}
	}
}

func (m *requirementsModel) scroll() {
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

func (m requirementsModel) View() string {
	if m.err != nil {
		return fmt.Sprintf(" the requirements could not be read: %v\n\n%s\n", m.err,
			footer{way: backOrQuit}.line())
	}

	var b strings.Builder
	end := len(m.rows)
	if m.height > 0 && m.top+m.height < end {
		end = m.top + m.height
	}
	for i := m.top; i < end; i++ {
		line := m.rows[i].text
		if i == m.cursor {
			line = cursorIn(line)
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}

	b.WriteByte('\n')
	b.WriteString(rule(contentWidth(m.width)) + "\n")
	// The list scrolls and the footer stays: a terminal too short for
	// both is still a terminal the selected row is explained on.
	for _, line := range m.footer() {
		b.WriteString(line + "\n")
	}
	b.WriteByte('\n')
	b.WriteString(footer{
		movement: "↑↓ move",
		actions:  []string{"r re-run"},
		way:      backOrQuit,
	}.line() + "\n")
	return b.String()
}

// footer is the selected requirement named, and the explanation under a
// hanging indent. It is wrapped, never truncated: a sentence longer than
// the pane is still the whole sentence.
func (m requirementsModel) footer() []string {
	if m.cursor < 0 || m.cursor >= len(m.rows) {
		return nil
	}
	r := m.rows[m.cursor]
	return wrap(r.name+" — "+r.explain, contentWidth(m.width), " ", "        ")
}
