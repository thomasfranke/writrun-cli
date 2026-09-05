package term

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

// headless runs fn with a deadline: a form that would hang without a
// terminal fails the test instead of the suite. fn reports through the
// returned value only — nothing touches t from the goroutine, so a
// late-unblocking form cannot log into a finished test. On timeout the
// goroutine is abandoned; the buffered channel lets it exit whenever
// the form finally returns.
func headless[T any](t *testing.T, fn func() T) T {
	t.Helper()
	ch := make(chan T, 1)
	go func() { ch <- fn() }()
	select {
	case v := <-ch:
		return v
	case <-time.After(10 * time.Second):
		t.Fatal("the form did not return — it is waiting on a terminal")
		panic("unreachable")
	}
}

type answer[T any] struct {
	val T
	err error
}

func TestProbesReportNoTerminalUnderTheTestRunner(t *testing.T) {
	tm := New()
	// go test wires stdin to the null device, never a TTY.
	if tm.InteractiveIn() {
		t.Error("InteractiveIn = true; want false under the test runner")
	}
	// stdout is a pipe under go test; only a by-hand ./term.test run
	// on a real terminal differs.
	if tm.InteractiveOut() {
		t.Skip("stdout is a terminal — running outside go test")
	}
}

func TestOverridesCountAsInteractive(t *testing.T) {
	// The overrides stand in for the terminal, so the guarded flows
	// (Ctx.AskConfirm and friends) stay exercisable headless.
	tm := Terminal{In: strings.NewReader(""), Out: &bytes.Buffer{}}
	if !tm.InteractiveIn() {
		t.Error("InteractiveIn = false with an In override; want true")
	}
	if !tm.InteractiveOut() {
		t.Error("InteractiveOut = false with an Out override; want true")
	}
}

func TestSelectConfirmsTheHighlightedOption(t *testing.T) {
	got := headless(t, func() answer[int] {
		tm := Terminal{In: strings.NewReader("\r"), Out: &bytes.Buffer{}}
		i, err := tm.Select("pick one", []string{"first", "second"})
		return answer[int]{i, err}
	})
	if got.err != nil {
		t.Fatalf("Select = %v", got.err)
	}
	if got.val != 0 {
		t.Errorf("index = %d, want 0 (enter confirms the highlighted option)", got.val)
	}
}

func TestSelectMovesWithArrowKeys(t *testing.T) {
	got := headless(t, func() answer[int] {
		// Down arrow (ESC [ B), then enter.
		tm := Terminal{In: strings.NewReader("\x1b[B\r"), Out: &bytes.Buffer{}}
		i, err := tm.Select("pick one", []string{"first", "second"})
		return answer[int]{i, err}
	})
	if got.err != nil {
		t.Fatalf("Select = %v", got.err)
	}
	if got.val != 1 {
		t.Errorf("index = %d, want 1 (arrow moved before enter)", got.val)
	}
}

func TestConfirmAnswersYes(t *testing.T) {
	got := headless(t, func() answer[bool] {
		tm := Terminal{In: strings.NewReader("y"), Out: &bytes.Buffer{}}
		ok, err := tm.Confirm("proceed?")
		return answer[bool]{ok, err}
	})
	if got.err != nil {
		t.Fatalf("Confirm = %v", got.err)
	}
	if !got.val {
		t.Error("answer = false, want true after y")
	}
}

func TestInputReturnsWhatWasTyped(t *testing.T) {
	got := headless(t, func() answer[string] {
		tm := Terminal{In: strings.NewReader("[Feat][Ci] Debounce it\r"), Out: &bytes.Buffer{}}
		s, err := tm.Input("the summary:")
		return answer[string]{s, err}
	})
	if got.err != nil {
		t.Fatalf("Input = %v", got.err)
	}
	if got.val != "[Feat][Ci] Debounce it" {
		t.Errorf("answer = %q, want the typed line", got.val)
	}
}

func TestSpinRunsTheWorkAndCarriesItsError(t *testing.T) {
	got := headless(t, func() answer[bool] {
		tm := Terminal{Out: &bytes.Buffer{}}
		ran := false
		err := tm.Spin("waiting", func() error { ran = true; return nil })
		return answer[bool]{ran, err}
	})
	if got.err != nil || !got.val {
		t.Errorf("Spin ran=%v err=%v", got.val, got.err)
	}
}

func TestSpinCarriesTheWorkError(t *testing.T) {
	boom := errors.New("boom")
	got := headless(t, func() answer[bool] {
		tm := Terminal{Out: &bytes.Buffer{}}
		return answer[bool]{true, tm.Spin("waiting", func() error { return boom })}
	})
	if !errors.Is(got.err, boom) {
		t.Errorf("Spin = %v, want the work's own error", got.err)
	}
}

func TestSpinWithoutATerminalRunsTheWorkPlainly(t *testing.T) {
	// Zero value under go test: stdout is no terminal, so the work runs
	// with no spinner and no escape sequences anywhere.
	tm := New()
	if tm.InteractiveOut() {
		t.Skip("stdout is a terminal — running outside go test")
	}
	ran := false
	if err := tm.Spin("waiting", func() error { ran = true; return nil }); err != nil || !ran {
		t.Errorf("Spin ran=%v err=%v", ran, err)
	}
}
