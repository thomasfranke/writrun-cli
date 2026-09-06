package finishcmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/thomasfranke/writrun-cli/internal/command"
)

// caught is every signal the guard answers: the Ctrl-C at a run whose
// silence reads as a hang, and the supervisor's kill.
//
// The set stops there. A signal this command takes off its default
// disposition is a signal this command must then answer for, and these
// two are the ones huh already answers while it holds the terminal, so
// standing down for the question costs nothing. SIGQUIT keeps its
// default, which is the runtime's goroutine dump. SIGKILL cannot be
// caught at all, so the guard narrows the window rather than closing
// it (spec-0021, edge cases).
var caught = []os.Signal{syscall.SIGINT, syscall.SIGTERM}

// armable is that set minus whatever this process already ignores. A
// signal the process ignores does not end the run, so answering it
// would put the completion writes back under a command that then went
// on to succeed — and the re-raise would find the same ignore and never
// kill anything. A shell starts a background job with SIGINT ignored,
// so `writrun finish &` is exactly that case.
func armable() []os.Signal {
	armed := make([]os.Signal, 0, len(caught))
	for _, sig := range caught {
		if !signal.Ignored(sig) {
			armed = append(armed, sig)
		}
	}
	return armed
}

// guard is the undo on the one path ordinary control flow does not
// reach. From the moment the journal holds what it would put back until
// the command reaches an end on its own, a caught signal runs the same
// `journal.restore` every other non-success end runs and then lets the
// signal kill the process (spec-0021).
type guard struct {
	ctx  *command.Ctx
	d    Deps
	undo *journal

	// armed is what this process can actually be killed by, fixed once
	// so standing down and starting again register the same set.
	armed []os.Signal
	// ch carries the signals. One buffered slot is enough: a second
	// signal arriving while the restore runs must not restart it, and
	// the runtime drops what does not fit.
	ch chan os.Signal
	// done is closed by disarm; stopped is closed by the watcher on
	// its way out, so disarm can wait for it.
	done    chan struct{}
	stopped chan struct{}
}

// arm starts answering signals. It is called once the journal holds
// what it would put back and before the first write, so a signal
// arriving with nothing written finds nothing to undo.
func arm(ctx *command.Ctx, d Deps, undo *journal) *guard {
	g := &guard{
		ctx:     ctx,
		d:       d,
		undo:    undo,
		armed:   armable(),
		ch:      make(chan os.Signal, 1),
		done:    make(chan struct{}),
		stopped: make(chan struct{}),
	}
	g.notify(g.ch)
	go g.watch()
	return g
}

// notify registers a channel for the armed set. An empty set is not a
// call to `signal.Notify` with no signals, which asks for every one of
// them.
func (g *guard) notify(ch chan os.Signal) {
	if len(g.armed) == 0 {
		return
	}
	signal.Notify(ch, g.armed...)
}

// watch is the guard's own goroutine: one signal, one undo, one death.
func (g *guard) watch() {
	defer close(g.stopped)
	for {
		// An end reached on its own wins a signal still queued behind
		// it: the command finished, and putting the files back after
		// that would undo writes that must stand. A bare select would
		// choose between the two at random.
		select {
		case <-g.done:
			return
		default:
		}
		select {
		case <-g.done:
			return
		case sig := <-g.ch:
			g.catch(sig)
			g.d.Die(sig)
			return
		}
	}
}

// catch runs the undo and reports what it could not put back, the way
// every other non-success end reports it — the frame is not on this
// path, so the sentence is written here rather than returned.
//
// A restore that put everything back has already said so, one line per
// file, and a signal that arrived before the first write has nothing to
// say at all: `restore` hands the cause back untouched in both cases,
// and anything else is the sentence naming the files left changed
// (spec-0021, step 4).
func (g *guard) catch(sig os.Signal) {
	cause := fmt.Errorf("interrupted: %s", sig)
	if err := g.undo.restore(g.ctx, g.d, cause); err != cause {
		fmt.Fprintf(g.ctx.Stderr, "writrun finish: %v\n", err)
	}
}

// whileAsking runs the confirmation with the guard stood down, because
// huh answers these two signals for itself while it holds the terminal:
// bubbletea turns SIGINT into an abort and SIGTERM into a quit, the
// question returns an error either way, and the ordinary undo runs on
// the way up. Standing down is what keeps the process from dying with
// the terminal still in raw mode (spec-0021, step 5).
//
// It stands down by stopping the delivery rather than by reading a
// flag: `signal.Stop` guarantees the channel receives nothing more,
// which no flag read after a signal is already off the channel can. A
// second registration takes the disposition for the length of the
// question, so standing down never hands a signal back to the default
// and kills the process outright.
//
// One restore is guaranteed by the journal, not by this: a signal the
// prompt answered can still be handed over once the question has
// returned, and the journal answers the second asking with the work
// already done.
func (g *guard) whileAsking(ask func() error) error {
	parked := make(chan os.Signal, 1)
	g.notify(parked)
	defer signal.Stop(parked)
	signal.Stop(g.ch)
	select {
	case <-g.ch:
	default:
	}
	defer g.notify(g.ch)
	return ask()
}

// disarm ends the window: it stops the delivery and waits for the
// watcher to go, so nothing this command already ended can be undone
// behind it. A signal caught in that same instant wins, and the wait is
// then the process waiting to die of it.
func (g *guard) disarm() {
	signal.Stop(g.ch)
	close(g.done)
	<-g.stopped
}

// Die ends the process the way it would have ended with no handler
// installed: the signal's disposition goes back to the default and the
// signal is raised again, so the shell reports the conventional 128+n
// rather than a status invented here (spec-0021, step 3).
//
// It does not return. Delivery is asynchronous, so the raise is
// followed by a block rather than a return into a command that has
// already put the tree back.
func Die(sig os.Signal) {
	signal.Reset(sig)
	self, err := os.FindProcess(os.Getpid())
	if err == nil && self.Signal(sig) == nil {
		select {}
	}
	// A platform that cannot raise the signal has no 128+n convention
	// to honour; ending non-zero is the whole of what is left.
	os.Exit(1)
}
