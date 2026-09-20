package finishcmd

import (
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

// keeper is this package's own registration for every signal the guard
// answers, armed before the first case and stopped by nothing.
//
// The cases prove the guard with real signals, so the binary's own
// disposition for them is fixture state. A process kills itself on
// SIGTERM only where no channel is registered for it, and both
// `signal.Stop` on the last channel and `signal.Reset` put it back
// there — asynchronously delivered, a signal raised before either can
// arrive after it, and the package then dies with `signal: terminated`
// and no case named (report-0052). A second registration nothing owns
// closes that: an individual Stop is no longer the last.
//
// kept counts what reached it, which is how a case can tell a signal
// that was delivered from one that killed nothing because it never
// arrived.
var (
	keeper chan os.Signal
	kept   atomic.Int32
)

// TestMain arms the keeper for the life of the run.
func TestMain(m *testing.M) {
	keeper = make(chan os.Signal, 8)
	keep()
	go func() {
		for range keeper {
			kept.Add(1)
		}
	}()
	os.Exit(m.Run())
}

// keep arms the keeper. It is called again wherever a case resets a
// signal, because `signal.Reset` undoes every registration for it —
// the keeper's included.
func keep() { signal.Notify(keeper, caught...) }

// A signal a case raises may not depend on that case having registered
// for it. This is the whole invariant: with the keeper armed the raise
// is delivered and reported; without it the process takes the
// disposition that kills, and this case fails by killing the binary
// (spec-0049).
//
// **It runs in a child of this binary, and the reason is the bug it
// would otherwise be.** A raise in the test process is still in flight
// while the next case's guard is armed, and a guard that catches it
// undoes that case's writes — a first attempt turned nine unrelated
// cases red that way, all of them reporting a spec left `approved`.
// The child runs this case alone, so the only signal its keeper can
// count is the one raised here, and nothing it leaks outlives it.
const signalChild = "WRITRUN_FINISH_SIGNAL_CHILD"

func TestARaiseNoCaseRegisteredForDoesNotKillTheBinary(t *testing.T) {
	if os.Getenv(signalChild) == "1" {
		before := kept.Load()
		raise(t, syscall.SIGTERM)
		deadline := time.After(10 * time.Second)
		for kept.Load() == before {
			select {
			case <-deadline:
				t.Fatal("the signal was never delivered — nothing held the disposition")
			default:
				time.Sleep(time.Millisecond)
			}
		}
		return
	}
	child := exec.Command(os.Args[0], "-test.run", "^"+t.Name()+"$")
	child.Env = append(os.Environ(), signalChild+"=1")
	if out, err := child.CombinedOutput(); err != nil {
		t.Fatalf("the binary did not survive a raise nothing of the case's own was registered for: %v\n%s", err, out)
	}
}

// raise sends the signal to this process. The keeper above is what
// makes that safe: it is registered for the life of the run, so no
// case's own teardown can put the signal back to the disposition that
// kills.
func raise(t *testing.T, sig syscall.Signal) {
	t.Helper()
	self, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("finding this process: %v", err)
	}
	if err := self.Signal(sig); err != nil {
		t.Fatalf("raising %s: %v", sig, err)
	}
}

// signalDuringPreflight makes the gates stand for the seconds the
// window is open: the signal arrives while `preflight.sh` runs, the
// script then reports what a child killed by the same signal reports,
// and the case reads what the guard did with the tree. This is
// report-0018's reproduction, in the package.
func signalDuringPreflight(t *testing.T, h *harness, sig syscall.Signal, verdict error) {
	t.Helper()
	h.scripts.replies[preflightScript] = reply{
		err: verdict,
		does: func() {
			raise(t, sig)
			if got := h.died.waitFor(t); got != sig {
				t.Errorf("the guard died of %v, want %v", got, sig)
			}
		},
	}
}

