// Package term is the production implementation of the frame's
// terminal port, on the Charm stack — decision 0009.
package term

import (
	"io"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	xterm "golang.org/x/term"
)

// Terminal renders questions with huh and answers TTY probes for the
// frame's color and interaction rules. The zero value uses the real
// terminal; In and Out exist so tests can drive the forms headless.
type Terminal struct {
	In  io.Reader
	Out io.Writer
}

// New returns the production terminal.
func New() Terminal { return Terminal{} }

func (Terminal) InteractiveIn() bool  { return xterm.IsTerminal(int(os.Stdin.Fd())) }
func (Terminal) InteractiveOut() bool { return xterm.IsTerminal(int(os.Stdout.Fd())) }

func (t Terminal) run(form *huh.Form) error {
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
		huh.NewSelect[int]().Title(title).Options(opts...).Value(&choice),
	)))
	return choice, err
}

// Confirm renders a yes/no question.
func (t Terminal) Confirm(question string) (bool, error) {
	ok := false
	err := t.run(huh.NewForm(huh.NewGroup(
		huh.NewConfirm().Title(question).Value(&ok),
	)))
	return ok, err
}

// Spin runs work behind a spinner; the work's own error is the verdict.
func (t Terminal) Spin(label string, work func() error) error {
	var werr error
	s := spinner.New().Title(label).Action(func() { werr = work() })
	if t.Out != nil {
		s = s.Output(t.Out)
	}
	if err := s.Run(); err != nil {
		return err
	}
	return werr
}
