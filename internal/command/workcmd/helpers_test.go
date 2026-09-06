package workcmd

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/kit"
)

// root is the repository every case runs in.
const root = "/repo"

// answer is one run of the lister with every section present, in the
// order and the wording it prints them — the answer this command reads
// and never rewrites.
const answer = `In progress — resume before selecting anything new:
  task-0004  Take a task without memorising the flow

Available — any of these may be taken:
  task-0007  high    Launch the configured agent with writrun work
  task-0011  low     Something smaller

Order is a suggestion for a person and binding for an agent.

In flight — an open pull request already exists:
  task-0003  #12 by @someone  Package the kit
  task-0005  #14 by @other    Refresh the kit
             paused — spec-0005 is amended by #21; the work waits on the re-approval

Held back:
  task-0009  spec-0009 is draft

Open reports — waiting to be triaged, never selected:
  report-0002  A thing that was noticed
`

// nothing is the lister's other answer, the one it exits 1 with.
const nothing = "Nothing is available.\n"

// brief is what brief.sh printed for the selected task — the text the
// launched agent has to receive unedited.
const brief = `task-0007  ready  high  specs: spec-0007 approved

== work/tasks/task-0007-work-command.md ==

---
id: task-0007
status: ready
---

# Launch the configured agent with writrun work
`

// exitErr is a child's own verdict as a port hands it up.
type exitErr int

func (e exitErr) Error() string { return fmt.Sprintf("exit status %d", int(e)) }
func (e exitErr) ExitCode() int { return int(e) }

// scriptRun is one invocation the exec port was asked for.
type scriptRun struct {
	dir    string
	script string
	args   []string
}

// fakeScripts is the exec port faked down to the two scripts this
// command wraps: each answers with its own canned output and verdict,
// and every invocation is recorded — so a case can assert that nothing
// else was ever run.
type fakeScripts struct {
	listing  string
	listErr  error
	brief    string
	briefOut string
	briefErr error
	runs     []scriptRun
}

func (f *fakeScripts) run(dir string, stdout, stderr io.Writer, name string, args ...string) error {
	f.runs = append(f.runs, scriptRun{dir: dir, script: name, args: args})
	switch name {
	case listScript:
		fmt.Fprint(stdout, f.listing)
		return f.listErr
	case briefScript:
		fmt.Fprint(stdout, f.brief)
		fmt.Fprint(stderr, f.briefOut)
		return f.briefErr
	}
	return fmt.Errorf("unexpected script %q", name)
}

// scripts is the port itself, answering with the whole listing and a
// complete brief.
func scripts() *fakeScripts {
	return &fakeScripts{listing: answer, brief: brief}
}

// ran reports the scripts a run asked for, in order.
func (f *fakeScripts) ran() []string {
	var out []string
	for _, r := range f.runs {
		out = append(out, r.script)
	}
	return out
}

// configuredAgent is the git port faked down to the one question this
// command asks it.
func configuredAgent(value string) gitx.Runner {
	return func(dir string, args ...string) (string, error) {
		if strings.Join(args, " ") != "config --get "+configKey {
			return "", fmt.Errorf("unexpected git %v", args)
		}
		return value + "\n", nil
	}
}

// noAgent is git answering that the key is not set: exit 1, nothing on
// stdout.
func noAgent() gitx.Runner {
	return func(dir string, args ...string) (string, error) { return "", exitErr(1) }
}

// gitFails is a git that could not answer at all.
func gitFails(err error) gitx.Runner {
	return func(dir string, args ...string) (string, error) { return "", err }
}

// result is one run of the command: what it printed, what it launched
// and what it decided.
type result struct {
	stdout   string
	stderr   string
	launched []Launch
	err      error
}

// work runs the command over one fixture.
func work(t *testing.T, git gitx.Runner, s *fakeScripts, launcher *FakeLauncher, args ...string) result {
	t.Helper()
	var stdout, stderr bytes.Buffer
	ctx := &command.Ctx{Stdout: &stdout, Stderr: &stderr, Root: root}
	if launcher == nil {
		launcher = &FakeLauncher{}
	}
	var runner kit.Runner = s.run
	err := run(ctx, Deps{Git: git, Scripts: runner, Launch: launcher.Run}, args)
	return result{
		stdout:   stdout.String(),
		stderr:   stderr.String(),
		launched: launcher.Launched,
		err:      err,
	}
}

// prompted is the argument the launch carried the brief in — the last
// one, after whatever the configured command declared for itself.
func (r result) prompted(t *testing.T) string {
	t.Helper()
	if len(r.launched) != 1 {
		t.Fatalf("launched %d times; want exactly one", len(r.launched))
	}
	args := r.launched[0].Args
	if len(args) == 0 {
		t.Fatal("the launch carried no arguments, so it carried no brief")
	}
	return args[len(args)-1]
}

// wantsNoLaunch fails unless nothing was started.
func (r result) wantsNoLaunch(t *testing.T) {
	t.Helper()
	if len(r.launched) != 0 {
		t.Fatalf("launched %v; want nothing started", r.launched)
	}
}

// wantsError fails unless the refusal carries want.
func (r result) wantsError(t *testing.T, want string) {
	t.Helper()
	if r.err == nil {
		t.Fatalf("run exited 0; want a refusal naming %q\nstdout:\n%s", want, r.stdout)
	}
	if !strings.Contains(r.err.Error(), want) {
		t.Fatalf("err = %q; want it to name %q", r.err, want)
	}
}
