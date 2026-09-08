package updatecmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/kit"
	"github.com/thomasfranke/writrun-cli/internal/kitpaths"
	"github.com/thomasfranke/writrun-cli/internal/kittag"
	"github.com/thomasfranke/writrun-cli/internal/pointer"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// verb is what a refresh does to one path.
type verb string

const (
	added   verb = "add"
	changed verb = "change"
	removed verb = "remove"
	seeded  verb = "seed"
)

// change is one file the refresh touches, named before it is touched.
type change struct {
	rel  string // slash-separated, relative to the repository root
	verb verb
}

// refresh is the whole plan, computed before anything is written and
// shown before the confirmation.
type refresh struct {
	disk     vfs.FS
	root     string
	template string
	from, to string

	changes []change
	// kept are the adopter-owned files the tag ships and the repository
	// already has. They are named because a reader is being promised
	// they survive, not because anything happens to them.
	kept []string
	// legacy says AGENTS.md still carries the fenced section a kit
	// before v0.0.04 grafted. A refresh names it and rewrites nothing:
	// from v0.0.04 that file is the project's, whole.
	legacy bool
	// moves are the adopter's files this refresh carries from the kit's
	// home to the project's, once. stranded are the ones it will not,
	// because both addresses already hold a file.
	moves, stranded []move
}

// plan walks the fetched template and decides every write without
// performing one. What the template ships is what a refresh writes,
// minus the adopter's own paths — so a tag that adds a file needs no
// change here (docs/technical/engineering/coupling.md).
func plan(disk vfs.FS, root, template, from, to string) (*refresh, error) {
	r := &refresh{disk: disk, root: root, template: template, from: from, to: to}

	want, err := readTree(disk, template)
	if err != nil {
		return nil, err
	}
	for _, rel := range sortedKeys(want) {
		switch {
		case kitpaths.Seeds(rel):
			if _, err := disk.Stat(localOf(root, rel)); err == nil {
				r.kept = append(r.kept, rel)
				continue
			}
			r.changes = append(r.changes, change{rel: rel, verb: seeded})
		case kitpaths.Untouched(rel):
			continue
		default:
			have, readErr := disk.ReadFile(localOf(root, rel))
			switch {
			case readErr != nil:
				r.changes = append(r.changes, change{rel: rel, verb: added})
			case !bytes.Equal(have, want[rel]):
				r.changes = append(r.changes, change{rel: rel, verb: changed})
			}
		}
	}

	gone, err := r.dropped(want)
	if err != nil {
		return nil, err
	}
	r.changes = append(r.changes, gone...)

	// The tag is recorded from what was fetched, never copied from the
	// template's own file — the same rule the adoption follows. It is
	// listed whether or not the file differs, because the refresh's
	// whole point is that the recorded tag moves.
	if !named(r.changes, kittag.Rel) {
		r.changes = append(r.changes, change{rel: kittag.Rel, verb: changed})
	}

	agents, err := disk.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err == nil {
		r.legacy = pointer.Legacy(agents)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("reading AGENTS.md: %w", err)
	}
	r.moves, r.stranded = migrations(disk, root)

	return r, nil
}

// dropped names the kit's files the repository still holds and the tag
// no longer ships. It reaches only where the kit's own files live:
// everything under `.writrun/` that is not the adopter's, and the two
// `.github` folders where the kit namespaces what is its.
func (r *refresh) dropped(want map[string][]byte) ([]change, error) {
	var out []change
	seen := map[string]bool{}
	roots := append([]string{".writrun"}, kitpaths.NamespacedDirs()...)
	for _, dir := range roots {
		have, err := readTree(r.disk, filepath.Join(r.root, dir))
		if err != nil {
			return nil, err
		}
		for _, rel := range sortedKeys(have) {
			full := dir + "/" + rel
			if _, there := want[full]; there || seen[full] {
				continue
			}
			if !kitpaths.Removable(full) {
				continue
			}
			seen[full] = true
			out = append(out, change{rel: full, verb: removed})
		}
	}
	return out, nil
}

// readTree reads every file under dir, keyed by its slash-separated
// path relative to dir. A directory that is not there is an empty tree,
// not a failure: a tag may add a folder the adopted kit never had.
func readTree(disk vfs.FS, dir string) (map[string][]byte, error) {
	out := map[string][]byte{}
	err := disk.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
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
		content, readErr := disk.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		out[filepath.ToSlash(rel)] = content
		return nil
	})
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	return out, nil
}

func localOf(root, rel string) string {
	return filepath.Join(root, filepath.FromSlash(rel))
}

func named(changes []change, rel string) bool {
	for _, c := range changes {
		if c.rel == rel {
			return true
		}
	}
	return false
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
	if len(r.moves) > 0 || len(r.stranded) > 0 {
		return false
	}
	return len(r.changes) == 1 && r.changes[0].rel == kittag.Rel
}

