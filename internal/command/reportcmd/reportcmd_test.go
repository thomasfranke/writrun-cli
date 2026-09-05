package reportcmd

import (
	"errors"
	"io/fs"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
)

const (
	title = "The layout table omits two fixtures"
	body  = "Found while adding a suite: the fixtures table names five of them."
	file  = "work/reports/report-0009-a-thing.md"
)

func TestNewDeclaresTheCommand(t *testing.T) {
	c := New(Deps{})
	if c.Name != "report" {
		t.Errorf("name = %q, want report", c.Name)
	}
	if c.Need != command.NeedAdopted {
		t.Errorf("need = %v; want an adopted repository", c.Need)
	}
	if c.Summary == "" || c.Run == nil {
		t.Error("the command carries no summary or no work")
	}
}

// The generator mints the file and the reporter's paragraph lands in
// it — the whole act, in one run of one script.
func TestTheGeneratorMintsAndTheObservationLands(t *testing.T) {
	h := newHarness(t, minted(file))
	h.ctx.Yes = true

	if err := h.report(title, "--body", body, "--slug", "a-thing", "--doc-ref", "technical/testing/suites.md"); err != nil {
		t.Fatalf("report = %v", err)
	}
	if len(h.scripts.calls) != 1 {
		t.Fatalf("%d calls, want one", len(h.scripts.calls))
	}
	if h.scripts.calls[0].script != generator {
		t.Errorf("script = %q, want %q", h.scripts.calls[0].script, generator)
	}
	if h.scripts.calls[0].root != "/repo" {
		t.Errorf("root = %q, want the repository root", h.scripts.calls[0].root)
	}
	want := "report " + title + " --slug a-thing --doc-ref technical/testing/suites.md"
	if got := h.argsOf(t, 0); got != want {
		t.Errorf("args = %q, want %q", got, want)
	}
	recorded := h.read(t, file)
	if !strings.Contains(recorded, body) {
		t.Errorf("the recorded file carries no observation:\n%s", recorded)
	}
	if strings.Contains(recorded, "TODO") {
		t.Errorf("the placeholder survived the observation:\n%s", recorded)
	}
	if !strings.Contains(recorded, "id: report-0009") || !strings.Contains(recorded, "status: open") {
		t.Errorf("the generator's front matter did not survive:\n%s", recorded)
	}
	if !strings.Contains(h.out.String(), "Created "+file) {
		t.Errorf("stdout = %q; want the generator's own report", h.out.String())
	}
}

// Nothing is invented for the generator: an omitted slug and an
// omitted doc-ref are the script's defaults, not this command's.
func TestOmittedFlagsAreNotPassed(t *testing.T) {
	h := newHarness(t, minted(file))
	h.ctx.Yes = true

	if err := h.report(title, "--body", body); err != nil {
		t.Fatalf("report = %v", err)
	}
	if got := h.argsOf(t, 0); got != "report "+title {
		t.Errorf("args = %q; want the kind and the title alone", got)
	}
}

// Triage is a judgement, so no flag here sets a route: the routes are
// not options this command has (docs/product/reports/report.md).
func TestNoRouteFlagExists(t *testing.T) {
	for _, flag := range []string{"--tracked", "--fixed", "--declined", "--authored", "--priority", "--origin"} {
		h := newHarness(t, minted(file))
		h.ctx.Yes = true
		err := h.report(title, "--body", body, flag)
		if err == nil {
			t.Errorf("%s was accepted; a route is not this command's to set", flag)
		}
		if len(h.scripts.calls) != 0 {
			t.Errorf("%s reached the generator", flag)
		}
	}
}

// The write is preceded by the question rules.md demands, and without
// a terminal the question aborts naming the flag that answers it.
func TestTheQuestionPrecedesTheWrite(t *testing.T) {
	h := newHarness(t, minted(file))
	err := h.report(title, "--body", body)
	if err == nil {
		t.Fatal("an unanswerable question recorded a report anyway")
	}
	if !strings.Contains(err.Error(), "--yes") {
		t.Errorf("err = %v; want the flag that answers it named", err)
	}
	if len(h.scripts.calls) != 0 {
		t.Errorf("%d calls; nothing runs before the question is answered", len(h.scripts.calls))
	}
}

func TestDeclineRecordsNothing(t *testing.T) {
	h := newHarness(t, minted(file))
	h.term.In = true
	h.term.ConfirmAnswer = false

	err := h.report(title, "--body", body)
	if !errors.Is(err, command.ErrDeclined) {
		t.Fatalf("err = %v, want ErrDeclined", err)
	}
	if len(h.scripts.calls) != 0 {
		t.Errorf("%d calls; a decline reaches nothing", len(h.scripts.calls))
	}
}

// The title and the observation are typed where they were not given,
// and the question shows what is about to be recorded.
func TestBothAnswersAreAskedAndTheQuestionShowsTheTitle(t *testing.T) {
	h := newHarness(t, minted(file))
	h.term.In = true
	h.term.ConfirmAnswer = true
	h.term.InputAnswer = "typed at the prompt"

	if err := h.report(); err != nil {
		t.Fatalf("report = %v", err)
	}
	if len(h.term.Asked) != 3 {
		t.Fatalf("asked %v; want the title, the observation and the confirmation", h.term.Asked)
	}
	if got := h.argsOf(t, 0); got != "report typed at the prompt" {
		t.Errorf("args = %q; want the typed title", got)
	}
	if !strings.Contains(h.out.String(), "typed at the prompt") ||
		!strings.Contains(h.out.String(), "work/reports/") {
		t.Errorf("stdout = %q; want the title and where it lands shown", h.out.String())
	}
	if !strings.Contains(h.read(t, file), "typed at the prompt") {
		t.Error("the typed observation never reached the file")
	}
}

