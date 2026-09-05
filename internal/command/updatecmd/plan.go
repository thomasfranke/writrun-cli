package updatecmd

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/fence"
	"github.com/thomasfranke/writrun-cli/internal/kitpaths"
)

// verb is what a refresh does to one path.
type verb string

const (
	added   verb = "add"
	changed verb = "change"
	removed verb = "remove"
)

// change is one file the refresh touches, named before it is touched.
type change struct {
	rel  string
	verb verb
	src  string // where it comes from in the fetched template; empty on a removal
	mode fs.FileMode
}

// refresh is the whole plan, computed before anything is written and
// shown before the confirmation.
type refresh struct {
	root     string
	template string
	from, to string

	changes []change
	// dirs are the kit-owned directories replaced whole; a directory
	// the new tag no longer ships is removed rather than emptied.
	dirs []string

	agentsPath string
	agents     []byte // the merged document; nil when the section is already current
}

func plan(root, template, from, to string, agents []byte) (*refresh, error) {
	r := &refresh{
		root:       root,
		template:   template,
		from:       from,
		to:         to,
		agentsPath: filepath.Join(root, "AGENTS.md"),
	}

	for _, dir := range kitpaths.RefreshDirs {
		cs, err := diffTree(root, template, dir)
		if err != nil {
			return nil, err
		}
		r.changes = append(r.changes, cs...)
		r.dirs = append(r.dirs, dir)
	}

	for _, rel := range kitpaths.Workflows {
		c, err := diffFile(root, template, rel)
		if err != nil {
			return nil, err
		}
		if c != nil {
			r.changes = append(r.changes, *c)
		}
	}

	// The tag is recorded from what was fetched, never copied from the
	// template's own file — the same rule the adoption follows.
	r.changes = append(r.changes, change{rel: ".writrun/VERSION", verb: changed})

	merged, err := mergeAgents(template, agents)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(merged, agents) {
		r.agents = merged
	}
	return r, nil
}

// mergeAgents refreshes the fenced section of the project's AGENTS.md,
// carrying its `yours` blocks across.
func mergeAgents(template string, agents []byte) ([]byte, error) {
	raw, err := os.ReadFile(filepath.Join(template, "AGENTS.md"))
	if err != nil {
		return nil, fmt.Errorf("reading the template's AGENTS.md: %w", err)
	}
	section, err := fence.Section(raw)
	if err != nil {
		return nil, err
	}
	return fence.Replace(agents, section)
}

// diffTree names every file that differs between the fetched tree and
// the repository's copy of one kit-owned directory.
func diffTree(root, template, dir string) ([]change, error) {
	want, err := readTree(filepath.Join(template, dir))
	if err != nil {
		return nil, err
	}
	have, err := readTree(filepath.Join(root, dir))
	if err != nil {
		return nil, err
	}

	var out []change
	for _, rel := range sortedKeys(want) {
		src := filepath.Join(template, dir, rel)
		info, statErr := os.Stat(src)
		if statErr != nil {
			return nil, statErr
		}
		c := change{rel: filepath.ToSlash(filepath.Join(dir, rel)), src: src, mode: info.Mode()}
		old, there := have[rel]
		switch {
		case !there:
			c.verb = added
		case !bytes.Equal(old, want[rel]):
			c.verb = changed
		default:
			continue
		}
		out = append(out, c)
	}
	for _, rel := range sortedKeys(have) {
		if _, there := want[rel]; !there {
			out = append(out, change{rel: filepath.ToSlash(filepath.Join(dir, rel)), verb: removed})
		}
	}
	return out, nil
}

// diffFile is diffTree for a single path; nil means the two agree.
func diffFile(root, template, rel string) (*change, error) {
	src := filepath.Join(template, rel)
	want, wantErr := os.ReadFile(src)
	have, haveErr := os.ReadFile(filepath.Join(root, rel))
	switch {
	case wantErr != nil && haveErr != nil:
		return nil, nil
	case wantErr != nil:
		return &change{rel: rel, verb: removed}, nil
	case haveErr != nil:
		info, err := os.Stat(src)
		if err != nil {
			return nil, err
		}
		return &change{rel: rel, verb: added, src: src, mode: info.Mode()}, nil
	case bytes.Equal(want, have):
		return nil, nil
	default:
		info, err := os.Stat(src)
		if err != nil {
			return nil, err
		}
		return &change{rel: rel, verb: changed, src: src, mode: info.Mode()}, nil
	}
}

