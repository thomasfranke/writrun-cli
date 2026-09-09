package screen

import (
	"io"

	tea "github.com/charmbracelet/bubbletea"
)

// Open runs the entry screen, and the queue behind it, until the reader
// chooses a command or leaves. It returns the action the caller is to
// perform; a zero Action is `q` — the screen left and nothing runs.
//
// The screen closes *before* the command runs, so the command owns the
// terminal it asks its questions on. That is why this is a value handed
// back rather than a call made from inside the model: a huh form
// rendering underneath a live Bubble Tea program is two programs
// holding one terminal.
//
// queue is called when the reader opens the queue, and returns the
// lister's output. It is a callback rather than a field because running
// the kit's scripts is the caller's act, not this package's — nothing
// here knows a script path (docs/technical/engineering/coupling.md).
func Open(e Entry, queue func() (string, error), in io.Reader, w io.Writer) (Action, error) {
	for {
		entry, err := run(newEntry(e), in, w)
		if err != nil {
			return Action{}, err
		}
		m, ok := entry.(entryModel)
		if !ok {
			return Action{}, nil
		}
		if !m.queue {
			return m.action, nil
		}

		listing, err := queue()
		if err != nil {
			return Action{}, err
		}
		q, err := run(newModel(Parse(listing)), in, w)
		if err != nil {
			return Action{}, err
		}
		done, ok := q.(model)
		if !ok {
			return Action{}, nil
		}
		// `esc` on the queue is the way back, and the only key that
		// returns nothing and does not end the screen: the reader asked
		// to see the queue, not to leave.
		if done.back {
			continue
		}
		return done.action, nil
	}
}

// run gives each screen the alternate buffer. A screen owns the display
// while it is open and gives it back untouched when it closes: the
// entry screen and the queue replace one another rather than stacking,
// and the command a key dispatches starts on a terminal the screen left
// as it found it. Inline, each screen would leave its last frame in the
// scrollback and the next would draw beneath it.
func run(m tea.Model, in io.Reader, w io.Writer) (tea.Model, error) {
	p := tea.NewProgram(m, tea.WithInput(in), tea.WithOutput(w), tea.WithAltScreen())
	return p.Run()
}
