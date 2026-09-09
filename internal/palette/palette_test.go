package palette

import (
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// A test binary writes to a pipe, so lipgloss decides on its own that
// the run has no colour and every role renders plain — which would make
// the two cases below pass whatever this package does. Naming a profile
// takes that decision away, so a role that paints when it should not
// has somewhere to show it.
func TestMain(m *testing.M) {
	lipgloss.SetColorProfile(termenv.ANSI)
	os.Exit(m.Run())
}

// roles is every role this package offers, so one added without a line
// here is a role neither promise below holds.
func roles(p Palette) map[string]func(string) string {
	return map[string]func(string) string{
		"Heading":  p.Heading,
		"Declared": p.Declared,
		"Dim":      p.Dim,
		"Breaks":   p.Breaks,
		"Advises":  p.Advises,
		"Unread":   p.Unread,
	}
}

// The zero value paints nothing, which is what lets a caller print
// without asking whether colour is on. A role that painted anyway would
// put escape codes into a piped run and into a `NO_COLOR` reader's
// terminal, and no caller branches to catch it.
func TestTheZeroValueHandsTheTextBack(t *testing.T) {
	for name, role := range roles(Palette{}) {
		if got := role("task-0029"); got != "task-0029" {
			t.Errorf("%s painted a disabled run: %q", name, got)
		}
	}
	for name, role := range roles(New(false)) {
		if got := role("task-0029"); got != "task-0029" {
			t.Errorf("%s painted a run that disabled colour: %q", name, got)
		}
	}
}

// Colour is the second signal, never the only one: a role wraps the
// text and never replaces it, so the reader who never sees the colour
// reads the same answer.
func TestEveryRoleKeepsItsTextReadable(t *testing.T) {
	for name, role := range roles(New(true)) {
		painted := role("task-0029")
		if !strings.Contains(painted, "task-0029") {
			t.Errorf("%s lost the text it was given: %q", name, painted)
		}
		if painted == "task-0029" {
			t.Errorf("%s painted nothing on a run that allows colour", name)
		}
	}
}
