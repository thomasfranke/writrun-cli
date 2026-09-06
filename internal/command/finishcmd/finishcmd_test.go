package finishcmd

import (
	"errors"
	"path"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/queue"
)

func TestNewDeclaresTheCommand(t *testing.T) {
	c := New(Deps{})
	if c.Name != "finish" {
		t.Errorf("name = %q, want finish", c.Name)
	}
	if c.Need != command.NeedAdopted {
		t.Errorf("need = %v; want an adopted repository", c.Need)
	}
	if c.Summary == "" || c.Run == nil {
		t.Error("the command carries no summary or no work")
	}
}

// The command the frame dispatches is the work itself — the wiring
// New builds is what runs, not a second path beside it.
func TestTheDeclaredCommandRunsTheWork(t *testing.T) {
	h := newHarness(t)
	if err := New(h.deps()).Run(h.ctx, nil); err != nil {
		t.Fatalf("finish = %v", err)
	}
	if !h.gh.reached("pr ready 45") {
		t.Error("the command the frame would dispatch did nothing")
	}
}

// The whole sequence, in the one order spec-0010 fixes: deltas, the two
// writes, provenance, preflight, the question, the forge.
func TestTheGreenPathRunsTheSequenceInOrder(t *testing.T) {
	h := newHarness(t)
	if err := h.finish(); err != nil {
		t.Fatalf("finish = %v", err)
	}

	var order []string
	for _, c := range h.scripts.calls {
		order = append(order, c.script)
	}
	want := []string{deltasScript, provenanceScript, preflightScript}
	if strings.Join(order, "\n") != strings.Join(want, "\n") {
		t.Errorf("order =\n%s\nwant\n%s", strings.Join(order, "\n"), strings.Join(want, "\n"))
	}
	if got, _ := h.scripts.ran(deltasScript); got != "spec-0010 origin/main...HEAD" {
		t.Errorf("check_deltas args = %q", got)
	}
	if got, _ := h.scripts.ran(provenanceScript); got != "task-0011" {
		t.Errorf("record_provenance args = %q; want the task and nothing else", got)
	}
	if got, _ := h.scripts.ran(preflightScript); got != "task-0011 origin/main...HEAD" {
		t.Errorf("preflight args = %q", got)
	}

	if got := queue.Field([]byte(h.read(t, specPath)), "status"); got != "implemented" {
		t.Errorf("spec status = %q, want implemented", got)
	}
	if got := queue.Field([]byte(h.read(t, taskPath)), "completed"); got != stamped {
		t.Errorf("completed = %q, want %q", got, stamped)
	}
	if !h.gh.reached("pr ready 45") {
		t.Errorf("the pull request was never marked ready; gh saw %v", h.gh.calls)
	}
	if len(h.term.Asked) != 1 || !strings.Contains(h.term.Asked[0], "#45") {
		t.Errorf("asked %v; want one question naming the pull request", h.term.Asked)
	}
}

// The composition is shown before the question, so a yes answers
// something the human has seen (product/pull-requests/shape.md).
func TestTheCompositionIsShownBeforeTheQuestion(t *testing.T) {
	h := newHarness(t)
	if err := h.finish(); err != nil {
		t.Fatalf("finish = %v", err)
	}
	out := h.out.String()
	for _, want := range []string{"task-0011", "spec-0010", "#45", "[TASK-0011] Finish a task"} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout does not show %q:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "wrote completed: "+stamped+" on "+taskPath) {
		t.Errorf("stdout does not report the completion write:\n%s", out)
	}
}

// The task's status line is what this command may never write — not on
// the green path, and not anywhere (spec-0010, acceptance criteria).
func TestTheTasksStatusLineIsNeverWritten(t *testing.T) {
	h := newHarness(t)
	before := h.read(t, taskPath)
	if err := h.finish(); err != nil {
		t.Fatalf("finish = %v", err)
	}
	after := h.read(t, taskPath)
	if statusLineOf(after) != statusLineOf(before) {
		t.Errorf("status line = %q, was %q", statusLineOf(after), statusLineOf(before))
	}
	if statusLineOf(after) != "status: in-progress" {
		t.Errorf("status line = %q", statusLineOf(after))
	}
	// The completion write is the only difference between the two.
	if strings.Replace(before, "completed: null", "completed: "+stamped, 1) != after {
		t.Errorf("the task changed in more than its completed date:\n%s", after)
	}
}

