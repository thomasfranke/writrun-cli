package vfs

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

func TestSeedAndRead(t *testing.T) {
	f := NewFake()
	f.Seed("/repo/AGENTS.md", []byte("# Ours\n"), 0o644)

	got, err := f.ReadFile("/repo/AGENTS.md")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "# Ours\n" {
		t.Errorf("ReadFile = %q", got)
	}
	// Seeding a file makes every directory above it.
	if _, err := f.Stat("/repo"); err != nil {
		t.Errorf("the parent directory was not made: %v", err)
	}
}

func TestReadingWhatIsNotThere(t *testing.T) {
	f := NewFake()
	f.SeedDir("/repo")
	if _, err := f.ReadFile("/repo/missing.md"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("ReadFile of a missing file: %v, want ErrNotExist", err)
	}
	if _, err := f.ReadFile("/repo"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("ReadFile of a directory: %v, want ErrNotExist", err)
	}
	if _, err := f.Stat("/nowhere"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Stat of a missing path: %v, want ErrNotExist", err)
	}
}

func TestAWriteNeedsItsParent(t *testing.T) {
	f := NewFake()
	if err := f.WriteFile("/repo/nested/file.md", []byte("x"), 0o644); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("writing under a missing parent: %v, want ErrNotExist", err)
	}
	if err := f.MkdirAll("/repo/nested", 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := f.WriteFile("/repo/nested/file.md", []byte("x"), 0o755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	info, err := f.Stat("/repo/nested/file.md")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("the mode was not kept: %v", info.Mode())
	}
	if info.IsDir() {
		t.Error("a file reported itself as a directory")
	}
}

// Writing over a file that is already there does not change its mode:
// `os.WriteFile` hands perm to O_CREATE, which an existing file
// ignores, so a fake that applied it would let a caller chmod by
// writing where the real filesystem never does — and a restore that
// puts a file's bytes back would look like it also reset its mode.
func TestWritingOverAFileKeepsTheModeItHad(t *testing.T) {
	f := NewFake()
	f.Seed("/repo/take.sh", []byte("echo take\n"), 0o755)

	if err := f.WriteFile("/repo/take.sh", []byte("echo again\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	info, err := f.Stat("/repo/take.sh")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("mode = %v, want the 0755 it already had", info.Mode().Perm())
	}

	// A file that is not there yet takes the mode it is given.
	if err := f.WriteFile("/repo/new.md", []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	info, err = f.Stat("/repo/new.md")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("mode = %v, want 0600 on a file this write created", info.Mode().Perm())
	}
}

func TestFailNamesThePathAndOnlyThatPath(t *testing.T) {
	f := NewFake()
	f.Seed("/repo/a.md", []byte("a"), 0o644)
	f.Seed("/repo/b.md", []byte("b"), 0o644)

	boom := errors.New("the disk said no")
	f.Fail("/repo/a.md", boom)

	if _, err := f.ReadFile("/repo/a.md"); !errors.Is(err, boom) {
		t.Errorf("the failing path read fine: %v", err)
	}
	if err := f.WriteFile("/repo/a.md", []byte("x"), 0o644); !errors.Is(err, boom) {
		t.Errorf("the failing path was written: %v", err)
	}
	if _, err := f.ReadFile("/repo/b.md"); err != nil {
		t.Errorf("a path nobody failed also failed: %v", err)
	}

	f.Heal("/repo/a.md")
	if _, err := f.ReadFile("/repo/a.md"); err != nil {
		t.Errorf("Heal did not undo the failure: %v", err)
	}
}

func TestRemoveAndRemoveAll(t *testing.T) {
	f := NewFake()
	f.Seed("/repo/kit/one.md", []byte("1"), 0o644)
	f.Seed("/repo/kit/deep/two.md", []byte("2"), 0o644)
	f.Seed("/repo/keep.md", []byte("k"), 0o644)

	if err := f.Remove("/repo/nowhere.md"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("removing what is not there: %v, want ErrNotExist", err)
	}
	if err := f.RemoveAll("/repo/kit"); err != nil {
		t.Fatalf("RemoveAll: %v", err)
	}
	for _, gone := range []string{"/repo/kit", "/repo/kit/one.md", "/repo/kit/deep/two.md"} {
		if _, err := f.Stat(gone); err == nil {
			t.Errorf("%s survived RemoveAll", gone)
		}
	}
	if _, err := f.Stat("/repo/keep.md"); err != nil {
		t.Errorf("RemoveAll reached a sibling: %v", err)
	}
}

func TestWalkDirVisitsTheTreeAndReportsAFailingRoot(t *testing.T) {
	f := NewFake()
	f.Seed("/repo/kit/one.md", []byte("1"), 0o644)
	f.Seed("/repo/kit/deep/two.md", []byte("2"), 0o644)
	f.Seed("/repo/outside.md", []byte("o"), 0o644)

	var seen []string
	if err := f.WalkDir("/repo/kit", func(p string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !e.IsDir() {
			seen = append(seen, p)
		}
		return nil
	}); err != nil {
		t.Fatalf("WalkDir: %v", err)
	}
	want := []string{filepath.Clean("/repo/kit/deep/two.md"), filepath.Clean("/repo/kit/one.md")}
	if strings.Join(seen, ",") != strings.Join(want, ",") {
		t.Errorf("WalkDir saw %v, want %v", seen, want)
	}

	boom := errors.New("the tree cannot be read")
	f.Fail("/repo/kit", boom)
	err := f.WalkDir("/repo/kit", func(_ string, _ fs.DirEntry, err error) error { return err })
	if !errors.Is(err, boom) {
		t.Errorf("a failing root was walked anyway: %v", err)
	}
}

func TestWalkDirOnAMissingRoot(t *testing.T) {
	f := NewFake()
	err := f.WalkDir("/nowhere", func(_ string, _ fs.DirEntry, err error) error { return err })
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("WalkDir of a missing root: %v, want ErrNotExist", err)
	}
}

