package finishcmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

const (
	root     = "/repo"
	taskPath = "work/tasks/task-0011-finish-command.md"
	specPath = "work/specs/spec-0010-finish-command.md"
	// stamped is the instant the fake clock stands at, so a test can
	// assert the exact date the completion wrote.
	stamped = "2026-09-05T11:22:33Z"
)

// scriptExit is a script's own verdict, shaped like the *exec.ExitError
// the production runner returns: an error carrying an exit code.
type scriptExit int

func (e scriptExit) Error() string { return fmt.Sprintf("exit status %d", int(e)) }
func (e scriptExit) ExitCode() int { return int(e) }

// reply is one canned run: what the script wrote, and how it ended.
type reply struct {
	out    string
	errOut string
	err    error
}

// call is one invocation the fake recorded.
type call struct {
	root   string
	script string
	args   []string
}

// fakeScripts is the fake beside the kit.Runner port. The replies are
// keyed by script, not ordered, because finish runs three different
// authorities and a test cares which one spoke.
type fakeScripts struct {
	replies map[string]reply
	calls   []call
}

func (f *fakeScripts) run(root string, stdout, stderr io.Writer, script string, args ...string) error {
	f.calls = append(f.calls, call{root: root, script: script, args: args})
	r := f.replies[script]
	fmt.Fprint(stdout, r.out)
	fmt.Fprint(stderr, r.errOut)
	return r.err
}

// ran reports whether a script was invoked, and with which arguments.
func (f *fakeScripts) ran(script string) (string, bool) {
	for _, c := range f.calls {
		if c.script == script {
			return strings.Join(c.args, " "), true
		}
	}
	return "", false
}

// fakeGit answers the two questions finish asks git: which branch this
// is, and which bases exist.
type fakeGit struct {
	branch string
	refs   map[string]bool
	err    error
	calls  []string
}

func (g *fakeGit) run(dir string, args ...string) (string, error) {
	g.calls = append(g.calls, strings.Join(args, " "))
	if g.err != nil {
		return "", g.err
	}
	switch {
	case len(args) > 1 && args[0] == "rev-parse" && args[1] == "--abbrev-ref":
		return g.branch + "\n", nil
	case len(args) > 1 && args[0] == "rev-parse" && args[1] == "--verify":
		ref := args[len(args)-1]
		if g.refs[ref] {
			return ref + "\n", nil
		}
		return "", errors.New("git rev-parse --verify: unknown revision")
	}
	return "", nil
}

// fakeGh is the forge, stubbed: what `pr view` answers, and whether
// `pr ready` was reached.
type fakeGh struct {
	view     string
	viewErr  error
	readyErr error
	calls    []string
}

func (g *fakeGh) run(args ...string) (string, error) {
	g.calls = append(g.calls, strings.Join(args, " "))
	switch {
	case len(args) > 1 && args[0] == "pr" && args[1] == "view":
		return g.view, g.viewErr
	case len(args) > 1 && args[0] == "pr" && args[1] == "ready":
		return "", g.readyErr
	}
	return "", nil
}

func (g *fakeGh) reached(what string) bool {
	for _, c := range g.calls {
		if strings.HasPrefix(c, what) {
			return true
		}
	}
	return false
}

// taskFixture is a queue task, its fields spelled the way the schema
// spells them.
func taskFixture(status, specRef, completed string) string {
	return "---\n" +
		"id: task-0011\n" +
		"status: " + status + "\n" +
		"blocked_reason: null\n" +
		"taken_by: null\n" +
		"spec_ref: [" + specRef + "]\n" +
		"doc_ref: null\n" +
		"origin: rule\n" +
		"priority: medium\n" +
		"depends_on: []\n" +
		"milestone: null\n" +
		"created: 2026-09-03T22:30:23Z\n" +
		"queued: 2026-09-04T05:05:03Z\n" +
		"completed: " + completed + "\n" +
		"merged: null\n" +
		"provenance: []\n" +
		"---\n\n" +
		"# Finish a task with writrun finish\n\n" +
		"One paragraph of brief. It quotes `status: done` nowhere at column 0.\n"
}

