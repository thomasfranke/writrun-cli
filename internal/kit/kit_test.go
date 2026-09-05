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
	if err := Run(root, &out, &out, script, "one"); err != nil {
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
	if err := Run(root, &out, &errb, "both.sh"); err != nil {
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
	err := Run(root, os.Stdout, os.Stderr, "fail.sh")
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
