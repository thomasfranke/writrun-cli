package screen

import (
	"io"

	tea "github.com/charmbracelet/bubbletea"
)

// Open runs the screen over the lister's output and returns the action
// the user chose. The program owns the terminal only while it is open:
// it has quit by the time this returns, which is what lets the caller
// run a command that asks its own questions.
//
// in and out are the terminal; the suite passes its own so the model is
// driven without one.
func Open(out string, in io.Reader, w io.Writer) (Action, error) {
	m := newModel(Parse(out))
	p := tea.NewProgram(m, tea.WithInput(in), tea.WithOutput(w))
	final, err := p.Run()
	if err != nil {
		return Action{}, err
	}
	done, ok := final.(model)
	if !ok {
		return Action{}, nil
	}
	return done.action, nil
}
