package screen

import (
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea"
)

// Runner runs the command a key chose. It is a callback because running
// a command is the caller's act, not this package's — nothing here
// knows a command's behaviour, only its name
// (docs/technical/engineering/coupling.md).
//
// It answers nothing, because the command already has: whatever it
// found, it printed on the terminal this released for it, and a second
// account of the same thing is one the reader has to reconcile.
type Runner func(Action)

// session is the whole screen: the entry screen and the queue as two
// states of one program rather than two programs taking turns.
//
// # One program, because a keyboard has one reader
//
// The screens used to be a program each, run one after another, and a
// command ran after the last of them had quit. That reads as correct
// and is not: Bubble Tea reads the terminal on a goroutine it cannot
// always stop at the moment it quits, so a keypress meant for the next
// thing could still be taken by the screen that had already left. The
// cost was not a misdrawn frame — `finish`'s confirmation answered
// itself with the `enter` that had chosen `finish`, and a command that
// opens a pull request must never be told yes by the key that named it.
//
// So there is one program for the session, and a command runs through
// `tea.Exec`, which pauses this one and gives the terminal back before
// the command asks anything. Two readers never exist at once, because
// there is only ever one program.
type session struct {
	entry   entryModel
	queue   model
	inQueue bool

	// listing reads the queue when the reader opens it, and again after
	// a command, because a command is the thing that changes it.
	listing func() (string, error)
	run     Runner

	// in and out are the terminal's, handed to a command through the
	// released terminal rather than reopened here.
	in  io.Reader
	out io.Writer

	// err is the last thing that went wrong reading the queue, shown in
	// place of the rows rather than ending the session: a queue that
	// cannot be read is not a reason to close the door.
	err error
}

// ranMsg says a command finished and the screen is coming back.
type ranMsg struct{}

// dispatch is one command, run while the program is paused.
//
// It waits for a line before returning, because the alternate screen
// comes back the moment it does and takes the command's output with it.
// A reader who has to ask what just happened was shown nothing.
//
// The three setters are the interface's and are deliberately empty: the
// command writes to the process's own streams, which is exactly what
// the released terminal is.
type dispatch struct {
	run    Runner
	action Action
	in     io.Reader
	out    io.Writer
}

// The alternate screen, entered and left by hand.
//
// `tea.Exec` gives the terminal back in the state it found it — the
// normal buffer — so a command left inline would print into the
// scrollback and the screen would come back over the top of it. A
// command is a screen like the others: it takes the whole terminal and
// gives it back untouched, and the reader's scrollback is not where its
// answer lands.
//
// Leaving before returning is what keeps the two in step: Bubble Tea
// restores from the normal buffer, which is where it released from.
//
// The cost is that output taller than the window scrolls off the top
// and there is no scrollback to reach it — an alternate buffer keeps
// none. `writrun <command>` on its own is where a long answer is read
// at leisure, and a screen that pages its own output is the fix if this
// bites.
const (
	altOn  = "\033[?1049h\033[H\033[2J"
	altOff = "\033[?1049l"
)

func (d dispatch) Run() error {
	fmt.Fprint(d.out, altOn)
	// What is running, said before it runs. A command that reaches the
	// forge takes seconds, and the screen it replaced is gone by then —
	// an empty terminal is indistinguishable from a hung one.
	//
	// It does not animate, and that is the terminal's arithmetic rather
	// than a preference: from here until the command returns, the
	// command owns the terminal, and a spinner would be a second writer
	// interleaving with its output. The same reason a screen cannot
	// stay open behind a question.
	fmt.Fprintf(d.out, "  running %s…\n\n", d.action.Command)
	d.run(d.action)
	fmt.Fprint(d.out, "\n— enter to return to the screen —")
	waitForLine(d.in)
	fmt.Fprint(d.out, altOff)
	return nil
}

