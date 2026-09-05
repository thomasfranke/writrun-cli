package gitx

import (
	"strings"
	"testing"
)

func TestRunReturnsStdout(t *testing.T) {
	dir := t.TempDir()
	if _, err := Run(dir, "init", "-q"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	out, err := Run(dir, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		t.Fatalf("rev-parse: %v", err)
	}
	if strings.TrimSpace(out) != "true" {
		t.Errorf("rev-parse = %q, want true", out)
	}
}

func TestAFailureCarriesGitsOwnWords(t *testing.T) {
	// Outside a repository, git says so on stderr; the refusal's reason
	// has to reach the user unedited.
	_, err := Run(t.TempDir(), "rev-parse", "--git-dir")
	if err == nil {
		t.Fatal("git succeeded outside a repository")
	}
	if !strings.Contains(err.Error(), "git rev-parse --git-dir") {
		t.Errorf("the error does not name the invocation: %v", err)
	}
	if !strings.Contains(strings.ToLower(err.Error()), "repository") {
		t.Errorf("git's own words did not reach the error: %v", err)
	}
}

func TestAFailureThatSpokeOnStdoutCarriesThatToo(t *testing.T) {
	// A shell alias is the one way to get git to fail *after* writing
	// to stdout, which is the branch where both streams are joined.
	dir := t.TempDir()
	out, err := Run(dir, "-c", "alias.loud=!echo the reason; exit 3", "loud")
	if err == nil {
		t.Fatalf("the alias succeeded: %q", out)
	}
	if !strings.Contains(err.Error(), "the reason") {
		t.Errorf("what git wrote on stdout did not reach the error: %v", err)
	}
}

func TestAFailureThatSaidNothingStillNamesItself(t *testing.T) {
	// Nothing on either stream: the invocation is all there is to say,
	// and saying nothing at all would be a refusal with no reason.
	_, err := Run(t.TempDir(), "-c", "alias.quiet=!exit 3", "quiet")
	if err == nil {
		t.Fatal("the alias succeeded")
	}
	if !strings.Contains(err.Error(), "alias.quiet") {
		t.Errorf("the error does not name the invocation: %v", err)
	}
}
