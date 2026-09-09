package screen

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Runner runs the command a key chose. It is a callback because running
// a command is the caller's act, not this package's — nothing here
// knows a command's behaviour, only its name
// (docs/technical/engineering/coupling.md).
//
// out is where the command writes. A nil out hands it the process's own
// streams — the terminal, released, which is what a command that asks
// questions needs. A non-nil out captures instead, which only a command
// that declares it asks nothing may be given.
//
// It answers nothing, because the command already has: whatever it
// found, it wrote, and a second account of the same thing is one the
// reader has to reconcile.
type Runner func(a Action, out io.Writer)

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
	entry entryModel
	queue model
	pager pager
	// where says which of the screens is showing.
	where where

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

	// running is what the spinner is spinning for, and frame is where
	// it has got to. pending is the command it will start on the first
	// tick — see exec for why it does not start sooner.
	running string
	frame   int
	pending *Action

	// cameFromQueue says which screen the pager will hand back to.
	cameFromQueue bool
	// height is the window's, kept so a pager built mid-session is born
	// the right size rather than waiting for the next resize.
	height int
}

// where is the screen the reader is looking at.
type where int

const (
	atEntry where = iota
	atQueue
	atRunning
	atPager
)

// ranMsg says a command that owned the terminal finished.
type ranMsg struct{}

// wroteMsg carries a captured command's whole output back to the
// program that never gave up the terminal for it.
type wroteMsg struct {
	command string
	output  string
}

// tickMsg turns the spinner. It exists only while a captured command is
// running: nothing else animates, so nothing else has a clock.
type tickMsg struct{}

// spinner is the frames and the pace. A captured command is the only
// thing this screen waits on, and waiting with no sign of life is
// indistinguishable from hanging — which is what the maintainer saw
// before there was one.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

const spinnerPace = 90 * time.Millisecond

func tick() tea.Cmd {
	return tea.Tick(spinnerPace, func(time.Time) tea.Msg { return tickMsg{} })
}

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
	// label is what the screen says it is running, and run is the
	// running. A func rather than an Action because more than one
	// screen hands the terminal over — the session for a command, the
	// settings screen for a change — and the discipline of doing so is
	// the thing worth having in one place.
	label string
	run   func()
	in    io.Reader
	out   io.Writer
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
	fmt.Fprintf(d.out, "  running %s…\n\n", d.label)
	d.run()
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
		pg, _ := s.pager.Update(msg)
		s.pager = pg.(pager)
		s.height = s.pager.height
		return s, nil
	case ranMsg:
		// A command that owned the terminal has been read there already,
		// and the screen it was chosen from is still the one showing —
		// this path never left it. What is left is the queue, which the
		// command may have changed.
		if s.where == atQueue {
			s.loadQueue()
		}
		return s, nil
	case wroteMsg:
		// A captured command answers into a screen of its own.
		s.where = atPager
		s.running = ""
		s.pager = newPager(msg.command, msg.output)
		s.pager.height = s.height
		s.pager.clamp()
		return s, nil
	case tickMsg:
		if s.where != atRunning {
			return s, nil
		}
		// The first tick is where a captured command starts, one frame
		// after the screen said it would. A command started in the same
		// breath can answer before that frame is ever painted — a fast
		// one did, on a machine quicker than the one this was written
		// on — and then the screen never says what it is doing, which
		// is the whole reason the frame exists.
		if a := s.pending; a != nil {
			s.pending = nil
			return s, tea.Batch(tick(), s.capture(*a))
		}
		s.frame = (s.frame + 1) % len(spinnerFrames)
		return s, tick()
	}
	switch s.where {
	case atRunning:
		// Nothing answers a key here. The command is not the screen's
		// to interrupt: it is already writing, and a half-run command
		// abandoned is a worse answer than a slow one.
		return s, nil
	case atPager:
		return s.updatePager(msg)
	case atQueue:
		return s.updateQueue(msg)
	}
	return s.updateEntry(msg)
}

func (s session) updatePager(msg tea.Msg) (tea.Model, tea.Cmd) {
	p, _ := s.pager.Update(msg)
	s.pager = p.(pager)
	switch {
	case s.pager.left:
		return s, tea.Quit
	case s.pager.back:
		s.pager.back = false
		// Back to where the command was chosen, and the queue re-read
		// where that is what is behind it.
		if s.cameFromQueue {
			s.where = atQueue
			s.loadQueue()
		} else {
			s.where = atEntry
		}
		return s, nil
	}
	return s, nil
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
		s.where = atQueue
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
		s.where = atEntry
		return s, nil
	case s.queue.action != (Action{}):
		a := s.queue.action
		s.queue.action = Action{}
		return s, s.exec(a)
	}
	return s, nil
}

// exec runs a command one of two ways, and the command's own
// declaration picks which.
//
// A command that asks nothing is captured: the program keeps the
// terminal, so it can spin while it waits and page the answer
// afterwards. A command that asks is handed the terminal through
// `tea.Exec`, because a captured question waits on a reader who cannot
// see it — and because two programs must never hold one keyboard.
func (s *session) exec(a Action) tea.Cmd {
	s.cameFromQueue = s.where == atQueue
	if s.asksNothing(a.Command) {
		s.where = atRunning
		s.running = a.Command
		s.frame = 0
		s.pending = &a
		return tick()
	}
	run := s.run
	d := dispatch{
		label: a.Command,
		run:   func() { run(a, nil) },
		in:    s.in,
		out:   s.out,
	}
	return tea.Exec(d, func(error) tea.Msg { return ranMsg{} })
}

// capture runs the command off the update loop and answers with
// everything it wrote.
//
// The buffer is guarded because the command may write from more than
// one goroutine of its own — a script runner streaming stdout and
// stderr is two — and a torn line is a line the reader cannot trust.
func (s *session) capture(a Action) tea.Cmd {
	run, name := s.run, a.Command
	return func() tea.Msg {
		w := &safeBuffer{}
		run(a, w)
		return wroteMsg{command: name, output: w.String()}
	}
}

type safeBuffer struct {
	mu sync.Mutex
	b  strings.Builder
}

func (s *safeBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *safeBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

// asksNothing reads the declaration off the row that names the command,
// so the table is asked rather than copied. A command the entry screen
// does not list — `take` and `work` reached from the queue do appear —
// is treated as asking, which is the safe answer.
func (s session) asksNothing(name string) bool {
	for _, r := range s.entry.rows {
		if r.command == name {
			return r.asksNothing
		}
	}
	return false
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
	switch s.where {
	case atRunning:
		return fmt.Sprintf("\n  %s running %s…\n",
			spinnerFrames[s.frame], s.running)
	case atPager:
		return s.pager.View()
	case atQueue:
		if s.err != nil {
			return fmt.Sprintf(" the queue could not be read: %v\n\n esc back · q quit\n", s.err)
		}
		return s.queue.View()
	}
	return s.entry.View()
}
