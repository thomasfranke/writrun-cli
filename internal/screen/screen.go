package screen

import (
	"io"

	tea "github.com/charmbracelet/bubbletea"
)

// Open runs the screen until the reader leaves it. The entry screen and
// the queue are two states of it, and a command chosen on either runs
// and hands the screen back — reading the next thing is a keypress, not
// another `writrun`.
//
// queue is called when the reader opens the queue, and again after a
// command, and returns the lister's output. run is called with the
// command a key chose. Both are callbacks rather than fields because
// running the kit's scripts is the caller's act, not this package's —
// nothing here knows a script path or a command's behaviour
// (docs/technical/engineering/coupling.md).
//
// A command still owns the terminal alone while it asks its questions;
// see session for why that is a pause rather than a close.
func Open(e Entry, queue func() (string, error), run Runner, in io.Reader, w io.Writer) error {
	// The alternate buffer is the session's, taken once and given back
	// once. Each screen replaces the last rather than stacking, and the
	// terminal a command is handed is the one it would have had.
	p := tea.NewProgram(
		newSession(e, queue, run, in, w),
		tea.WithInput(in),
		tea.WithOutput(w),
		tea.WithAltScreen(),
	)
	_, err := p.Run()
	return err
}