// specFixture is a queue spec whose Outcome says whatever the case
// needs it to say.
func specFixture(status, outcome string) string {
	return "---\n" +
		"id: spec-0010\n" +
		"task_ref: task-0011\n" +
		"status: " + status + "\n" +
		"created: 2026-09-03T22:30:42Z\n" +
		"---\n\n" +
		"# spec-0010 — the contract\n\n" +
		"- **Goal:** something the task implements.\n\n" +
		"## Proposed product changes\n\n- none\n\n" +
		"## Outcome\n\n" + outcome + "\n"
}

// draftPR is what `gh pr view --json …` answers for this branch's open
// draft.
const draftPR = `{"number":45,"title":"[TASK-0011] Finish a task","state":"OPEN","isDraft":true}`

// denyAfterFirstWrite lets one write to a path land and refuses every
// later one. The fail table can say "this path is not writable", which
// is a different state: what a failed undo needs is a write that
// succeeded and a put-back that cannot, and only a decorator can say
// that (spec-0017, edge cases).
type denyAfterFirstWrite struct {
	vfs.FS
	path string
	err  error
	seen bool
}

func (f *denyAfterFirstWrite) WriteFile(name string, data []byte, perm fs.FileMode) error {
	if name == f.path {
		if f.seen {
			return f.err
		}
		f.seen = true
	}
	return f.FS.WriteFile(name, data, perm)
}

// harness is one finish: every port faked, the streams captured.
type harness struct {
	scripts *fakeScripts
	files   *vfs.Fake
	// fs is what the command is wired to — the fake tree, unless a
	// case wrapped it. Seeds and reads go to files either way, so a
	// decorator never has to answer for the fixture.
	fs   vfs.FS
	git  *fakeGit
	gh   *fakeGh
	term *command.FakeTerminal
	ctx  *command.Ctx
	out  bytes.Buffer
	errb bytes.Buffer
}

// deniedAfterWrite makes the second write to rel fail — the undo, when
// the first write was the completion edit.
func (h *harness) deniedAfterWrite(rel string, err error) {
	h.fs = &denyAfterFirstWrite{FS: h.files, path: path.Join(root, rel), err: err}
}

// newHarness is the green path, ready to be spoiled one port at a time:
// every script exits 0, the branch names task-0011, origin/main is
// there, the pull request is an open draft, and the human says yes.
func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{
		scripts: &fakeScripts{replies: map[string]reply{}},
		files:   vfs.NewFake(),
		git: &fakeGit{
			branch: "task/0011-finish-command",
			refs:   map[string]bool{"refs/remotes/origin/main": true},
		},
		gh:   &fakeGh{view: draftPR},
		term: &command.FakeTerminal{In: true, ConfirmAnswer: true},
	}
	h.fs = h.files
	h.seed(taskPath, taskFixture("in-progress", "spec-0010", "null"))
	h.seed(specPath, specFixture("approved", "What was built, and what diverged."))
	h.ctx = &command.Ctx{Stdout: &h.out, Stderr: &h.errb, Terminal: h.term, Root: root, Adopted: true}
	return h
}

func (h *harness) seed(rel, content string) {
	h.files.Seed(path.Join(root, rel), []byte(content), 0o644)
}

func (h *harness) read(t *testing.T, rel string) string {
	t.Helper()
	b, err := h.files.ReadFile(path.Join(root, rel))
	if err != nil {
		t.Fatalf("reading %s: %v", rel, err)
	}
	return string(b)
}

// deps is the wiring, with the clock stopped so the stamped date is
// assertable.
func (h *harness) deps() Deps {
	at, _ := time.Parse(time.RFC3339, stamped)
	return Deps{
		Scripts: h.scripts.run,
		Files:   h.fs,
		Git:     h.git.run,
		Gh:      h.gh.run,
		Now:     func() time.Time { return at },
	}
}

func (h *harness) finish(args ...string) error {
	return run(h.ctx, h.deps(), args)
}

// exitOf is the exit code the frame would report for err — the same
// read internal/command.Run makes on a wrapped script's verdict.
func exitOf(err error) int {
	if err == nil {
		return 0
	}
	var verdict interface{ ExitCode() int }
	if errors.As(err, &verdict) && verdict.ExitCode() > 0 {
		return verdict.ExitCode()
	}
	return 1
}

// statusLineOf is the task's `status:` line, the one line this command
// may never write.
func statusLineOf(content string) string {
	for _, l := range strings.Split(content, "\n") {
		if strings.HasPrefix(l, "status:") {
			return l
		}
	}
	return ""
}