// Every exit code check_deltas.sh defines stops the command where it
// stands: no write, no provenance, no preflight, no forge, and the
// script's own code carried up (spec-0010, acceptance criteria).
func TestADeltaFailureStopsBeforeAnyWrite(t *testing.T) {
	for _, code := range []int{1, 2, 3} {
		h := newHarness(t)
		h.scripts.replies[deltasScript] = reply{
			errOut: "MISSING: spec-0010's promised change to 'docs/x.md' not found in diff\n",
			err:    scriptExit(code),
		}
		before := h.read(t, taskPath)
		err := h.finish()
		if got := exitOf(err); got != code {
			t.Fatalf("exit = %d, want the script's own %d", got, code)
		}
		if len(h.scripts.calls) != 1 {
			t.Errorf("%d scripts ran; nothing runs after a failed delta check", len(h.scripts.calls))
		}
		if h.read(t, taskPath) != before {
			t.Error("the task was written after a failed delta check")
		}
		if got := queue.Field([]byte(h.read(t, specPath)), "status"); got != "approved" {
			t.Errorf("spec status = %q; the spec was written after a failed delta check", got)
		}
		if len(h.gh.calls) != 0 {
			t.Errorf("the forge was reached: %v", h.gh.calls)
		}
		if !strings.Contains(h.errb.String(), "MISSING") {
			t.Errorf("stderr = %q; want the script's own reason", h.errb.String())
		}
	}
}

// A runner that failed before the script spoke has no verdict to map:
// the command names the script instead of inventing an exit code.
func TestARunnerFailureNamesTheScript(t *testing.T) {
	h := newHarness(t)
	h.scripts.replies[deltasScript] = reply{err: errors.New("fork/exec: permission denied")}
	err := h.finish()
	if err == nil {
		t.Fatal("a runner failure returned no error")
	}
	if !strings.Contains(err.Error(), deltasScript) {
		t.Errorf("the error does not name the script: %v", err)
	}
}

// An Outcome nobody wrote is a refusal, before the spec is marked
// implemented and before anything else runs (spec-0010, step 2).
func TestAnEmptyOutcomeRefusesTheImplementation(t *testing.T) {
	for _, outcome := range []string{"", "_(fill after execution)_", "TODO"} {
		h := newHarness(t)
		h.seed(specPath, specFixture("approved", outcome))
		err := h.finish()
		if err == nil {
			t.Fatalf("an Outcome of %q was accepted", outcome)
		}
		if !strings.Contains(err.Error(), "Outcome") || !strings.Contains(err.Error(), "spec-0010") {
			t.Errorf("the refusal does not name the spec's Outcome: %v", err)
		}
		if got := queue.Field([]byte(h.read(t, specPath)), "status"); got != "approved" {
			t.Errorf("spec status = %q; an unfilled Outcome was marked implemented", got)
		}
		if got := queue.Field([]byte(h.read(t, taskPath)), "completed"); got != "null" {
			t.Errorf("completed = %q; it was written past a refusal", got)
		}
		if _, ran := h.scripts.ran(preflightScript); ran {
			t.Error("preflight ran past a refusal")
		}
		if len(h.gh.calls) != 0 {
			t.Errorf("the forge was reached: %v", h.gh.calls)
		}
	}
}

// Two specs on one branch: one call carrying both, and both marked
// implemented (spec-0010, edge cases).
func TestSeveralSpecsAreOneDeltaCall(t *testing.T) {
	h := newHarness(t)
	h.seed(taskPath, taskFixture("in-progress", "spec-0010, spec-0011", "null"))
	h.seed("work/specs/spec-0011-another.md",
		strings.ReplaceAll(specFixture("approved", "Also built."), "spec-0010", "spec-0011"))
	if err := h.finish(); err != nil {
		t.Fatalf("finish = %v", err)
	}
	if got, _ := h.scripts.ran(deltasScript); got != "spec-0010,spec-0011 origin/main...HEAD" {
		t.Errorf("check_deltas args = %q; want one call carrying both specs", got)
	}
	for _, p := range []string{specPath, "work/specs/spec-0011-another.md"} {
		if got := queue.Field([]byte(h.read(t, p)), "status"); got != "implemented" {
			t.Errorf("%s status = %q, want implemented", p, got)
		}
	}
}

// One spec of two with an empty Outcome refuses the whole branch — the
// first spec is not half-recorded.
func TestOneUnfilledOutcomeRefusesBeforeEitherSpecIsWritten(t *testing.T) {
	h := newHarness(t)
	h.seed(taskPath, taskFixture("in-progress", "spec-0010, spec-0011", "null"))
	h.seed("work/specs/spec-0011-another.md",
		strings.ReplaceAll(specFixture("approved", "_(fill after execution)_"), "spec-0010", "spec-0011"))
	if err := h.finish(); err == nil {
		t.Fatal("the branch was finished with one Outcome unwritten")
	}
	if got := queue.Field([]byte(h.read(t, specPath)), "status"); got != "approved" {
		t.Errorf("spec-0010 status = %q; it was written before its sibling was judged", got)
	}
}

