// Package forge shells out to `gh` — the forge reached only where the
// wrapped flows already reach it.
package forge

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Client invokes gh. The zero value uses the gh on PATH.
type Client struct {
	// Bin overrides the gh executable; tests point it at a stub.
	Bin string
	// Dir is the working directory of every invocation.
	Dir string
}

// Run executes one gh invocation and returns its stdout. A failure
// carries gh's own stderr, so the forge's reason reaches the user
// unedited.
func (c Client) Run(args ...string) (string, error) {
	bin := c.Bin
	if bin == "" {
		bin = "gh"
	}
	cmd := exec.Command(bin, args...)
	cmd.Dir = c.Dir
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errb.String())
		if msg == "" {
			return "", fmt.Errorf("gh %s: %w", strings.Join(args, " "), err)
		}
		return "", fmt.Errorf("gh %s: %s", strings.Join(args, " "), msg)
	}
	return out.String(), nil
}
