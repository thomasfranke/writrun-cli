package amendcmd

import (
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"strings"

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

// numOf is ql_task_num's rule, applied to either vocabulary: the digits
// of an id, its zero-padding stripped, so `0011`, `spec-0011` and
// `task/0011-amend-command` all resolve to the same number. Empty when
// the input names none.
var numOf = func() func(string) string {
	strip := regexp.MustCompile(`^(task|spec)[-/]`)
	digits := regexp.MustCompile(`[0-9]+`)
	return func(s string) string {
		s = strip.ReplaceAllString(strings.TrimSpace(s), "")
		d := digits.FindString(s)
		return strings.TrimLeft(d, "0")
	}
}()

// queueFile finds the one file under dir whose id is that number,
// whatever width the name was written at. A miss is an error naming
// what was looked for — a queue file that cannot be found is never a
// reason to carry on with nothing.
func queueFile(files vfs.FS, root, dir, kind, id string) (string, error) {
	num := numOf(id)
	if num == "" {
		return "", fmt.Errorf("%q names no %s id", id, kind)
	}
	found := ""
	err := files.WalkDir(path.Join(root, dir), func(p string, e fs.DirEntry, err error) error {
		if err != nil || e.IsDir() || found != "" {
			return err
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".md") || !strings.HasPrefix(name, kind+"-") {
			return nil
		}
		if numOf(strings.TrimSuffix(name, ".md")) == num {
			found = dir + "/" + name
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", dir, err)
	}
	if found == "" {
		return "", fmt.Errorf("%s-%s resolves to no file under %s/", kind, num, dir)
	}
	return found, nil
}

// frontMatter returns the lines of the leading `---` block. A file that
// does not open with one has no front matter at all, and every reader
// here says so rather than guessing where it ends.
func frontMatter(content []byte) ([]string, bool) {
	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil, false
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return lines[1:i], true
		}
	}
	return nil, false
}

// field reads one front-matter field. A body line spelling `status:` at
// column 0 is prose, so only the block above counts — the same rule
// ql_fm_field keeps, and the same rule check_amendment_reference.sh
// relies on when it decides what a change returned to draft.
func field(content []byte, name string) string {
	lines, ok := frontMatter(content)
	if !ok {
		return ""
	}
	for _, l := range lines {
		if rest, found := strings.CutPrefix(l, name+":"); found {
			return strings.TrimSpace(rest)
		}
	}
	return ""
}

// setField rewrites one front-matter field in place and reports whether
// the file changed. A field the front matter does not already carry is
// an error rather than an addition: this command edits the queue's
// schema, it does not extend it.
func setField(content []byte, name, value string) ([]byte, bool, error) {
	lines := strings.Split(string(content), "\n")
	fmLines, ok := frontMatter(content)
	if !ok {
		return nil, false, fmt.Errorf("no closed front matter to write %s into", name)
	}
	for i := 1; i <= len(fmLines); i++ {
		if !strings.HasPrefix(lines[i], name+":") {
			continue
		}
		want := name + ": " + value
		if lines[i] == want {
			return content, false, nil
		}
		lines[i] = want
		return []byte(strings.Join(lines, "\n")), true, nil
	}
	return nil, false, fmt.Errorf("the front matter carries no %s field", name)
}

// specRefs reads a task's `spec_ref` list.
func specRefs(content []byte) []string {
	raw := strings.TrimSpace(field(content, "spec_ref"))
	raw = strings.TrimPrefix(raw, "[")
	raw = strings.TrimSuffix(raw, "]")
	var ids []string
	for _, part := range strings.Split(raw, ",") {
		if id := strings.TrimSpace(part); id != "" && id != "null" {
			ids = append(ids, id)
		}
	}
	return ids
}

// suspended lists the tasks this amendment suspends: the ones in flight
// whose `spec_ref` names the spec. The queue is the authority on both
// halves — check_amendment_reference.sh reads exactly these two fields
// out of exactly these files, so the composition and the gate agree
// about who is waiting.
//
// The ids come back in the walk's own order, which is the directory's,
// so a second run composes the same body as the first.
func suspended(files vfs.FS, root, specID string) ([]string, error) {
	num := numOf(specID)
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
		if !inFlight[field(content, "status")] {
			return nil
		}
		for _, ref := range specRefs(content) {
			if numOf(ref) != num {
				continue
			}
			if id := field(content, "id"); id != "" {
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
		add(numOf(branch))
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
