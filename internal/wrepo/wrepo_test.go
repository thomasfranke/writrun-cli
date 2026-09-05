package wrepo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

func TestFind(t *testing.T) {
	mk := func(t *testing.T, layout ...string) string {
		t.Helper()
		root := t.TempDir()
		for _, p := range layout {
			if err := os.MkdirAll(filepath.Join(root, p), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		return root
	}

	t.Run("adopted repository from its root", func(t *testing.T) {
		root := mk(t, ".git", ".writrun", "sub/dir")
		got, adopted, err := Find(vfs.OS{}, root)
		if err != nil || !adopted || got != root {
			t.Fatalf("Find = %q %v %v; want %q true nil", got, adopted, err, root)
		}
	})

	t.Run("adopted repository from a subdirectory", func(t *testing.T) {
		root := mk(t, ".git", ".writrun", "sub/dir")
		got, adopted, err := Find(vfs.OS{}, filepath.Join(root, "sub", "dir"))
		if err != nil || !adopted || got != root {
			t.Fatalf("Find = %q %v %v; want %q true nil", got, adopted, err, root)
		}
	})

	t.Run("git repository not adopted", func(t *testing.T) {
		root := mk(t, ".git")
		got, adopted, err := Find(vfs.OS{}, root)
		if err != nil || adopted || got != root {
			t.Fatalf("Find = %q %v %v; want %q false nil", got, adopted, err, root)
		}
	})

	t.Run("a .writrun file is not adoption", func(t *testing.T) {
		root := mk(t, ".git")
		if err := os.WriteFile(filepath.Join(root, ".writrun"), nil, 0o644); err != nil {
			t.Fatal(err)
		}
		_, adopted, err := Find(vfs.OS{}, root)
		if err != nil || adopted {
			t.Fatalf("adopted = %v err = %v; a plain file is not the kit", adopted, err)
		}
	})

	t.Run("worktree's .git file is the toplevel", func(t *testing.T) {
		root := mk(t, ".writrun")
		if err := os.WriteFile(filepath.Join(root, ".git"), []byte("gitdir: elsewhere"), 0o644); err != nil {
			t.Fatal(err)
		}
		got, adopted, err := Find(vfs.OS{}, root)
		if err != nil || !adopted || got != root {
			t.Fatalf("Find = %q %v %v; want %q true nil", got, adopted, err, root)
		}
	})

	t.Run("not a git repository at all", func(t *testing.T) {
		dir := mk(t, "just/dirs")
		_, _, err := Find(vfs.OS{}, filepath.Join(dir, "just", "dirs"))
		if err == nil {
			t.Fatal("want an error outside any git repository")
		}
	})
}
