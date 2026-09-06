package authorcmd

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
	root = "/repo"
	// authorBranch carries more subject words than `normalize` keeps.
	// A one-word fixture cannot tell a slug that round-trips from one
	// that is silently re-composed into a different branch.
	authorBranch = "docs/the-declaration-is-a-section"
	newTaskPath  = "work/tasks/task-0016-declare-derived-work.md"
	newSpecPath  = "work/specs/spec-0014-declare-derived-work.md"
	// title is one summary the declared style accepts, carrying no
	// task tag — the shape an authoring title has.
	title = "[DOCS] The merge is the assenting act"
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

// fakeScripts is the fake beside the kit.Runner port. The replies are
// keyed by script because author runs five different authorities and a
// test cares which one spoke.
type fakeScripts struct {
	replies map[string]reply
	calls   []string
	// env records what each script was handed through the environment,
	// keyed by script: the seam this fake exists to witness.
	env map[string][]string
}

func (f *fakeScripts) run(_ string, stdout, stderr io.Writer, env []string, script string, args ...string) error {
	f.calls = append(f.calls, strings.TrimSpace(script+" "+strings.Join(args, " ")))
	if f.env == nil {
		f.env = map[string][]string{}
	}
	f.env[script] = env
	r := f.replies[script]
	fmt.Fprint(stdout, r.out)
	fmt.Fprint(stderr, r.errOut)
	return r.err
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

func (f *fakeScripts) ran(script string) bool {
	for _, c := range f.calls {
		if strings.HasPrefix(c, script) {
			return true
		}
	}
	return false
}

// fakeGit answers what the change is and records what was done to it.
type fakeGit struct {
	branch string
	dirty  string
	// refs is every ref that resolves; a ref absent from it is git's
	// "unknown revision".
	refs map[string]bool
	// remotes is what `git remote` lists; empty is a repository with
	// no forge, where nothing is fetched.
	remotes  []string
	fetchErr error
	// files is the diff's listing; added is keyed by pathspec.
	files []string
	added map[string][]string

	pushErr error
	calls   []string
}

func (g *fakeGit) run(_ string, args ...string) (string, error) {
	joined := strings.Join(args, " ")
	g.calls = append(g.calls, joined)
	switch {
	case joined == "status --porcelain":
		return g.dirty, nil
	case joined == "remote":
		return strings.Join(g.remotes, "\n") + "\n", nil
	case args[0] == "fetch":
		return "", g.fetchErr
	case joined == "rev-parse --abbrev-ref HEAD":
		return g.branch + "\n", nil
	case len(args) > 1 && args[0] == "rev-parse" && args[1] == "--verify":
		ref := args[len(args)-1]
		if g.refs[ref] {
			return ref + "\n", nil
		}
		return "", errors.New("git rev-parse --verify: unknown revision")
	case len(args) > 3 && args[0] == "diff" && args[2] == "--diff-filter=A":
		return strings.Join(g.added[args[len(args)-1]], "\n") + "\n", nil
	case len(args) > 1 && args[0] == "diff":
		return strings.Join(g.files, "\n") + "\n", nil
	case args[0] == "push":
		return "", g.pushErr
	}
	return "", nil
}

func (g *fakeGit) did(what string) bool {
	for _, c := range g.calls {
		if strings.HasPrefix(c, what) {
			return true
		}
	}
	return false
}

// fakeGh is the forge, stubbed: what it was asked, and whether it
// refuses.
type fakeGh struct {
	authErr   error
	createErr error
	created   string
	calls     []string
}

func (g *fakeGh) run(args ...string) (string, error) {
	g.calls = append(g.calls, strings.Join(args, " "))
	switch {
	case len(args) > 1 && args[0] == "auth":
		return "", g.authErr
	case len(args) > 1 && args[0] == "pr" && args[1] == "create":
		g.created = strings.Join(args, " ")
		return "https://forge.test/pull/9\n", g.createErr
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

// taskFixture is a derived task, its fields spelled the way the schema
// spells them.
func taskFixture(id, specRef, heading string) string {
	return "---\n" +
		"id: " + id + "\n" +
		"status: backlog\n" +
		"blocked_reason: null\n" +
		"taken_by: null\n" +
		"spec_ref: [" + specRef + "]\n" +
		"doc_ref: product/pull-requests/author.md\n" +
		"origin: rule\n" +
		"priority: medium\n" +
		"depends_on: []\n" +
		"milestone: null\n" +
		"created: 2026-09-03T22:30:23Z\n" +
		"queued: null\n" +
		"completed: null\n" +
		"merged: null\n" +
		"provenance: []\n" +
		"---\n\n" +
		"# " + heading + "\n\n" +
		"One paragraph of brief.\n"
}

// specFixture is a derived spec, born `draft` the way derivation
// writes it.
func specFixture(id, taskRef, heading string) string {
	return "---\n" +
		"id: " + id + "\n" +
		"task_ref: " + taskRef + "\n" +
		"status: draft\n" +
		"created: 2026-09-03T22:30:41Z\n" +
		"---\n\n" +
		"# " + id + " — " + heading + "\n\n" +
		"- **Goal:** something the task implements.\n"
}

// harness is one author: every port faked, the streams captured.
type harness struct {
	scripts *fakeScripts
	files   *vfs.Fake
	git     *fakeGit
	gh      *fakeGh
	term    *command.FakeTerminal
	ctx     *command.Ctx
	out     bytes.Buffer
	errb    bytes.Buffer
}

// newHarness is the green path, ready to be spoiled one port at a time:
// a clean tree on a local `docs/` branch, a diff that writes one rule
// and derives one task and one spec from it, every check exiting 0, and
// a human who says yes.
func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{
		scripts: &fakeScripts{replies: map[string]reply{
			settingScript: {out: "bracketed\n"},
		}},
		files: vfs.NewFake(),
		git: &fakeGit{
			branch:  authorBranch,
			remotes: []string{"origin"},
			refs:    map[string]bool{"refs/remotes/origin/main": true},
			files:   []string{"docs/product/pull-requests/author.md", newTaskPath, newSpecPath},
			added: map[string][]string{
				tasksDir + "/task-*.md": {newTaskPath},
				specsDir + "/spec-*.md": {newSpecPath},
			},
		},
		gh:   &fakeGh{},
		term: &command.FakeTerminal{In: true, ConfirmAnswer: true},
	}
	h.seed(templatePath, template)
	h.seed(newTaskPath, taskFixture("task-0016", "spec-0014", "Declare the derived work"))
	h.seed(newSpecPath, specFixture("spec-0014", "task-0016", "The declaration is the section"))
	h.ctx = &command.Ctx{Stdout: &h.out, Stderr: &h.errb, Terminal: h.term, Root: root, Adopted: true}
	return h
}

func (h *harness) seed(rel, content string) {
	h.files.Seed(path.Join(root, rel), []byte(content), 0o644)
}

func (h *harness) deps() Deps {
	return Deps{Scripts: h.scripts.run, Files: h.files, Git: h.git.run, Gh: h.gh.run}
}

// author runs the command with `--title` already answered, which is the
// one question that is not the confirmation.
func (h *harness) author(args ...string) error {
	return run(h.ctx, h.deps(), append([]string{"--title", title}, args...))
}

// template is the kit's body template, reduced to the lines this
// package transforms — the leading instructions, the two halves, and
// the markers around them.
const template = `<!--
Shipped by WritRun — the canonical PR body template.

Two kinds of PR land here. Keep the section that applies, delete the other.
-->

## What

## Why

<!-- writrun:begin -->

## Derived work

<!-- AUTHORING PRs ONLY. Every task and spec this change creates.
     If the rule derives no work, write "none" and say why in Notes. -->

| Task | Spec | What it implements |
|---|---|---|
| task-NNNN | spec-NNNN | |

## Spec

<!-- IMPLEMENTATION PRs ONLY. The spec this PR implements. -->

Implements spec-NNNN.

## How to verify

<!-- Implementation PRs: the writrun-check-spec-deltas result. -->

<!-- writrun:end -->

## Notes
`

// splitShell is the reader a person's shell is, for the one line this
// package prints for them to paste: single-quoted arguments, where a
// quote inside one is written by closing, escaping and reopening.
func splitShell(line string) []string {
	var out []string
	var cur strings.Builder
	inWord, quoted := false, false
	for i := 0; i < len(line); i++ {
		c := line[i]
		switch {
		case c == '\'':
			quoted = !quoted
			inWord = true
		case c == '\\' && !quoted && i+1 < len(line):
			i++
			cur.WriteByte(line[i])
			inWord = true
		case (c == ' ' || c == '\t') && !quoted:
			if inWord {
				out = append(out, cur.String())
				cur.Reset()
				inWord = false
			}
		default:
			cur.WriteByte(c)
			inWord = true
		}
	}
	if inWord {
		out = append(out, cur.String())
	}
	return out
}

// replay runs a printed command line the way a shell would: split it,
// drop the binary's own name, and hand the rest to the frame — which is
// what reads `--yes`, so a resume that carries it must go through here
// rather than around it. The terminal is gone, because the run that
// printed the line is the run that had none.
func replay(t *testing.T, h *harness, line string) error {
	t.Helper()
	fields := splitShell(strings.TrimSpace(line))
	if len(fields) == 0 || fields[0] != "writrun" {
		t.Fatalf("the printed line does not open with the binary: %q", line)
	}
	h.term.In = false
	var err error
	f := command.Frame{
		Commands: []command.Command{{
			Name: "author",
			Need: command.NeedAdopted,
			Run: func(ctx *command.Ctx, args []string) error {
				err = run(ctx, h.deps(), args)
				return err
			},
		}},
		Stdout:   &h.out,
		Stderr:   &h.errb,
		Terminal: h.term,
		FindRepo: func(string) (string, bool, error) { return root, true, nil },
		Getenv:   func(string) string { return "" },
		Getwd:    func() (string, error) { return root, nil },
	}
	if code := command.Run(f, fields[1:]); code != 0 {
		if err == nil {
			err = fmt.Errorf("the frame refused the line, exit %d", code)
		}
		return err
	}
	return nil
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