// render prints what will change before anything changes.
func (r *refresh) render(w io.Writer) {
	fmt.Fprintf(w, "writrun update — WritRun %s → %s; nothing is written before the confirmation:\n\n", r.from, r.to)
	if r.empty() {
		fmt.Fprintf(w, "  Only the recorded tag differs; every kit-owned file already matches %s.\n\n", r.to)
		return
	}
	for _, m := range r.moves {
		fmt.Fprintf(w, "  move         %s → %s — yours, and it leaves the kit's home\n", m.from, m.to)
	}
	for _, m := range r.stranded {
		fmt.Fprintf(w, "  left         %s — %s already answers; nothing is merged\n", m.from, m.to)
	}
	if len(r.moves) > 0 || len(r.stranded) > 0 {
		fmt.Fprintln(w)
	}
	tops := map[string]map[verb]int{}
	var order []string
	for _, c := range r.changes {
		top := group(c.rel)
		if _, there := tops[top]; !there {
			tops[top] = map[verb]int{}
			order = append(order, top)
		}
		tops[top][c.verb]++
	}
	sort.Strings(order)
	for _, top := range order {
		parts := make([]string, 0, 4)
		for _, v := range []verb{added, changed, removed, seeded} {
			if n := tops[top][v]; n > 0 {
				parts = append(parts, fmt.Sprintf("%d to %s", n, v))
			}
		}
		fmt.Fprintf(w, "  %-34s %s\n", top, strings.Join(parts, ", "))
	}
	for _, rel := range r.kept {
		fmt.Fprintf(w, "  %-34s yours; this tag ships one and it is left alone\n", rel)
	}
	if r.legacy {
		fmt.Fprintln(w, "\n  AGENTS.md still carries a writrun:begin/writrun:end section. From")
		fmt.Fprintf(w, "  WritRun %s the flow lives in %s and that file is yours,\n", r.to, pointer.Target)
		fmt.Fprintln(w, "  whole — so this refresh does not touch it. Cutting the stale section")
		fmt.Fprintln(w, "  is yours to do.")
	}
	fmt.Fprintf(w, "\n  untouched    %s\n\n", strings.Join(kitpaths.Untouchable, ", "))
}

// group is the heading a change is counted under: the first two path
// segments for a nested file, the path itself for a shallow one.
func group(rel string) string {
	if i := strings.IndexByte(rel, '/'); i >= 0 {
		if j := strings.IndexByte(rel[i+1:], '/'); j >= 0 {
			return rel[:i+1+j] + "/"
		}
	}
	return rel
}

// apply performs exactly the rendered plan.
func (r *refresh) apply() error {
	// The migration runs first: a repository left half-refreshed is
	// recoverable, and one whose answers were overwritten at the old
	// address before they moved is not.
	for _, m := range r.moves {
		if err := r.migrate(m); err != nil {
			return err
		}
	}
	for _, c := range r.changes {
		dst := localOf(r.root, c.rel)
		if c.verb == removed {
			if err := r.disk.Remove(dst); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("removing %s: %w", c.rel, err)
			}
			continue
		}
		if c.rel == kittag.Rel {
			continue // recorded below, from what was fetched
		}
		if err := copyFile(r.disk, localOf(r.template, c.rel), dst); err != nil {
			return fmt.Errorf("writing %s: %w", c.rel, err)
		}
	}
	if err := r.disk.WriteFile(kittag.Path(r.root), []byte(r.to+"\n"), 0o644); err != nil {
		return fmt.Errorf("recording the tag: %w", err)
	}
	return nil
}

func copyFile(disk vfs.FS, src, dst string) error {
	info, err := disk.Stat(src)
	if err != nil {
		return err
	}
	if err := disk.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	content, err := disk.ReadFile(src)
	if err != nil {
		return err
	}
	return disk.WriteFile(dst, content, info.Mode().Perm())
}

// move is one adopter file leaving the kit's home for the project's.
// Both paths are slash-separated and relative to the repository root.
type move struct{ from, to string }

// migrations are the adopter's files still sitting at the addresses
// WritRun v0.0.05 moved them out of. A repository adopted before that
// tag keeps its answers under `.writrun/`, which the kit's own scripts
// no longer read — so the refresh carries them across once, and never
// looks again.
//
// Where both addresses hold the file the new one wins and the old one
// is named, never merged: two answers about one setting are the
// adopter's to reconcile, and a merge would pick for them
// (WritRun spec-0090).
func migrations(disk vfs.FS, root string) (moves, stranded []move) {
	for _, m := range []move{
		{kit.LegacySettings, kit.Settings},
		{kit.LegacyGates, kit.Gates},
		{kit.LegacyConventions, kit.Conventions},
	} {
		if !present(disk, localOf(root, m.from)) {
			continue
		}
		if present(disk, localOf(root, m.to)) {
			stranded = append(stranded, m)
			continue
		}
		moves = append(moves, m)
	}
	return moves, stranded
}

func present(disk vfs.FS, path string) bool {
	_, err := disk.Stat(path)
	return err == nil
}

// migrate carries one address across, byte for byte. The port has no
// rename, and a copy followed by a removal is the same act through it.
func (r *refresh) migrate(m move) error {
	src, dst := localOf(r.root, m.from), localOf(r.root, m.to)
	info, err := r.disk.Stat(src)
	if err != nil {
		return fmt.Errorf("reading %s: %w", m.from, err)
	}
	if !info.IsDir() {
		if err := copyFile(r.disk, src, dst); err != nil {
			return fmt.Errorf("writing %s: %w", m.to, err)
		}
		return r.disk.Remove(src)
	}
	err = r.disk.WalkDir(src, func(p string, e fs.DirEntry, err error) error {
		if err != nil || e.IsDir() {
			return err
		}
		rel, relErr := filepath.Rel(src, p)
		if relErr != nil {
			return relErr
		}
		return copyFile(r.disk, p, filepath.Join(dst, rel))
	})
	if err != nil {
		return fmt.Errorf("writing %s: %w", m.to, err)
	}
	return r.disk.RemoveAll(src)
}