// The bug report-0018 reproduced: a signal between the completion
// writes and the confirmation used to run none of the undo, because the
// undo is ordinary control flow and a signal returns through none of
// it. Both files go back and the process dies of the signal
// (spec-0021, acceptance criteria).
func TestASignalInTheWindowPutsTheWritesBack(t *testing.T) {
	h := newHarness(t)
	signalDuringPreflight(t, h, syscall.SIGTERM, scriptExit(143))

	if err := h.finish(); err == nil {
		t.Fatal("finish = nil; want the stopped run's verdict")
	}
	if got := h.read(t, specPath); !strings.Contains(got, "status: approved") {
		t.Error("the spec was left implemented after the signal")
	}
	if got := h.read(t, taskPath); !strings.Contains(got, "completed: null") {
		t.Error("the task was left carrying its completion date after the signal")
	}
	if h.gh.reached("pr ready") {
		t.Error("the forge was reached after the signal")
	}
}

// SIGINT is the same window: the Ctrl-C at a run whose silence reads as
// a hang, with no prompt holding the terminal.
func TestAnInterruptInTheWindowPutsTheWritesBack(t *testing.T) {
	if signal.Ignored(syscall.SIGINT) {
		t.Skip("this binary was started with SIGINT ignored — a background job, where there is no interrupt to answer")
	}
	h := newHarness(t)
	signalDuringPreflight(t, h, syscall.SIGINT, scriptExit(130))

	if err := h.finish(); err == nil {
		t.Fatal("finish = nil; want the stopped run's verdict")
	}
	if got := h.read(t, specPath); !strings.Contains(got, "status: approved") {
		t.Error("the spec was left implemented after the interrupt")
	}
	if got := h.read(t, taskPath); !strings.Contains(got, "completed: null") {
		t.Error("the task was left carrying its completion date after the interrupt")
	}
}

// The child gets the signal too, so the script's own non-zero verdict
// asks for the undo a second time on the way up. The journal answers
// once: the tree is put back, and it is put back once.
func TestOneSignalRunsOneUndoEvenWhenTheScriptAlsoFails(t *testing.T) {
	h := newHarness(t)
	signalDuringPreflight(t, h, syscall.SIGTERM, scriptExit(143))

	if err := h.finish(); err == nil {
		t.Fatal("finish = nil; want the stopped run's verdict")
	}
	if n := strings.Count(h.out.String(), "restored "+specPath); n != 1 {
		t.Errorf("the spec was reported restored %d times, want 1:\n%s", n, h.out.String())
	}
	if n := strings.Count(h.out.String(), "restored "+taskPath); n != 1 {
		t.Errorf("the task was reported restored %d times, want 1:\n%s", n, h.out.String())
	}
	if n := h.died.seen.Load(); n != 1 {
		t.Errorf("the guard died %d times, want 1", n)
	}
}

// A script that exits 0 on its way out is not the answer either: the
// undo ran on the signal, whatever the child reported (spec-0021, edge
// cases).
func TestTheUndoRunsWhateverTheScriptReportsOnItsWayOut(t *testing.T) {
	h := newHarness(t)
	signalDuringPreflight(t, h, syscall.SIGTERM, nil)

	_ = h.finish()
	if got := h.read(t, specPath); !strings.Contains(got, "status: approved") {
		t.Error("the spec was left implemented after a signal the script survived")
	}
}

// A restore that cannot be made is reported on this path the way it is
// reported on every other one — naming the file left changed and
// telling the user to put it back by hand. The frame is not on this
// path, so the guard writes the sentence itself (spec-0021, step 4).
func TestARestoreThatFailsOnTheSignalPathSaysSo(t *testing.T) {
	h := newHarness(t)
	h.deniedAfterWrite(specPath, os.ErrPermission)
	signalDuringPreflight(t, h, syscall.SIGTERM, scriptExit(143))

	_ = h.finish()
	got := h.errb.String()
	if !strings.Contains(got, specPath) {
		t.Errorf("stderr does not name the file left changed:\n%s", got)
	}
	if !strings.Contains(got, "The working tree is left changed; put those files back by hand") {
		t.Errorf("stderr does not say the tree is left changed:\n%s", got)
	}
}