// A task with no spec: nothing to check, nothing to mark implemented,
// and the completion still written and still gated (spec-0010, edge
// cases).
func TestATaskWithNoSpecSkipsTheDeltasAndStillCompletes(t *testing.T) {
	h := newHarness(t)
	h.seed(taskPath, taskFixture("in-progress", "", "null"))
	if err := h.finish(); err != nil {
		t.Fatalf("finish = %v", err)
	}
	if _, ran := h.scripts.ran(deltasScript); ran {
		t.Error("check_deltas ran for a task carrying no spec")
	}
	if _, ran := h.scripts.ran(preflightScript); !ran {
		t.Error("preflight did not run for a task carrying no spec")
	}
	if got := queue.Field([]byte(h.read(t, taskPath)), "completed"); got != stamped {
		t.Errorf("completed = %q, want %q", got, stamped)
	}
	if !strings.Contains(h.out.String(), "carries no spec") {
		t.Errorf("stdout does not say why no deltas were checked:\n%s", h.out.String())
	}
	if !h.gh.reached("pr ready 45") {
		t.Error("the pull request was not marked ready")
	}
}

// The ledger runs unconditionally, and its vocabulary is passed through
// unread in the script's own key order (spec-0010, step 3).
func TestTheLedgerFlagsArePassedThrough(t *testing.T) {
	h := newHarness(t)
	err := h.finish("task-0011", "--by", "agent", "--model", "claude-opus-5",
		"--login", "octocat", "--input", "562", "--output", "1753",
		"--cache-read", "37266", "--cache-write", "3665")
	if err != nil {
		t.Fatalf("finish = %v", err)
	}
	want := "task-0011 by=agent login=octocat model=claude-opus-5 input=562 output=1753 cache_read=37266 cache_write=3665"
	if got, _ := h.scripts.ran(provenanceScript); got != want {
		t.Errorf("record_provenance args =\n%q\nwant\n%q", got, want)
	}
}

// A ledger the script refuses stops the run there: preflight does not
// follow, and the forge is never reached.
func TestALedgerFailureStopsBeforePreflight(t *testing.T) {
	h := newHarness(t)
	h.scripts.replies[provenanceScript] = reply{
		errOut: "record_provenance.sh: by= is required\n",
		err:    scriptExit(1),
	}
	if got := exitOf(h.finish()); got != 1 {
		t.Errorf("exit = %d, want the script's own 1", got)
	}
	if _, ran := h.scripts.ran(preflightScript); ran {
		t.Error("preflight ran past a failed ledger write")
	}
	if len(h.gh.calls) != 0 {
		t.Errorf("the forge was reached: %v", h.gh.calls)
	}
}

// Preflight non-zero: the pull request is not marked ready, and the
// stage's own code is what the frame reports (spec-0010, acceptance
// criteria).
func TestPreflightNonZeroNeverReachesTheForge(t *testing.T) {
	for _, code := range []int{1, 3, 4} {
		h := newHarness(t)
		h.scripts.replies[preflightScript] = reply{
			errOut: "PREFLIGHT STOPPED at 1/3 front matter — exit 1.\n",
			err:    scriptExit(code),
		}
		if got := exitOf(h.finish()); got != code {
			t.Errorf("exit = %d, want preflight's own %d", got, code)
		}
		if len(h.gh.calls) != 0 {
			t.Errorf("the forge was reached after a failed preflight: %v", h.gh.calls)
		}
		if len(h.term.Asked) != 0 {
			t.Errorf("asked %v; a failed preflight asks nothing", h.term.Asked)
		}
	}
}

// A no leaves nothing behind: not the forge, and not the two writes
// that ran before the question. The whole sentence, asserted over both
// files (spec-0017, acceptance criteria; product/pull-requests/shape.md).
func TestADeclineLeavesTheTreeAsItFoundIt(t *testing.T) {
	h := newHarness(t)
	h.term.ConfirmAnswer = false
	beforeTask, beforeSpec := h.read(t, taskPath), h.read(t, specPath)

	err := h.finish()
	if !errors.Is(err, command.ErrDeclined) {
		t.Fatalf("finish = %v, want ErrDeclined", err)
	}
	if h.gh.reached("pr ready") {
		t.Errorf("the pull request was marked ready after a no: %v", h.gh.calls)
	}
	if got := h.read(t, taskPath); got != beforeTask {
		t.Errorf("the task was left changed after a no:\n%s", got)
	}
	if got := h.read(t, specPath); got != beforeSpec {
		t.Errorf("the spec was left changed after a no:\n%s", got)
	}
	if !strings.Contains(h.out.String(), "restored "+specPath) ||
		!strings.Contains(h.out.String(), "restored "+taskPath) {
		t.Errorf("stdout does not report both files put back:\n%s", h.out.String())
	}
}

