package queue

import (
	"errors"
	"path"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

const (
	root     = "/repo"
	tasksDir = "work/tasks"
	specsDir = "work/specs"
)

// canonical is a queue file in the shape check_front_matter.sh accepts.
func canonical(id, status string) string {
	return "---\nid: " + id + "\nstatus: " + status + "\nspec_ref: [spec-0012]\n---\n\n# A title\n\nA body paragraph.\n"
}

// Num is ql_task_num, and the kind it is given is the only prefix it
// strips. Each row was read off the script itself:
//
//	ql_task_num task-0012    -> 12
//	ql_task_num spec-0012    -> ''
//	ql_task_num task-0000    -> ''
//	ql_task_num task-abc-12  -> ''
//	ql_task_num report-0020  -> ''
func TestNumIsTheKitsRuleForOneKind(t *testing.T) {
	cases := []struct {
		kind Kind
		in   string
		want string
	}{
		{Task, "task-0012", "12"},
		{Task, "task-0012-queue-reader", "12"},
		{Task, "task/0012-queue-reader", "12"},
		{Task, "task/12", "12"},
		{Task, "0012", "12"},
		{Task, "12", "12"},
		{Task, "  task-0012  ", "12"},
		{Spec, "spec-0012", "12"},
		{Spec, "0012", "12"},
		// The other vocabulary keeps its prefix and is truncated at the
		// first letter, so it names no number of this kind.
		{Task, "spec-0012", ""},
		{Spec, "task-0012", ""},
		{Spec, "task/0012-queue-reader", ""},
		// A third vocabulary is no different: `report-` is not stripped.
		{Task, "report-0020", ""},
		{Spec, "report-0020", ""},
		// The padding is the whole of the number.
		{Task, "task-0000", ""},
		{Task, "0000", ""},
		{Task, "00", ""},
		// The digits must lead: `s/[^0-9].*$//` keeps nothing else.
		{Task, "task-abc-0012", ""},
		{Task, "abc0012", ""},
		{Task, "nothing", ""},
		{Task, "", ""},
	}
	for _, c := range cases {
		if got := Num(c.kind, c.in); got != c.want {
			t.Errorf("Num(%q, %q) = %q, want %q", c.kind, c.in, got, c.want)
		}
	}
}

func TestDeclaresIsTheKindTheIdSpells(t *testing.T) {
	cases := []struct {
		in   string
		want Kind
	}{
		{"spec-0011", Spec},
		{"task-0012", Task},
		{"task/0012-amend-command", Task},
		{"spec/0011", Spec},
		{"  task-0012  ", Task},
		{"0011", ""},
		{"report-0020", ""},
		{"taskish-0012", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := Declares(c.in); got != c.want {
			t.Errorf("Declares(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestResolveFindsTheFileWhateverWidthTheIdWasWrittenAt(t *testing.T) {
	files := vfs.NewFake()
	rel := specsDir + "/spec-0011-amend-command.md"
	files.Seed(path.Join(root, rel), []byte(canonical("spec-0011", "approved")), 0o644)
	files.Seed(path.Join(root, specsDir, "README.md"), []byte("# Specs\n"), 0o644)

	for _, id := range []string{"spec-0011", "spec-11", "0011", "11"} {
		got, err := Resolve(files, root, specsDir, Spec, id)
		if err != nil {
			t.Fatalf("Resolve(%q) = %v", id, err)
		}
		if got != rel {
			t.Errorf("Resolve(%q) = %q, want %q", id, got, rel)
		}
	}
	if _, err := Resolve(files, root, specsDir, Spec, "spec-0099"); err == nil {
		t.Error("a spec that is not there resolved to something")
	} else if !strings.Contains(err.Error(), "spec-99 resolves to no file") {
		t.Errorf("error = %v; want it to name what was looked for", err)
	}
	if _, err := Resolve(files, root, specsDir, Spec, "words"); err == nil {
		t.Error("an id naming no number resolved to something")
	}
}

// The task and spec counters run independently and are routinely one
// apart, so an id of the other kind names no number and the error says
// which file its digits alone would have landed on.
func TestResolveRefusesAnIdOfTheOtherKind(t *testing.T) {
	files := vfs.NewFake()
	files.Seed(path.Join(root, specsDir, "spec-0012-release.md"),
		[]byte(canonical("spec-0012", "approved")), 0o644)

	got, err := Resolve(files, root, specsDir, Spec, "task-0012")
	if err == nil {
		t.Fatalf("Resolve resolved a task id to %q", got)
	}
	for _, want := range []string{"task-0012", "spec-12", "different"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %v; want it to name %q", err, want)
		}
	}
	if _, err := Resolve(files, root, specsDir, Spec, "task/0012-amend-command"); err == nil {
		t.Error("a task branch name resolved to a spec")
	}
	// A third vocabulary declares no kind this binary resolves, so it
	// is refused as naming no id rather than as a near miss.
	_, err = Resolve(files, root, specsDir, Spec, "report-0020")
	if err == nil || !strings.Contains(err.Error(), "names no spec id") {
		t.Errorf("report-0020 = %v; want it refused as naming no spec id", err)
	}
	// `task-abc-0012` names no number at all, so there is no file to
	// name in the refusal.
	_, err = Resolve(files, root, specsDir, Spec, "task-abc-0012")
	if err == nil || !strings.Contains(err.Error(), "names no spec id") {
		t.Errorf("task-abc-0012 = %v; want it refused as naming no spec id", err)
	}
	// The bare number still resolves: it declares no kind.
	if _, err := Resolve(files, root, specsDir, Spec, "0012"); err != nil {
		t.Errorf("a bare number was refused: %v", err)
	}
}

// A queue directory that cannot be walked is an error, not an empty
// answer. What a command says about it is the command's.
func TestResolveReturnsTheWalksOwnError(t *testing.T) {
	files := vfs.NewFake()
	files.Seed(path.Join(root, tasksDir, "task-0012-a.md"), []byte(canonical("task-0012", "ready")), 0o644)
	files.FailOp("walk", path.Join(root, tasksDir), errors.New("permission denied"))
	if _, err := Resolve(files, root, tasksDir, Task, "task-0012"); err == nil {
		t.Error("a queue that could not be walked resolved anyway")
	}
}

// A body line spelling `status:` at column 0 is prose, not front matter
// — the rule the whole queue is read by.
func TestFieldReadsTheFrontMatterAlone(t *testing.T) {
	content := []byte("---\nid: spec-0011\nstatus: approved\n---\n\nstatus: draft\n")
	if got := Field(content, "status"); got != "approved" {
		t.Errorf("status = %q, want approved", got)
	}
	if got := Field([]byte("no front matter\nstatus: draft\n"), "status"); got != "" {
		t.Errorf("status = %q; a file with no front matter carries no field", got)
	}
	if got := Field(content, "nothing"); got != "" {
		t.Errorf("a field nobody wrote = %q", got)
	}
	if got := Field([]byte("---\nstatus:approved\n---\n"), "status"); got != "approved" {
		t.Errorf("status = %q; `key:value` is a field to `sub(\"^\" f \": *\", \"\")`", got)
	}
}

// Every one of these is a point four copies of this reader disagreed
// on, and every answer is the kit's own, read off `ql_fm_field` and
// `fm_block` on the same input.
func TestTheFenceAndTheBlockAreTheKitsAndNotATrimmedOne(t *testing.T) {
	cases := []struct {
		what, content, want string
	}{
		// `NR == 1 { if ($0 != "---") exit }` — a CRLF file opens with
		// `---\r`, which is not the fence.
		{"CRLF throughout", "---\r\nid: task-0012\r\nstatus: ready\r\n---\r\n", ""},
		// The same exactness at the other end: `/^---$/`.
		{"a CRLF closing fence", "---\nid: task-0012\nstatus: ready\n---\r\n", ""},
		{"a fence carrying a trailing space", "--- \nid: task-0012\nstatus: ready\n---\n", ""},
		{"an indented fence", " ---\nid: task-0012\nstatus: ready\n---\n", ""},
		// `fm_block` fails when the closing fence never comes, so the
		// file holds no fields — not the ones above the body it never
		// stopped reading.
		{"an unclosed block", "---\nid: task-0012\nstatus: ready\n\n# A title\n", ""},
		{"no front matter at all", "# A README\n\nstatus: open\n", ""},
		// `ql_fm_field` exits on the first match; `get` is `head -n1`.
		{"two status lines", "---\nid: task-0012\nstatus: ready\nstatus: done\n---\n", "ready"},
		// `sub("^" f ": *", "")` wants the colon against the name.
		{"a space before the colon", "---\nid: task-0012\nstatus : ready\n---\n", ""},
		// A field line's own carriage return is trailing whitespace,
		// which `sub(/[[:space:]]*$/, "")` takes off the value.
		{"a CRLF field line under an LF fence", "---\nid: task-0012\r\nstatus: ready\r\n---\n", "ready"},
	}
	for _, c := range cases {
		if got := Field([]byte(c.content), "status"); got != c.want {
			t.Errorf("%s: status = %q, want %q", c.what, got, c.want)
		}
	}
}

// awk and sed read a line of any length. A reader that capped one and
// carried on dropped every field after the cap and said nothing.
func TestNoLineIsTooLongToReadPast(t *testing.T) {
	long := strings.Repeat("x", 2*1024*1024)
	content := []byte("---\nid: task-0012\nnote: " + long + "\nstatus: ready\n---\n\n# " + long + "\n\n# A title\n")
	if got := Field(content, "status"); got != "ready" {
		t.Errorf("status = %q; the field after a 2 MiB line was dropped", got)
	}
	if got := Heading(content); got != long {
		t.Errorf("heading was truncated to %d bytes", len(got))
	}
}

// ql_resting's tokenizer: `tr -d '[]' | tr ',' ' '` and then the words.
func TestListIsTheKitsTokenizer(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"[spec-0011]", "spec-0011"},
		{"[spec-0011, spec-0012]", "spec-0011,spec-0012"},
		{"[spec-0011 spec-0012]", "spec-0011,spec-0012"},
		{"[spec-0011,]", "spec-0011"},
		{"[]", ""},
		{"", ""},
		// `null` is a word, not a filter: a list holding it is a file
		// check_front_matter.sh refuses, and the word is what says so.
		{"[null]", "null"},
		{"null", "null"},
	}
	for _, c := range cases {
		got := List([]byte("---\nid: task-0012\nspec_ref: "+c.in+"\n---\n"), "spec_ref")
		if strings.Join(got, ",") != c.want {
			t.Errorf("List(%q) = %v, want %q", c.in, got, c.want)
		}
	}
	if got := List([]byte("---\nid: task-0012\n---\n"), "spec_ref"); got != nil {
		t.Errorf("a field nobody wrote = %v, want nil", got)
	}
}

func TestHeadingIsTheFirstTitleWhole(t *testing.T) {
	content := []byte(canonical("task-0012", "ready"))
	if got := Heading(content); got != "A title" {
		t.Errorf("heading = %q", got)
	}
	if got := Heading([]byte("---\nid: task-0012\n---\n\n# spec-0012 — Read the queue in one place\n")); got != "spec-0012 — Read the queue in one place" {
		t.Errorf("heading = %q; the whole line is what this returns", got)
	}
	if got := Heading([]byte("no title here\n## Not this one\n")); got != "" {
		t.Errorf("heading = %q; want nothing where there is no title", got)
	}
}

func TestSetWritesOneFieldAndLeavesTheRest(t *testing.T) {
	content := []byte(canonical("spec-0011", "approved"))
	next, changed, err := Set(content, "status", "draft")
	if err != nil || !changed {
		t.Fatalf("Set = %v, changed=%v", err, changed)
	}
	if Field(next, "status") != "draft" {
		t.Error("the field was not written")
	}
	if !strings.Contains(string(next), "A body paragraph.") {
		t.Error("the body was disturbed")
	}
	if strings.Contains(string(next), "\r") {
		t.Error("an LF file gained a carriage return")
	}
	if _, changed, _ := Set(next, "status", "draft"); changed {
		t.Error("a field already carrying the value reported a change")
	}
	if _, _, err := Set(content, "nowhere", "x"); err == nil {
		t.Error("a field the schema does not carry was added")
	}
	if _, _, err := Set([]byte("no front matter\n"), "status", "draft"); err == nil {
		t.Error("a file with no front matter was written into")
	}
	if _, _, err := Set([]byte("---\nstatus: approved\n"), "status", "draft"); err == nil {
		t.Error("an unclosed block was written into")
	}
}

// `ql_set_field` prints `field ": " value` and nothing else, so a field
// line that carried a carriage return loses it. Read off the script:
//
//	printf -- '---\nid: task-0012\r\nstatus: ready\r\n---\n' > f
//	ql_set_field f status approved
//	-> ---\nid: task-0012\r\nstatus: approved\n---\n
func TestSetDropsTheCarriageReturnTheKitDrops(t *testing.T) {
	mixed := []byte("---\nid: spec-0011\r\nstatus: approved\r\n---\n\nA body paragraph.\n")
	next, changed, err := Set(mixed, "status", "draft")
	if err != nil || !changed {
		t.Fatalf("Set = %v, changed=%v", err, changed)
	}
	if !strings.Contains(string(next), "status: draft\n") || strings.Contains(string(next), "status: draft\r") {
		t.Errorf("the rewritten line kept a carriage return the kit drops:\n%q", string(next))
	}
	if !strings.Contains(string(next), "id: spec-0011\r\n") {
		t.Error("a line this write was not about lost its carriage return")
	}
	if Field(next, "status") != "draft" {
		t.Error("the field was not written")
	}
}

// A wholly-CRLF file has no front matter, so there is nothing to write
// into it — the refusal, not a careful rewrite (report-0024).
func TestSetRefusesACrlfFile(t *testing.T) {
	crlf := []byte(strings.ReplaceAll(canonical("spec-0011", "approved"), "\n", "\r\n"))
	if _, changed, err := Set(crlf, "status", "draft"); err == nil || changed {
		t.Errorf("Set wrote into a file the repository calls malformed: changed=%v err=%v", changed, err)
	}
}
