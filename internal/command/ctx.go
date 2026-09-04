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

	Terminal Terminal

	// Root is the git toplevel; empty when the need was NeedAny and no
	// repository was found. Adopted says whether `.writrun/` is there.
	Root    string
	Adopted bool

	// Yes is --yes: every question already answered.
	Yes bool
	// Color is the reporting rule already decided: stdout is a
	// terminal, NO_COLOR is unset, --no-color was not given.
	Color bool
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

// AskSelect is the selection flow: a preset answers it, a terminal
// renders it, and anything else aborts naming the flag that would have
// answered.
func (c *Ctx) AskSelect(title string, options []string, preset string, flag string) (int, error) {
	if preset != "" {
		for i, o := range options {
			if o == preset {
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
