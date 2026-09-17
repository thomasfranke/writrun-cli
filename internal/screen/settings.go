package screen

import (
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Settings is what the settings screen shows: the sections the kit
// gives, and under each the keys it documents with the values its
// reader answered.
//
// It arrives already read. Nothing here knows the file's shape, which
// key exists, or what a value may be — those are the kit's, and a copy
// would be a second authority that drifts
// (docs/product/screens/README.md).
type Settings struct {
	// Identity is the first line: the product, its version, the tag it
	// pins, the branch. Context names the file the sections were read
	// from, and Source the check that judges a change to it — drawn at
	// the right of the same line. The caller composes all three from
	// facts it already holds.
	Identity string
	Context  string
	Source   string
	// Checker is the check named in the explanation under the rule. It
	// is the same check Source names, without the words around it.
	Checker string
	Groups  []SettingGroup
}

// SettingGroup is one section and the keys under it.
type SettingGroup struct {
	Name string
	Rows []Setting
}

// Setting is one key: what the reader is shown, what identifies it to
// the caller, and the value read for it.
//
// The two names are not the same and must not be conflated. The
// drawing shows `auto_push`; the kit knows it as
// `stage_2.auto_push`, and that is what a change is asked for. A
// screen that handed back what it displayed would name a key the kit
// cannot find.
type Setting struct {
	// Name is shown.
	Name string
	// Key identifies, and is what Change is given.
	Key   string
	Value string
}

// Change is asked for a key and does whatever changing it means —
// asking for the value, writing, and letting the kit judge. It runs on
// a terminal this screen has released, because it asks questions and a
// question needs the keyboard to itself.
type Change func(key string)

// OpenSettings runs the settings screen until the reader leaves it.
//
// load is called at the start and again after every change, so the
// screen shows the file rather than a remembered copy of it: the kit's
// checker may have refused, and a refusal puts the previous bytes back.
func OpenSettings(load func() (Settings, error), change Change, in io.Reader, w io.Writer) error {
	s, err := load()
	if err != nil {
		return err
	}
	p := tea.NewProgram(
		newSettings(s, load, change, in, w),
		tea.WithInput(in),
		tea.WithOutput(w),
		tea.WithAltScreen(),
	)
	_, err = p.Run()
	return err
}

type settingsModel struct {
	chrome chrome
	// checker is the check named in the explanation, which is the one
	// named in the header: a reader is told what judges a change where
	// they are about to make one.
	checker string
	rows    []settingRow
	cursor  int
	height  int
	width   int
	top     int

	load   func() (Settings, error)
	change Change
	in     io.Reader
	out    io.Writer

	// err is what went wrong reading the settings, shown in place of
	// the rows: a file that cannot be read is not a reason to close the
	// screen on somebody mid-change.
	err error
}

// settingRow is one rendered line. A row with no key is a heading or a
// blank — shown, and skipped by the selection.
type settingRow struct {
	text string
	// name is shown on the detail line; key is what a change is asked
	// for. See Setting.
	name  string
	key   string
	value string
}

func newSettings(s Settings, load func() (Settings, error), change Change, in io.Reader, w io.Writer) settingsModel {
	m := settingsModel{load: load, change: change, in: in, out: w}
	m.fill(s)
	return m
}

// settingsChrome is the lines the screen keeps around the rows: the two
// header lines, the blank, the rule, the explanation's two lines, the
// blank, the footer.
const settingsChrome = 8

func (m *settingsModel) fill(s Settings) {
	var rows []settingRow
	push := func(text string) { rows = append(rows, settingRow{text: text}) }

	m.chrome = chrome{identity: s.Identity, context: s.Context, source: s.Source}
	m.checker = s.Checker

	width := 0
	for _, g := range s.Groups {
		for _, r := range g.Rows {
			if len(r.Name) > width {
				width = len(r.Name)
			}
		}
	}
	for _, g := range s.Groups {
		push("")
		push(g.Name)
		for _, r := range g.Rows {
			rows = append(rows, settingRow{
				text:  "  " + pad(r.Name, width) + "    " + r.Value,
				name:  r.Name,
				key:   r.Key,
				value: r.Value,
			})
		}
	}
	m.rows = rows
	// The reader's place is kept across a reload where the key it was
	// on still exists — a refused change should not move them.
	if m.cursor < 0 || m.cursor >= len(rows) || rows[m.cursor].key == "" {
		m.cursor = firstSetting(rows)
	}
	m.scroll()
}

func firstSetting(rows []settingRow) int {
	for i, r := range rows {
		if r.key != "" {
			return i
		}
	}
	return -1
}

func (m settingsModel) Init() tea.Cmd { return nil }

func (m settingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		switch {
		case msg.Height == 0:
			m.height = 0
		case msg.Height > settingsChrome:
			m.height = msg.Height - settingsChrome
		default:
			m.height = 1
		}
		m.scroll()
		return m, nil
	case changedMsg:
		s, err := m.load()
		m.err = err
		if err == nil {
			m.fill(s)
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.move(-1)
		case "down", "j":
			m.move(1)
		case keyTake: // enter
			if key := m.selected(); key != "" {
				return m, m.exec(key)
			}
		case keyQuit, "ctrl+c", keyBack:
			return m, tea.Quit
		}
	}
	return m, nil
}

