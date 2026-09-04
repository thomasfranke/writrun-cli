package kit

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRunnerRunsTheScriptFromTheRoot(t *testing.T) {
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
	r := Runner{Root: root, Stdout: &out, Stderr: &out}
	if err := r.Run(script, "one"); err != nil {
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

func TestRunnerPassesTheScriptsExitThrough(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "fail.sh"),
		[]byte("#!/usr/bin/env bash\nexit 3\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	r := Runner{Root: root, Stdout: os.Stdout, Stderr: os.Stderr}
	err := r.Run("fail.sh")
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 3 {
		t.Fatalf("err = %v; want the script's own exit 3", err)
	}
}
