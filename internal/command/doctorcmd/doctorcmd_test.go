package doctorcmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
)

// runDoctor is one whole run through the command's own entry point.
func runDoctor(t *testing.T, f *fixture, args ...string) (string, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	ctx := &command.Ctx{Stdout: &stdout, Stderr: &stderr, Root: f.root, Adopted: true}
	err := run(ctx, f.deps(), args)
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q; the report is one stream", stderr.String())
	}
	return stdout.String(), err
}

// exitCode is the code an error carries, or -1 for one carrying none —
// the reading the frame makes of a command's verdict.
func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var v interface{ ExitCode() int }
	if errors.As(err, &v) {
		return v.ExitCode()
	}
	return -1
}

func TestTheCommandIsDeclaredAsRequiringAnAdoptedRepository(t *testing.T) {
	c := New(Deps{})
	if c.Name != "doctor" {
		t.Errorf("name = %q, want doctor", c.Name)
	}
	if c.Need != command.NeedAdopted {
		t.Errorf("need = %v, want NeedAdopted", c.Need)
	}
	if c.Summary == "" {
		t.Error("summary is empty; --help prints one line per command")
	}
}

func TestEveryAssumptionHoldingExitsZero(t *testing.T) {
	f := newFixture(t, "3")
	out, err := runDoctor(t, f)
	if err != nil {
		t.Fatalf("run = %v (exit %d), want 0", err, exitCode(err))
	}
	for _, want := range []string{
		"Stage 3 is declared — stages 0–3 examined.",
		"Stage 0 — environment: all clear.",
		"Stage 1 — files: all clear.",
		"Stage 2 — the forge: all clear.",
		"Stage 3 — Issues: all clear.",
		"Every assumption up to stage 3 holds.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output misses %q:\n%s", want, out)
		}
	}
}

func TestAFindingThatBreaksAFlowExitsNonZero(t *testing.T) {
	f := newFixture(t, "1")
	f.path["sed"] = false
	out, err := runDoctor(t, f)
	if exitCode(err) != 1 {
		t.Fatalf("exit = %d, want 1", exitCode(err))
	}
	if !strings.Contains(out, "sed is not on the PATH") {
		t.Errorf("output does not name the missing requirement:\n%s", out)
	}
	if !strings.Contains(out, "1 finding(s): 1 breaking a flow") {
		t.Errorf("the summary does not agree with the exit status:\n%s", out)
	}
}

func TestARecommendationAloneExitsZero(t *testing.T) {
	f := newFixture(t, "3")
	f.forge.replies["api repos/{owner}/{repo}/rules/branches/main --jq .[].type"] = "\n"
	f.forge.replies["api repos/{owner}/{repo}/rules/branches/main --jq .[].ruleset_id"] = "\n"
	out, err := runDoctor(t, f)
	if err != nil {
		t.Fatalf("run = %v (exit %d), want 0 for a recommendation", err, exitCode(err))
	}
	if !strings.Contains(out, "advises  main is governed by no ruleset") {
		t.Errorf("the recommendation is not reported as one:\n%s", out)
	}
	if !strings.Contains(out, "none breaking a flow") {
		t.Errorf("the summary does not say nothing breaks:\n%s", out)
	}
}

func TestAnUnreadableForgeIsSaidAndExitsZero(t *testing.T) {
	f := newFixture(t, "3")
	f.forge.fails["auth status"] = errors.New("not logged in")
	out, err := runDoctor(t, f)
	if err != nil {
		t.Fatalf("run = %v (exit %d), want 0 for an unreachable forge", err, exitCode(err))
	}
	if !strings.Contains(out, "unread   gh is not authenticated") {
		t.Errorf("the stand-down is not reported:\n%s", out)
	}
	if !strings.Contains(out, "unread   whether Issues are enabled was not read") {
		t.Errorf("stage 3 does not say what it could not check:\n%s", out)
	}
}

// A stage above the declared one is named all the same: a clean report
// and an unexamined one must not read alike (spec-0004, edge cases).
func TestTheStandDownAboveTheDeclaredStageIsSaid(t *testing.T) {
	f := newFixture(t, "1")
	out, err := runDoctor(t, f)
	if err != nil {
		t.Fatalf("run = %v, want 0", err)
	}
	for _, want := range []string{
		"Stage 2 — the forge: not examined — the repository declares stage 1.",
		"Stage 3 — Issues: not examined — the repository declares stage 1.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output misses %q:\n%s", want, out)
		}
	}
}

// The report says what doctor is, on the line a reader starts at: no
// run of it repairs anything (product/adoption/doctor.md).
func TestTheReportSaysItRepairsNothing(t *testing.T) {
	f := newFixture(t, "3")
	out, _ := runDoctor(t, f)
	if !strings.Contains(out, "doctor reports; it repairs nothing.") {
		t.Errorf("output does not say what doctor is:\n%s", out)
	}
}

func TestAWrappedChecksOwnWordsAreCarriedUnderTheFinding(t *testing.T) {
	f := newFixture(t, "1")
	f.scripts.verdict[frontMatterScript] = exitErr(1)
	f.scripts.said[frontMatterScript] = "REJECTED: work/tasks/task-0001.md\nREJECTED: work/specs/spec-0001.md\n"
	out, err := runDoctor(t, f)
	if exitCode(err) != 1 {
		t.Fatalf("exit = %d, want 1", exitCode(err))
	}
	for _, want := range []string{
		"           | REJECTED: work/tasks/task-0001.md",
		"           | REJECTED: work/specs/spec-0001.md",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output misses %q:\n%s", want, out)
		}
	}
}

func TestDoctorTakesNoArgument(t *testing.T) {
	f := newFixture(t, "1")
	cases := []struct{ name, arg string }{
		{"a stage", "2"},
		{"a fix flag", "--fix"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := runDoctor(t, f, c.arg)
			if err == nil {
				t.Fatalf("run(%q) = nil; want a refusal", c.arg)
			}
			if exitCode(err) > 0 {
				t.Errorf("the refusal carries exit %d; the frame should report it", exitCode(err))
			}
		})
	}
}

// The exit status is the only thing doctor's own error carries: the
// findings are on stdout, so the frame prints nothing over them.
func TestTheVerdictCarriesACodeAndNothingToRestate(t *testing.T) {
	if got := exitCode(verdict(1)); got != 1 {
		t.Errorf("exit = %d, want 1", got)
	}
	if verdict(1).Error() != "exit status 1" {
		t.Errorf("error = %q", verdict(1).Error())
	}
}

func TestDetailLinesDropASpacingOnlyReport(t *testing.T) {
	if got := detailLines("\n\n  \n"); got != nil {
		t.Errorf("detailLines = %q, want none", got)
	}
	if got := detailLines("\none\ntwo\n"); len(got) != 2 {
		t.Errorf("detailLines = %q, want the two lines", got)
	}
}
