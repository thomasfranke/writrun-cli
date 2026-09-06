package amendcmd

import (
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/queue"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// The queue's two folders, relative to the repository root. They are
// the methodology's own layout, not this command's choice.
const (
	tasksDir = "work/tasks"
	specsDir = "work/specs"
)

// inFlight is the pair of task statuses check_amendment_reference.sh
// calls flight: the work a returned approval suspends. Every other
// status is the ordinary pre-implementation amendment, which owes no
// reference and costs nothing (spec-0011, edge cases).
var inFlight = map[string]bool{"in-progress": true, "in-review": true}

// suspended lists the tasks this amendment suspends: the ones in flight
// whose `spec_ref` names the spec. The queue is the authority on both
// halves — check_amendment_reference.sh reads exactly these two fields
// out of exactly these files, so the composition and the gate agree
// about who is waiting.
//
// The ids come back in the walk's own order, which is the directory's,
// so a second run composes the same body as the first.
func suspended(files vfs.FS, root, specID string) ([]string, error) {
	num := queue.Num(queue.Spec, specID)
	var ids []string
	err := files.WalkDir(path.Join(root, tasksDir), func(p string, e fs.DirEntry, err error) error {
		if err != nil || e.IsDir() {
			return err
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".md") || !strings.HasPrefix(name, "task-") {
			return nil
		}
		content, err := files.ReadFile(p)
		if err != nil {
			return fmt.Errorf("reading %s: %w", tasksDir+"/"+name, err)
		}
		if !inFlight[queue.Field(content, "status")] {
			return nil
		}
		for _, ref := range queue.List(content, "spec_ref") {
			// A ref naming no spec number matches nothing, including a
			// spec whose own id names none: two unreadable ids are not
			// the same id.
			if n := queue.Num(queue.Spec, ref); n == "" || n != num {
				continue
			}
			if id := queue.Field(content, "id"); id != "" {
				ids = append(ids, id)
			}
			return nil
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// titleTag is one `[TASK-NNNN]` tag at the head of what is left of a
// title. Case is not judged, the same way the kit does not judge it.
var titleTag = regexp.MustCompile(`^\[(?i:task)-([0-9]+)\]`)

// carriedOf is ql_carried_of's rule: the task numbers a pull request
// works — its head branch's own `task/NNNN-…`, plus every `[TASK-NNNN]`
// tag leading its title, deduplicated. Both are a fork's to write, so
// only digits survive.
func carriedOf(branch, title string) []string {
	var nums []string
	add := func(num string) {
		if num == "" {
			return
		}
		for _, have := range nums {
			if have == num {
				return
			}
		}
		nums = append(nums, num)
	}
	if strings.HasPrefix(branch, "task/") {
		add(queue.Num(queue.Task, branch))
	}
	rest := strings.TrimSpace(title)
	for {
		m := titleTag.FindStringSubmatch(rest)
		if m == nil {
			break
		}
		add(strings.TrimLeft(m[1], "0"))
		rest = strings.TrimSpace(rest[len(m[0]):])
	}
	return nums
}