// The undo reaches every end that is not a success, not only the
// question: a refused ledger, a non-zero preflight, and a forge that
// will not answer all leave the queue as they found it (spec-0017,
// step 1).
func TestEveryFailureAfterTheWritesPutsThemBack(t *testing.T) {
	cases := []struct {
		name  string
		spoil func(*harness)
	}{
		{"the ledger refuses", func(h *harness) {
			h.scripts.replies[provenanceScript] = reply{errOut: "by= is required\n", err: scriptExit(1)}
		}},
		{"preflight is non-zero", func(h *harness) {
			h.scripts.replies[preflightScript] = reply{errOut: "PREFLIGHT STOPPED\n", err: scriptExit(3)}
		}},
		{"the forge will not answer", func(h *harness) {
			h.gh.viewErr = errors.New("gh pr view: no pull requests found for branch")
		}},
		{"the act itself fails", func(h *harness) {
			h.gh.readyErr = errors.New("gh pr ready: HTTP 403")
		}},
		{"the pull request is merged", func(h *harness) {
			h.gh.view = `{"number":45,"title":"x","state":"MERGED","isDraft":false}`
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(t)
			beforeTask, beforeSpec := h.read(t, taskPath), h.read(t, specPath)
			tc.spoil(h)
			if err := h.finish(); err == nil {
				t.Fatal("finish = nil; the case was supposed to fail")
			}
			if got := h.read(t, taskPath); got != beforeTask {
				t.Errorf("the task was left changed:\n%s", got)
			}
			if got := h.read(t, specPath); got != beforeSpec {
				t.Errorf("the spec was left changed:\n%s", got)
			}
		})
	}
}

// The script's own verdict survives the undo: a restored tree does not
// turn preflight's 3 into a 1.
func TestTheUndoCarriesTheScriptsVerdictUp(t *testing.T) {
	h := newHarness(t)
	h.scripts.replies[preflightScript] = reply{errOut: "PREFLIGHT STOPPED\n", err: scriptExit(3)}
	if got := exitOf(h.finish()); got != 3 {
		t.Errorf("exit = %d, want preflight's own 3", got)
	}
}

// A rerun after a no behaves as the first run did — the decline left
// nothing for a second run to report as already done (spec-0017,
// acceptance criteria).
func TestARerunAfterADeclineBehavesLikeTheFirstRun(t *testing.T) {
	h := newHarness(t)
	h.term.ConfirmAnswer = false
	if err := h.finish(); !errors.Is(err, command.ErrDeclined) {
		t.Fatalf("the first finish = %v, want ErrDeclined", err)
	}

	h.out.Reset()
	h.term.ConfirmAnswer = true
	if err := h.finish(); err != nil {
		t.Fatalf("the second finish = %v", err)
	}
	out := h.out.String()
	if strings.Contains(out, "unchanged") {
		t.Errorf("the rerun reported a file as already written:\n%s", out)
	}
	if !strings.Contains(out, "wrote completed: "+stamped+" on "+taskPath) {
		t.Errorf("the rerun did not write the completion:\n%s", out)
	}
	if got := queue.Field([]byte(h.read(t, specPath)), "status"); got != "implemented" {
		t.Errorf("spec status = %q, want implemented", got)
	}
	if !h.gh.reached("pr ready 45") {
		t.Error("the rerun never marked the pull request ready")
	}
}

// A task with no spec has one write, not two, and the same guarantee
// holds over it (spec-0017, edge cases).
func TestATaskWithNoSpecIsRestoredToo(t *testing.T) {
	h := newHarness(t)
	h.seed(taskPath, taskFixture("in-progress", "", "null"))
	h.term.ConfirmAnswer = false
	before := h.read(t, taskPath)
	if err := h.finish(); !errors.Is(err, command.ErrDeclined) {
		t.Fatalf("finish = %v, want ErrDeclined", err)
	}
	if got := h.read(t, taskPath); got != before {
		t.Errorf("the task was left changed after a no:\n%s", got)
	}
}

