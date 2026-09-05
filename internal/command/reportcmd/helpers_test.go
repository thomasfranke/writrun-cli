package reportcmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// scriptExit is a script's own verdict, shaped like the *exec.ExitError
// the production runner returns: an error carrying an exit code.
type scriptExit int

func (e scriptExit) Error() string { return fmt.Sprintf("exit status %d", int(e)) }
func (e scriptExit) ExitCode() int { return int(e) }

// generated is what the real generator writes: the front matter it
// owns, the heading, and the placeholder paragraph this command
// substitutes.
const generated = `---
id: report-0009
status: open
task_ref: []
doc_ref: null
created: 2026-09-05T10:00:00Z
triaged: null
---

# Something that was noticed

TODO: what was observed, and the evidence at hand. What should be done
about it is triage's output, never this file's content.
`

// reply is one canned run of the generator: what it wrote, how it
// ended, and the file it left behind.
type reply struct {
	out    string
	errOut string
	err    error
	// file and contents are what a real run would have created; the
	// fake seeds them so the substitution has something to edit.
	file     string
	contents string
}

// created is a reply that minted the file the path names.
func minted(path string) reply {
	return reply{
		out:      "Created " + path + " (report-0009)\nMinted above the queue, its history, and every open pull request.\n",
		file:     path,
		contents: generated,
	}
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
	files   *vfs.Fake
	root    string
}

func (f *fakeScripts) run(root string, stdout, stderr io.Writer, script string, args ...string) error {
	f.calls = append(f.calls, call{root: root, script: script, args: args})
	if len(f.replies) == 0 {
		return nil
	}
	r := f.replies[0]
	f.replies = f.replies[1:]
	fmt.Fprint(stdout, r.out)
	fmt.Fprint(stderr, r.errOut)
	if r.file != "" {
		f.files.Seed(f.root+"/"+r.file, []byte(r.contents), 0o644)
	}
	return r.err
}

// harness is one report: the fake generator, the fake filesystem, the
// fake terminal, the streams.
type harness struct {
	scripts *fakeScripts
	files   *vfs.Fake
	term    *command.FakeTerminal
	ctx     *command.Ctx
	out     bytes.Buffer
	errb    bytes.Buffer
}

func newHarness(t *testing.T, replies ...reply) *harness {
	t.Helper()
	h := &harness{files: vfs.NewFake(), term: &command.FakeTerminal{}}
	h.files.SeedDir("/repo/work/reports")
	h.scripts = &fakeScripts{replies: replies, files: h.files, root: "/repo"}
	h.ctx = &command.Ctx{Stdout: &h.out, Stderr: &h.errb, Terminal: h.term, Root: "/repo"}
	return h
}

// report runs the command as the frame would, with the question
// already answered unless a case says otherwise.
func (h *harness) report(args ...string) error {
	return run(h.ctx, Deps{Scripts: h.scripts.run, Files: h.files}, args)
}

// argsOf is the argument list of the nth recorded call, joined so a
// test can assert on the whole invocation at once.
func (h *harness) argsOf(t *testing.T, n int) string {
	t.Helper()
	if len(h.scripts.calls) <= n {
		t.Fatalf("call %d was never made; %d call(s) recorded", n, len(h.scripts.calls))
	}
	return strings.Join(h.scripts.calls[n].args, " ")
}

// read is the recorded file as it stands after the run.
func (h *harness) read(t *testing.T, path string) string {
	t.Helper()
	raw, err := h.files.ReadFile("/repo/" + path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(raw)
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
