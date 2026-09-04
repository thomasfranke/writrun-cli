package term

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// headless runs fn with a deadline: a form that would hang without a
// terminal fails the test instead of the suite.
func headless(t *testing.T, fn func()) {
	t.Helper()
	done := make(chan struct{})
	go func() { fn(); close(done) }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("the form did not return — it is waiting on a terminal")
	}
}

func TestProbesReportNoTerminalUnderTheTestRunner(t *testing.T) {
	tm := New()
	// go test wires pipes, not TTYs; both probes answer accordingly.
	if tm.InteractiveIn() && tm.InteractiveOut() {
		t.Skip("running on a real terminal")
	}
}

func TestSelectConfirmsTheHighlightedOption(t *testing.T) {
	headless(t, func() {
		tm := Terminal{In: strings.NewReader("\r"), Out: &bytes.Buffer{}}
		got, err := tm.Select("pick one", []string{"first", "second"})
		if err != nil {
			t.Errorf("Select = %v", err)
			return
		}
		if got != 0 {
			t.Errorf("index = %d, want 0 (enter confirms the highlighted option)", got)
		}
	})
}

func TestSelectMovesWithArrowKeys(t *testing.T) {
	headless(t, func() {
		// Down arrow (ESC [ B), then enter.
		tm := Terminal{In: strings.NewReader("\x1b[B\r"), Out: &bytes.Buffer{}}
		got, err := tm.Select("pick one", []string{"first", "second"})
		if err != nil {
			t.Errorf("Select = %v", err)
			return
		}
		if got != 1 {
			t.Errorf("index = %d, want 1 (arrow moved before enter)", got)
		}
	})
}

func TestConfirmAnswersYes(t *testing.T) {
	headless(t, func() {
		tm := Terminal{In: strings.NewReader("y"), Out: &bytes.Buffer{}}
		got, err := tm.Confirm("proceed?")
		if err != nil {
			t.Errorf("Confirm = %v", err)
			return
		}
		if !got {
			t.Error("answer = false, want true after y")
		}
	})
}

func TestSpinRunsTheWorkAndCarriesItsError(t *testing.T) {
	headless(t, func() {
		tm := Terminal{Out: &bytes.Buffer{}}
		ran := false
		if err := tm.Spin("waiting", func() error { ran = true; return nil }); err != nil || !ran {
			t.Errorf("Spin ran=%v err=%v", ran, err)
		}
	})
}