// changedMsg says a change was attempted and the screen is coming back.
// Whether it was kept is the file's answer, which is why the rows are
// read again rather than edited here.
type changedMsg struct{}

func (m settingsModel) exec(key string) tea.Cmd {
	change := m.change
	d := dispatch{
		label:    "config " + key,
		identity: m.chrome.identity,
		run:      func() { change(key) },
		in:       m.in,
		out:      m.out,
	}
	return tea.Exec(d, func(error) tea.Msg { return changedMsg{} })
}

func (m settingsModel) selected() string {
	if m.cursor < 0 || m.cursor >= len(m.rows) {
		return ""
	}
	return m.rows[m.cursor].key
}

// move steps to the next selectable row and stops at the ends rather
// than wrapping — a list read top to bottom should not silently start
// again.
func (m *settingsModel) move(d int) {
	if m.cursor < 0 {
		return
	}
	for i := m.cursor + d; i >= 0 && i < len(m.rows); i += d {
		if m.rows[i].key != "" {
			m.cursor = i
			m.scroll()
			return
		}
	}
}

func (m *settingsModel) scroll() {
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

func (m settingsModel) View() string {
	width := contentWidth(m.width)
	if m.err != nil {
		return m.unreadableView(width)
	}

	var b strings.Builder
	for _, line := range m.chrome.lines(width) {
		b.WriteString(line + "\n")
	}

	end := window(len(m.rows), m.top, m.height)
	for i := m.top; i < end; i++ {
		b.WriteString(cursorBefore(i == m.cursor) + m.rows[i].text + "\n")
	}

	b.WriteByte('\n')
	b.WriteString(rule(width) + "\n")
	if m.cursor >= 0 && m.cursor < len(m.rows) {
		r := m.rows[m.cursor]
		for _, line := range wrap(r.name+" — "+r.value+". Every key above sits where "+
			m.checker+" says it lives; nothing here is a list this binary keeps.",
			width, " ", "        ") {
			b.WriteString(line + "\n")
		}
	}
	b.WriteByte('\n')
	b.WriteString(footer{
		movement: "↑↓ move",
		actions:  []string{"enter change"},
		way:      backOrQuit,
	}.line() + "\n")
	return b.String()
}

// unreadableView is the settings that could not be read: the failure
// marked as `doctor` marks one, and the command that writes the file
// named — an error alone says what broke and not what to do
// (spec-0042).
func (m settingsModel) unreadableView(width int) string {
	var b strings.Builder
	for _, line := range (chrome{identity: m.chrome.identity, context: "CONFIG · unreadable"}).lines(width) {
		b.WriteString(line + "\n")
	}
	b.WriteString("\n  ✗  the settings could not be read\n")
	for _, line := range wrap(fmt.Sprint(m.err), width, "      ", "      ") {
		b.WriteString(line + "\n")
	}
	b.WriteString("\n\n")
	for _, line := range wrap("The settings file is the project's own, and writing it "+
		"is `init`'s act, never this screen's. The kit's reader documents its "+
		"defaults and keeps working without one, so this is a repository that was "+
		"never adopted — or one whose home was moved.", width, " ", "        ") {
		b.WriteString(line + "\n")
	}
	b.WriteByte('\n')
	b.WriteString(footer{way: backOrQuit}.line() + "\n")
	return b.String()
}
