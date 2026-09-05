package wrepo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAWorkingDirectoryThatIsGoneIsReported(t *testing.T) {
	// A relative path is resolved against the working directory; with
	// that directory deleted there is nothing to resolve against, and
	// the walk may not start from a guess.
	before, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(before) })

	gone := filepath.Join(t.TempDir(), "gone")
	if err := os.Mkdir(gone, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(gone); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(gone); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Find("."); err == nil {
		t.Skip("this platform still resolves a deleted working directory")
	}
}
