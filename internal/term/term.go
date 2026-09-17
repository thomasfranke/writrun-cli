// Package term is the production implementation of the frame's
// terminal port, on the Charm stack — decision 0009.
package term

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	xterm "golang.org/x/term"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/screen"
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
//
// It is the screen component rather than a form, because the drawings
// put the highlighted option's own sentence under the list and a form
// puts its one description above it
// (docs/product/screens/tasks/take.excalidraw, spec-0040).
func (t Terminal) Select(title string, options []command.Option) (int, error) {
	rows := make([]screen.Line, 0, len(options))
	for _, o := range options {
		rows = append(rows, screen.Line{Text: "  " + o.Label, Detail: o.Detail, Selects: true})
	}
	return t.choose(screen.Choice{Title: title, Rows: rows, Verb: "choose"})
}

// Plan renders a composed plan and answers whether the reader confirmed
// it. `esc` is the decline, and it is the same key the questions cancel
// on (docs/product/screens/README.md).
func (t Terminal) Plan(p command.Plan) (bool, error) {
	rows := make([]screen.Line, 0, len(p.Rows))
	for _, r := range p.Rows {
		rows = append(rows, screen.Line{Text: r.Text, Detail: r.Detail, Selects: r.Selects})
	}
	i, err := t.choose(screen.Choice{Rows: rows, Verb: p.Verb})
	if errors.Is(err, command.ErrDeclined) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return i >= 0, nil
}

// choose runs the component on this terminal, and reads a cancellation
// as the decline it is — the frame already knows that word, and a
// second name for it would be a second answer (command.ErrDeclined).
func (t Terminal) choose(c screen.Choice) (int, error) {
	in, out := io.Reader(os.Stdin), io.Writer(os.Stdout)
	if t.In != nil {
		in = t.In
	}
	if t.Out != nil {
		out = t.Out
	}
	i, err := screen.Choose(c, in, out)
	if errors.Is(err, screen.ErrCancelled) {
		return -1, command.ErrDeclined
	}
	return i, err
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

// PlanLines and QuestionLines are what the screen draws for a plan and
// for a question, without a terminal to draw it on.
//
// They are here because the shapes are here: a command composes a
// `command.Plan`, this package turns it into the screen's own, and the
// suite holds that drawing to the frame it is checked against. A
// command package reading the screen's types directly would be a second
// conversion (docs/technical/engineering/boundaries.md).
func PlanLines(p command.Plan, width, on int) []string {
	rows := make([]screen.Line, 0, len(p.Rows))
	for _, r := range p.Rows {
		rows = append(rows, screen.Line{Text: r.Text, Detail: r.Detail, Selects: r.Selects})
	}
	return screen.Draw(screen.Choice{Rows: rows, Verb: p.Verb}, width, on)
}

func QuestionLines(title string, options []command.Option, width, on int) []string {
	rows := make([]screen.Line, 0, len(options))
	for _, o := range options {
		rows = append(rows, screen.Line{Text: "  " + o.Label, Detail: o.Detail, Selects: true})
	}
	return screen.Draw(screen.Choice{Title: title, Rows: rows, Verb: "choose"}, width, on)
}
