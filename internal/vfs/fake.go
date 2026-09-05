package vfs

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Fake is the fake beside the port: a tree in memory, and a way to say
// *this path's calls fail*. A test that wants a failing write names the
// path and the error instead of changing a mode on the machine running
// it — which is what the read-only-file tests did, and why they proved
// less than they looked like they proved.
type Fake struct {
	entries map[string]*node
	fails   map[string]error
	tmpSeq  int
}

type node struct {
	dir  bool
	data []byte
	mode fs.FileMode
}

// NewFake returns an empty tree holding only its root.
func NewFake() *Fake {
	f := &Fake{entries: map[string]*node{}, fails: map[string]error{}}
	f.entries[string(filepath.Separator)] = &node{dir: true, mode: fs.ModeDir | 0o755}
	return f
}

// Fail makes every call touching path return err, until Heal.
func (f *Fake) Fail(path string, err error) { f.fails[clean(path)] = err }

// Heal undoes a Fail.
func (f *Fake) Heal(path string) { delete(f.fails, clean(path)) }

// Seed writes a file and every directory above it, bypassing the fail
// table — a fixture is not the act under test.
func (f *Fake) Seed(name string, data []byte, mode fs.FileMode) {
	p := clean(name)
	f.seedDirs(filepath.Dir(p))
	f.entries[p] = &node{data: append([]byte(nil), data...), mode: mode}
}

// SeedDir makes a directory and every one above it.
func (f *Fake) SeedDir(name string) { f.seedDirs(clean(name)) }

// Paths lists every path in the tree, sorted — what a test asserts
// against when the question is what survived.
func (f *Fake) Paths() []string {
	out := make([]string, 0, len(f.entries))
	for p := range f.entries {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

func (f *Fake) seedDirs(dir string) {
	dir = clean(dir)
	for {
		if _, there := f.entries[dir]; !there {
			f.entries[dir] = &node{dir: true, mode: fs.ModeDir | 0o755}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return
		}
		dir = parent
	}
}

func clean(p string) string { return filepath.Clean(p) }

func (f *Fake) failure(op, p string) error {
	if err, there := f.fails[p]; there {
		return &fs.PathError{Op: op, Path: p, Err: err}
	}
	return nil
}

func (f *Fake) ReadFile(name string) ([]byte, error) {
	p := clean(name)
	if err := f.failure("read", p); err != nil {
		return nil, err
	}
	n, there := f.entries[p]
	if !there || n.dir {
		return nil, &fs.PathError{Op: "read", Path: p, Err: fs.ErrNotExist}
	}
	return append([]byte(nil), n.data...), nil
}

func (f *Fake) WriteFile(name string, data []byte, perm fs.FileMode) error {
	p := clean(name)
	if err := f.failure("write", p); err != nil {
		return err
	}
	parent, there := f.entries[filepath.Dir(p)]
	if !there || !parent.dir {
		return &fs.PathError{Op: "write", Path: p, Err: fs.ErrNotExist}
	}
	f.entries[p] = &node{data: append([]byte(nil), data...), mode: perm}
	return nil
}

func (f *Fake) Stat(name string) (fs.FileInfo, error) {
	p := clean(name)
	if err := f.failure("stat", p); err != nil {
		return nil, err
	}
	n, there := f.entries[p]
	if !there {
		return nil, &fs.PathError{Op: "stat", Path: p, Err: fs.ErrNotExist}
	}
	return info{name: filepath.Base(p), n: n}, nil
}

func (f *Fake) MkdirAll(dir string, perm fs.FileMode) error {
	p := clean(dir)
	if err := f.failure("mkdir", p); err != nil {
		return err
	}
	f.seedDirs(p)
	return nil
}

func (f *Fake) Remove(name string) error {
	p := clean(name)
	if err := f.failure("remove", p); err != nil {
		return err
	}
	if _, there := f.entries[p]; !there {
		return &fs.PathError{Op: "remove", Path: p, Err: fs.ErrNotExist}
	}
	delete(f.entries, p)
	return nil
}

func (f *Fake) RemoveAll(dir string) error {
	p := clean(dir)
	if err := f.failure("removeall", p); err != nil {
		return err
	}
	for _, q := range f.Paths() {
		if q == p || strings.HasPrefix(q, p+string(filepath.Separator)) {
			delete(f.entries, q)
		}
	}
	return nil
}

func (f *Fake) WalkDir(root string, fn fs.WalkDirFunc) error {
	r := clean(root)
	if err := f.failure("walk", r); err != nil {
		return fn(r, nil, err)
	}
	if _, there := f.entries[r]; !there {
		return fn(r, nil, &fs.PathError{Op: "walk", Path: r, Err: fs.ErrNotExist})
	}
	for _, p := range f.Paths() {
		if p != r && !strings.HasPrefix(p, r+string(filepath.Separator)) {
			continue
		}
		n := f.entries[p]
		if err := fn(p, entry{name: filepath.Base(p), n: n}, nil); err != nil {
			if err == filepath.SkipDir {
				continue
			}
			return err
		}
	}
	return nil
}

func (f *Fake) MkdirTemp(dir, pattern string) (string, error) {
	base := dir
	if base == "" {
		base = string(filepath.Separator) + "tmp"
	}
	if err := f.failure("mkdirtemp", clean(base)); err != nil {
		return "", err
	}
	f.tmpSeq++
	p := clean(filepath.Join(base, strings.ReplaceAll(pattern, "*", "")+itoa(f.tmpSeq)))
	f.seedDirs(p)
	return p, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// info and entry are the two shapes the port hands back; both read from
// the same node, so a fake file cannot describe itself two ways.
type info struct {
	name string
	n    *node
}

func (i info) Name() string { return i.name }
func (i info) Size() int64  { return int64(len(i.n.data)) }
func (i info) Mode() fs.FileMode {
	if i.n.dir {
		return fs.ModeDir | i.n.mode.Perm()
	}
	return i.n.mode
}
func (i info) ModTime() time.Time { return time.Time{} }
func (i info) IsDir() bool        { return i.n.dir }
func (i info) Sys() any           { return nil }

type entry struct {
	name string
	n    *node
}

func (e entry) Name() string               { return e.name }
func (e entry) IsDir() bool                { return e.n.dir }
func (e entry) Type() fs.FileMode          { return info{name: e.name, n: e.n}.Mode().Type() }
func (e entry) Info() (fs.FileInfo, error) { return info{name: e.name, n: e.n}, nil }
