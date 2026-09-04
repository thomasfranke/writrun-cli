// Package term is the production implementation of the frame's
// terminal port, on the Charm stack — decision 0009.
package term

import (
	"context"
	"io"
	"os"

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
