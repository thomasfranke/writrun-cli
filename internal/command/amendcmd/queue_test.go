package amendcmd

import (
	"errors"
	"path"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// How the front matter, the id and the file are read is
// internal/queue's, and where each disputed reading came from is held
// in its tests. What is held here is what this command does with the
// answers.

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
