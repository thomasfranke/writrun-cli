package kittag

import (
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

// The two comparisons are separate because their answers on an
// unreadable tag are: Compare calls it a move forward whichever side it
// is on, and SameRelease calls it a difference. A caller asking only
// whether two tags match inherits no ordering.
func TestSameReleaseCarriesNoOrdering(t *testing.T) {
	if Compare("main", "v0.0.3") != 1 || Compare("v0.0.3", "main") != 1 {
		t.Error("an unreadable tag ordered as a downgrade on one side")
	}
	if SameRelease("main", "v0.0.3") || SameRelease("v0.0.3", "main") {
		t.Error("an unreadable tag matched a release")
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
