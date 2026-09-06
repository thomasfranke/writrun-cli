package takecmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
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

// fakeScripts is the fake beside the kit.Runner port: canned replies in
// order, every invocation recorded.
type fakeScripts struct {
	replies []reply
	calls   []call
}

func (f *fakeScripts) run(root string, stdout, stderr io.Writer, _ []string, script string, args ...string) error {
	f.calls = append(f.calls, call{root: root, script: script, args: args})
	if len(f.replies) == 0 {
		return nil
	}
	r := f.replies[0]
	f.replies = f.replies[1:]
	fmt.Fprint(stdout, r.out)
	fmt.Fprint(stderr, r.errOut)
	return r.err
}

// harness is one take: the fake runner, the fake terminal, the streams.
type harness struct {
	scripts *fakeScripts
	term    *command.FakeTerminal
	ctx     *command.Ctx
	out     bytes.Buffer
	errb    bytes.Buffer
}

func newHarness(t *testing.T, replies ...reply) *harness {
	t.Helper()
	h := &harness{
		scripts: &fakeScripts{replies: replies},
		term:    &command.FakeTerminal{},
	}
	h.ctx = &command.Ctx{Stdout: &h.out, Stderr: &h.errb, Terminal: h.term, Root: "/repo"}
	return h
}

func (h *harness) take(args ...string) error {
	return run(h.ctx, Deps{Scripts: h.scripts.run}, args)
}

// lastArgs is the argument list of the nth recorded call, joined so a
// test can assert on the whole invocation at once.
func (h *harness) argsOf(t *testing.T, n int) string {
	t.Helper()
	if len(h.scripts.calls) <= n {
		t.Fatalf("call %d was never made; %d call(s) recorded", n, len(h.scripts.calls))
	}
	return strings.Join(h.scripts.calls[n].args, " ")
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