// A signal before the first completion write has nothing to put back,
// so it puts nothing back and says nothing — the journal holds the two
// files as it found them and they still hold that (spec-0021,
// acceptance criteria).
func TestASignalBeforeTheFirstWriteRestoresNothingAndSaysNothing(t *testing.T) {
	h := newHarness(t)
	undo := &journal{}
	undo.remember(specPath, []byte(h.read(t, specPath)))
	undo.remember(taskPath, []byte(h.read(t, taskPath)))
	g := &guard{ctx: h.ctx, d: h.deps(), undo: undo}

	g.catch(syscall.SIGTERM)

	if got := h.out.String(); got != "" {
		t.Errorf("stdout = %q, want nothing said", got)
	}
	if got := h.errb.String(); got != "" {
		t.Errorf("stderr = %q, want nothing said", got)
	}
	if got := h.read(t, specPath); !strings.Contains(got, "status: approved") {
		t.Error("the spec was written by an undo that had nothing to undo")
	}
}

// The writes stand once the pull request is marked ready: a signal
// after that point finds the journal sealed and takes nothing back
// (spec-0021, acceptance criteria).
func TestASignalAfterThePullRequestIsReadyTakesNothingBack(t *testing.T) {
	h := newHarness(t)
	if err := h.finish(); err != nil {
		t.Fatalf("finish = %v", err)
	}
	before := h.out.String()
	undo := &journal{}
	undo.remember(specPath, []byte("what the spec said before"))
	undo.seal()
	g := &guard{ctx: h.ctx, d: h.deps(), undo: undo}

	g.catch(syscall.SIGTERM)

	if got := h.read(t, specPath); !strings.Contains(got, "status: implemented") {
		t.Error("a sealed journal took the completion write back")
	}
	if h.out.String() != before {
		t.Errorf("a sealed journal reported an undo:\n%s", strings.TrimPrefix(h.out.String(), before))
	}
}

// unignore puts a signal back the way the test binary found it.
// `signal.Reset` alone does not do it: it undoes `signal.Notify` and
// nothing else, so a signal left ignored stays ignored for every case
// after — and the ones asserting the guard answers a signal then wait
// for a death that cannot come. Registering first is what gives Reset
// something to undo.
func unignore(t *testing.T, sigs ...os.Signal) {
	t.Helper()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, sigs...)
	signal.Stop(ch)
	signal.Reset(sigs...)
	for _, sig := range sigs {
		if signal.Ignored(sig) {
			t.Fatalf("%v is still ignored — every case after this one would wait for a death that cannot come", sig)
		}
	}
	// After the check, never before it: `signal.Notify` un-ignores what
	// it registers, so arming the keeper first would make the check
	// above pass on a signal that is still ignored.
	keep()
}

// A signal this process ignores is not armed for. It cannot end the
// run, so answering it would take the completion writes back under a
// command that then went on to succeed — and the re-raise would find
// the same ignore and kill nothing. `writrun finish &` in a script is
// that case: a shell starts a background job with SIGINT ignored.
func TestAnIgnoredSignalIsNotArmedFor(t *testing.T) {
	signal.Ignore(syscall.SIGINT)
	defer unignore(t, syscall.SIGINT)

	got := armable()
	if len(got) != 1 || got[0] != syscall.SIGTERM {
		t.Errorf("armable() = %v, want the terminate alone", got)
	}
}

// Nothing catchable is left to catch when both are ignored, and an
// empty set is never registered: `signal.Notify` with no signal asks
// for every one of them, which is the opposite of standing down.
func TestAGuardWithNothingToArmForRegistersNothing(t *testing.T) {
	signal.Ignore(caught...)
	defer unignore(t, caught...)

	h := newHarness(t)
	undo := &journal{}
	undo.remember(specPath, []byte("what the spec said before"))
	g := arm(h.ctx, h.deps(), undo)
	defer g.disarm()

	if len(g.armed) != 0 {
		t.Fatalf("armed = %v, want nothing", g.armed)
	}
	raise(t, syscall.SIGTERM)
	if n := h.died.seen.Load(); n != 0 {
		t.Errorf("the guard died %d times of a signal this process ignores", n)
	}
}