// Under --yes there is no question and so no decline: the writes stand
// and the pull request is marked ready (spec-0017, edge cases).
func TestYesLeavesTheWritesStanding(t *testing.T) {
	h := newHarness(t)
	h.term.In = false
	h.ctx.Yes = true
	if err := h.finish(); err != nil {
		t.Fatalf("finish = %v", err)
	}
	if got := queue.Field([]byte(h.read(t, specPath)), "status"); got != "implemented" {
		t.Errorf("spec status = %q; --yes undid a write it never questioned", got)
	}
	if got := queue.Field([]byte(h.read(t, taskPath)), "completed"); got != stamped {
		t.Errorf("completed = %q, want %q", got, stamped)
	}
	if strings.Contains(h.out.String(), "restored") {
		t.Errorf("a successful run put its own writes back:\n%s", h.out.String())
	}
}

// An undo the filesystem refuses says so and exits non-zero — never a
// quiet "declined — nothing changed" over a tree that did change
// (spec-0017, edge cases).
func TestARestoreThatFailsSaysSo(t *testing.T) {
	h := newHarness(t)
	h.term.ConfirmAnswer = false
	// Writable once, then not: the completion edit lands and the undo
	// cannot put it back.
	h.deniedAfterWrite(specPath, errors.New("read-only file system"))

	err := h.finish()
	if err == nil {
		t.Fatal("a failed restore exited 0")
	}
	if errors.Is(err, command.ErrDeclined) {
		t.Errorf("the failure still reads as a plain decline: %v", err)
	}
	if !strings.Contains(err.Error(), specPath) || !strings.Contains(err.Error(), "by hand") {
		t.Errorf("the failure does not name the file left changed: %v", err)
	}
	if got := queue.Field([]byte(h.read(t, specPath)), "status"); got != "implemented" {
		t.Errorf("spec status = %q; the fake did not leave the write standing", got)
	}
}

// The ledger's append is undone with everything else, on the path where
// the completion date was already there. The worker dates the task by
// hand — AGENTS.md says that is who writes it — so step 2 writes
// nothing to the task, and a journal that remembered only its own
// writes left the entry `record_provenance.sh` appended sitting in the
// tree under the words "declined — nothing changed" (spec-0017;
// product/pull-requests/shape.md).
func TestADeclineUndoesTheLedgerAppendWhenTheDateWasAlreadyThere(t *testing.T) {
	h := newHarness(t)
	h.seed(taskPath, taskFixture("in-progress", "spec-0010", "2026-09-01T00:00:00Z"))
	h.ledgerAppends("by: human, login: octocat")
	h.term.ConfirmAnswer = false
	beforeTask, beforeSpec := h.read(t, taskPath), h.read(t, specPath)

	if err := h.finish(); !errors.Is(err, command.ErrDeclined) {
		t.Fatalf("finish = %v, want ErrDeclined", err)
	}
	if got := h.read(t, taskPath); got != beforeTask {
		t.Errorf("the ledger entry was left behind after a no:\n%s", got)
	}
	if got := h.read(t, specPath); got != beforeSpec {
		t.Errorf("the spec was left changed after a no:\n%s", got)
	}
}

// The same undo on the path where the date was written by this run: the
// entry goes back with the date, because it records an act that did not
// happen (spec-0017; report-0017 carries the tension with
// record_provenance.sh's append-only contract to triage).
func TestADeclineUndoesTheLedgerAppendBesideTheDateItWrote(t *testing.T) {
	h := newHarness(t)
	h.ledgerAppends("by: agent, model: claude-opus-5, login: octocat")
	h.term.ConfirmAnswer = false
	beforeTask := h.read(t, taskPath)

	if err := h.finish(); !errors.Is(err, command.ErrDeclined) {
		t.Fatalf("finish = %v, want ErrDeclined", err)
	}
	if got := h.read(t, taskPath); got != beforeTask {
		t.Errorf("the task was left changed after a no:\n%s", got)
	}
	if strings.Contains(h.read(t, taskPath), "octocat") {
		t.Error("the ledger entry survived the undo")
	}
}

// A success keeps it. The undo fires on no path the pull request
// reached, so the entry the ledger appended stands with the two writes.
func TestASuccessKeepsTheLedgerAppend(t *testing.T) {
	h := newHarness(t)
	h.ledgerAppends("by: agent, model: claude-opus-5, login: octocat")

	if err := h.finish(); err != nil {
		t.Fatalf("finish = %v", err)
	}
	if !strings.Contains(h.read(t, taskPath), "octocat") {
		t.Errorf("a successful run undid the ledger entry:\n%s", h.read(t, taskPath))
	}
	if strings.Contains(h.out.String(), "restored") {
		t.Errorf("a successful run put its own writes back:\n%s", h.out.String())
	}
}

