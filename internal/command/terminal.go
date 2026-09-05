package command

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
	// Select renders an arrow-key selection and returns the chosen index.
	Select(title string, options []string) (int, error)
	// Confirm renders a yes/no question.
	Confirm(question string) (bool, error)
	// Input renders a free-text question and returns what was typed —
	// the only shape of question that is typed rather than navigated
	// (docs/product/rules.md).
	Input(question string) (string, error)
	// Spin runs work behind a spinner while the terminal waits.
	Spin(label string, work func() error) error
}
