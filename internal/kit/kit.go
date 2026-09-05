// Package kit runs the adopted repository's own `.writrun/` scripts as
// child processes — the execution authority the binary wraps and never
// reimplements.
package kit

import (
	"io"
	"os/exec"
	"path/filepath"
)

// Runner is one script invocation: the type every consumer names, so
// the wiring hands Run over without converting between identical
// declarations of it. The streams belong to the caller, because a
// command that shows a script's reporting and one that reads it back
// need the same runner.
type Runner func(root string, stdout, stderr io.Writer, script string, args ...string) error

// Run executes one script with bash, from the repository root, and
// returns the script's own verdict — an *exec.ExitError carries its
// exit code, which is the whole answer a wrapping command maps.
func Run(root string, stdout, stderr io.Writer, script string, args ...string) error {
	cmd := exec.Command("bash", append([]string{filepath.Join(root, script)}, args...)...)
	cmd.Dir = root
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}
