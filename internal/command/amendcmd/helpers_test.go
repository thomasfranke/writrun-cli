package amendcmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

const (
	root       = "/repo"
	specPath   = "work/specs/spec-0011-amend-command.md"
	taskPath   = "work/tasks/task-0012-amend-command.md"
	otherTask  = "work/tasks/task-0013-another-thing.md"
	tmplPath   = ".writrun/templates/pull_request_template.md"
	amendTitle = "Reopen the amendment gate"
)

// scriptExit is a script's own verdict, shaped like the *exec.ExitError
// the production runner returns: an error carrying an exit code.
type scriptExit int

func (e scriptExit) Error() string { return fmt.Sprintf("exit status %d", int(e)) }
func (e scriptExit) ExitCode() int { return int(e) }

// fakeScripts is the fake beside the kit.Runner port. amend runs two
// scripts — the settings reader and the door — so both the reply and
// the verdict are keyed by script and a test can spoil one without
// spoiling the other.
type fakeScripts struct {
	replies map[string]string
	// fail is what a named script returns instead of running.
	fail  map[string]error
	calls []string
	// env records what each script was handed through the environment,
	// keyed by script: the seam this fake exists to witness.
	env map[string][]string
}

func (f *fakeScripts) run(_ string, stdout, _ io.Writer, env []string, script string, args ...string) error {
	f.calls = append(f.calls, strings.Join(append([]string{script}, args...), " "))
	if f.env == nil {
		f.env = map[string][]string{}
	}
	f.env[script] = env
	if err := f.fail[script]; err != nil {
		return err
	}
	fmt.Fprint(stdout, f.replies[strings.Join(args, " ")])
	return nil
}

// handed is one variable as the named script received it, and whether
// it was there at all.
func (f *fakeScripts) handed(script, key string) (string, bool) {
	for _, e := range f.env[script] {
		if strings.HasPrefix(e, key+"=") {
			return strings.TrimPrefix(e, key+"="), true
		}
	}
	return "", false
}

// fakeGit records every invocation and answers the four questions amend
// asks: is the tree dirty, does a ref exist, and the two acts.
type fakeGit struct {
	dirty  string
	refs   map[string]bool
	fail   map[string]error
	calls  []string
	onSwap func()
}

func (g *fakeGit) run(_ string, args ...string) (string, error) {
	joined := strings.Join(args, " ")
	g.calls = append(g.calls, joined)
	for prefix, err := range g.fail {
		if strings.HasPrefix(joined, prefix) {
			return "", err
		}
	}
	switch {
	case len(args) > 1 && args[0] == "status":
		return g.dirty, nil
	case len(args) > 2 && args[0] == "rev-parse" && args[1] == "--verify":
		ref := args[len(args)-1]
		if g.refs[ref] {
			return ref + "\n", nil
		}
		return "", errors.New("git rev-parse --verify: unknown revision")
	case len(args) > 1 && args[0] == "switch":
		if g.onSwap != nil {
			g.onSwap()
		}
	}
	return "", nil
}

func (g *fakeGit) ran(prefix string) bool {
	for _, c := range g.calls {
		if strings.HasPrefix(c, prefix) {
			return true
		}
	}
	return false
}

func (g *fakeGit) arg(prefix string) string {
	for _, c := range g.calls {
		if strings.HasPrefix(c, prefix) {
			return c
		}
	}
	return ""
}

// fakeGh is the forge, stubbed: what `pr list` answers, and what `pr
// create` was handed.
type fakeGh struct {
	list      string
	listErr   error
	createErr error
	calls     [][]string
}

func (g *fakeGh) run(args ...string) (string, error) {
	g.calls = append(g.calls, args)
	switch {
	case len(args) > 1 && args[0] == "pr" && args[1] == "list":
		return g.list, g.listErr
	case len(args) > 1 && args[0] == "pr" && args[1] == "create":
		return "https://forge/pull/99\n", g.createErr
	}
	return "", nil
}

