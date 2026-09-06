package kittag

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

func TestPathIsTheFileTheAdoptionWrote(t *testing.T) {
	want := filepath.Join("repo", ".writrun", "VERSION")
	if got := Path("repo"); got != want {
		t.Errorf("Path(%q) = %q, want %q", "repo", got, want)
	}
}

func TestReadTrimsWhatTheFileHolds(t *testing.T) {
	root := t.TempDir()
	write(t, root, "  v0.0.03 \n")
	tag, err := Read(vfs.OS{}, root)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if tag != "v0.0.03" {
		t.Errorf("tag = %q, want %q", tag, "v0.0.03")
	}
}

// An unrecorded tag is not this package's refusal: the three callers
// word it three ways, so Read hands back the empty string and no error.
func TestReadAnswersAnEmptyFileWithNoError(t *testing.T) {
	root := t.TempDir()
	write(t, root, "   \n")
	tag, err := Read(vfs.OS{}, root)
	if err != nil || tag != "" {
		t.Errorf("Read = %q, %v; want %q and no error", tag, err, "")
	}
}

func TestReadReportsAFileThatIsNotThere(t *testing.T) {
	if _, err := Read(vfs.OS{}, t.TempDir()); err == nil {
		t.Error("a missing .writrun/VERSION read without an error")
	}
}

// The error is the filesystem's own, with nothing of this package's
// added: `update` prints it behind its own "reading .writrun/VERSION:",
// and a second wrapping here would say the file twice. Read must hand
// back exactly what the port returned, whatever the fault was.
func TestReadWrapsNothingAroundTheFilesystemsError(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(*vfs.Fake, string)
	}{
		{"permission denied", func(f *vfs.Fake, root string) {
			f.Seed(Path(root), []byte("v0.0.3\n"), 0o644)
			f.FailOp("read", Path(root), fs.ErrPermission)
		}},
		{"a directory where the file goes", func(f *vfs.Fake, root string) {
			f.SeedDir(Path(root))
		}},
		{"the file is not there", func(f *vfs.Fake, root string) {
			f.SeedDir(filepath.Join(root, ".writrun"))
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := "repo"
			f := vfs.NewFake()
			tc.setup(f, root)

			tag, err := Read(f, root)
			if err == nil {
				t.Fatalf("Read = %q, nil; want an error", tag)
			}
			if tag != "" {
				t.Errorf("tag = %q, want the empty string beside an error", tag)
			}
			_, direct := f.ReadFile(Path(root))
			if direct == nil {
				t.Fatal("the fixture is not failing the read at all")
			}
			if err.Error() != direct.Error() {
				t.Errorf("Read error = %q; the port's own is %q — something wrapped it",
					err.Error(), direct.Error())
			}
			var pathErr *fs.PathError
			if !errors.As(err, &pathErr) {
				t.Fatalf("error %v is not the port's *fs.PathError", err)
			}
			if pathErr.Path != Path(root) {
				t.Errorf("error names %q, want %q", pathErr.Path, Path(root))
			}
		})
	}
}

// The permission case again, at the real filesystem rather than the
// fake: a mode-000 file is the state a reader actually meets.
func TestReadReportsAFileItMayNotOpen(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root reads a mode-000 file regardless")
	}
	root := t.TempDir()
	write(t, root, "v0.0.3\n")
	if err := os.Chmod(Path(root), 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(Path(root), 0o644) })

	tag, err := Read(vfs.OS{}, root)
	if err == nil {
		t.Fatalf("Read = %q, nil; want a permission error", tag)
	}
	if tag != "" {
		t.Errorf("tag = %q, want the empty string beside an error", tag)
	}
	if !errors.Is(err, fs.ErrPermission) {
		t.Errorf("error = %v, want one that is fs.ErrPermission", err)
	}
}

