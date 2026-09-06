package authorcmd

import (
	"errors"
	"strings"
	"testing"
)

// The title author asks a human for is free text, and until now nothing
// judged it before the push: the branch reached the forge and the door
// refused it there (task-0024).
func TestTheComposedTitleIsJudgedBeforeThePush(t *testing.T) {
	h := newHarness(t)
	h.scripts.replies[observanceScript] = reply{
		errOut: "REJECTED: the title's type 'Wibble' is outside the vocabulary",
		err:    scriptExit(1),
	}

	err := h.author()
	if exitCode(err) != 1 {
		t.Fatalf("author = %v; want the door's own exit 1", err)
	}
	if h.git.did("push") {
		t.Error("the branch was pushed past a refused check")
	}
	if h.gh.reached("pr create") {
		t.Error("a pull request was opened past a refused check")
	}
	if !strings.Contains(h.errb.String(), "outside the vocabulary") {
		t.Errorf("the door's own words were not carried: %q", h.errb.String())
	}
}

// The door runs after the three checks that read the diff, because the
// composition it judges does not exist until they have passed.
func TestTheDoorIsTheLastCheckAndRunsOnTheResolvedRange(t *testing.T) {
	h := newHarness(t)
	if err := h.author(); err != nil {
		t.Fatalf("author = %v", err)
	}

	var order []string
	for _, c := range h.scripts.calls {
		for _, s := range []string{frontMatterScript, docShapesScript, stateScript, observanceScript} {
			if strings.HasPrefix(c, s) {
				order = append(order, s)
			}
		}
	}
	want := []string{frontMatterScript, docShapesScript, stateScript, observanceScript}
	if strings.Join(order, "\n") != strings.Join(want, "\n") {
		t.Fatalf("the checks ran %v; want %v", order, want)
	}
	if got := h.scripts.calls[len(h.scripts.calls)-1]; got != observanceScript+" origin/main...HEAD" {
		t.Errorf("the door was called %q; want the resolved range alone in argv", got)
	}
}

// The title and the body reach the door through the environment, which
// is where check_observance.sh reads them and the only way it accepts
// them.
func TestTheDoorReadsTheCompositionFromTheEnvironment(t *testing.T) {
	h := newHarness(t)
	if err := h.author(); err != nil {
		t.Fatalf("author = %v", err)
	}

	got, ok := h.scripts.handed(observanceScript, "PR_TITLE")
	if !ok {
		t.Fatal("the door was handed no PR_TITLE")
	}
	if got != title {
		t.Errorf("PR_TITLE = %q; want the composed title %q", got, title)
	}
	body, ok := h.scripts.handed(observanceScript, "PR_BODY")
	if !ok {
		t.Fatal("the door was handed no PR_BODY")
	}
	if !strings.Contains(body, "Derived work") {
		t.Errorf("PR_BODY = %q; want the body the pull request opens with", body)
	}
	// The credit half of the same call reads the body, so a body
	// carrying a trailer while agent_coauthor is false is caught here
	// too — one call, two verdicts.
	if strings.Contains(strings.Join(h.scripts.calls, " "), title) {
		t.Errorf("the title reached a script through argv: %v", h.scripts.calls)
	}
}

// A runner that failed before the script spoke is not a verdict.
func TestARunnerThatNeverReachedTheDoorIsNamed(t *testing.T) {
	h := newHarness(t)
	h.scripts.replies[observanceScript] = reply{err: errors.New("bash is not installed")}

	err := h.author()
	if err == nil || !strings.Contains(err.Error(), observanceScript) {
		t.Fatalf("author = %v; want the failure to name the check", err)
	}
	if h.git.did("push") {
		t.Error("the branch was pushed past a check that never ran")
	}
}
