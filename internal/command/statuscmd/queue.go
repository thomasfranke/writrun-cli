package statuscmd

import (
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/kittag"
	"github.com/thomasfranke/writrun-cli/internal/queue"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// The queue's directories, as the methodology lays them out. They are
// read, never written.
const (
	specsDir = queue.SpecsDir
)

// branchPrefix is the branch naming convention the queue's flow fixes:
// `task/NNNN-short-name`. What the number inside it is, is
// `ql_task_num`'s answer and not this command's (internal/queue).
const branchPrefix = "task/"

// currentBranch is step 1's first half: the branch, in git's words.
func currentBranch(git gitx.Runner, root string) (string, error) {
	out, err := git(root, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("reading the current branch: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// branchLabel names the branch a reader is on. Git spells a detached
// HEAD `HEAD`, which reads as a branch called HEAD unless it is said
// plainly (spec-0013, edge cases).
func branchLabel(branch string) string {
	if branch == "" || branch == "HEAD" {
		return "detached HEAD — no branch"
	}
	return branch
}

// spec is one entry of a task's spec_ref: the id it names, and the
// status the spec file records — or the fact that no file holds it.
type spec struct {
	id     string
	status string
	found  bool
}

// task is what the branch resolved to. `named` says the branch spelled
// a task id; `found` says the queue holds it. The two differ on a
// branch named like a task the queue does not have, and that difference
// is reported rather than closed by inventing the task.
type task struct {
	named  bool
	found  bool
	id     string
	status string
	title  string
	specs  []spec
}

// line is one labelled row of the answer.
type line struct{ label, text string }

// lines renders steps 1 and 2: the task the branch carries, then its
// specs with their statuses.
func (t task) lines() []line {
	switch {
	case !t.named:
		return []line{{"Task", "none — this branch carries no task"}}
	case !t.found:
		return []line{{"Task", t.id + " — the queue holds no such task"}}
	}
	out := []line{{"Task", strings.TrimSpace(fmt.Sprintf("%s  %s  %s", t.id, t.status, t.title))}}
	if len(t.specs) == 0 {
		return append(out, line{"Spec", "none — this task names no spec"})
	}
	for i, s := range t.specs {
		label := "Spec"
		if i > 0 {
			label = ""
		}
		text := s.id + "  " + s.status
		if !s.found {
			text = s.id + " — no file under " + specsDir + "/"
		}
		out = append(out, line{label, text})
	}
	return out
}

// resolveTask is steps 1 and 2: the branch read as a task id, that id
// resolved to a queue file, and the file's specs read after it.
//
// A queue file that cannot be read is a task the answer does not hold,
// which is the line this command already gives for a branch naming one
// the queue does not have. Naming the read is `finish`'s to do, on the
// same file, when the same branch is finished.
func resolveTask(files vfs.FS, root, branch string) task {
	if !strings.HasPrefix(branch, branchPrefix) {
		return task{}
	}
	num := queue.Num(queue.Task, branch)
	if num == "" {
		return task{}
	}
	t := task{named: true, id: taskID(num)}

	rel, err := queue.Resolve(files, root, queue.Task, branch)
	if err != nil {
		return t
	}
	data, err := files.ReadFile(path.Join(root, rel))
	if err != nil {
		return t
	}
	t.found = true
	if id := queue.Field(data, "id"); id != "" {
		t.id = id
	}
	t.status = value(queue.Field(data, "status"), "no status")
	t.title = queue.Heading(data)
	for _, id := range queue.List(data, "spec_ref") {
		t.specs = append(t.specs, readSpec(files, root, id))
	}
	return t
}

// taskID spells a number the way the queue spells it: four digits, the
// width every queue file is named at. A wider number keeps its width —
// the padding is a spelling, not a limit.
func taskID(num string) string {
	if len(num) < 4 {
		num = strings.Repeat("0", 4-len(num)) + num
	}
	return "task-" + num
}

// readSpec names one spec and the status its file records.
func readSpec(files vfs.FS, root, id string) spec {
	s := spec{id: id}
	rel, err := queue.Resolve(files, root, queue.Spec, id)
	if err != nil {
		return s
	}
	data, err := files.ReadFile(path.Join(root, rel))
	if err != nil {
		return s
	}
	s.found = true
	s.status = value(queue.Field(data, "status"), "no status")
	return s
}

// openReports is step 4: the reports whose status is `open`, counted.
// Anything beside them in the directory — the README, a file carrying
// no front matter — holds no status and is not a report.
func openReports(files vfs.FS, root string) string {
	dir := path.Join(root, queue.ReportsDir)
	open := 0
	err := files.WalkDir(dir, func(p string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(p, ".md") {
			return nil
		}
		data, readErr := files.ReadFile(p)
		if readErr != nil {
			return nil
		}
		if queue.Field(data, "status") == "open" {
			open++
		}
		return nil
	})
	if err != nil {
		return fmt.Sprintf("%s/ could not be read: %v", queue.ReportsDir, err)
	}
	if open == 0 {
		return "none open"
	}
	return fmt.Sprintf("%d open, waiting to be triaged", open)
}

// recordedTag is the tag the kit records. The file and its parsing are
// kittag's, shared with `update` and `doctor`; what an unrecorded tag
// means to a status line is this command's.
func recordedTag(files vfs.FS, root string) (string, error) {
	tag, err := kittag.Read(files, root)
	if err != nil {
		return "", fmt.Errorf(".writrun/VERSION could not be read")
	}
	if tag == "" {
		return "", fmt.Errorf(".writrun/VERSION records no tag")
	}
	return tag, nil
}

// value is the field, or what to say where the file records none.
func value(field, absent string) string {
	if strings.TrimSpace(field) == "" || field == "null" {
		return absent
	}
	return field
}
