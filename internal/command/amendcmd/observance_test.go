package amendcmd

import (
	"errors"
	"strings"
	"testing"
)

// verdictOf reads a script's own exit code off the error a command
// returned; 0 says nothing carried one.
func verdictOf(err error) int {
	var v interface{ ExitCode() int }
	if errors.As(err, &v) {
		return v.ExitCode()
	}
	return 0
}

// The door judges the composition before the first write. `--type
// wibble` composes a title the project's own vocabulary refuses, and
// until now the refusal arrived on the forge — after the branch was
// pushed and the pull request was open (task-0024, report-0022).
func TestTheComposedTitleIsJudgedBeforeAnythingIsWritten(t *testing.T) {
	h := newHarness(t)
	h.scripts.fail[observanceScript] = scriptExit(1)

	err := h.amend("--type", "wibble")
	if verdictOf(err) != 1 {
		t.Fatalf("amend = %v; want the door's own exit 1", err)
	}
	if h.git.ran("switch") {
		t.Error("a branch was cut past a refused check")
	}
	if h.gh.reached("pr create") {
		t.Error("a pull request was opened past a refused check")
	}
	if got := h.read(t, specPath); !strings.Contains(got, "status: approved") {
		t.Error("the spec was written past a refused check")
	}
	for _, q := range h.term.Asked {
		if strings.HasPrefix(q, "Return spec-0011") {
			t.Errorf("the confirmation was asked past a refused check: %q", q)
		}
	}
}

// --yes answers a question. A check is not a question, so the refusal
// fires on the unattended path too (spec-0023, edge cases).
func TestTheDoorRefusesWithoutATerminalToo(t *testing.T) {
	h := newHarness(t)
	h.scripts.fail[observanceScript] = scriptExit(1)
	h.ctx.Yes = true
	h.term.In = false

	if verdictOf(h.amend("--type", "wibble")) != 1 {
		t.Fatal("--yes carried the composition past the door")
	}
	if h.git.ran("switch") {
		t.Error("a branch was cut past a refused check")
	}
}

// The title and the body reach the door through the environment, which
// is where check_observance.sh reads them and the only way it accepts
// them: argv carries the range and nothing else.
func TestTheDoorReadsTheCompositionFromTheEnvironment(t *testing.T) {
	h := newHarness(t)
	if err := h.amend(); err != nil {
		t.Fatalf("amend = %v", err)
	}

	title, ok := h.scripts.handed(observanceScript, "PR_TITLE")
	if !ok {
		t.Fatal("the door was handed no PR_TITLE")
	}
	if want := h.gh.created()["--title"]; title != want {
		t.Errorf("PR_TITLE = %q; want the title the pull request opened with, %q", title, want)
	}
	body, ok := h.scripts.handed(observanceScript, "PR_BODY")
	if !ok {
		t.Fatal("the door was handed no PR_BODY")
	}
	if want := h.gh.created()["--body"]; body != want {
		t.Errorf("PR_BODY = %q; want the body the pull request opened with", body)
	}

	asked := false
	for _, c := range h.scripts.calls {
		if !strings.HasPrefix(c, observanceScript) {
			continue
		}
		asked = true
		if c != observanceScript+" "+composedRange {
			t.Errorf("the door was called %q; want the range alone in argv", c)
		}
	}
	if !asked {
		t.Errorf("the door was never asked about the composition: %v", h.scripts.calls)
	}
}

// A runner that is not wired is a check that cannot run, and amend does
// not compose past one.
func TestAmendRefusesWithNoRunnerToAskTheDoorWith(t *testing.T) {
	h := newHarness(t)
	d := h.deps()
	d.Scripts = nil

	err := run(h.ctx, d, []string{"spec-0011", "--title", amendTitle})
	if err == nil || !strings.Contains(err.Error(), "cannot be judged") {
		t.Fatalf("amend = %v; want a refusal naming the check it could not run", err)
	}
	if h.git.ran("switch") {
		t.Error("a branch was cut without the door being asked")
	}
}

// A runner that failed before the script spoke is not a verdict: it is
// named, rather than passed up as somebody's exit code.
func TestARunnerThatNeverReachedTheDoorIsNamed(t *testing.T) {
	h := newHarness(t)
	h.scripts.fail[observanceScript] = errors.New("bash is not installed")

	err := h.amend()
	if err == nil || !strings.Contains(err.Error(), observanceScript) {
		t.Fatalf("amend = %v; want the failure to name the check", err)
	}
	if verdictOf(err) != 0 {
		t.Errorf("amend = %v; want no exit code invented for a script that never ran", err)
	}
}