// An observation the reporter left blank leaves the generator's
// placeholder standing — the file is recorded, and it says so.
func TestABlankObservationLeavesThePlaceholder(t *testing.T) {
	h := newHarness(t, minted(file))
	h.term.In = true
	h.term.ConfirmAnswer = true
	h.term.InputAnswer = ""

	if err := h.report(title); err != nil {
		t.Fatalf("report = %v", err)
	}
	if got := h.read(t, file); got != generated {
		t.Errorf("the file was edited:\n%s", got)
	}
}

// The generator's refusal is this command's refusal: its exit code,
// its words, and no file touched.
func TestAGeneratorRefusalPassesThroughUnedited(t *testing.T) {
	h := newHarness(t, reply{errOut: "Unknown flag: --nope\n", err: scriptExit(3)})
	h.ctx.Yes = true

	err := h.report(title, "--body", body)
	if err == nil {
		t.Fatal("a refusal returned no error")
	}
	if got := exitOf(err); got != 3 {
		t.Errorf("exit = %d, want the script's own 3", got)
	}
	if !strings.Contains(h.errb.String(), "Unknown flag") {
		t.Errorf("stderr = %q; want the script's reason", h.errb.String())
	}
	if _, err := h.files.ReadFile("/repo/" + file); err == nil {
		t.Error("a refusal left a file behind")
	}
}

// A runner that failed before the script spoke has no verdict to pass
// through, so the failure names what was being run.
func TestARunnerFailureNamesTheGenerator(t *testing.T) {
	h := newHarness(t, reply{err: errors.New("bash: not found")})
	h.ctx.Yes = true

	err := h.report(title, "--body", body)
	if err == nil || !strings.Contains(err.Error(), generator) {
		t.Fatalf("err = %v; want the generator named", err)
	}
}

// A generator that created nothing it named leaves nothing to fill,
// and the command says so rather than guessing a path.
func TestAGeneratorThatNamesNoFileIsAnError(t *testing.T) {
	h := newHarness(t, reply{out: "Nothing to say.\n"})
	h.ctx.Yes = true

	err := h.report(title, "--body", body)
	if err == nil || !strings.Contains(err.Error(), "named no file") {
		t.Fatalf("err = %v; want the missing path reported", err)
	}
}

// The file is minted before the body is written, so a failure after
// that point names the file the reporter finishes by hand.
func TestAFailedWriteNamesTheRecordedFile(t *testing.T) {
	h := newHarness(t, minted(file))
	h.ctx.Yes = true
	h.files.FailOp("write", "/repo/"+file, fs.ErrPermission)

	err := h.report(title, "--body", body)
	if err == nil {
		t.Fatal("a failed write returned no error")
	}
	if !strings.Contains(err.Error(), file) || !strings.Contains(err.Error(), "by hand") {
		t.Errorf("err = %v; want the file and the way to finish it named", err)
	}
}

func TestAFailedReadNamesTheRecordedFile(t *testing.T) {
	h := newHarness(t, minted(file))
	h.ctx.Yes = true
	h.files.FailOp("read", "/repo/"+file, fs.ErrPermission)

	err := h.report(title, "--body", body)
	if err == nil || !strings.Contains(err.Error(), file) {
		t.Fatalf("err = %v; want the recorded file named", err)
	}
}

func TestTwoTitlesAreRefused(t *testing.T) {
	h := newHarness(t, minted(file))
	h.ctx.Yes = true

	err := h.report(title, "another title", "--body", body)
	if err == nil || !strings.Contains(err.Error(), "one") {
		t.Fatalf("err = %v; want two titles refused", err)
	}
	if len(h.scripts.calls) != 0 {
		t.Error("a refused invocation reached the generator")
	}
}

func TestSplitLiftsTheTitleOutOfTheFlags(t *testing.T) {
	cases := []struct {
		name  string
		args  []string
		title string
		flags string
	}{
		{"title first", []string{"a title", "--body", "b"}, "a title", "--body b"},
		{"title last", []string{"--body", "b", "a title"}, "a title", "--body b"},
		{"a valued flag swallows its value", []string{"--slug", "a title", "x"}, "x", "--slug a title"},
		{"a bare flag stands alone", []string{"--nope", "t"}, "t", "--nope"},
		{"a valued flag at the end", []string{"t", "--doc-ref"}, "t", "--doc-ref"},
		{"nothing given", nil, "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, flags, err := split(c.args)
			if err != nil {
				t.Fatalf("split = %v", err)
			}
			if got != c.title {
				t.Errorf("title = %q, want %q", got, c.title)
			}
			if joined := strings.Join(flags, " "); joined != c.flags {
				t.Errorf("flags = %q, want %q", joined, c.flags)
			}
		})
	}
}

func TestCreatedReadsBackThePathTheGeneratorNamed(t *testing.T) {
	cases := []struct {
		name string
		said string
		want string
		ok   bool
	}{
		{"the generator's line", "Created work/reports/report-0009-a.md (report-0009)\n", "work/reports/report-0009-a.md", true},
		{"no id in parentheses", "Created work/reports/report-0009-a.md\n", "work/reports/report-0009-a.md", true},
		{"a line among others", "Minted.\nCreated work/reports/report-0001-b.md (report-0001)\nMinted above the queue.\n", "work/reports/report-0001-b.md", true},
		{"a task is not a report", "Created work/tasks/task-0009-a.md (task-0009)\n", "", false},
		{"a path that climbs out", "Created work/reports/../../etc/passwd (report-0009)\n", "", false},
		{"nothing said", "", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := created(c.said)
			if ok != c.ok || got != c.want {
				t.Errorf("created = (%q, %v), want (%q, %v)", got, ok, c.want, c.ok)
			}
		})
	}
}