// waitForLine consumes exactly up to and including one return.
//
// Both `\n` and `\r` end it, because which one arrives is the terminal
// mode's answer and not the reader's. Enter sends `\r`; a terminal in
// canonical mode translates it to `\n` and a terminal in raw mode does
// not. A command that asks a question puts the terminal in raw mode to
// do it — so waiting on `\n` alone hangs after exactly those commands,
// and the way back reads as absent when it is simply deaf.
//
// A buffered reader would be the obvious way and is the wrong one: it
// reads ahead, and what it read ahead of the return is dropped with it
// — keys the reader typed for the screen, swallowed on the way back to
// it. One byte at a time takes what this is owed and leaves the rest
// where the screen will find it.
func waitForLine(r io.Reader) {
	var b [1]byte
	for {
		n, err := r.Read(b[:])
		if err != nil {
			return
		}
		if n > 0 && (b[0] == '\n' || b[0] == '\r') {
			return
		}
	}
}

func (d dispatch) SetStdin(io.Reader)  {}
func (d dispatch) SetStdout(io.Writer) {}
func (d dispatch) SetStderr(io.Writer) {}

func newSession(e Entry, listing func() (string, error), run Runner, in io.Reader, out io.Writer) session {
	return session{entry: newEntry(e), listing: listing, run: run, in: in, out: out}
}

func (s session) Init() tea.Cmd { return nil }

func (s session) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Both screens are sized, not just the one showing: the other is
		// one keypress away and should not have to be resized first.
		e, _ := s.entry.Update(msg)
		s.entry = e.(entryModel)
		q, _ := s.queue.Update(msg)
		s.queue = q.(model)
		return s, nil
	case ranMsg:
		// The command's own words have been read already. What is left
		// is the queue, which the command may have changed.
		if s.inQueue {
			s.loadQueue()
		}
		return s, nil
	}
	if s.inQueue {
		return s.updateQueue(msg)
	}
	return s.updateEntry(msg)
}

// updateEntry delegates to the entry screen and then reads what it
// decided. The sub-model still ends itself with `tea.Quit` — that is
// its own behaviour, and correct when it runs alone — so the cmd is
// dropped here and this decides what ends instead.
func (s session) updateEntry(msg tea.Msg) (tea.Model, tea.Cmd) {
	e, _ := s.entry.Update(msg)
	s.entry = e.(entryModel)

	switch {
	case s.entry.left:
		return s, tea.Quit
	case s.entry.queue:
		s.entry.queue = false
		s.inQueue = true
		s.loadQueue()
		return s, nil
	case s.entry.action != (Action{}):
		a := s.entry.action
		s.entry.action = Action{}
		return s, s.exec(a)
	}
	return s, nil
}

func (s session) updateQueue(msg tea.Msg) (tea.Model, tea.Cmd) {
	q, _ := s.queue.Update(msg)
	s.queue = q.(model)

	switch {
	case s.queue.left:
		return s, tea.Quit
	case s.queue.back:
		s.queue.back = false
		s.inQueue = false
		return s, nil
	case s.queue.action != (Action{}):
		a := s.queue.action
		s.queue.action = Action{}
		return s, s.exec(a)
	}
	return s, nil
}

func (s *session) exec(a Action) tea.Cmd {
	d := dispatch{run: s.run, action: a, in: s.in, out: s.out}
	return tea.Exec(d, func(error) tea.Msg { return ranMsg{} })
}

// loadQueue reads the queue and keeps the reader's place where it can.
func (s *session) loadQueue() {
	listing, err := s.listing()
	if err != nil {
		s.err = err
		s.queue = newModel(nil)
		return
	}
	s.err = nil
	cursor, top, height := s.queue.cursor, s.queue.top, s.queue.height
	s.queue = newModel(Parse(listing))
	s.queue.height = height
	// The rows may be fewer than they were, so the old place is only
	// kept where it still exists.
	if cursor >= 0 && cursor < len(s.queue.rows) && s.queue.rows[cursor].Selectable() {
		s.queue.cursor, s.queue.top = cursor, top
		s.queue.scroll()
	}
}

func (s session) View() string {
	if s.inQueue {
		if s.err != nil {
			return fmt.Sprintf(" the queue could not be read: %v\n\n esc back · q quit\n", s.err)
		}
		return s.queue.View()
	}
	return s.entry.View()
}
