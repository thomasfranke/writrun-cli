package chapter

import (
	"io/fs"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

func seeded(files ...string) *vfs.Fake {
	f := vfs.NewFake()
	f.SeedDir("repo/docs/product")
	for _, name := range files {
		f.Seed("repo/docs/product/"+name, []byte("# x\n"), 0o644)
	}
	return f
}

func TestARealChapterIsFound(t *testing.T) {
	for _, name := range []string{"adoption.md", "doctor.md", "a/nested.md"} {
		if !In(seeded(name), "repo/docs/product") {
			t.Errorf("%q was not read as a chapter", name)
		}
	}
}

// The README is the table of chapters to come, not a chapter.
func TestTheReadmeAloneIsNoChapter(t *testing.T) {
	for _, name := range []string{"README.md", "readme.md", "ReadMe.md"} {
		if In(seeded(name), "repo/docs/product") {
			t.Errorf("%q was read as a chapter", name)
		}
	}
}

func TestAnEmptyFolderHoldsNoChapter(t *testing.T) {
	if In(seeded(), "repo/docs/product") {
		t.Error("an empty folder held a chapter")
	}
}

func TestSomethingThatIsNotMarkdownIsNoChapter(t *testing.T) {
	if In(seeded("notes.txt", "diagram.png"), "repo/docs/product") {
		t.Error("a non-markdown file was read as a chapter")
	}
}

// A folder that is not there, or cannot be walked, holds no chapter —
// the caller is asking whether the project wrote a doc, not why it did
// not, and it words the gap itself.
func TestAFolderThatCannotBeWalkedHoldsNoChapter(t *testing.T) {
	if In(vfs.NewFake(), "repo/docs/product") {
		t.Error("a missing folder held a chapter")
	}
	f := seeded("adoption.md")
	f.FailOp("walk", "repo/docs/product", fs.ErrPermission)
	if In(f, "repo/docs/product") {
		t.Error("an unwalkable folder held a chapter")
	}
}
