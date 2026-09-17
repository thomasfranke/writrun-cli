package command

import (
	"errors"
	"fmt"
	"io"
)

// ErrDeclined is the user's no. AskConfirm returns it so a command
// cannot forget the answer: the error travels up and the frame turns
// it into a non-zero exit having changed nothing (spec-0001).
var ErrDeclined = errors.New("declined")

// Ctx is what the frame hands a running command: the resolved
// repository, the streams, the interaction helpers, and the flags every
// command shares.
type Ctx struct {
	Stdout io.Writer
	Stderr io.Writer
	// Stdin is the third stream, and only a command that opens a screen
	// of its own needs it: a question is asked through Terminal, which
	// holds its own reader. Nil where the frame was built without one,
	// which is every test that asks nothing.
	Stdin io.Reader

	Terminal Terminal

	// Root is the git toplevel; empty when the need was NeedAny and no
	// repository was found. Adopted says whether `.writrun/` is there.
	Root    string
	Adopted bool

	// Version is the client's own version, the string `--version`
	// prints. It is handed down rather than read again so that a
	// command naming the binary and `--version` naming it cannot
	// disagree about what is running (spec-0043).
	Version string

	// Yes is --yes: every question already answered.
	Yes bool
	// Color is the reporting rule already decided: stdout is a
	// terminal, NO_COLOR is unset, --no-color was not given.
	Color bool

	// Again runs this binary again, as a process of its own, with the
	// arguments given — the frame's own flags already prepended, so a
	// caller here names the command and nothing else.
	//
	// It is for a command that opens a screen and then has to hand the
	// terminal to a question: a question is a terminal program, and a
	// terminal program's input reader can outlive it, taking the next
	// key for something that has ended (decision 0015, report-0040). A
	// screen whose question is a process has no such reader to leave
	// behind.
	//
	// nil is a frame built without the port. The caller then asks in
	// this process, which is what it did before the port existed.
	Again func(args []string) error
}

// AskConfirm is the confirmation flow: --yes answers it, a terminal
// asks it, and anything else aborts naming the flag — a question never
// hangs. nil is the go-ahead; a decline is ErrDeclined, so proceeding
// takes an explicit yes and a forgotten check cannot exit 0.
func (c *Ctx) AskConfirm(question string) error {
	if c.Yes {
		return nil
	}
	if !c.Terminal.InteractiveIn() {
		return fmt.Errorf("no terminal to ask %q — pass --yes", question)
	}
	ok, err := c.Terminal.Confirm(question)
	if err != nil {
		return err
	}
	if !ok {
		return ErrDeclined
	}
	return nil
}

// AskInput is the free-text flow: a flag answers it, a terminal types
// it, and anything else aborts naming the flag that would have
// answered. --yes does not answer it — a value nobody wrote is not an
// answer a flag can stand in for.
func (c *Ctx) AskInput(question string, preset string, flag string) (string, error) {
	if preset != "" {
		return preset, nil
	}
	if !c.Terminal.InteractiveIn() {
		return "", fmt.Errorf("no terminal to ask %q — pass %s", question, flag)
	}
	return c.Terminal.Input(question)
}

// AskSelect is the selection flow: a preset answers it, a terminal
// renders it, and anything else aborts naming the flag that would have
// answered.
//
// Every option carries its own sentence, shown under the list while the
// cursor rests on it. A question that names its options and explains
// none of them asks a reader to choose by the shape of a word
// (spec-0040).
func (c *Ctx) AskSelect(title string, options []Option, preset string, flag string) (int, error) {
	if preset != "" {
		for i, o := range options {
			if o.Label == preset {
				return i, nil
			}
		}
		return -1, fmt.Errorf("%s is not one of the options for %s", preset, flag)
	}
	if !c.Terminal.InteractiveIn() {
		return -1, fmt.Errorf("no terminal to ask %q — pass %s", title, flag)
	}
	return c.Terminal.Select(title, options)
}

// AskPlan is the confirmation flow for a composed plan: a terminal
// navigates it, and everything else prints it and answers the question
// beside it.
//
// The printed form is the plan's own rows, so what a run with no
// terminal writes is what it wrote before there was a screen — the
// screen is where the rows are read, never where they are composed
// (spec-0040).
func (c *Ctx) AskPlan(p Plan) error {
	if c.Yes || !c.Terminal.InteractiveIn() || !c.Terminal.InteractiveOut() {
		for _, line := range p.Lines() {
			fmt.Fprintln(c.Stdout, line)
		}
		return c.AskConfirm(p.Question)
	}
	ok, err := c.Terminal.Plan(p)
	if err != nil {
		return err
	}
	if !ok {
		return ErrDeclined
	}
	return nil
}
