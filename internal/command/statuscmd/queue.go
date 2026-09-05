package statuscmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// The queue's directories, as the methodology lays them out. They are
// read, never written.
const (
	tasksDir   = "work/tasks"
	specsDir   = "work/specs"
	reportsDir = "work/reports"
)

// taskBranch is the branch naming convention the queue's flow fixes:
// `task/NNNN-short-name`. The leading zeros are part of how the queue
// spells an id and no part of the number, so `task/0014-x` and
// `task/14-x` name one task — the reading the methodology's own scripts
// make of a branch.
var taskBranch = regexp.MustCompile(`^task/0*([0-9]+)`)

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
func resolveTask(files vfs.FS, root, branch string) task {
	m := taskBranch.FindStringSubmatch(branch)
	if m == nil {
		return task{}
	}
	number, err := strconv.Atoi(m[1])
	if err != nil {
		return task{}
	}
	t := task{named: true, id: fmt.Sprintf("task-%04d", number)}

	path, ok := queueFile(files, filepath.Join(root, tasksDir), "task-", number)
	if !ok {
		return t
	}
	data, err := files.ReadFile(path)
	if err != nil {
		return t
	}
	t.found = true
	fm := frontMatter(data)
	if id := fm["id"]; id != "" {
		t.id = id
	}
	t.status = value(fm["status"], "no status")
	t.title = heading(data)
	for _, id := range list(fm["spec_ref"]) {
		t.specs = append(t.specs, readSpec(files, root, id))
	}
	return t
}

// readSpec names one spec and the status its file records.
func readSpec(files vfs.FS, root, id string) spec {
	s := spec{id: id}
	number, ok := idNumber(id, "spec-")
	if !ok {
		return s
	}
	path, ok := queueFile(files, filepath.Join(root, specsDir), "spec-", number)
	if !ok {
		return s
	}
	data, err := files.ReadFile(path)
	if err != nil {
		return s
	}
	s.found = true
	s.status = value(frontMatter(data)["status"], "no status")
	return s
}

// openReports is step 4: the reports whose status is `open`, counted.
// Anything beside them in the directory — the README, a file carrying
// no front matter — holds no status and is not a report.
func openReports(files vfs.FS, root string) string {
	dir := filepath.Join(root, reportsDir)
	open := 0
	err := files.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		data, readErr := files.ReadFile(path)
		if readErr != nil {
			return nil
		}
		if frontMatter(data)["status"] == "open" {
			open++
		}
		return nil
	})
	if err != nil {
		return fmt.Sprintf("%s/ could not be read: %v", reportsDir, err)
	}
	if open == 0 {
		return "none open"
	}
	return fmt.Sprintf("%d open, waiting to be triaged", open)
}

// recordedTag is the tag the kit records. The file is read here rather
// than through a shared reader: `update` reads the same file to decide
// a refresh, this command reads it to name a difference, and one small
// duplication is cheaper than a package two commands would have to
// agree on before either could land.
func recordedTag(files vfs.FS, root string) (string, error) {
	path := filepath.Join(root, ".writrun", "VERSION")
	raw, err := files.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf(".writrun/VERSION could not be read")
	}
	tag := strings.TrimSpace(string(raw))
	if tag == "" {
		return "", fmt.Errorf(".writrun/VERSION records no tag")
	}
	return tag, nil
}

// sameRelease reports whether two tags name one release. The components
// are read as numbers, so `v0.0.03` and `v0.0.3` are the same release
// and a mismatch is never announced over a spelling. Two tags neither
// side can read are the same only when they are the same text.
func sameRelease(a, b string) bool {
	if a == b {
		return true
	}
	x, okA := numbers(a)
	y, okB := numbers(b)
	if !okA || !okB {
		return false
	}
	for i := 0; i < len(x) || i < len(y); i++ {
		var xi, yi int
		if i < len(x) {
			xi = x[i]
		}
		if i < len(y) {
			yi = y[i]
		}
		if xi != yi {
			return false
		}
	}
	return true
}

// numbers reads a tag's components; ok is false for anything that is
// not `vN.N.N`.
func numbers(tag string) ([]int, bool) {
	t := strings.TrimPrefix(strings.TrimSpace(tag), "v")
	if t == "" {
		return nil, false
	}
	var out []int
	for _, p := range strings.Split(t, ".") {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, false
		}
		out = append(out, n)
	}
	return out, true
}

// queueFile finds the file holding one queue id. The number is what
// matches, not the filename: the subject slug after the id is the
// author's and no part of the identity.
func queueFile(files vfs.FS, dir, prefix string, number int) (string, bool) {
	found := ""
	_ = files.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || found != "" {
			return nil
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		name := strings.TrimSuffix(filepath.Base(path), ".md")
		if n, ok := idNumber(name, prefix); ok && n == number {
			found = path
		}
		return nil
	})
	return found, found != ""
}

// idNumber reads the number a queue id carries: `task-0014-status` is
// 14, and so is `task-14`.
func idNumber(name, prefix string) (int, bool) {
	if !strings.HasPrefix(name, prefix) {
		return 0, false
	}
	digits := name[len(prefix):]
	end := strings.IndexFunc(digits, func(r rune) bool { return r < '0' || r > '9' })
	if end >= 0 {
		digits = digits[:end]
	}
	if digits == "" {
		return 0, false
	}
	n, err := strconv.Atoi(digits)
	if err != nil {
		return 0, false
	}
	return n, true
}

// frontMatter reads the leading `---` block of a queue file into its
// fields. A file without one has no fields, which is the right answer
// for a README sitting beside the queue.
func frontMatter(data []byte) map[string]string {
	out := map[string]string{}
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	opened := false
	for sc.Scan() {
		line := sc.Text()
		if !opened {
			if strings.TrimSpace(line) != "---" {
				return out
			}
			opened = true
			continue
		}
		if strings.TrimSpace(line) == "---" {
			return out
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok || strings.HasPrefix(key, " ") {
			continue
		}
		out[strings.TrimSpace(key)] = strings.TrimSpace(val)
	}
	return out
}

// heading is the file's first `# ` line — the title the queue gave it.
func heading(data []byte) string {
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		if line := sc.Text(); strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

// list reads a front-matter list — `[spec-0013]`, `[]`, or a bare
// value — into its entries.
func list(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "null" || raw == "[]" {
		return nil
	}
	raw = strings.TrimSuffix(strings.TrimPrefix(raw, "["), "]")
	var out []string
	for _, p := range strings.Split(raw, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// value is the field, or what to say where the file records none.
func value(field, absent string) string {
	if strings.TrimSpace(field) == "" || field == "null" {
		return absent
	}
	return field
}