// readTree reads every file under dir, keyed by its path relative to
// dir. A directory that is not there is an empty tree, not a failure:
// a tag may add a folder the adopted kit never had.
func readTree(dir string) (map[string][]byte, error) {
	out := map[string][]byte{}
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			return relErr
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		out[rel] = content
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return out, nil
}

func sortedKeys(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// empty reports a refresh with nothing to do — the tag moved but the
// files it owns did not.
func (r *refresh) empty() bool {
	return len(r.changes) == 1 && r.changes[0].rel == ".writrun/VERSION" && r.agents == nil
}

// render prints what will change before anything changes.
func (r *refresh) render(w io.Writer) {
	fmt.Fprintf(w, "writrun update — WritRun %s → %s; nothing is written before the confirmation:\n\n", r.from, r.to)
	if r.empty() {
		fmt.Fprintf(w, "  Only the recorded tag differs; every kit-owned file already matches %s.\n\n", r.to)
		return
	}
	counts := map[verb]int{}
	tops := map[string]map[verb]int{}
	var order []string
	for _, c := range r.changes {
		counts[c.verb]++
		top := c.rel
		if i := strings.IndexByte(top, '/'); i >= 0 {
			if j := strings.IndexByte(top[i+1:], '/'); j >= 0 {
				top = top[:i+1+j] + "/"
			}
		}
		if _, there := tops[top]; !there {
			tops[top] = map[verb]int{}
			order = append(order, top)
		}
		tops[top][c.verb]++
	}
	sort.Strings(order)
	for _, top := range order {
		parts := make([]string, 0, 3)
		for _, v := range []verb{added, changed, removed} {
			if n := tops[top][v]; n > 0 {
				parts = append(parts, fmt.Sprintf("%d to %s", n, v))
			}
		}
		fmt.Fprintf(w, "  %-34s %s\n", top, strings.Join(parts, ", "))
	}
	if r.agents != nil {
		fmt.Fprintln(w, "  AGENTS.md                          the fenced section; the lines marked `yours` are carried across")
	} else {
		fmt.Fprintln(w, "  AGENTS.md                          the fenced section already matches — left alone")
	}
	fmt.Fprintf(w, "\n  untouched    %s\n\n", strings.Join(kitpaths.Untouchable, ", "))
}

// apply performs exactly the rendered plan: the kit-owned directories
// replaced whole, the named files rewritten, the tag recorded, and the
// fenced section swapped last.
func (r *refresh) apply() error {
	for _, dir := range r.dirs {
		src := filepath.Join(r.template, dir)
		dst := filepath.Join(r.root, dir)
		if _, err := os.Stat(src); err != nil {
			// The tag no longer ships it; the kit's copy goes with it.
			if err := os.RemoveAll(dst); err != nil {
				return fmt.Errorf("removing %s: %w", dir, err)
			}
			continue
		}
		if err := os.RemoveAll(dst); err != nil {
			return fmt.Errorf("replacing %s: %w", dir, err)
		}
		if err := copyTree(src, dst); err != nil {
			return fmt.Errorf("replacing %s: %w", dir, err)
		}
	}

	for _, c := range r.changes {
		if c.rel == ".writrun/VERSION" {
			continue
		}
		if !strings.HasPrefix(c.rel, ".github/") {
			continue // the directories above already carried it
		}
		dst := filepath.Join(r.root, c.rel)
		if c.verb == removed {
			if err := os.Remove(dst); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("removing %s: %w", c.rel, err)
			}
			continue
		}
		if err := copyFile(c.src, dst, c.mode); err != nil {
			return fmt.Errorf("writing %s: %w", c.rel, err)
		}
	}

	if err := os.WriteFile(filepath.Join(r.root, ".writrun", "VERSION"), []byte(r.to+"\n"), 0o644); err != nil {
		return fmt.Errorf("recording the tag: %w", err)
	}
	if r.agents != nil {
		if err := os.WriteFile(r.agentsPath, r.agents, 0o644); err != nil {
			return fmt.Errorf("refreshing AGENTS.md: %w", err)
		}
	}
	return nil
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(src, path)
		if relErr != nil {
			return relErr
		}
		target := filepath.Join(dst, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			return infoErr
		}
		return copyFile(path, target, info.Mode())
	})
}

func copyFile(src, dst string, mode fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, content, mode.Perm())
}
