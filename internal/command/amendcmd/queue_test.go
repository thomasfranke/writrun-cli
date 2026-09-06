package amendcmd

import (
	"errors"
	"path"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// numOf answers a number and nothing else — ql_task_num's rule, which
// strips `task-` and `spec-` alike because `carriedOf` feeds it head
// branches and `suspended` feeds it spec refs. So `numOf("task-0012")`
// really is "12", and that is not a licence to resolve a task id under
// work/specs/: kindOf below is what refuses the mismatch, and
// TestQueueFileRefusesAnIdOfTheOtherKind is where the refusal is held.
func TestNumOf(t *testing.T) {
	cases := []struct{ in, want string }{
		{"spec-0011", "11"},
		{"0011", "11"},
		{"11", "11"},
		{"task/0012-amend-command", "12"},
		{"task-0012", "12"},
		{"spec-0011-amend-command", "11"},
		{"nothing", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := numOf(c.in); got != c.want {
			t.Errorf("numOf(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// kindOf is what numOf throws away: the vocabulary the id declared for
// itself. A bare number declares none and is read as whatever is being
// looked for.
func TestKindOf(t *testing.T) {
	cases := []struct{ in, want string }{
		{"spec-0011", "spec"},
		{"task-0012", "task"},
		{"task/0012-amend-command", "task"},
		{"spec/0011", "spec"},
		{"  task-0012  ", "task"},
		{"0011", ""},
		{"11", ""},
		{"", ""},
		{"taskish-0012", ""},
	}
	for _, c := range cases {
		if got := kindOf(c.in); got != c.want {
			t.Errorf("kindOf(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// The task and spec counters run independently and are routinely one
// apart, so `amend task-0012` used to resolve to spec-0012 — a real
// file, about different work, returned to draft and pushed without a
// word. An id that declares the other kind is refused.
func TestQueueFileRefusesAnIdOfTheOtherKind(t *testing.T) {
	files := vfs.NewFake()
	files.Seed(path.Join(root, specPath), []byte(specFixture("approved")), 0o644)
	files.Seed(path.Join(root, "work/specs/spec-0012-release-distribution.md"),
		[]byte(specFixture("approved")), 0o644)

	got, err := queueFile(files, root, specsDir, "spec", "task-0012")
	if err == nil {
		t.Fatalf("queueFile resolved a task id to %q", got)
	}
	for _, want := range []string{"task-0012", "spec-12", "different"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %v; want it to name %q", err, want)
		}
	}
	if _, err := queueFile(files, root, specsDir, "spec", "task/0012-amend-command"); err == nil {
		t.Error("a task branch name resolved to a spec")
	}
	// The bare number still resolves: it declares no kind.
	if _, err := queueFile(files, root, specsDir, "spec", "0011"); err != nil {
		t.Errorf("a bare number was refused: %v", err)
	}
}

func TestQueueFileResolvesWhateverWidthTheIdWasWrittenAt(t *testing.T) {
	files := vfs.NewFake()
	files.Seed(path.Join(root, specPath), []byte(specFixture("approved")), 0o644)
	for _, id := range []string{"spec-0011", "0011", "11"} {
		got, err := queueFile(files, root, specsDir, "spec", id)
		if err != nil {
			t.Fatalf("queueFile(%q) = %v", id, err)
		}
		if got != specPath {
			t.Errorf("queueFile(%q) = %q, want %q", id, got, specPath)
		}
	}
	if _, err := queueFile(files, root, specsDir, "spec", "spec-0099"); err == nil {
		t.Error("a spec that is not there resolved to something")
	}
	if _, err := queueFile(files, root, specsDir, "spec", "words"); err == nil {
		t.Error("an id naming no number resolved to something")
	}
}

// A body line spelling `status:` at column 0 is prose, not front matter
// — the rule the whole queue is read by.
func TestFieldReadsTheFrontMatterAlone(t *testing.T) {
	content := []byte("---\nid: spec-0011\nstatus: approved\n---\n\nstatus: draft\n")
	if got := field(content, "status"); got != "approved" {
		t.Errorf("status = %q, want approved", got)
	}
	if got := field([]byte("no front matter\nstatus: draft\n"), "status"); got != "" {
		t.Errorf("status = %q; a file with no front matter carries no field", got)
	}
	if got := field([]byte("---\nid: spec-0011\n"), "id"); got != "" {
		t.Errorf("id = %q; an unclosed block is not front matter", got)
	}
}

// A CRLF queue file opens with `---\r`, which the kit's own ql_fm_field
// reads as no front matter and check_front_matter.sh calls MALFORMED.
// A reader here that trimmed the fence would read a status off a file
// the repository refuses, and amend would rewrite it (report-0024).
func TestTheFenceIsTheKitsFenceAndNotATrimmedOne(t *testing.T) {
	crlf := []byte("---\r\nid: spec-0011\r\nstatus: approved\r\n---\r\n")
	if got := field(crlf, "status"); got != "" {
		t.Errorf("status = %q; a CRLF file has no front matter to the kit, so it has none here", got)
	}
	if _, ok := frontMatter(crlf); ok {
		t.Error("a CRLF file reported front matter the kit's reader does not find")
	}
	if _, ok := frontMatter([]byte(" ---\nid: spec-0011\n---\n")); ok {
		t.Error("an indented fence opened a block; ql_fm_field requires column 0")
	}
}

func TestSetField(t *testing.T) {
	content := []byte(specFixture("approved"))
	next, changed, err := setField(content, "status", "draft")
	if err != nil || !changed {
		t.Fatalf("setField = %v, changed=%v", err, changed)
	}
	if field(next, "status") != "draft" {
		t.Error("the field was not written")
	}
	if !strings.Contains(string(next), "A body paragraph.") {
		t.Error("the body was disturbed")
	}
	if _, changed, _ := setField(next, "status", "draft"); changed {
		t.Error("a field already carrying the value reported a change")
	}
	if strings.Contains(string(next), "\r") {
		t.Error("an LF file gained a carriage return")
	}
	if _, _, err := setField(content, "nowhere", "x"); err == nil {
		t.Error("a field the schema does not carry was added")
	}
	if _, _, err := setField([]byte("no front matter\n"), "status", "draft"); err == nil {
		t.Error("a file with no front matter was written into")
	}
}

func TestSpecRefs(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"spec_ref: [spec-0011]", []string{"spec-0011"}},
		{"spec_ref: [spec-0011, spec-0012]", []string{"spec-0011", "spec-0012"}},
		{"spec_ref: []", nil},
		{"spec_ref: null", nil},
	}
	for _, c := range cases {
		got := specRefs([]byte("---\nid: task-0012\n" + c.in + "\n---\n"))
		if strings.Join(got, ",") != strings.Join(c.want, ",") {
			t.Errorf("specRefs(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

// Only a task in flight is suspended: every other status is the
// ordinary pre-implementation amendment (spec-0011, edge cases).
func TestSuspendedFindsOnlyFlight(t *testing.T) {
	files := vfs.NewFake()
	seed := func(name, id, status, ref string) {
		files.Seed(path.Join(root, tasksDir, name), []byte(taskFixture(id, status, ref)), 0o644)
	}
	seed("task-0012-a.md", "task-0012", "in-progress", "spec-0011")
	seed("task-0013-b.md", "task-0013", "in-review", "spec-0011")
	seed("task-0014-c.md", "task-0014", "ready", "spec-0011")
	seed("task-0015-d.md", "task-0015", "backlog", "spec-0011")
	seed("task-0016-e.md", "task-0016", "done", "spec-0011")
	seed("task-0017-f.md", "task-0017", "in-progress", "spec-0012")
	files.Seed(path.Join(root, tasksDir, "README.md"), []byte("# The queue\n"), 0o644)

	got, err := suspended(files, root, "spec-0011")
	if err != nil {
		t.Fatalf("suspended = %v", err)
	}
	if strings.Join(got, ",") != "task-0012,task-0013" {
		t.Errorf("suspended = %v; want the two in flight", got)
	}
}

// The spec is matched by number, so a `spec_ref` written at another
// width still resolves — the rule ql_task_num keeps.
func TestSuspendedMatchesBySpecNumber(t *testing.T) {
	files := vfs.NewFake()
	files.Seed(path.Join(root, tasksDir, "task-0012-a.md"),
		[]byte(taskFixture("task-0012", "in-progress", "spec-11")), 0o644)
	got, err := suspended(files, root, "spec-0011")
	if err != nil || strings.Join(got, ",") != "task-0012" {
		t.Errorf("suspended = %v, %v", got, err)
	}
}

// A task file that cannot be read stops the answer: a queue half-read
// is a suspension list that quietly forgets somebody.
func TestSuspendedRefusesAQueueItCannotRead(t *testing.T) {
	files := vfs.NewFake()
	p := path.Join(root, tasksDir, "task-0012-a.md")
	files.Seed(p, []byte(taskFixture("task-0012", "in-progress", "spec-0011")), 0o644)
	files.FailOp("read", p, errors.New("permission denied"))
	if _, err := suspended(files, root, "spec-0011"); err == nil {
		t.Error("a queue that could not be read answered anyway")
	}
}

// ql_carried_of's rule: the branch's own id plus every leading tag,
// deduplicated, case not judged.
func TestCarriedOf(t *testing.T) {
	cases := []struct {
		branch, title string
		want          string
	}{
		{"task/0012-amend-command", "[TASK-0012] A thing", "12"},
		{"docs/something", "[TASK-0012][TASK-0014] A thing", "12,14"},
		{"task/0012-a", "[task-0014] A thing", "12,14"},
		{"docs/something", "A thing with no tag", ""},
		{"report/observed", "[TASK-0001] Ignored? no, read", "1"},
		{"", "", ""},
	}
	for _, c := range cases {
		got := strings.Join(carriedOf(c.branch, c.title), ",")
		if got != c.want {
			t.Errorf("carriedOf(%q, %q) = %q, want %q", c.branch, c.title, got, c.want)
		}
	}
}

// The listing both the seam and `gh` produce, read the same way.
func TestParsePulls(t *testing.T) {
	got := parsePulls(strings.Join([]string{
		"42\ttask/0012-a\tsomeone\t[TASK-0012] A thing",
		"43\tdocs/other\t[TASK-0013] Three fields is the check's own shape",
		"",
		"not-a-number\tbranch\tauthor\ttitle",
		"44",
	}, "\n"))
	if len(got) != 2 {
		t.Fatalf("parsePulls = %v; want the two readable lines", got)
	}
	if got[0].number != 42 || got[0].branch != "task/0012-a" || !strings.Contains(got[0].title, "A thing") {
		t.Errorf("first = %+v", got[0])
	}
	if got[1].number != 43 || !strings.Contains(got[1].title, "Three fields") {
		t.Errorf("second = %+v", got[1])
	}
}

// A task no open pull request works keeps its place with no number:
// check_amendment_reference.sh calls that a stale flight state, and the
// composition says so rather than dropping the task.
func TestMatchKeepsATaskNoPullRequestWorks(t *testing.T) {
	got := match([]string{"task-0012", "task-0013"}, []pull{
		{number: 42, branch: "task/0012-a", title: "[TASK-0012] A thing"},
	})
	if len(got) != 2 {
		t.Fatalf("match = %v", got)
	}
	if got[0].number != 42 {
		t.Errorf("task-0012 = %+v; want #42", got[0])
	}
	if got[1].number != 0 {
		t.Errorf("task-0013 = %+v; want no number", got[1])
	}
}

// A wholly-CRLF queue file opens with `---\r`, which the kit's reader
// does not see as front matter and check_front_matter.sh calls
// MALFORMED. Reading a status off it and writing one back is what
// report-0024 named; the refusal is the answer, not a careful rewrite.
func TestAcrlfFileIsRefusedRatherThanRewritten(t *testing.T) {
	crlf := []byte(strings.ReplaceAll(specFixture("approved"), "\n", "\r\n"))
	if _, changed, err := setField(crlf, "status", "draft"); err == nil || changed {
		t.Errorf("setField wrote into a file the repository calls malformed: changed=%v err=%v", changed, err)
	}
}

// Where the fence is the kit's and a field line still carries a
// carriage return, the line keeps it: rewriting it bare would make it
// the one LF line in its neighbourhood, which is a change nobody asked
// for.
func TestSetFieldKeepsTheLineEndingItFound(t *testing.T) {
	mixed := []byte("---\nid: spec-0011\nstatus: approved\r\n---\n\nA body paragraph.\n")
	next, changed, err := setField(mixed, "status", "draft")
	if err != nil || !changed {
		t.Fatalf("setField = %v, changed=%v", err, changed)
	}
	if !strings.Contains(string(next), "status: draft\r\n") {
		t.Errorf("the rewritten line lost its carriage return:\n%q", string(next))
	}
	if field(next, "status") != "draft" {
		t.Error("the field was not written")
	}
	if _, changed, _ := setField(next, "status", "draft"); changed {
		t.Error("a line already carrying the value reported a change")
	}
}