// A write that fails after truncating the file has still left
// something, and it is put back — the journal remembers the file before
// the write is attempted, not the write after it succeeded.
func TestAWriteThatMangledTheFileIsStillPutBack(t *testing.T) {
	h := newHarness(t)
	before := h.read(t, specPath)
	h.mangledOnWrite(specPath, errors.New("input/output error"))

	err := h.finish()
	if err == nil {
		t.Fatal("finish = nil; the write was supposed to fail")
	}
	if got := h.read(t, specPath); got != before {
		t.Errorf("the half-written spec was left in the tree:\n%s", got)
	}
	if h.gh.reached("pr ") {
		t.Error("the forge was reached after a failed write")
	}
}

// A file somebody else changed while `preflight.sh` ran is not this
// run's to revert: the undo leaves it, says so, and the failure names
// it rather than reporting a tree it did not put back.
func TestAFileChangedUnderTheRunIsLeftAlone(t *testing.T) {
	h := newHarness(t)
	h.scripts.replies[preflightScript] = reply{errOut: "PREFLIGHT STOPPED\n", err: scriptExit(3)}
	h.editedDuring(preflightScript, specPath, specFixture("implemented", "Rewritten in an editor while the gates ran."))

	err := h.finish()
	if err == nil {
		t.Fatal("finish = nil; preflight was supposed to fail")
	}
	if !strings.Contains(h.read(t, specPath), "Rewritten in an editor") {
		t.Error("the undo reverted an edit that was not this run's")
	}
	if !strings.Contains(err.Error(), specPath) || !strings.Contains(err.Error(), "by hand") {
		t.Errorf("the failure does not name the file left changed: %v", err)
	}
	if !strings.Contains(h.out.String(), "left "+specPath+" alone") {
		t.Errorf("the run did not say it left the file alone:\n%s", h.out.String())
	}
	// The task is this run's own and goes back regardless.
	if got := queue.Field([]byte(h.read(t, taskPath)), "completed"); got != "null" {
		t.Errorf("completed = %q, want null", got)
	}
}

// A file the undo cannot even read is a file it will not write over.
func TestAFileTheUndoCannotReadIsNotWrittenOver(t *testing.T) {
	h := newHarness(t)
	h.term.ConfirmAnswer = false
	h.scripts.replies[preflightScript] = reply{does: func() {
		h.files.Fail(path.Join(root, specPath), errors.New("read-only file system"))
	}}

	err := h.finish()
	if err == nil {
		t.Fatal("a restore that could not read its file exited 0")
	}
	if errors.Is(err, command.ErrDeclined) {
		t.Errorf("the failure still reads as a plain decline: %v", err)
	}
	if !strings.Contains(err.Error(), specPath) {
		t.Errorf("the failure does not name the file: %v", err)
	}
}

// --yes answers the question without a terminal, and the act happens.
func TestYesAnswersTheQuestion(t *testing.T) {
	h := newHarness(t)
	h.term.In = false
	h.ctx.Yes = true
	if err := h.finish(); err != nil {
		t.Fatalf("finish = %v", err)
	}
	if len(h.term.Asked) != 0 {
		t.Errorf("asked %v under --yes", h.term.Asked)
	}
	if !h.gh.reached("pr ready 45") {
		t.Error("the pull request was not marked ready under --yes")
	}
}

// Without a terminal and without --yes the question aborts naming the
// flag rather than hanging (product/rules.md).
func TestNoTerminalAbortsNamingTheFlag(t *testing.T) {
	h := newHarness(t)
	h.term.In = false
	err := h.finish()
	if err == nil || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("finish = %v; want an abort naming --yes", err)
	}
	if h.gh.reached("pr ready") {
		t.Error("the pull request was marked ready with nobody to ask")
	}
}

// A pull request already out of draft is a complete state, not a
// failure: nothing is asked and nothing is sent.
func TestAnAlreadyReadyPullRequestIsNotMarkedAgain(t *testing.T) {
	h := newHarness(t)
	h.gh.view = `{"number":45,"title":"[TASK-0011] Finish a task","state":"OPEN","isDraft":false}`
	if err := h.finish(); err != nil {
		t.Fatalf("finish = %v", err)
	}
	if h.gh.reached("pr ready") {
		t.Error("a pull request already ready for review was marked again")
	}
	if len(h.term.Asked) != 0 {
		t.Errorf("asked %v; there was nothing to confirm", h.term.Asked)
	}
	if !strings.Contains(h.out.String(), "already ready for review") {
		t.Errorf("stdout does not say so:\n%s", h.out.String())
	}
}

