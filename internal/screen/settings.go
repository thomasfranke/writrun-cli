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
	// Header is the line above the sections, composed by the caller
	// from facts it already holds.
	Header string
	Groups []SettingGroup
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
	rows   []settingRow
	cursor int
	height int
	top    int

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

func (m *settingsModel) fill(s Settings) {
	var rows []settingRow
	push := func(text string) { rows = append(rows, settingRow{text: text}) }

	push(" " + s.Header)

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
		push(" " + g.Name)
		for _, r := range g.Rows {
			rows = append(rows, settingRow{
				text:  "   " + pad(r.Name, width) + "  " + r.Value,
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
		// Four lines are the separator, the detail's two, and the
		// footer; one more is the blank above them.
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
		label: "config " + key,
		run:   func() { change(key) },
		in:    m.in,
		out:   m.out,
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
	if m.err != nil {
		return fmt.Sprintf(" the settings could not be read: %v\n\n q back\n", m.err)
	}

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
	if m.cursor >= 0 && m.cursor < len(m.rows) {
		r := m.rows[m.cursor]
		b.WriteString(" " + r.name + " — " + r.value + "\n")
	}
	b.WriteByte('\n')
	b.WriteString("↑↓ move · enter change · q back\n")
	return b.String()
}
