package command

// FakeTerminal is the fake beside the Terminal port. Zero value: no
// terminal on either stream, every question refused.
type FakeTerminal struct {
	In  bool
	Out bool

	SelectIndex   int
	SelectErr     error
	ConfirmAnswer bool
	ConfirmErr    error
	InputAnswer   string
	InputErr      error

	// Asked records every question rendered, in order.
	Asked []string
}

func (f *FakeTerminal) InteractiveIn() bool  { return f.In }
func (f *FakeTerminal) InteractiveOut() bool { return f.Out }

func (f *FakeTerminal) Select(title string, options []string) (int, error) {
	f.Asked = append(f.Asked, title)
	return f.SelectIndex, f.SelectErr
}

func (f *FakeTerminal) Confirm(question string) (bool, error) {
	f.Asked = append(f.Asked, question)
	return f.ConfirmAnswer, f.ConfirmErr
}

func (f *FakeTerminal) Input(question string) (string, error) {
	f.Asked = append(f.Asked, question)
	return f.InputAnswer, f.InputErr
}

func (f *FakeTerminal) Spin(label string, work func() error) error {
	return work()
}
