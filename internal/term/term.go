// Package term is the production implementation of the frame's
// terminal port, on the Charm stack — decision 0009.
package term

import (
	"context"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	xterm "golang.org/x/term"
)

// Terminal renders questions with huh and answers TTY probes for the
// frame's color and interaction rules. The zero value uses the real
// terminal; In and Out exist so tests can drive the forms headless —
// an override stands in for the terminal, so the probes count it as
// one and the guarded flows stay exercisable end to end.
type Terminal struct {
	In  io.Reader
	Out io.Writer
}

// New returns the production terminal.
func New() Terminal { return Terminal{} }

func (t Terminal) InteractiveIn() bool {
	if t.In != nil {
		return true
	}
	return xterm.IsTerminal(int(os.Stdin.Fd()))
}

func (t Terminal) InteractiveOut() bool {
	if t.Out != nil {
		return true
	}
	return xterm.IsTerminal(int(os.Stdout.Fd()))
}

// keys are huh's, with the way out given a name.
//
// huh binds leaving to `ctrl+c` and gives that binding no help text, so
// a form's footer reads `enter submit` and stops there: it says how to
// go forward and never how to go back. A reader who does not want to
// answer has no key they can see, which is how `report` came to be
// reported as having no way out — it had one, spelled nowhere.
//
// `esc` joins `ctrl+c` because it is the key the screens already use
// for going back, and naming it makes the two agree
// (docs/product/screens/README.md). huh's footer renders the field's
// own bindings and not the form's, so binding it is not enough to show
// it — `wayOut` is where it is said.
func keys() *huh.KeyMap {
	k := huh.NewDefaultKeyMap()
	k.Quit = key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc", "cancel"),
	)
	return k
}

// wayOut is the line under every question. huh will not put the
// cancel key in its footer, so it is written where huh does render:
// the field's description, one line under the title.
const wayOut = "esc cancels"

func (t Terminal) run(form *huh.Form) error {
	form = form.WithKeyMap(keys())
	if t.In != nil {
		form = form.WithInput(t.In)
	}
	if t.Out != nil {
		form = form.WithOutput(t.Out)
	}
	return form.Run()
}

// Select renders an arrow-key selection and returns the chosen index.
func (t Terminal) Select(title string, options []string) (int, error) {
	choice := 0
	opts := make([]huh.Option[int], len(options))
	for i, o := range options {
		opts[i] = huh.NewOption(o, i)
	}
	err := t.run(huh.NewForm(huh.NewGroup(
		huh.NewSelect[int]().Title(title).Description(wayOut).Options(opts...).Value(&choice),
	)))
	return choice, err
}

// Confirm renders a yes/no question.
func (t Terminal) Confirm(question string) (bool, error) {
	ok := false
	err := t.run(huh.NewForm(huh.NewGroup(
		huh.NewConfirm().Title(question).Description(wayOut).Value(&ok),
	)))
	return ok, err
}

// Input renders a free-text question and returns what was typed.
func (t Terminal) Input(question string) (string, error) {
	answer := ""
	err := t.run(huh.NewForm(huh.NewGroup(
		huh.NewInput().Title(question).Description(wayOut).Value(&answer),
	)))
	return answer, err
}

// Spin runs work behind a spinner; the work's own error is the
// verdict. Where stdout is no terminal the spinner would be escape
// sequences in machine-read output, so only the work runs. The work is
// never abandoned: an interrupted spinner still waits for it, and only
// then does the interruption surface.
func (t Terminal) Spin(label string, work func() error) error {
	if !t.InteractiveOut() {
		return work()
	}
	var werr error
	finished := make(chan struct{})
	go func() {
		werr = work()
		close(finished)
	}()
	s := spinner.New().Title(label).ActionWithErr(func(ctx context.Context) error {
		select {
		case <-finished:
		case <-ctx.Done():
		}
		return nil
	})
	if t.Out != nil {
		s = s.Output(t.Out)
	}
	runErr := s.Run()
	<-finished
	if werr != nil {
		return werr
	}
	return runErr
}
