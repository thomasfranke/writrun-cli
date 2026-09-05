package takecmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
)

const title = "[Feat][Ci] Debounce the mirror updates"

func TestNewDeclaresTheCommand(t *testing.T) {
	c := New(Deps{})
	if c.Name != "take" {
		t.Errorf("name = %q, want take", c.Name)
	}
	if c.Need != command.NeedAdopted {
		t.Errorf("need = %v; want an adopted repository", c.Need)
	}
	if c.Summary == "" || c.Run == nil {
		t.Error("the command carries no summary or no work")
	}
}

// Exit 0 — the act is done, and the script's own report is the report.
func TestExitZeroReportsTheScriptsOwnWords(t *testing.T) {
	h := newHarness(t, reply{out: "Took task-0001: task/0001-a-thing pushed, draft pull request open.\n"})
	if err := h.take("task-0001", "--title", title); err != nil {
		t.Fatalf("take = %v", err)
	}
	if !strings.Contains(h.out.String(), "Took task-0001") {
		t.Errorf("stdout = %q; want the script's own report", h.out.String())
	}
	if len(h.scripts.calls) != 1 {
		t.Fatalf("%d calls, want one", len(h.scripts.calls))
	}
	if h.scripts.calls[0].script != takeScript {
		t.Errorf("script = %q, want %q", h.scripts.calls[0].script, takeScript)
	}
	if h.scripts.calls[0].root != "/repo" {
		t.Errorf("root = %q, want the repository root", h.scripts.calls[0].root)
	}
	if got := h.argsOf(t, 0); got != "task-0001 --title "+title {
		t.Errorf("args = %q; want the id and the title", got)
	}
	if len(h.term.Asked) != 0 {
		t.Errorf("asked %v; a done take asks nothing", h.term.Asked)
	}
}

// Exit 1 — a refusal, passed through with the script's exit code and
// nothing added to what it said.
func TestExitOneRefusalPassesThroughUnedited(t *testing.T) {
	h := newHarness(t, reply{errOut: "REFUSED: task-0001 is 'backlog'\n", err: scriptExit(1)})
	err := h.take("task-0001", "--title", title)
	if err == nil {
		t.Fatal("a refusal returned no error")
	}
	if got := exitOf(err); got != 1 {
		t.Errorf("exit = %d, want the script's own 1", got)
	}
	if !strings.Contains(h.errb.String(), "REFUSED: task-0001 is 'backlog'") {
		t.Errorf("stderr = %q; want the script's reason", h.errb.String())
	}
	if len(h.scripts.calls) != 1 {
		t.Errorf("%d calls; a refusal is never rerun", len(h.scripts.calls))
	}
	if len(h.term.Asked) != 0 {
		t.Errorf("asked %v; a refusal asks nothing", h.term.Asked)
	}
}

// Exit 2 — composed and waiting: the composition is the script's, the
// question is this command's, and a yes reruns the exact act.
func TestExitTwoAsksThenRerunsWithConfirm(t *testing.T) {
	h := newHarness(t,
		reply{out: "Composed, and waiting: auto_pr is false.\nbranch: task/0001-a-thing\n", err: scriptExit(2)},
		reply{out: "Took task-0001.\n"},
	)
	h.term.In = true
	h.term.ConfirmAnswer = true

	if err := h.take("task-0001", "--title", title); err != nil {
		t.Fatalf("take = %v", err)
	}
	if !strings.Contains(h.out.String(), "branch: task/0001-a-thing") {
		t.Errorf("stdout = %q; want the composition shown", h.out.String())
	}
	if len(h.term.Asked) != 1 {
		t.Fatalf("asked %v, want one question", h.term.Asked)
	}
	if len(h.scripts.calls) != 2 {
		t.Fatalf("%d calls, want the take and its confirmed rerun", len(h.scripts.calls))
	}
	if got := h.argsOf(t, 1); got != "task-0001 --title "+title+" --confirm" {
		t.Errorf("rerun args = %q; want the same act plus --confirm", got)
	}
}

func TestDecliningTheCompositionReachesNothing(t *testing.T) {
	h := newHarness(t, reply{out: "Composed, and waiting.\n", err: scriptExit(2)})
	h.term.In = true
	h.term.ConfirmAnswer = false

	err := h.take("task-0001", "--title", title)
	if !errors.Is(err, command.ErrDeclined) {
		t.Fatalf("err = %v, want ErrDeclined", err)
	}
	if len(h.scripts.calls) != 1 {
		t.Errorf("%d calls; a decline reruns nothing", len(h.scripts.calls))
	}
}

// --yes answers the question, and the composition is still printed
// first (spec-0008, edge cases).
func TestYesAnswersTheQuestionAndStillShowsTheComposition(t *testing.T) {
	h := newHarness(t,
		reply{out: "Composed, and waiting.\nbranch: task/0001-a-thing\n", err: scriptExit(2)},
		reply{out: "Took task-0001.\n"},
	)
	h.ctx.Yes = true

	if err := h.take("task-0001", "--title", title); err != nil {
		t.Fatalf("take = %v", err)
	}
	if len(h.term.Asked) != 0 {
		t.Errorf("asked %v; --yes answers it", h.term.Asked)
	}
	if !strings.Contains(h.out.String(), "branch: task/0001-a-thing") {
		t.Errorf("stdout = %q; want the composition printed before the confirmed rerun", h.out.String())
	}
	if len(h.scripts.calls) != 2 {
		t.Fatalf("%d calls, want the take and its confirmed rerun", len(h.scripts.calls))
	}
}

