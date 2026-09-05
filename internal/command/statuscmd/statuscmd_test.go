package statuscmd

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/kit"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// ask runs the command over one fixture and returns what it answered.
func ask(t *testing.T, files vfs.FS, git gitx.Runner, checks kit.Runner, args ...string) (string, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	ctx := &command.Ctx{Stdout: &stdout, Stderr: &stderr, Root: root}
	if checks == nil {
		checks = scripts(preflightOK, nil, nil)
	}
	err := run(ctx, Deps{Tag: "v0.0.03", Git: git, Files: files, Scripts: checks}, args)
	if stderr.Len() > 0 {
		t.Errorf("stderr = %q; status answers on stdout", stderr.String())
	}
	return stdout.String(), err
}

func TestOnATaskBranchTheAnswerNamesTheTaskItsSpecAndTheChecks(t *testing.T) {
	var rec call
	out, err := ask(t, fixture(), onBranch("task/0014-status-command"), scripts(preflightOK, nil, &rec))
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	wants(t, out, "Branch", "task/0014-status-command")
	wants(t, out, "Task", "task-0014  in-progress  Answer where the work stands from the current branch")
	wants(t, out, "Spec", "spec-0013  approved")
	wants(t, out, "Checks", "all pass")
	wants(t, out, "Reports", "none open")
	wants(t, out, "Kit", "WritRun v0.0.03 — the tag this client pins")
}

func TestTheChecksAreTheRepositorysOwnPreflightRunFromTheRoot(t *testing.T) {
	var rec call
	if _, err := ask(t, fixture(), onBranch("task/0014-status-command"), scripts(preflightOK, nil, &rec)); err != nil {
		t.Fatalf("run = %v", err)
	}
	if rec.root != root {
		t.Errorf("root = %q; want the repository root", rec.root)
	}
	if rec.script != ".writrun/scripts/stage-1-tasks-and-specs/preflight.sh" {
		t.Errorf("script = %q; want the completion gates", rec.script)
	}
	if len(rec.args) != 0 {
		t.Errorf("args = %v; want the script's own defaults", rec.args)
	}
	if rec.runs != 1 {
		t.Errorf("runs = %d; want the checks run once", rec.runs)
	}
}

func TestAFailingCheckIsNamedInTheScriptsOwnWords(t *testing.T) {
	out, err := ask(t, fixture(), onBranch("task/0014-status-command"), scripts(preflightStops, exitErr(3), nil))
	if err != nil {
		t.Fatalf("run = %v; a failing check is the answer, not a failure of status", err)
	}
	got := field(out, "Checks")
	if !strings.Contains(got, "1/3 front matter") {
		t.Errorf("Checks = %q; want the failing stage named", got)
	}
	if !strings.Contains(got, "exit 3") {
		t.Errorf("Checks = %q; want the failure named", got)
	}
}

func TestPreflightsOwnRefusalIsNamedWithItsCode(t *testing.T) {
	own := "PREFLIGHT: task id '99' resolves to no file under work/tasks/\n"
	out, err := ask(t, fixture(), onBranch("task/0014-status-command"), scripts(own, exitErr(4), nil))
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	got := field(out, "Checks")
	if !strings.Contains(got, "resolves to no file") || !strings.Contains(got, "4") {
		t.Errorf("Checks = %q; want preflight's own refusal and its code", got)
	}
}

func TestAScriptThatCouldNotRunIsSaidSo(t *testing.T) {
	out, err := ask(t, fixture(), onBranch("task/0014-status-command"), scripts("", errors.New("bash is not installed"), nil))
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	if got := field(out, "Checks"); !strings.Contains(got, "bash is not installed") {
		t.Errorf("Checks = %q; want the cause preserved", got)
	}
}

