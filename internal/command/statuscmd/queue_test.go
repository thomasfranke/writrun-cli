package statuscmd

import (
	"errors"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// The front matter, the id and the file are internal/queue's to read,
// and its tests hold the disputed readings. What is held here is what
// this command makes of the answers.

// The branch names a task by the id it spells, at the width the queue
// spells it — `ql_task_num`'s number, padded back to four digits.
func TestTaskIdIsSpelledAtTheQueuesWidth(t *testing.T) {
	cases := map[string]string{
		"14":    "task-0014",
		"1":     "task-0001",
		"1234":  "task-1234",
		"12345": "task-12345",
	}
	for num, want := range cases {
		if got := taskID(num); got != want {
			t.Errorf("taskID(%q) = %q, want %q", num, got, want)
		}
	}
}

// A branch is read as a task only where it is spelled as one, and the
// number inside it is the kit's. `task/0000-x` names no task: the
// padding is the whole of its number (`ql_task_num task-0000` is
// empty).
func TestOnlyATaskBranchNamesATask(t *testing.T) {
	files := vfs.NewFake()
	files.Seed(root+"/work/tasks/task-0014-status-command.md", []byte(taskFile), 0o644)
	files.Seed(root+"/work/specs/spec-0013-status.md", []byte(specFile), 0o644)

	cases := map[string]struct{ named, found bool }{
		"task/0014-status-command": {true, true},
		"task/14-status-command":   {true, true},
		"task/0099-nothing-here":   {true, false},
		"task/0000-nothing":        {false, false},
		"task/abc-0014":            {false, false},
		"docs/something":           {false, false},
		"main":                     {false, false},
		"":                         {false, false},
	}
	for branch, want := range cases {
		got := resolveTask(files, root, branch)
		if got.named != want.named || got.found != want.found {
			t.Errorf("resolveTask(%q) = named %v, found %v; want %v, %v",
				branch, got.named, got.found, want.named, want.found)
		}
	}
}

// A queue file the walk cannot reach is a task the answer does not
// hold, in the words this command already gives a branch naming one the
// queue has not.
func TestAQueueThatCannotBeWalkedIsATaskNotHeld(t *testing.T) {
	files := vfs.NewFake()
	files.Seed(root+"/work/tasks/task-0014-status-command.md", []byte(taskFile), 0o644)
	files.FailOp("walk", root+"/work/tasks", errors.New("permission denied"))
	got := resolveTask(files, root, "task/0014-status-command")
	if !got.named || got.found {
		t.Fatalf("resolveTask = %+v; want the task named and not found", got)
	}
	if lines := got.lines(); lines[0].text != "task-0014 — the queue holds no such task" {
		t.Errorf("line = %q", lines[0].text)
	}
}

func TestValueNamesWhatTheFileDoesNotRecord(t *testing.T) {
	if got := value("approved", "no status"); got != "approved" {
		t.Errorf("value = %q", got)
	}
	if got := value("null", "no status"); got != "no status" {
		t.Errorf("value = %q; want the absence named", got)
	}
	if got := value("  ", "no status"); got != "no status" {
		t.Errorf("value = %q; want the absence named", got)
	}
}