// Exit 3 — the git or forge failure, with the resume the script named.
func TestExitThreePassesTheReasonAndTheResumeThrough(t *testing.T) {
	resume := "  bash take_task.sh task-0001 --title \"" + title + "\" --slug a-thing --resume --confirm\n"
	h := newHarness(t, reply{errOut: "gh pr create failed:\nFinish it with:\n" + resume, err: scriptExit(3)})

	err := h.take("task-0001", "--title", title)
	if got := exitOf(err); got != 3 {
		t.Errorf("exit = %d, want the script's own 3", got)
	}
	if !strings.Contains(h.errb.String(), "--resume") {
		t.Errorf("stderr = %q; want the exact resume invocation", h.errb.String())
	}
}

// A confirmed rerun that composes again would leave the act half done
// with nobody saying so, so it is named.
func TestAConfirmedRerunThatComposesAgainIsNamed(t *testing.T) {
	h := newHarness(t,
		reply{err: scriptExit(2)},
		reply{err: scriptExit(2)},
	)
	h.ctx.Yes = true
	err := h.take("task-0001", "--title", title)
	if err == nil || !strings.Contains(err.Error(), "composed again") {
		t.Fatalf("err = %v; want the second composition named", err)
	}
}

// An error carrying no exit code is the runner's, not the script's.
func TestARunnerFailureIsNotAScriptVerdict(t *testing.T) {
	h := newHarness(t, reply{err: errors.New("bash: no such file")})
	err := h.take("task-0001", "--title", title)
	if err == nil || !strings.Contains(err.Error(), takeScript) {
		t.Fatalf("err = %v; want the script named", err)
	}
	if got := exitOf(err); got != 1 {
		t.Errorf("exit = %d, want 1 for a runner failure", got)
	}
}

func TestTheTypedTitleIsPassedThrough(t *testing.T) {
	h := newHarness(t, reply{})
	h.term.In = true
	h.term.InputAnswer = title

	if err := h.take("task-0001"); err != nil {
		t.Fatalf("take = %v", err)
	}
	if got := h.argsOf(t, 0); got != "task-0001 --title "+title {
		t.Errorf("args = %q; want the typed title", got)
	}
	if len(h.term.Asked) != 1 {
		t.Fatalf("asked %v, want the one free-text question", h.term.Asked)
	}
}

// Nothing is validated here: an empty title is the script's refusal to
// give, not this command's (spec-0008, steps).
func TestAnEmptyTypedTitleStillReachesTheScript(t *testing.T) {
	h := newHarness(t, reply{errOut: "REFUSED: --title is required\n", err: scriptExit(1)})
	h.term.In = true
	h.term.InputAnswer = ""

	if got := exitOf(h.take("task-0001")); got != 1 {
		t.Errorf("exit = %d, want the script's refusal", got)
	}
	if got := h.argsOf(t, 0); got != "task-0001 --title " {
		t.Errorf("args = %q; want the empty title handed over", got)
	}
}

func TestWithoutATerminalTheTitleAbortsNamingItsFlag(t *testing.T) {
	h := newHarness(t)
	err := h.take("task-0001")
	if err == nil || !strings.Contains(err.Error(), "--title") {
		t.Fatalf("err = %v; want --title named", err)
	}
	if len(h.scripts.calls) != 0 {
		t.Errorf("%d calls; an unanswered question runs nothing", len(h.scripts.calls))
	}
}

func TestSlugIsPassedThrough(t *testing.T) {
	h := newHarness(t, reply{})
	if err := h.take("task-0001", "--title", title, "--slug", "a-thing"); err != nil {
		t.Fatalf("take = %v", err)
	}
	if got := h.argsOf(t, 0); got != "task-0001 --title "+title+" --slug a-thing" {
		t.Errorf("args = %q; want the slug passed through", got)
	}
}

func TestSplit(t *testing.T) {
	cases := []struct {
		name      string
		args      []string
		wantID    string
		wantFlags string
		wantErr   string
	}{
		{"the id before the flags", []string{"task-9", "--title", "t"}, "task-9", "--title t", ""},
		{"the id after the flags", []string{"--title", "t", "task-9"}, "task-9", "--title t", ""},
		{"a title that looks like an id", []string{"--title", "task-9"}, "", "--title task-9", ""},
		{"no id at all", []string{"--slug", "s"}, "", "--slug s", ""},
		{"an attached value", []string{"--title=t", "task-9"}, "task-9", "--title=t", ""},
		{"two ids", []string{"task-9", "task-8"}, "", "", "two task ids"},
		{"a dangling flag", []string{"--title"}, "", "--title", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id, flags, err := split(tc.args)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v; want it to name %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("err = %v", err)
			}
			if id != tc.wantID {
				t.Errorf("id = %q, want %q", id, tc.wantID)
			}
			if got := strings.Join(flags, " "); got != tc.wantFlags {
				t.Errorf("flags = %q, want %q", got, tc.wantFlags)
			}
		})
	}
}

func TestAnUnknownFlagAbortsBeforeAnythingRuns(t *testing.T) {
	h := newHarness(t)
	if err := h.take("task-0001", "--bogus"); err == nil {
		t.Fatal("an unknown flag was accepted")
	}
	if len(h.scripts.calls) != 0 {
		t.Errorf("%d calls; an unusable argument runs nothing", len(h.scripts.calls))
	}
}