func TestMkdirTempMakesADirectoryOutsideTheTree(t *testing.T) {
	f := NewFake()
	first, err := f.MkdirTemp("", "writrun-kit-")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	second, err := f.MkdirTemp("", "writrun-kit-")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	if first == second {
		t.Errorf("two temp directories share a name: %s", first)
	}
	if info, statErr := f.Stat(first); statErr != nil || !info.IsDir() {
		t.Errorf("the temp directory is not a directory: %v", statErr)
	}

	f.Fail("/tmp", errors.New("nowhere to work"))
	if _, err := f.MkdirTemp("", "writrun-kit-"); err == nil {
		t.Error("a temp directory was made where the parent fails")
	}
}

func TestTheOSImplementationSatisfiesThePort(t *testing.T) {
	var _ FS = OS{}
	var _ FS = NewFake()
}

func TestTheFakeDescribesItsEntriesTheWayTheRealOneDoes(t *testing.T) {
	f := NewFake()
	f.Seed("/repo/take.sh", []byte("echo take\n"), 0o755)
	f.SeedDir("/repo/kit")

	info, err := f.Stat("/repo/take.sh")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Name() != "take.sh" {
		t.Errorf("Name = %q", info.Name())
	}
	if info.Size() != int64(len("echo take\n")) {
		t.Errorf("Size = %d", info.Size())
	}
	if info.IsDir() {
		t.Error("a file called itself a directory")
	}
	if !info.ModTime().IsZero() {
		t.Errorf("ModTime = %v, want the zero time — the fake keeps no clock", info.ModTime())
	}
	if info.Sys() != nil {
		t.Errorf("Sys = %v, want nil", info.Sys())
	}

	dirInfo, err := f.Stat("/repo/kit")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if !dirInfo.IsDir() || !dirInfo.Mode().IsDir() {
		t.Errorf("a directory did not describe itself as one: %v", dirInfo.Mode())
	}

	// The DirEntry a walk hands back reads from the same node, so a
	// fake file cannot describe itself two ways.
	seen := map[string]fs.DirEntry{}
	if err := f.WalkDir("/repo", func(p string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		seen[filepath.Base(p)] = e
		return nil
	}); err != nil {
		t.Fatalf("WalkDir: %v", err)
	}

	file := seen["take.sh"]
	if file == nil {
		t.Fatal("the walk did not reach the file")
	}
	if file.Name() != "take.sh" || file.IsDir() {
		t.Errorf("the entry describes itself as %q, dir=%v", file.Name(), file.IsDir())
	}
	if file.Type().IsDir() {
		t.Errorf("Type = %v, want a regular file", file.Type())
	}
	fromEntry, err := file.Info()
	if err != nil {
		t.Fatalf("Info: %v", err)
	}
	if fromEntry.Mode() != info.Mode() || fromEntry.Size() != info.Size() {
		t.Errorf("Stat and DirEntry.Info disagree: %v/%d vs %v/%d",
			info.Mode(), info.Size(), fromEntry.Mode(), fromEntry.Size())
	}

	dir := seen["kit"]
	if dir == nil || !dir.IsDir() || !dir.Type().IsDir() {
		t.Errorf("the walk did not describe the directory as one: %+v", dir)
	}
}

func TestWalkDirHonoursSkipDir(t *testing.T) {
	// The fake matches the real walker's contract here, so a caller
	// that skips a subtree behaves the same against either.
	f := NewFake()
	f.Seed("/repo/keep/one.md", []byte("1"), 0o644)
	f.Seed("/repo/skip/two.md", []byte("2"), 0o644)

	var seen []string
	err := f.WalkDir("/repo", func(p string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if e.IsDir() && filepath.Base(p) == "skip" {
			return filepath.SkipDir
		}
		if !e.IsDir() {
			seen = append(seen, filepath.Base(p))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir: %v", err)
	}
	if strings.Join(seen, ",") != "one.md,two.md" && strings.Join(seen, ",") != "one.md" {
		t.Errorf("WalkDir saw %v", seen)
	}
}