// A closed or merged pull request is named for what it is rather than
// pushed at.
func TestAClosedPullRequestIsRefused(t *testing.T) {
	h := newHarness(t)
	h.gh.view = `{"number":45,"title":"x","state":"MERGED","isDraft":false}`
	err := h.finish()
	if err == nil || !strings.Contains(err.Error(), "merged") {
		t.Fatalf("finish = %v; want a refusal naming the state", err)
	}
	if h.gh.reached("pr ready") {
		t.Error("a merged pull request was marked ready")
	}
}

// The forge's own words reach the user, whether the read or the act
// failed.
func TestTheForgesRefusalReachesTheUser(t *testing.T) {
	h := newHarness(t)
	h.gh.viewErr = errors.New("gh pr view: no pull requests found for branch")
	err := h.finish()
	if err == nil || !strings.Contains(err.Error(), "no pull requests found") {
		t.Fatalf("finish = %v; want gh's own words", err)
	}

	h = newHarness(t)
	h.gh.readyErr = errors.New("gh pr ready: HTTP 403")
	err = h.finish()
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("finish = %v; want gh's own words", err)
	}
}

// Unreadable JSON from the forge is a failure, never an empty pull
// request acted on as if it were real.
func TestUnreadablePullRequestJSONIsAFailure(t *testing.T) {
	h := newHarness(t)
	h.gh.view = "not json"
	if err := h.finish(); err == nil {
		t.Fatal("garbage from gh was accepted")
	}
	if h.gh.reached("pr ready") {
		t.Error("the forge was acted on from unreadable JSON")
	}
}

// A second run restamps nothing: the date the first run declared is the
// declaration, and the spec is already implemented.
func TestASecondRunRewritesNothing(t *testing.T) {
	h := newHarness(t)
	h.seed(taskPath, taskFixture("in-review", "spec-0010", "2026-09-01T09:00:00Z"))
	h.seed(specPath, specFixture("implemented", "What was built."))
	before := h.read(t, taskPath)
	if err := h.finish(); err != nil {
		t.Fatalf("finish = %v", err)
	}
	if h.read(t, taskPath) != before {
		t.Error("the completed date was restamped on a second run")
	}
	if !strings.Contains(h.out.String(), "unchanged") {
		t.Errorf("stdout does not report the file as unchanged:\n%s", h.out.String())
	}
}

// The task is the branch's when none is given, and the branch's
// spelling is preflight.sh's own.
func TestTheTaskIsInferredFromTheBranch(t *testing.T) {
	h := newHarness(t)
	if err := h.finish(); err != nil {
		t.Fatalf("finish = %v", err)
	}
	if got, _ := h.scripts.ran(preflightScript); !strings.HasPrefix(got, "task-0011 ") {
		t.Errorf("preflight args = %q; the branch's task was not inferred", got)
	}

	h = newHarness(t)
	h.git.branch = "report/a-finding"
	err := h.finish()
	if err == nil || !strings.Contains(err.Error(), "names no task") {
		t.Fatalf("finish = %v; want a refusal naming the branch", err)
	}
}

// An id the queue does not carry is a refusal naming what was looked
// for, never a run against nothing.
func TestAnUnknownTaskIsRefused(t *testing.T) {
	h := newHarness(t)
	err := h.finish("task-0099")
	if err == nil || !strings.Contains(err.Error(), "work/tasks") {
		t.Fatalf("finish = %v; want a refusal naming where it looked", err)
	}
	if len(h.scripts.calls) != 0 {
		t.Errorf("%d scripts ran for a task that is not there", len(h.scripts.calls))
	}
}

// Two ids is a usage error — finish completes one task.
func TestTwoTaskIdsAreRefused(t *testing.T) {
	h := newHarness(t)
	err := h.finish("task-0011", "task-0012")
	if err == nil || !strings.Contains(err.Error(), "takes one") {
		t.Fatalf("finish = %v; want a usage refusal", err)
	}
}

// The range: --range answers it, the pushed main is the default, the
// local main is the fallback, and neither is an abort naming the flag.
func TestTheRangeIsResolvedLikePreflightResolvesItsOwn(t *testing.T) {
	h := newHarness(t)
	if err := h.finish("--range", "HEAD~3...HEAD"); err != nil {
		t.Fatalf("finish = %v", err)
	}
	if got, _ := h.scripts.ran(deltasScript); got != "spec-0010 HEAD~3...HEAD" {
		t.Errorf("check_deltas args = %q; --range was not honoured", got)
	}

	h = newHarness(t)
	h.git.refs = map[string]bool{"refs/heads/main": true}
	if err := h.finish(); err != nil {
		t.Fatalf("finish = %v", err)
	}
	if got, _ := h.scripts.ran(deltasScript); got != "spec-0010 main...HEAD" {
		t.Errorf("check_deltas args = %q; the local main was not the fallback", got)
	}

	h = newHarness(t)
	h.git.refs = map[string]bool{}
	err := h.finish()
	if err == nil || !strings.Contains(err.Error(), "--range") {
		t.Fatalf("finish = %v; want an abort naming --range", err)
	}
}