func TestOffATaskBranchTheChecksAreSkippedAndTheRestIsStillAnswered(t *testing.T) {
	files := fixture()
	files.Seed(root+"/work/reports/report-0007-a-thing.md", []byte(report("report-0007", "open")), 0o644)
	var rec call
	out, err := ask(t, files, onBranch("main"), scripts(preflightOK, nil, &rec))
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	wants(t, out, "Branch", "main")
	wants(t, out, "Task", "none — this branch carries no task")
	wants(t, out, "Checks", "")
	wants(t, out, "Spec", "")
	wants(t, out, "Reports", "1 open, waiting to be triaged")
	wants(t, out, "Kit", "WritRun v0.0.03 — the tag this client pins")
	if rec.runs != 0 {
		t.Errorf("the checks ran on a branch carrying no task")
	}
}

func TestADetachedHeadIsSaidPlainly(t *testing.T) {
	out, err := ask(t, fixture(), onBranch("HEAD"), nil)
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	wants(t, out, "Branch", "detached HEAD — no branch")
	wants(t, out, "Task", "none — this branch carries no task")
}

func TestABranchNamingATaskTheQueueDoesNotHoldIsNamedAsUnknown(t *testing.T) {
	var rec call
	out, err := ask(t, fixture(), onBranch("task/0099-invented"), scripts(preflightOK, nil, &rec))
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	wants(t, out, "Task", "task-0099 — the queue holds no such task")
	wants(t, out, "Spec", "")
	if rec.runs != 0 {
		t.Errorf("the checks ran for a task nothing resolved to")
	}
}

func TestASpecTheQueueDoesNotHoldIsNamedNeverInvented(t *testing.T) {
	files := fixture()
	files.Seed(root+"/work/specs/spec-0013-status-command.md", []byte("no front matter here\n"), 0o644)
	out, err := ask(t, files, onBranch("task/0014-status-command"), nil)
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	wants(t, out, "Spec", "spec-0013  no status")

	files = vfs.NewFake()
	files.Seed(root+"/.writrun/VERSION", []byte("v0.0.03\n"), 0o644)
	files.Seed(root+"/work/tasks/task-0014-status-command.md", []byte(taskFile), 0o644)
	out, err = ask(t, files, onBranch("task/0014-status-command"), nil)
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	wants(t, out, "Spec", "spec-0013 — no file under work/specs/")
}

func TestATaskNamingNoSpecSaysSo(t *testing.T) {
	files := fixture()
	files.Seed(root+"/work/tasks/task-0014-status-command.md",
		[]byte(strings.Replace(taskFile, "spec_ref: [spec-0013]", "spec_ref: []", 1)), 0o644)
	out, err := ask(t, files, onBranch("task/0014-status-command"), nil)
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	wants(t, out, "Spec", "none — this task names no spec")
}

func TestEverySpecTheTaskNamesIsAnswered(t *testing.T) {
	files := fixture()
	files.Seed(root+"/work/tasks/task-0014-status-command.md",
		[]byte(strings.Replace(taskFile, "spec_ref: [spec-0013]", "spec_ref: [spec-0013, spec-0021]", 1)), 0o644)
	files.Seed(root+"/work/specs/spec-0021-another.md",
		[]byte("---\nid: spec-0021\nstatus: draft\n---\n\n# spec-0021\n"), 0o644)
	out, err := ask(t, files, onBranch("task/0014-status-command"), nil)
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	if !strings.Contains(out, "spec-0013  approved") || !strings.Contains(out, "spec-0021  draft") {
		t.Errorf("both specs were not answered:\n%s", out)
	}
}

func TestOpenReportsAreCountedAndNothingElseIs(t *testing.T) {
	files := fixture()
	files.Seed(root+"/work/reports/report-0001-a.md", []byte(report("report-0001", "open")), 0o644)
	files.Seed(root+"/work/reports/report-0002-b.md", []byte(report("report-0002", "open")), 0o644)
	files.Seed(root+"/work/reports/report-0003-c.md", []byte(report("report-0003", "declined")), 0o644)
	files.Seed(root+"/work/reports/report-0004-d.md", []byte(report("report-0004", "tracked")), 0o644)
	out, err := ask(t, files, onBranch("main"), nil)
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	wants(t, out, "Reports", "2 open, waiting to be triaged")
}

