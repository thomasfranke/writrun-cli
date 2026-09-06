package agentx

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// script writes an executable into dir and returns its path.
func script(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/usr/bin/env bash\n"+body), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunStartsTheCommandFromTheDirectoryWithItsArguments(t *testing.T) {
	dir := t.TempDir()
	// The streams are the process's own, so what the child did is read
	// from what it wrote, not from a buffer this test handed it.
	agent := script(t, dir, "agent.sh", `printf '%s\n%s\n' "$(pwd)" "$*" > "$1/record"`)

	if err := Run(dir, agent, dir, "--flag"); err != nil {
		t.Fatalf("Run = %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "record"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	resolved, _ := filepath.EvalSymlinks(dir)
	if lines[0] != resolved && lines[0] != dir {
		t.Errorf("pwd = %q; want the repository root", lines[0])
	}
	if lines[1] != dir+" --flag" {
		t.Errorf("args = %q; want them passed through", lines[1])
	}
}

func TestRunPassesTheCommandsOwnExitThrough(t *testing.T) {
	dir := t.TempDir()
	agent := script(t, dir, "fail.sh", "exit 3\n")

	err := Run(dir, agent)
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 3 {
		t.Fatalf("err = %v; want the agent's own exit 3", err)
	}
}

func TestACommandThatDoesNotExistFailsWithoutRunningAnything(t *testing.T) {
	dir := t.TempDir()
	if err := Run(dir, filepath.Join(dir, "no-such-agent")); err == nil {
		t.Fatal("a command that is not there exited 0")
	}
}
