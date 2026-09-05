package kitfetch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/gitx"
)

const tag = "v9.9.9"

func gitT(t *testing.T, dir string, args ...string) {
	t.Helper()
	full := append([]string{"-c", "user.name=suite", "-c", "user.email=suite@test", "-c", "commit.gpgsign=false"}, args...)
	if _, err := gitx.Run(dir, full...); err != nil {
		t.Fatalf("git %s: %v", strings.Join(args, " "), err)
	}
}

// makeRepo builds a git repository tagged tag; withTemplate says
// whether it is a WritRun one.
func makeRepo(t *testing.T, withTemplate bool) string {
	t.Helper()
	src := t.TempDir()
	gitT(t, src, "init", "-q")
	rel := "README.md"
	if withTemplate {
		rel = "template/AGENTS.md"
	}
	path := filepath.Join(src, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("# the kit\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitT(t, src, "add", ".")
	gitT(t, src, "commit", "-q", "-m", "the kit")
	gitT(t, src, "tag", tag)
	return src
}

func TestFetchReturnsTheTemplateAndCleansUp(t *testing.T) {
	got, err := Fetch(tag, makeRepo(t, true), gitx.Run)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(got.Template, "AGENTS.md")); statErr != nil {
		t.Errorf("the template is not where Fetch says: %v", statErr)
	}
	got.Cleanup()
	if _, statErr := os.Stat(got.Template); statErr == nil {
		t.Error("the checkout survived its own cleanup")
	}
}

func TestACloneWithNoTemplateIsNotAWritRunRepository(t *testing.T) {
	_, err := Fetch(tag, makeRepo(t, false), gitx.Run)
	if err == nil {
		t.Fatal("a repository with no template/ was accepted")
	}
	if !strings.Contains(err.Error(), "not a WritRun repository") {
		t.Errorf("the refusal does not say what is wrong: %v", err)
	}
}

func TestAnUnreachableSourceWritesNothing(t *testing.T) {
	_, err := Fetch(tag, filepath.Join(t.TempDir(), "not-a-repository"), gitx.Run)
	if err == nil {
		t.Fatal("an unreachable source was accepted")
	}
	if !strings.Contains(err.Error(), "nothing was written") {
		t.Errorf("the refusal does not say the tree is untouched: %v", err)
	}
}

func TestATagThatDoesNotExistIsAFailure(t *testing.T) {
	if _, err := Fetch("v0.0.0", makeRepo(t, true), gitx.Run); err == nil {
		t.Fatal("a tag that does not exist was accepted")
	}
}

func TestNowhereToWorkIsReportedBeforeTheClone(t *testing.T) {
	// The checkout goes outside the repository, so a temp directory
	// that cannot be made is the first thing that can fail.
	t.Setenv("TMPDIR", filepath.Join(t.TempDir(), "no-such-place"))
	if _, err := Fetch(tag, makeRepo(t, true), gitx.Run); err == nil {
		t.Fatal("a fetch with nowhere to work succeeded")
	}
}