// A signal arriving while the prompt holds the terminal is the
// prompt's: the guard is stood down for the length of the question, so
// it cannot kill the process behind a form that still has the terminal
// in raw mode (spec-0021, step 5).
func TestTheGuardStandsDownWhileTheQuestionHoldsTheTerminal(t *testing.T) {
	h := newHarness(t)
	undo := &journal{}
	undo.remember(specPath, []byte("what the spec said before"))
	g := arm(h.ctx, h.deps(), undo)

	err := g.whileAsking(func() error {
		raise(t, syscall.SIGTERM)
		select {
		case sig := <-h.died.got:
			t.Errorf("the guard died of %v while the question held the terminal", sig)
		case <-time.After(500 * time.Millisecond):
		}
		return nil
	})
	g.disarm()

	if err != nil {
		t.Fatalf("whileAsking = %v", err)
	}
	if got := h.read(t, specPath); strings.Contains(got, "what the spec said before") {
		t.Error("the guard put a file back behind the question")
	}
}

// The window reopens after the question: the guard takes the delivery
// back, so a signal between the yes and the forge is still answered.
func TestTheGuardTakesTheSignalsBackAfterTheQuestion(t *testing.T) {
	h := newHarness(t)
	g := arm(h.ctx, h.deps(), &journal{})
	if err := g.whileAsking(func() error { return nil }); err != nil {
		t.Fatalf("whileAsking = %v", err)
	}

	raise(t, syscall.SIGTERM)
	if got := h.died.waitFor(t); got != syscall.SIGTERM {
		t.Errorf("the guard died of %v, want SIGTERM", got)
	}
	g.disarm()
}

// The confirmation's own interrupt handling is left alone: the error
// the prompt returns carries the ordinary undo, and the tree is put
// back once whoever asked for it (spec-0021, step 5).
func TestASignalAtThePromptIsLeftToThePrompt(t *testing.T) {
	h := newHarness(t)
	h.term.ConfirmAnswer = false
	h.ctx.Terminal = &raisingTerminal{Terminal: h.term, t: t, sig: syscall.SIGTERM}

	if err := h.finish(); err == nil {
		t.Fatal("finish = nil; want the decline")
	}
	if got := h.read(t, specPath); !strings.Contains(got, "status: approved") {
		t.Error("the decline's own undo did not run")
	}
	if n := strings.Count(h.out.String(), "restored "+specPath); n != 1 {
		t.Errorf("the spec was reported restored %d times, want 1:\n%s", n, h.out.String())
	}
}

// disarm is the end of the window: the watcher is gone before it
// returns, so a signal after it cannot reach the journal.
func TestADisarmedGuardLeavesTheSignalAlone(t *testing.T) {
	// The probe keeps the disposition off the default, so the signal
	// this case sends reports rather than kills.
	probe := make(chan os.Signal, 1)
	signal.Notify(probe, syscall.SIGTERM)
	defer signal.Stop(probe)

	h := newHarness(t)
	undo := &journal{}
	undo.remember(specPath, []byte("what the spec said before"))
	g := arm(h.ctx, h.deps(), undo)
	g.disarm()

	raise(t, syscall.SIGTERM)
	select {
	case <-probe:
	case <-time.After(10 * time.Second):
		t.Fatal("the signal was never delivered")
	}
	if n := h.died.seen.Load(); n != 0 {
		t.Errorf("a disarmed guard died %d times, want 0", n)
	}
	if got := h.read(t, specPath); strings.Contains(got, "what the spec said before") {
		t.Error("a disarmed guard put a file back")
	}
}
