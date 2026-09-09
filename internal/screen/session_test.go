package screen

import (
	"errors"
	"io"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// The queue a case reads is the lister's own shape (screen_test.go), so
// a row here is selectable for the same reason a real one is.

func newTestSession(t *testing.T, read func() (string, error)) session {
	t.Helper()
	if read == nil {
		read = func() (string, error) { return listing, nil }
	}
	return newSession(sample(), read, func(Action, io.Writer) {}, strings.NewReader(""), &strings.Builder{})
}

func step(s session, msgs ...tea.Msg) (session, tea.Cmd) {
	var cmd tea.Cmd
	for _, m := range msgs {
		var out tea.Model
		out, cmd = s.Update(m)
		s = out.(session)
	}
	return s, cmd
}

var (
	esc  = tea.KeyMsg{Type: tea.KeyEsc}
	quit = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
)

// `enter` on `list` opens the queue and reads it; `esc` goes back to the
// entry screen. Neither ends the session.
func TestTheQueueOpensAndCloses(t *testing.T) {
	reads := 0
	s, cmd := step(newTestSession(t, func() (string, error) {
		reads++
		return listing, nil
	}), enter)
	if s.where != atQueue {
		t.Fatal("enter on list did not open the queue")
	}
	if reads != 1 {
		t.Errorf("the queue was read %d times, want once", reads)
	}
	if cmd != nil {
		t.Error("opening the queue asked for work; it is a state, not an act")
	}
	if !strings.Contains(s.View(), "task-0020") {
		t.Error("the queue's rows are not what the session shows")
	}

	s, cmd = step(s, esc)
	if s.where == atQueue {
		t.Error("esc did not return to the entry screen")
	}
	if cmd != nil {
		t.Error("esc ended the session; it is the way back, not the way out")
	}
	if !strings.Contains(s.View(), "enter run") {
		t.Error("the entry screen is not what the session shows after esc")
	}
}

// A command chosen on either screen is asked for and not performed here:
// the model decides, the program runs it.
func TestChoosingACommandAsksForItAndClearsTheChoice(t *testing.T) {
	// On the entry screen: down moves off `list` onto a command.
	s, cmd := step(newTestSession(t, nil), down, enter)
	if cmd == nil {
		t.Fatal("choosing a command asked for nothing")
	}
	if s.entry.action != (Action{}) {
		t.Errorf("the choice was left set: %+v — the next key would run it again", s.entry.action)
	}

	// On the queue: enter takes the selected id.
	s, cmd = step(newTestSession(t, nil), enter, enter)
	if cmd == nil {
		t.Fatal("choosing a task asked for nothing")
	}
	if s.queue.action != (Action{}) {
		t.Errorf("the choice was left set: %+v", s.queue.action)
	}
}

// q ends the session from either screen, and nothing else does.
func TestQLeavesFromEitherScreen(t *testing.T) {
	if _, cmd := step(newTestSession(t, nil), quit); cmd == nil {
		t.Error("q on the entry screen did not end the session")
	}
	if _, cmd := step(newTestSession(t, nil), enter, quit); cmd == nil {
		t.Error("q on the queue did not end the session")
	}
}

// A command is the thing that changes the queue, so the queue is read
// again when one finishes — but only when the reader is looking at it.
func TestTheQueueIsReReadAfterACommand(t *testing.T) {
	reads := 0
	read := func() (string, error) {
		reads++
		return listing, nil
	}
	s, _ := step(newTestSession(t, read), enter) // into the queue: one read
	step(s, ranMsg{})
	if reads != 2 {
		t.Errorf("the queue was read %d times, want a second after the command", reads)
	}

	reads = 0
	step(newTestSession(t, read), ranMsg{}) // on the entry screen
	if reads != 0 {
		t.Errorf("the queue was read %d times by a screen not showing it", reads)
	}
}

// A queue that cannot be read is said so and the session stays open: a
// lister that failed is not a reason to close the door.
func TestAQueueThatCannotBeReadIsNamed(t *testing.T) {
	s, cmd := step(newTestSession(t, func() (string, error) {
		return "", errors.New("the lister refused")
	}), enter)
	if cmd != nil {
		t.Error("a failed read ended the session")
	}
	if !strings.Contains(s.View(), "the lister refused") {
		t.Errorf("the cause was not shown:\n%s", s.View())
	}
}

// Both screens are sized, not just the one showing: the other is one
// keypress away and should not have to be resized first.
func TestBothScreensAreSized(t *testing.T) {
	s, _ := step(newTestSession(t, nil), tea.WindowSizeMsg{Width: 80, Height: 24})
	if s.entry.height == 0 || s.queue.height == 0 {
		t.Errorf("entry = %d, queue = %d; both want the window's height",
			s.entry.height, s.queue.height)
	}
}

// waitForLine takes what it is owed and no more.
//
// This is the defect it was written for: a buffered reader would take
// the whole string, answer with the first line, and drop the rest —
// which is keys the reader typed for the screen, swallowed on the way
// back to it.
func TestWaitForLineLeavesWhatComesAfterIt(t *testing.T) {
	r := strings.NewReader("\nqjk")
	waitForLine(r)
	rest, err := readAll(r)
	if err != nil {
		t.Fatalf("reading what was left = %v", err)
	}
	if rest != "qjk" {
		t.Errorf("what was left = %q, want the three keys typed after the line", rest)
	}
}

// A carriage return is a return too.
//
// Which byte Enter delivers is the terminal mode's answer, not the
// reader's: canonical mode translates `\r` to `\n` and raw mode hands
// `\r` over as typed. A command that asks a question leaves the
// terminal raw to do it, so waiting on `\n` alone hangs after exactly
// those commands — `report` was reported as having no way back, and it
// had one that could not hear.
func TestACarriageReturnEndsTheWaitToo(t *testing.T) {
	r := strings.NewReader("\rqjk")
	waitForLine(r)
	rest, _ := readAll(r)
	if rest != "qjk" {
		t.Errorf("what was left = %q, want the keys typed after the return", rest)
	}
}

// A reader that ends without a line does not wait forever.
func TestWaitForLineStopsAtTheEnd(t *testing.T) {
	waitForLine(strings.NewReader("no line here"))
}

func readAll(r *strings.Reader) (string, error) {
	var b strings.Builder
	buf := make([]byte, 8)
	for {
		n, err := r.Read(buf)
		b.Write(buf[:n])
		if err != nil {
			return b.String(), nil
		}
	}
}

// The setters are the ExecCommand interface's and are deliberately
// empty: a command writes to the process's own streams, which is what
// the released terminal is.
func TestTheDispatchTakesNoStreamsOfItsOwn(t *testing.T) {
	d := dispatch{}
	d.SetStdin(nil)
	d.SetStdout(nil)
	d.SetStderr(nil)
}

// The command is run with the action it was given, and the reader is
// asked before the screen takes the terminal back.
func TestTheDispatchRunsTheCommandThenWaits(t *testing.T) {
	var got Action
	var out strings.Builder
	d := dispatch{
		run:    func(a Action, _ io.Writer) { got = a },
		action: Action{Command: "take", Arg: "task-0021"},
		in:     strings.NewReader("\n"),
		out:    &out,
	}
	if err := d.Run(); err != nil {
		t.Fatalf("Run = %v", err)
	}
	if got != (Action{Command: "take", Arg: "task-0021"}) {
		t.Errorf("ran %+v, want the action it was given", got)
	}
	if !strings.Contains(out.String(), "enter to return") {
		t.Errorf("the way back was not offered: %q", out.String())
	}
}

// A command's own declaration picks how it is run.
//
// One that asks nothing is captured — the program keeps the terminal,
// which is what buys the spinner and the scrolling. One that asks is
// handed the terminal, because a captured question waits on a reader
// who cannot see it.
func TestTheDeclarationPicksHowACommandRuns(t *testing.T) {
	// `doctor` declares it asks nothing (entry_test.go's sample).
	s := newTestSession(t, nil)
	s.entry.cursor = 0
	for s.entry.selected() != "doctor" {
		s.entry.move(1)
		if s.entry.cursor < 0 {
			t.Fatal("doctor is not in the sample")
		}
	}
	out, cmd := s.Update(enter)
	s = out.(session)
	if s.where != atRunning {
		t.Errorf("where = %v, want the spinner while a captured command runs", s.where)
	}
	if cmd == nil {
		t.Error("nothing was asked for")
	}
	if !strings.Contains(s.View(), "running doctor") {
		t.Errorf("the spinner does not say what it waits for:\n%s", s.View())
	}

	// `take` asks, so it takes the terminal and the screen stays put.
	s2, _ := step(newTestSession(t, nil), down, enter)
	if s2.where == atRunning {
		t.Error("a command that asks was captured")
	}
}

// The answer becomes a screen, and going back from it lands where the
// command was chosen.
func TestTheAnswerBecomesAScreenAndHandsBack(t *testing.T) {
	s, _ := step(newTestSession(t, nil), tea.WindowSizeMsg{Width: 80, Height: 24})
	s, _ = step(s, wroteMsg{command: "doctor", output: "all clear\n"})
	if s.where != atPager {
		t.Fatalf("where = %v, want the answer's own screen", s.where)
	}
	if !strings.Contains(s.View(), "all clear") {
		t.Errorf("the answer is not what the screen shows:\n%s", s.View())
	}

	s, cmd := step(s, esc)
	if s.where != atEntry {
		t.Errorf("where = %v, want the screen it was chosen from", s.where)
	}
	if cmd != nil {
		t.Error("leaving the answer ended the session")
	}
}

// The spinner turns while it waits, and stops when there is nothing to
// wait for: a clock with no reason to tick is a program that never
// sleeps.
func TestTheSpinnerTurnsOnlyWhileSomethingRuns(t *testing.T) {
	s := newTestSession(t, nil)
	s.where = atRunning
	s.running = "doctor"
	before := s.frame
	s, cmd := step(s, tickMsg{})
	if s.frame == before {
		t.Error("the spinner did not turn")
	}
	if cmd == nil {
		t.Error("the spinner stopped while the command was still running")
	}

	idle := newTestSession(t, nil)
	if _, cmd := step(idle, tickMsg{}); cmd != nil {
		t.Error("the spinner kept ticking with nothing to wait for")
	}
}

// A key pressed while a command runs does nothing: the command is
// already writing, and a half-run command abandoned is a worse answer
// than a slow one.
func TestKeysAreIgnoredWhileACommandRuns(t *testing.T) {
	s := newTestSession(t, nil)
	s.where = atRunning
	s.running = "doctor"
	out, cmd := step(s, quit)
	if cmd != nil || out.where != atRunning {
		t.Error("a key interrupted a running command")
	}
}
