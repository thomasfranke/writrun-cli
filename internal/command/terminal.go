package command

import "strings"

// Terminal is the interaction port. internal/term carries the
// production implementation on the Charm stack; FakeTerminal, beside
// this interface, is the fake the tests inject.
type Terminal interface {
	// InteractiveIn reports whether stdin is a terminal — the condition
	// for rendering any question.
	InteractiveIn() bool
	// InteractiveOut reports whether stdout is a terminal — the
	// condition for color.
	InteractiveOut() bool
	// Select renders an arrow-key selection and returns the chosen
	// index. Every option carries the sentence shown under the list
	// while it is highlighted, so a reader is told what the row they
	// are about to choose means (spec-0040).
	Select(title string, options []Option) (int, error)
	// Plan renders a composed plan with a cursor on its rows and a
	// footer naming the act the next key performs, and answers whether
	// the reader confirmed it.
	Plan(p Plan) (bool, error)
	// Confirm renders a yes/no question.
	Confirm(question string) (bool, error)
	// Input renders a free-text question and returns what was typed —
	// the only shape of question that is typed rather than navigated
	// (docs/product/rules.md).
	Input(question string) (string, error)
	// Spin runs work behind a spinner while the terminal waits.
	Spin(label string, work func() error) error
}

// Option is one row of a navigated question: what the reader is shown,
// and what the screen says about it while the cursor rests there.
type Option struct {
	Label  string
	Detail string
}

// Plan is a composed plan waiting for a yes: the lines the command
// already writes, and the act `enter` performs on the selected one.
//
// It is the same screen a question is, because it is the same act — a
// reader moving a cursor over rows and answering with one key
// (docs/product/screens/adoption/uninstall.excalidraw, spec-0040).
type Plan struct {
	// Rows are every line the screen shows, in the order the command
	// composed them.
	Rows []PlanRow
	// Printed is what a run with no terminal writes. A nil Printed is
	// the rows' own text, which is the ordinary case; a command whose
	// screen shows a shorter form of what it prints — `author`'s body,
	// counted rather than quoted — gives both, because the printed form
	// is what a person reading a transcript has
	// (docs/product/screens/authoring/author.excalidraw, spec-0040).
	Printed []string
	// Verb is what `enter` does: `refresh`, `remove`, `mark ready`.
	Verb string
	// Question is the same act as a yes/no, asked where nothing
	// navigates.
	Question string
}

// PlanRow is one line of a plan.
type PlanRow struct {
	// Text carries its own indent, because the command composed it.
	Text string
	// Detail is the sentence the screen shows about this row.
	Detail string
	// Selects says the cursor stops here.
	Selects bool
}

// Lines are what a run with no terminal prints — the same bytes, in
// the same order, as before there was a screen.
func (p Plan) Lines() []string {
	if p.Printed != nil {
		return p.Printed
	}
	out := make([]string, 0, len(p.Rows))
	for _, r := range p.Rows {
		out = append(out, r.Text)
	}
	return out
}

// Written is a block a command already composes, split into the lines
// printing them back would write. It is how a command that shows one
// form and prints another keeps the printed one byte for byte.
func Written(block string) []string {
	lines := strings.Split(block, "\n")
	if n := len(lines); n > 0 && lines[n-1] == "" {
		lines = lines[:n-1]
	}
	return lines
}
