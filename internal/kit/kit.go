// Package kit runs the adopted repository's own `.writrun/` scripts as
// child processes — the execution authority the binary wraps and never
// reimplements.
package kit

import (
	"io"
	"os/exec"
	"path/filepath"
)

// Runner executes scripts relative to the adopted repository's root.
type Runner struct {
	// Root is the git toplevel of the adopted repository.
	Root string
	// Stdout and Stderr receive the script's own reporting, unedited.
	Stdout io.Writer
	Stderr io.Writer
}

// Run executes one script with bash, from the repository root, and
// returns the script's own verdict — an *exec.ExitError carries its
// exit code.
func (r Runner) Run(script string, args ...string) error {
	cmd := exec.Command("bash", append([]string{filepath.Join(r.Root, script)}, args...)...)
	cmd.Dir = r.Root
	cmd.Stdout = r.Stdout
	cmd.Stderr = r.Stderr
	return cmd.Run()
}

// Run is Runner.Run for a caller holding the root and the streams and
// nothing else — the shape a command package's exec port declares, so
// the wiring hands this function over without a closure.
func Run(root string, stdout, stderr io.Writer, script string, args ...string) error {
	return Runner{Root: root, Stdout: stdout, Stderr: stderr}.Run(script, args...)
}
