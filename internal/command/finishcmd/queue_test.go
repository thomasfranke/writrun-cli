package finishcmd

import (
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// The id spellings a person retypes all resolve to the same number —
// ql_task_num's rule.
func TestNumOf(t *testing.T) {
	cases := map[string]string{
		"task-0011":              "11",
		"0011":                   "11",
		"11":                     "11",
		"task/0011-finish":       "11",
		"spec-0010":              "10",
		"task-0011-finish-thing": "11",
		"report/a-finding":       "",
		"":                       "",
	}
	for in, want := range cases {
		if got := numOf(in); got != want {
			t.Errorf("numOf(%q) = %q, want %q", in, got, want)
		}
	}
}

// The file is found at whatever width its name was written, and a miss
// names where it looked.
func TestQueueFile(t *testing.T) {
	f := vfs.NewFake()
	f.Seed("/repo/work/tasks/task-0011-finish-command.md", []byte("---\nid: task-0011\n---\n"), 0o644)
	f.Seed("/repo/work/tasks/task-0110-other.md", []byte("---\nid: task-0110\n---\n"), 0o644)
	f.Seed("/repo/work/tasks/notes.txt", []byte("x"), 0o644)

	got, err := queueFile(f, "/repo", tasksDir, "task", "11")
	if err != nil {
		t.Fatalf("queueFile: %v", err)
	}
	if got != "work/tasks/task-0011-finish-command.md" {
		t.Errorf("queueFile = %q", got)
	}
	if got, err := queueFile(f, "/repo", tasksDir, "task", "task-0110"); err != nil || !strings.Contains(got, "0110") {
		t.Errorf("queueFile(0110) = %q, %v — the padding was not respected", got, err)
	}
	if _, err := queueFile(f, "/repo", tasksDir, "task", "task-0099"); err == nil {
		t.Error("a task that is not there resolved to something")
	}
	if _, err := queueFile(f, "/repo", tasksDir, "task", "nothing"); err == nil {
		t.Error("a string naming no id resolved to something")
	}
}

// Only the front-matter block counts: a body line spelling a field at
// column 0 is prose.
func TestFieldReadsTheFrontMatterOnly(t *testing.T) {
	content := []byte("---\nid: task-0011\nstatus: ready\n---\n\n# A task\n\nstatus: done\n")
	if got := field(content, "status"); got != "ready" {
		t.Errorf("status = %q, want ready", got)
	}
	if got := field(content, "nothing"); got != "" {
		t.Errorf("a field nobody wrote = %q", got)
	}
	if got := field([]byte("# no front matter\n\nstatus: done\n"), "status"); got != "" {
		t.Errorf("a file with no front matter answered %q", got)
	}
	if got := field([]byte("---\nid: x\n"), "id"); got != "" {
		t.Errorf("front matter that never closes answered %q", got)
	}
}

func TestSetField(t *testing.T) {
	content := []byte("---\nid: spec-0010\nstatus: approved\n---\n\n# body\n\nstatus: approved\n")

	next, changed, err := setField(content, "status", "implemented")
	if err != nil || !changed {
		t.Fatalf("setField = %v, changed %v", err, changed)
	}
	if got := field(next, "status"); got != "implemented" {
		t.Errorf("status = %q", got)
	}
	if !strings.HasSuffix(string(next), "# body\n\nstatus: approved\n") {
		t.Errorf("the body was rewritten:\n%s", next)
	}

	if _, changed, err := setField(next, "status", "implemented"); err != nil || changed {
		t.Errorf("rewriting the same value reported changed=%v, %v", changed, err)
	}
	if _, _, err := setField(content, "nonesuch", "x"); err == nil {
		t.Error("a field the schema does not carry was added")
	}
	if _, _, err := setField([]byte("# no front matter\n"), "status", "x"); err == nil {
		t.Error("a file with no front matter was written into")
	}
}

func TestSpecRefs(t *testing.T) {
	cases := map[string]int{
		"spec_ref: [spec-0010]":            1,
		"spec_ref: [spec-0010, spec-0011]": 2,
		"spec_ref: []":                     0,
		"spec_ref: null":                   0,
		"spec_ref: [spec-0010,spec-0011]":  2,
		"doc_ref: product/x.md":            0,
	}
	for line, want := range cases {
		content := []byte("---\nid: task-0011\n" + line + "\n---\n")
		if got := specRefs(content); len(got) != want {
			t.Errorf("specRefs(%q) = %v, want %d", line, got, want)
		}
	}
	got := specRefs([]byte("---\nspec_ref: [spec-0010, spec-0011]\n---\n"))
	if strings.Join(got, ",") != "spec-0010,spec-0011" {
		t.Errorf("specRefs = %v", got)
	}
}

func TestOutcomeFilled(t *testing.T) {
	cases := map[string]bool{
		"What was built, and what diverged.": true,
		"":                                   false,
		"_(fill after execution)_":           false,
		"TODO":                               false,
		"\n\n":                               false,
	}
	for outcome, want := range cases {
		if got := outcomeFilled([]byte(specFixture("approved", outcome))); got != want {
			t.Errorf("outcomeFilled(%q) = %v, want %v", outcome, got, want)
		}
	}
	// A section after Outcome does not leak into it.
	trailing := "---\nid: spec-0010\n---\n\n## Outcome\n\n## Notes\n\nSomething else.\n"
	if outcomeFilled([]byte(trailing)) {
		t.Error("a later section was read as the Outcome")
	}
	// A spec with no Outcome heading at all has none filled.
	if outcomeFilled([]byte("---\nid: spec-0010\n---\n\n# spec\n\nprose\n")) {
		t.Error("a spec carrying no Outcome heading reported one")
	}
}