func TestCompareOrdersTagsAsNumbers(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"v0.0.03", "v0.0.03", 0},
		{"v0.0.3", "v0.0.03", 0}, // the padding is not identity
		{"v0.0.10", "v0.0.9", 1}, // which string order gets wrong
		{"v0.0.9", "v0.0.10", -1},
		{"v0.1.0", "v0.0.99", 1},
		{"v1.0", "v1.0.0", 0}, // a missing component is zero
		{"v1.0.1", "v1.0", 1},
		{"not-a-tag", "v0.0.1", 1}, // unreadable is never a downgrade
	}
	for _, tc := range cases {
		if got := Compare(tc.a, tc.b); got != tc.want {
			t.Errorf("Compare(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestSameReleaseReadsTagsAsNumbers(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		same bool
	}{
		{"v0.0.03", "v0.0.03", true},
		{"v0.0.03", "v0.0.3", true},  // one release, two spellings
		{"v0.0.3", "v0.1.3", false},  // a real difference
		{"v0.0.10", "v0.0.1", false}, // not string order
		{"v1.0", "v1.0.0", true},     // a missing component is zero
		{"v1.0", "v1.0.1", false},    // and it is still read as zero
		{"main", "v0.0.03", false},   // unreadable against a tag
		{"main", "main", true},       // unreadable against itself
		{"nightly", "main", false},   // two unreadable, two things
		{"", "v0.0.03", false},       // no tag is not the pinned one
	} {
		if got := SameRelease(tc.a, tc.b); got != tc.same {
			t.Errorf("SameRelease(%q, %q) = %v; want %v", tc.a, tc.b, got, tc.same)
		}
	}
}

// An unreadable tag is a move forward whichever side it sits on, and it
// matches no release. Compare says so with 1 rather than -1; SameRelease
// says so with false, which is the same answer read as equality.
func TestAnUnreadableTagIsForwardAndMatchesNothing(t *testing.T) {
	if Compare("main", "v0.0.3") != 1 || Compare("v0.0.3", "main") != 1 {
		t.Error("an unreadable tag ordered as a downgrade on one side")
	}
	if SameRelease("main", "v0.0.3") || SameRelease("v0.0.3", "main") {
		t.Error("an unreadable tag matched a release")
	}
}

// SameRelease is Compare == 0 — one comparison, not two. The property
// is pinned rather than described so that a later change to either one
// has to answer for the other (report-0010, spec-0018).
func TestSameReleaseIsCompareEqualToZero(t *testing.T) {
	tags := []string{
		"", "v", "v0", "v1", "0.0.3", "v0.0.3", "v0.0.03", "v0.0.003",
		"v0.0.10", "v0.0.9", "v0.1.0", "v1.0", "v1.0.0", "v1.0.1",
		"v1.0.0.0", "v0..3", "v0.-1.3", " v1.0.0 ", "main", "nightly",
		"v1.x.0", "v0.0.3-rc.1", "V1.0.0", "v01.0.0", "v1.2.3.4.5",
	}
	for _, a := range tags {
		for _, b := range tags {
			if got, want := SameRelease(a, b), Compare(a, b) == 0; got != want {
				t.Errorf("SameRelease(%q, %q) = %v, Compare == 0 = %v", a, b, got, want)
			}
		}
	}
}

func TestComponentsRejectWhatIsNotANumber(t *testing.T) {
	for _, tag := range []string{"v", "", "v1.x.0", "v1..0"} {
		if _, ok := Components(tag); ok {
			t.Errorf("Components(%q) read as a version", tag)
		}
	}
	got, ok := Components("v0.1.03")
	if !ok || len(got) != 3 || got[0] != 0 || got[1] != 1 || got[2] != 3 {
		t.Errorf("Components(%q) = %v, %v; want [0 1 3]", "v0.1.03", got, ok)
	}
}

// Components takes what Compare can order; Readable takes what a
// repository may record. The bare number is the difference: it orders,
// and it is not a tag anyone recorded.
func TestReadableIsStricterThanComponents(t *testing.T) {
	for _, tag := range []string{"v0.0.03", "v0.0.3", "v1.10", "v10.0.100"} {
		if !Readable(tag) {
			t.Errorf("Readable(%q) = false, want true", tag)
		}
	}
	for _, tag := range []string{"", "v", "v1", "0.0.3", "v0.0.x", "vmain", "v1..2"} {
		if Readable(tag) {
			t.Errorf("Readable(%q) = true, want false", tag)
		}
	}
	if _, ok := Components("1.2.3"); !ok {
		t.Error("Components refused a tag it can order")
	}
	if Readable("1.2.3") {
		t.Error("Readable accepted a tag with no leading v")
	}
}

func write(t *testing.T, root, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, ".writrun"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(Path(root), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
