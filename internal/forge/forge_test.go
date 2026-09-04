package forge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func stub(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "gh")
	if err := os.WriteFile(path, []byte("#!/usr/bin/env bash\n"+body), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunReturnsStdout(t *testing.T) {
	c := Client{Bin: stub(t, `echo "pr list: $*"`)}
	out, err := c.Run("pr", "list")
	if err != nil {
		t.Fatalf("Run = %v", err)
	}
	if !strings.Contains(out, "pr list: pr list") {
		t.Fatalf("out = %q; want the arguments passed through", out)
	}
}

func TestRunCarriesGhsOwnStderrOnFailure(t *testing.T) {
	c := Client{Bin: stub(t, "echo 'no pull request found' >&2; exit 1")}
	_, err := c.Run("pr", "view")
	if err == nil || !strings.Contains(err.Error(), "no pull request found") {
		t.Fatalf("err = %v; want gh's own reason carried", err)
	}
}

func TestRunCarriesStdoutDetailOnFailure(t *testing.T) {
	// gh api puts the API's error detail on stdout with only a generic
	// line on stderr — both belong in the reason.
	c := Client{Bin: stub(t, `echo '{"message": "Not Found"}'; echo 'gh: HTTP 404' >&2; exit 1`)}
	_, err := c.Run("api", "repos/x/y")
	if err == nil || !strings.Contains(err.Error(), "HTTP 404") || !strings.Contains(err.Error(), "Not Found") {
		t.Fatalf("err = %v; want stderr and the stdout detail carried", err)
	}
}
