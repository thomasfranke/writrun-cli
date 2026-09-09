// Package palette is the one place this binary decides what a colour
// means. Every command that colours anything asks here, so the screens
// and the reports cannot drift into two vocabularies
// (docs/product/rules.md).
//
// # Colour never carries meaning alone
//
// Every element this package colours is already distinguishable without
// it: a finding keeps its level's word in its own column, a group keeps
// its heading, a selected row keeps its cursor. Colour is the second
// signal, never the only one — a reader with `NO_COLOR` set, a piped
// run, and a reader who cannot tell the hues apart all get the same
// answer.
package palette

import "github.com/charmbracelet/lipgloss"

// Palette paints one run's output. The zero value paints nothing, which
// is what a run with colour disabled uses: every method hands the string
// back unchanged, so a caller never branches on whether colour is on.
type Palette struct{ on bool }

// New returns a palette that paints only where the run allows it. The
// decision is the frame's — stdout is a terminal, `NO_COLOR` is unset,
// `--no-color` was not given — and this package neither repeats nor
// second-guesses it (internal/command/color.go).
func New(enabled bool) Palette { return Palette{on: enabled} }

// The roles, named for what the reader is looking at rather than for a
// colour. A role that outlives a hue keeps its callers.
var (
	headingStyle  = lipgloss.NewStyle().Bold(true)
	declaredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	breaksStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	advisesStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	unreadStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

func (p Palette) paint(s lipgloss.Style, text string) string {
	if !p.on {
		return text
	}
	return s.Render(text)
}

// Heading is a group's name: the entry screen's groups, the queue's
// sections.
func (p Palette) Heading(text string) string { return p.paint(headingStyle, text) }

// Declared is a value the project chose and the binary is reporting
// back — the stage, the conduct flags.
func (p Palette) Declared(text string) string { return p.paint(declaredStyle, text) }

// Dim is text that qualifies rather than answers: a footer, a note the
// lister appended, the detail beneath a selection.
func (p Palette) Dim(text string) string { return p.paint(dimStyle, text) }

// Breaks, Advises and Unread are `doctor`'s three levels, which the
// report already prints as words in their own column
// (docs/product/adoption/doctor.md).
func (p Palette) Breaks(text string) string  { return p.paint(breaksStyle, text) }
func (p Palette) Advises(text string) string { return p.paint(advisesStyle, text) }
func (p Palette) Unread(text string) string  { return p.paint(unreadStyle, text) }
