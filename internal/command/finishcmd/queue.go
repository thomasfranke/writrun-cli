package finishcmd

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

// numOf is ql_task_num's rule, applied to either vocabulary: the digits
// of an id, its zero-padding stripped, so `0011`, `task-0011` and
// `task/0011-finish-command` all resolve to the same file. Empty when
// the input names no number.
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
// ql_fm_field keeps.
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

// specRefs reads a task's `spec_ref` list. `[]` is a task with no spec,
// which is a complete answer and not a failure (spec-0010, edge cases).
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

// unfilled is what the generator writes into a fresh spec's Outcome,
// plus the placeholder a hand-written one tends to carry. Either one is
// an Outcome nobody has written.
var unfilled = map[string]bool{
	"_(fill after execution)_": true,
	"TODO":                     true,
}

// outcomeFilled reports whether the spec's `## Outcome` says anything.
// The section is the lines under the heading up to the next one; blank
// lines and the placeholder do not count as an answer.
func outcomeFilled(content []byte) bool {
	inside := false
	for _, l := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(l, "#") {
			inside = trimmed == "## Outcome"
			continue
		}
		if !inside || trimmed == "" {
			continue
		}
		if !unfilled[trimmed] {
			return true
		}
	}
	return false
}
