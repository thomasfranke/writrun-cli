// Package queue reads the queue's files the way the adopted
// repository's own scripts read them: the front matter, one field, an
// inline list, the title, and the id-to-file resolution.
//
// One reader, because there is one answer. Four command packages
// carried four copies of this and the copies had drifted apart on
// fifteen points — whether a CRLF file has front matter, whether an
// unclosed block has fields, whether the first or the last of two
// `status:` lines wins (report-0020, spec-0022, whose Outcome
// enumerates all fifteen). Every disputed point is settled against the
// kit, never against the majority copy:
// `queue_lib.sh`'s `ql_fm_field`, `ql_set_field` and `ql_task_num`, and
// `check_front_matter.sh`'s `fm_block`. If the binary and the scripts
// disagree, the scripts are right (docs/about.md).
//
// The reading is shared; the words are not. What a command says about
// what it read stays in that command.
package queue

import (
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// Kind is the vocabulary an id belongs to. The words are the queue's
// own, which is the only vocabulary this package carries
// (technical/engineering/boundaries.md); these two are the ones this
// binary resolves ids in.
type Kind string

const (
	Task Kind = "task"
	Spec Kind = "spec"
)

// declared matches the kind an id spells for itself.
var declared = regexp.MustCompile(`^(task|spec)[-/]`)

// Declares is the vocabulary an id names for itself — `spec` in
// `spec-0011`, `task` in `task/0012-amend-command`, and nothing at all
// in a bare `0011`, which is read as whichever kind is being looked
// for.
func Declares(s string) Kind {
	m := declared.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return ""
	}
	return Kind(m[1])
}

// Num is ql_task_num's rule for one kind: the prefix of that kind
// alone, then the zero-padding, then everything from the first
// non-digit on — `sed -E 's/^task-//; s/^task\///; s/^0+//;
// s/[^0-9].*$//'`. Empty when the input names no number of that kind.
//
// The kind is a parameter because the kit's rule is: an id that spells
// the other vocabulary keeps its prefix, is truncated at the letter,
// and names no number. So `Num(Spec, "task-0012")` is empty, and the
// refusal of a cross-kind id falls out of the parser rather than
// sitting beside it. `task-0000` and `report-0020` name no number
// either: the padding is the whole of the first, and the second spells
// a third vocabulary.
func Num(kind Kind, s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, string(kind)+"-")
	s = strings.TrimPrefix(s, string(kind)+"/")
	s = strings.TrimLeft(s, "0")
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return s[:i]
		}
	}
	return s
}

// Resolve finds the one file under dir whose id is that number,
// whatever width the name was written at. A miss is an error naming
// what was looked for — a queue file that cannot be found is never a
// reason to carry on with nothing.
//
// An id that declares the other kind names no number, and the error
// says which file its digits alone would have landed on. `finish
// spec-0012` would have worked task-0012 — a real task, about
// different work, carried to the write stage without a word. The two
// counters run independently and are routinely one apart, which is
// exactly when the wrong answer looks right (report-0020).
//
// The walk's own errors are returned rather than swallowed. A caller
// that has nothing to say about them says what it says about a file it
// did not find.
func Resolve(files vfs.FS, root, dir string, kind Kind, id string) (string, error) {
	num := Num(kind, id)
	if num == "" {
		if other := Declares(id); other != "" && other != kind {
			if n := Num(other, id); n != "" {
				return "", fmt.Errorf("%q names a %s, and this resolves a %s — "+
					"its number alone would land on %s-%s, which is a different file about different work",
					id, other, kind, kind, n)
			}
		}
		return "", fmt.Errorf("%q names no %s id", id, kind)
	}
	found := ""
	err := files.WalkDir(path.Join(root, dir), func(p string, e fs.DirEntry, err error) error {
		if err != nil || e.IsDir() || found != "" {
			return err
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".md") || !strings.HasPrefix(name, string(kind)+"-") {
			return nil
		}
		if Num(kind, strings.TrimSuffix(name, ".md")) == num {
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

// block returns the lines of the leading `---` block. A file that does
// not open with one, and a block that is never closed, both hold no
// front matter at all.
//
// The fence is matched exactly, never trimmed, because the kit matches
// it exactly: `ql_fm_field` is `NR == 1 { if ($0 != "---") exit }` and
// `/^---$/`, and `fm_block` is the same pair plus `END { exit closed ?
// 0 : 1 }`. A CRLF file opens with `---\r`, which the kit reads as no
// front matter and `check_front_matter.sh` calls MALFORMED — and a
// reader that trimmed would read a status off a file the repository
// refuses, and rewrite it (report-0024).
//
// The closing fence is required because `fm_block` requires it.
// `ql_fm_field` alone would read an unclosed block to the end of the
// file, which is the body — the silent misread of a file the
// repository already refuses (spec-0022, acceptance criteria).
func block(content []byte) ([]string, bool) {
	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 || lines[0] != "---" {
		return nil, false
	}
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			return lines[1:i], true
		}
	}
	return nil, false
}

// Field reads one front-matter field. A body line spelling `status:` at
// column 0 is prose, so only the block above counts.
//
// The name is matched at column 0 with its colon against it, as
// `ql_fm_field`'s `sub("^" f ": *", "")` is: `status : approved` spells
// no field, and `status:approved` spells one. The first of two lines
// carrying the same key wins, because the kit's readers stop at the
// first — `ql_fm_field` exits on it and `check_front_matter.sh`'s `get`
// is `head -n1`.
func Field(content []byte, name string) string {
	lines, ok := block(content)
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

// List reads an inline list field — `[]`, or `[spec-0009, spec-0010]`
// — into its entries, as `ql_resting` reads `spec_ref`: the brackets
// deleted, the commas made spaces, and the words that remain.
//
// `null` is one of those words and not a filter. A list holding it is a
// file `check_front_matter.sh` refuses, and handing the word on is what
// lets the script or the command that receives it say so — dropping it
// would make this reader decide the task carries no spec.
func List(content []byte, name string) []string {
	raw := strings.NewReplacer("[", "", "]", "", ",", " ").Replace(Field(content, name))
	if items := strings.Fields(raw); len(items) > 0 {
		return items
	}
	return nil
}

// Heading is the file's first `# ` line — the title the queue gave it,
// whole. What a command shows of it is that command's.
func Heading(content []byte) string {
	for _, l := range strings.Split(string(content), "\n") {
		if rest, found := strings.CutPrefix(l, "# "); found {
			return strings.TrimSpace(rest)
		}
	}
	return ""
}

// Set rewrites one front-matter field in place and reports whether the
// file changed. A field the front matter does not already carry is an
// error rather than an addition: this writes the queue's schema, it
// does not extend it.
//
// The line is written as `name: value` and nothing else, which is
// `ql_set_field`'s `print field ": " value`. A field line that carried
// a carriage return loses it, as it does under the kit.
func Set(content []byte, name, value string) ([]byte, bool, error) {
	lines := strings.Split(string(content), "\n")
	fm, ok := block(content)
	if !ok {
		return nil, false, fmt.Errorf("no closed front matter to write %s into", name)
	}
	for i := 1; i <= len(fm); i++ {
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