func TestAReportsDirectoryThatCannotBeReadIsNamed(t *testing.T) {
	files := fixture()
	files.FailOp("walk", root+"/work/reports", errors.New("permission denied"))
	out, err := ask(t, files, onBranch("main"), nil)
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	if got := field(out, "Reports"); !strings.Contains(got, "could not be read") {
		t.Errorf("Reports = %q; want the directory named as unreadable", got)
	}
}

func TestAKitAtAnotherTagNamesBothValues(t *testing.T) {
	files := fixture()
	files.Seed(root+"/.writrun/VERSION", []byte("v0.0.02\n"), 0o644)
	out, err := ask(t, files, onBranch("main"), nil)
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	got := field(out, "Kit")
	if !strings.Contains(got, "v0.0.02") || !strings.Contains(got, "v0.0.03") {
		t.Errorf("Kit = %q; want both tags named", got)
	}
}

func TestAKitVersionThatCannotBeReadIsNamed(t *testing.T) {
	for _, tc := range []struct{ name, seed string }{
		{"empty", "  \n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := fixture()
			files.Seed(root+"/.writrun/VERSION", []byte(tc.seed), 0o644)
			out, err := ask(t, files, onBranch("main"), nil)
			if err != nil {
				t.Fatalf("run = %v", err)
			}
			if got := field(out, "Kit"); !strings.Contains(got, ".writrun/VERSION") || !strings.Contains(got, "v0.0.03") {
				t.Errorf("Kit = %q; want the file named and the pinned tag kept", got)
			}
		})
	}

	files := vfs.NewFake()
	files.SeedDir(root + "/.writrun")
	files.Seed(root+"/work/reports/README.md", []byte("# Reports\n"), 0o644)
	out, err := ask(t, files, onBranch("main"), nil)
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	if got := field(out, "Kit"); !strings.Contains(got, "could not be read") {
		t.Errorf("Kit = %q; want the missing file named", got)
	}
}

func TestABranchGitCannotNameIsAnError(t *testing.T) {
	boom := errors.New("not a git repository")
	_, err := ask(t, fixture(), gitFails(boom), nil)
	if !errors.Is(err, boom) {
		t.Errorf("err = %v; want git's own words", err)
	}
}

func TestNothingIsWritten(t *testing.T) {
	files := fixture()
	before := files.Paths()
	if _, err := ask(t, files, onBranch("task/0014-status-command"), nil); err != nil {
		t.Fatalf("run = %v", err)
	}
	if after := files.Paths(); !reflect.DeepEqual(before, after) {
		t.Errorf("the tree changed:\nbefore %v\nafter  %v", before, after)
	}
}

func TestAnUnexpectedArgumentIsRefused(t *testing.T) {
	_, err := ask(t, fixture(), onBranch("main"), nil, "task-0014")
	if err == nil || !strings.Contains(err.Error(), "task-0014") {
		t.Errorf("err = %v; want the argument named", err)
	}
}

func TestAnUnknownFlagIsRefused(t *testing.T) {
	if _, err := ask(t, fixture(), onBranch("main"), nil, "--everything"); err == nil {
		t.Fatal("an unknown flag was accepted")
	}
}

func TestNewDeclaresTheCommand(t *testing.T) {
	files := fixture()
	c := New(Deps{Tag: "v0.0.03", Git: onBranch("main"), Files: files, Scripts: scripts(preflightOK, nil, nil)})
	if c.Name != "status" {
		t.Errorf("name = %q", c.Name)
	}
	if c.Need != command.NeedAdopted {
		t.Errorf("need = %v; want an adopted repository", c.Need)
	}
	if c.Summary == "" {
		t.Error("no summary for --help")
	}
	var stdout bytes.Buffer
	if err := c.Run(&command.Ctx{Stdout: &stdout, Stderr: io.Discard, Root: root}, nil); err != nil {
		t.Errorf("Run = %v", err)
	}
	if !strings.HasPrefix(stdout.String(), "Branch") {
		t.Errorf("the wired command answered %q", stdout.String())
	}
}
