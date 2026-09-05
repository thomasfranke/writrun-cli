package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/gitx"

	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// readOnly makes dir unwritable for the rest of the test, so a write
// into it fails the way a hooks directory somebody else owns would.
func readOnly(t *testing.T, dir string) {
	t.Helper()
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
}

func TestInstallReportsAnUnwritableDirectory(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root writes into a read-only directory all the same")
	}
	dir := t.TempDir()
	readOnly(t, dir)

	// The hooks directory cannot be created.
	if err := Install(vfs.OS{}, filepath.Join(dir, "hooks", "commit-msg")); err == nil {
		t.Error("installing under an unwritable parent was not an error")
	} else if !strings.Contains(err.Error(), "installing the commit-msg hook") {
		t.Errorf("the error does not name the act: %v", err)
	}

	// The directory is there, but the file cannot be written into it.
	if err := Install(vfs.OS{}, filepath.Join(dir, "commit-msg")); err == nil {
		t.Error("writing into an unwritable directory was not an error")
	}
}

func TestInspectReportsAnUnreadableHook(t *testing.T) {
	// A directory where the hook should be is neither ours, nor
	// foreign, nor absent — it is a fault worth naming.
	dir := filepath.Join(t.TempDir(), "commit-msg")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Inspect(vfs.OS{}, dir); err == nil {
		t.Error("a directory where the hook should be was read as a hook")
	}
}

func TestPathReportsAGitThatCannotAnswer(t *testing.T) {
	_, err := Path(t.TempDir(), GitRunner(gitx.Run))
	if err == nil {
		t.Fatal("the hooks directory was resolved outside a repository")
	}
	if !strings.Contains(err.Error(), "resolving the hooks directory") {
		t.Errorf("the error does not name the act: %v", err)
	}
}
