package statuscmd

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/kit"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// root is the repository every case reads.
const root = "/repo"

// taskFile and specFile are one queue entry as the methodology writes
// it — front matter, then the title the task line carries.
const taskFile = `---
id: task-0014
status: in-progress
spec_ref: [spec-0013]
completed: null
---

# Answer where the work stands from the current branch

Some body.
`

const specFile = `---
id: spec-0013
task_ref: task-0014
status: approved
---

# spec-0013 — Read the branch, the checks and the queue
`

// report is one report file at a status.
func report(id, status string) string {
	return fmt.Sprintf("---\nid: %s\nstatus: %s\n---\n\n# Something that was noticed\n", id, status)
}

// fixture is an adopted repository holding one task, its spec, the
// reports directory and a kit at the tag the client pins.
func fixture() *vfs.Fake {
	f := vfs.NewFake()
	f.Seed(root+"/.writrun/VERSION", []byte("v0.0.03\n"), 0o644)
	f.Seed(root+"/work/tasks/task-0014-status-command.md", []byte(taskFile), 0o644)
	f.Seed(root+"/work/specs/spec-0013-status-command.md", []byte(specFile), 0o644)
	f.Seed(root+"/work/reports/README.md", []byte("# Reports\n"), 0o644)
	return f
}

// onBranch is the git port faked down to the one question this command
// asks it.
func onBranch(branch string) gitx.Runner {
	return func(dir string, args ...string) (string, error) {
		if strings.Join(args, " ") != "rev-parse --abbrev-ref HEAD" {
			return "", fmt.Errorf("unexpected git %v", args)
		}
		return branch + "\n", nil
	}
}

// gitFails is a git that cannot answer at all.
func gitFails(err error) gitx.Runner {
	return func(dir string, args ...string) (string, error) { return "", err }
}

// call is what the exec port was asked for.
type call struct {
	root   string
	script string
	args   []string
	runs   int
}

// exitErr is a script's own verdict as the port hands it up.
type exitErr int

func (e exitErr) Error() string { return fmt.Sprintf("exit status %d", int(e)) }
func (e exitErr) ExitCode() int { return int(e) }

// scripts is the exec port faked: canned output, a canned verdict, and
// a record of what it was asked to run.
func scripts(out string, err error, rec *call) kit.Runner {
	return func(dir string, stdout, stderr io.Writer, name string, args ...string) error {
		if rec != nil {
			*rec = call{root: dir, script: name, args: args, runs: rec.runs + 1}
		}
		fmt.Fprint(stdout, out)
		return err
	}
}

// preflightOK and preflightStops are what the script prints on each of
// its two endings — its own words, so a case checks the reading rather
// than the wording.
const preflightOK = `== 1/3 front matter ==

PREFLIGHT OK — range origin/main...HEAD; deltas checked: spec-0013
`

const preflightStops = `== 1/3 front matter ==
work/tasks/task-0014-status-command.md: completed is not RFC3339

PREFLIGHT STOPPED at 1/3 front matter — exit 3. The stages after it did not run.
`

// field returns the text of the labelled line, or "" where the answer
// carries no such line.
func field(out, label string) string {
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, label+" ") || line == label {
			return strings.TrimSpace(strings.TrimPrefix(line, label))
		}
	}
	return ""
}

// wants fails unless the labelled line reads exactly want.
func wants(t *testing.T, out, label, want string) {
	t.Helper()
	if got := field(out, label); got != want {
		t.Errorf("%s = %q; want %q\nfull answer:\n%s", label, got, want, out)
	}
}
