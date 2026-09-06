package finishcmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
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

// reply is one canned run: what the script wrote, how it ended, and
// what it did to the tree while it ran. `does` is what makes the
// ledger's append and an editor's save modellable here at all — a
// script writes to the filesystem from outside the vfs port, exactly as
// `record_provenance.sh` does, and the undo has to answer for it.
type reply struct {
	out    string
	errOut string
	err    error
	does   func()
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

func (f *fakeScripts) run(root string, stdout, stderr io.Writer, _ []string, script string, args ...string) error {
	f.calls = append(f.calls, call{root: root, script: script, args: args})
	r := f.replies[script]
	fmt.Fprint(stdout, r.out)
	fmt.Fprint(stderr, r.errOut)
	if r.does != nil {
		r.does()
	}
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

// truncateThenFail is `os.WriteFile` losing the disk after O_TRUNC: the
// file is left holding part of what was meant for it and the call
// reports the failure. The fail table cannot say this — it refuses
// before touching anything — and it is the state a journal that
// recorded only completed writes had nothing to put back over
// (spec-0017, edge cases).
type truncateThenFail struct {
	vfs.FS
	path string
	err  error
	seen bool
}

func (f *truncateThenFail) WriteFile(name string, data []byte, perm fs.FileMode) error {
	// Once. The disk that lost the completion write is not the state
	// the undo is being asked about — what is, is whether the journal
	// has anything to put back over what the failed write left.
	if name != f.path || f.seen {
		return f.FS.WriteFile(name, data, perm)
	}
	f.seen = true
	if err := f.FS.WriteFile(name, data[:len(data)/2], perm); err != nil {
		return err
	}
	return f.err
}

// syncWriter is a captured stream two goroutines may write to. A
// caught signal puts one on the guard's goroutine beside the command's
// own, and production answers that with an *os.File, which is safe for
// concurrent use; a bytes.Buffer standing in for one has to be too, or
// the race detector reads the fixture as the defect.
type syncWriter struct {
	lock *sync.Mutex
	buf  bytes.Buffer
}

func (w *syncWriter) Write(p []byte) (int, error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.buf.Write(p)
}

func (w *syncWriter) String() string {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.buf.String()
}

func (w *syncWriter) Reset() {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.buf.Reset()
}

// syncFS is the same answer for the fake tree: production's filesystem
// is the disk, which two goroutines may reach at once, and the fake
// standing in for it takes the fixture's lock so they may too.
type syncFS struct {
	vfs.FS
	lock *sync.Mutex
}

func (f *syncFS) ReadFile(name string) ([]byte, error) {
	f.lock.Lock()
	defer f.lock.Unlock()
	return f.FS.ReadFile(name)
}

func (f *syncFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	return f.FS.WriteFile(name, data, perm)
}

// harness is one finish: every port faked, the streams captured.
type harness struct {
	// lock is the fixture's own, held by everything both goroutines
	// touch: the two streams and the fake tree.
	lock    sync.Mutex
	scripts *fakeScripts
	files   *vfs.Fake
	// fs is what the command is wired to — the fake tree, unless a
	// case wrapped it. Seeds and reads go to files either way, so a
	// decorator never has to answer for the fixture.
	fs   vfs.FS
	git  *fakeGit
	gh   *fakeGh
	term *command.FakeTerminal
	died *deaths
	ctx  *command.Ctx
	out  syncWriter
	errb syncWriter
}

// deniedAfterWrite makes the second write to rel fail — the undo, when
// the first write was the completion edit.
func (h *harness) deniedAfterWrite(rel string, err error) {
	h.fs = &denyAfterFirstWrite{FS: h.files, path: path.Join(root, rel), err: err}
}

// mangledOnWrite makes the write to rel land half of its bytes and then
// fail.
func (h *harness) mangledOnWrite(rel string, err error) {
	h.fs = &truncateThenFail{FS: h.files, path: path.Join(root, rel), err: err}
}

// ledgerAppends makes `record_provenance.sh` do what it really does:
// append one entry to the task file, from outside the vfs port, after
// the completion writes and before the gates.
func (h *harness) ledgerAppends(entry string) {
	h.scripts.replies[provenanceScript] = reply{
		out: "appended to " + taskPath + ": " + entry + "\n",
		does: func() {
			h.lock.Lock()
			defer h.lock.Unlock()
			p := path.Join(root, taskPath)
			b, err := h.files.ReadFile(p)
			if err != nil {
				return
			}
			next := strings.Replace(string(b), "provenance: []\n", "provenance:\n  - {"+entry+"}\n", 1)
			_ = h.files.WriteFile(p, []byte(next), 0o644)
		},
	}
}

// editedDuring makes a script's run stand for the seconds a human had
// while it ran: the file is saved over, by somebody who is not this
// command.
func (h *harness) editedDuring(script, rel, content string) {
	r := h.scripts.replies[script]
	r.does = func() {
		h.lock.Lock()
		defer h.lock.Unlock()
		_ = h.files.WriteFile(path.Join(root, rel), []byte(content), 0o644)
	}
	h.scripts.replies[script] = r
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
		died: newDeaths(),
	}
	h.fs = h.files
	h.out.lock, h.errb.lock = &h.lock, &h.lock
	h.seed(taskPath, taskFixture("in-progress", "spec-0010", "null"))
	h.seed(specPath, specFixture("approved", "What was built, and what diverged."))
	h.ctx = &command.Ctx{Stdout: &h.out, Stderr: &h.errb, Terminal: h.term, Root: root, Adopted: true}
	return h
}

func (h *harness) seed(rel, content string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.files.Seed(path.Join(root, rel), []byte(content), 0o644)
}

func (h *harness) read(t *testing.T, rel string) string {
	t.Helper()
	h.lock.Lock()
	b, err := h.files.ReadFile(path.Join(root, rel))
	h.lock.Unlock()
	if err != nil {
		t.Fatalf("reading %s: %v", rel, err)
	}
	return string(b)
}

// deps is the wiring, with the clock stopped so the stamped date is
// assertable and the death faked, because the production one kills the
// test binary.
func (h *harness) deps() Deps {
	at, _ := time.Parse(time.RFC3339, stamped)
	return Deps{
		Scripts: h.scripts.run,
		Files:   &syncFS{FS: h.fs, lock: &h.lock},
		Git:     h.git.run,
		Gh:      h.gh.run,
		Now:     func() time.Time { return at },
		Die:     h.died.record,
	}
}

// raisingTerminal is the confirmation with a signal arriving while it
// holds the terminal — huh's own path, where bubbletea answers the
// signal and the form returns rather than the guard acting.
type raisingTerminal struct {
	command.Terminal
	t   *testing.T
	sig syscall.Signal
}

func (r *raisingTerminal) Confirm(question string) (bool, error) {
	raise(r.t, r.sig)
	return r.Terminal.Confirm(question)
}

// deaths records what the guard asked to die of. The production Die
// never returns; this one does, so the watcher's goroutine ends and a
// case can assert on the signal it carried.
type deaths struct {
	got  chan os.Signal
	seen atomic.Int32
}

func newDeaths() *deaths { return &deaths{got: make(chan os.Signal, 4)} }

func (d *deaths) record(sig os.Signal) {
	d.seen.Add(1)
	d.got <- sig
}

// waitFor is the signal the guard died of, or a failure naming what it
// waited for — a case never blocks the suite on a handler that did not
// fire.
func (d *deaths) waitFor(t *testing.T) os.Signal {
	t.Helper()
	select {
	case sig := <-d.got:
		return sig
	case <-time.After(10 * time.Second):
		t.Fatal("the guard never died of the signal")
		panic("unreachable")
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