func (g *fakeGh) reached(what string) bool {
	for _, c := range g.calls {
		if strings.HasPrefix(strings.Join(c, " "), what) {
			return true
		}
	}
	return false
}

// created returns the arguments of the one `gh pr create`, keyed by
// flag, so a test asks for the title or the body by name.
func (g *fakeGh) created() map[string]string {
	out := map[string]string{}
	for _, c := range g.calls {
		if len(c) < 2 || c[0] != "pr" || c[1] != "create" {
			continue
		}
		for i := 2; i < len(c); i++ {
			if strings.HasPrefix(c[i], "--") && i+1 < len(c) && !strings.HasPrefix(c[i+1], "--") {
				out[c[i]] = c[i+1]
				i++
				continue
			}
			out[c[i]] = ""
		}
	}
	return out
}

// specFixture is a queue spec, its fields spelled the way the schema
// spells them.
func specFixture(status string) string {
	return "---\n" +
		"id: spec-0011\n" +
		"task_ref: task-0012\n" +
		"status: " + status + "\n" +
		"created: 2026-09-03T22:30:43Z\n" +
		"---\n\n" +
		"# spec-0011 — the contract\n\n" +
		"- **Goal:** something the task implements.\n\n" +
		"A body paragraph. It spells `status: draft` nowhere at column 0.\n"
}

// taskFixture is a queue task, in whatever state and referencing
// whatever spec the case needs.
func taskFixture(id, status, specRef string) string {
	return "---\n" +
		"id: " + id + "\n" +
		"status: " + status + "\n" +
		"blocked_reason: null\n" +
		"taken_by: null\n" +
		"spec_ref: [" + specRef + "]\n" +
		"doc_ref: null\n" +
		"origin: rule\n" +
		"priority: medium\n" +
		"depends_on: []\n" +
		"milestone: null\n" +
		"created: 2026-09-03T22:30:24Z\n" +
		"queued: 2026-09-04T05:05:03Z\n" +
		"completed: null\n" +
		"merged: null\n" +
		"provenance: []\n" +
		"---\n\n" +
		"# " + id + "\n\nOne paragraph of brief.\n"
}

// harness is one amend: every port faked, the streams captured.
type harness struct {
	scripts *fakeScripts
	files   *vfs.Fake
	git     *fakeGit
	gh      *fakeGh
	term    *command.FakeTerminal
	env     map[string]string
	ctx     *command.Ctx
	out     bytes.Buffer
	errb    bytes.Buffer
}

// newHarness is the green path, ready to be spoiled one port at a time:
// spec-0011 is approved, task-0012 is in flight on #42, the branch is
// free, the declared style is bracketed, and the human says yes.
func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{
		scripts: &fakeScripts{
			replies: map[string]string{"stage_2.pr_title_style": "bracketed\n"},
			fail:    map[string]error{},
		},
		files: vfs.NewFake(),
		git:   &fakeGit{refs: map[string]bool{"refs/remotes/origin/main": true}, fail: map[string]error{}},
		gh:    &fakeGh{},
		term:  &command.FakeTerminal{In: true, ConfirmAnswer: true},
		env:   map[string]string{},
	}
	h.seed(specPath, specFixture("approved"))
	h.seed(taskPath, taskFixture("task-0012", "in-progress", "spec-0011"))
	h.env["WRITRUN_PR_LIST"] = "42\ttask/0012-amend-command\tsomeone\t[TASK-0012] Amend the thing"
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

func (h *harness) deps() Deps {
	return Deps{
		Scripts: h.scripts.run,
		Files:   h.files,
		Git:     h.git.run,
		Gh:      h.gh.run,
		Getenv:  func(k string) string { return h.env[k] },
	}
}

// amend runs the command with the two answers a green path needs
// already given.
func (h *harness) amend(args ...string) error {
	return run(h.ctx, h.deps(), append([]string{"spec-0011", "--title", amendTitle}, args...))
}
