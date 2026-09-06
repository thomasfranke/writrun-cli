// Package agentx starts the adopter's configured agent as a child
// process that inherits this one's terminal. It is the production side
// of the launcher port `work` consumes: the one boundary where writrun
// hands control to a program it knows nothing about
// (docs/technical/architecture.md).
package agentx

import (
	"os"
	"os/exec"
)

// Run starts name with args from dir and waits for it. The three
// standard streams are the process's own rather than the frame's
// writers: an agent renders a terminal interface, reads keys, and asks
// its own questions, none of which survives a pipe into a buffer. A
// failure carries the command's own exit through an *exec.ExitError,
// which is the whole verdict `work` passes up (spec-0007, edge cases).
func Run(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
