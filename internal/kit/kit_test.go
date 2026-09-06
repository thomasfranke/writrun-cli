package kit

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRunRunsTheScriptFromTheRoot(t *testing.T) {
	root := t.TempDir()
	script := filepath.Join(".writrun", "scripts", "hello.sh")
	if err := os.MkdirAll(filepath.Dir(filepath.Join(root, script)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, script),
		[]byte("#!/usr/bin/env bash\necho \"pwd=$(pwd) arg=$1\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := Run(root, &out, &out, nil, script, "one"); err != nil {
		t.Fatalf("Run = %v", err)
	}
	got := out.String()
	if !bytes.Contains([]byte(got), []byte("arg=one")) {
		t.Fatalf("output = %q; want the argument passed through", got)
	}
	resolved, _ := filepath.EvalSymlinks(root)
	if !bytes.Contains([]byte(got), []byte("pwd="+resolved)) && !bytes.Contains([]byte(got), []byte("pwd="+root)) {
		t.Fatalf("output = %q; want the script run from the root", got)
	}
}

func TestRunSeparatesTheStreams(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "both.sh"),
		[]byte("#!/usr/bin/env bash\necho said; echo refused >&2\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if err := Run(root, &out, &errb, nil, "both.sh"); err != nil {
		t.Fatalf("Run = %v", err)
	}
	if out.String() != "said\n" {
		t.Errorf("stdout = %q, want the script's stdout alone", out.String())
	}
	if errb.String() != "refused\n" {
		t.Errorf("stderr = %q, want the script's stderr alone", errb.String())
	}
}

func TestRunPassesTheScriptsExitThrough(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "fail.sh"),
		[]byte("#!/usr/bin/env bash\nexit 3\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	err := Run(root, os.Stdout, os.Stderr, nil, "fail.sh")
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 3 {
		t.Fatalf("err = %v; want the script's own exit 3", err)
	}
}

// Runner is the type the commands name; Run has to satisfy it, or the
// wiring would need a conversion nobody would notice was missing.
func TestRunSatisfiesRunner(t *testing.T) {
	var r Runner = Run
	if r == nil {
		t.Fatal("Run is not a Runner")
	}
}

// envScript writes one variable's value, so a case reads what the child
// actually received rather than what the parent meant to send.
func envScript(t *testing.T, root, name, key string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, name),
		[]byte("#!/usr/bin/env bash\nprintf '%s' \"${"+key+":-<unset>}\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestRunHandsTheEnvironmentToTheScript(t *testing.T) {
	root := t.TempDir()
	envScript(t, root, "read.sh", "PR_TITLE")

	var out bytes.Buffer
	title := "[Docs][Product] The merge is the assenting act"
	if err := Run(root, &out, &out, []string{"PR_TITLE=" + title}, "read.sh"); err != nil {
		t.Fatalf("Run = %v", err)
	}
	if out.String() != title {
		t.Fatalf("PR_TITLE = %q; want the title handed through the environment", out.String())
	}
}

func TestRunLayersTheEnvironmentRatherThanReplacingIt(t *testing.T) {
	root := t.TempDir()
	envScript(t, root, "read.sh", "WRITRUN_PR_LIST")
	t.Setenv("WRITRUN_PR_LIST", "42\ttask/0012-x")

	// The kit's scripts read PATH, TMPDIR and the suite's own seams; a
	// runner that replaced the environment would hand them none of it.
	var out bytes.Buffer
	if err := Run(root, &out, &out, []string{"PR_TITLE=x"}, "read.sh"); err != nil {
		t.Fatalf("Run = %v", err)
	}
	if out.String() != "42\ttask/0012-x" {
		t.Fatalf("WRITRUN_PR_LIST = %q; want the inherited value to survive", out.String())
	}
}

func TestRunGivesTheCallersEntryPrecedence(t *testing.T) {
	root := t.TempDir()
	envScript(t, root, "read.sh", "PR_TITLE")
	t.Setenv("PR_TITLE", "whatever the surrounding job exported")

	var out bytes.Buffer
	if err := Run(root, &out, &out, []string{"PR_TITLE=the composed one"}, "read.sh"); err != nil {
		t.Fatalf("Run = %v", err)
	}
	if out.String() != "the composed one" {
		t.Fatalf("PR_TITLE = %q; want the caller's entry to win over the inherited one", out.String())
	}
}