// git failing to answer which branch this is stops the command with
// git's own words.
func TestAGitFailureIsReported(t *testing.T) {
	h := newHarness(t)
	h.git.err = errors.New("git rev-parse: not a git repository")
	err := h.finish()
	if err == nil || !strings.Contains(err.Error(), "not a git repository") {
		t.Fatalf("finish = %v; want git's own words", err)
	}
}

// A write the filesystem refuses is named for the file it was for, and
// nothing after it runs.
func TestAFailingWriteStopsTheSequence(t *testing.T) {
	h := newHarness(t)
	h.files.FailOp("write", root+"/"+specPath, errors.New("read-only file system"))
	err := h.finish()
	if err == nil || !strings.Contains(err.Error(), specPath) {
		t.Fatalf("finish = %v; want a failure naming the file", err)
	}
	if _, ran := h.scripts.ran(provenanceScript); ran {
		t.Error("the ledger was written past a failed spec write")
	}
}

// A `--flag=value` is one token, and the id beside it is still lifted
// out for the flag set.
func TestTheEqualsFormOfAFlagIsParsed(t *testing.T) {
	h := newHarness(t)
	if err := h.finish("--range=HEAD~1...HEAD", "task-0011"); err != nil {
		t.Fatalf("finish = %v", err)
	}
	if got, _ := h.scripts.ran(deltasScript); got != "spec-0010 HEAD~1...HEAD" {
		t.Errorf("check_deltas args = %q", got)
	}
}

// A spec_ref naming a spec the queue does not carry is a refusal, and
// it lands before anything is written.
func TestASpecTheQueueDoesNotCarryIsRefused(t *testing.T) {
	h := newHarness(t)
	h.seed(taskPath, taskFixture("in-progress", "spec-0099", "null"))
	err := h.finish()
	if err == nil || !strings.Contains(err.Error(), "work/specs") {
		t.Fatalf("finish = %v; want a refusal naming where it looked", err)
	}
	if got := queue.Field([]byte(h.read(t, taskPath)), "completed"); got != "null" {
		t.Errorf("completed = %q; it was written past a refusal", got)
	}
}

// A spec the filesystem will not read stops the command naming the
// file, never a spec read as empty and judged on that.
func TestASpecThatCannotBeReadIsRefused(t *testing.T) {
	h := newHarness(t)
	h.files.FailOp("read", root+"/"+specPath, errors.New("input/output error"))
	err := h.finish()
	if err == nil || !strings.Contains(err.Error(), specPath) {
		t.Fatalf("finish = %v; want a failure naming the spec", err)
	}
}

// A queue file whose front matter carries no field to write is refused
// rather than extended — the schema is the methodology's, not this
// command's to add to.
func TestAFileMissingTheFieldIsRefused(t *testing.T) {
	h := newHarness(t)
	h.seed(specPath, strings.Replace(specFixture("approved", "Built."), "status: approved\n", "", 1))
	err := h.finish()
	if err == nil || !strings.Contains(err.Error(), "status") {
		t.Fatalf("finish = %v; want a refusal naming the missing field", err)
	}
	if _, ran := h.scripts.ran(provenanceScript); ran {
		t.Error("the sequence went on past a file it could not write")
	}
}

// A task file with no front matter at all is refused: there is nothing
// to read an id out of, and nothing to write a date into.
func TestATaskWithNoFrontMatterIsRefused(t *testing.T) {
	h := newHarness(t)
	h.seed(taskPath, "# A task with no front matter\n")
	err := h.finish()
	if err == nil || !strings.Contains(err.Error(), "id") {
		t.Fatalf("finish = %v; want a refusal naming what is missing", err)
	}
	if len(h.scripts.calls) != 0 {
		t.Errorf("%d scripts ran against a file with no front matter", len(h.scripts.calls))
	}
}

// An unparseable flag is the flag package's refusal, not a run with a
// value nobody meant.
func TestAnUnknownFlagIsRefused(t *testing.T) {
	h := newHarness(t)
	if err := h.finish("--nonesuch"); err == nil {
		t.Fatal("an unknown flag was accepted")
	}
	if len(h.scripts.calls) != 0 {
		t.Errorf("%d scripts ran under an unknown flag", len(h.scripts.calls))
	}
}
