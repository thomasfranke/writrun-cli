// Package gitx is the production git runner: one invocation, its
// stdout returned, and a failure carrying everything git said. Every
// command that shells out to git goes through this, so a refusal's
// reason reaches the user unedited whichever command hit it.
package gitx

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Run executes one git invocation in dir and returns its stdout. A
// failure carries git's own words — stderr first, then stdout.
func Run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errb.String())
		if o := strings.TrimSpace(out.String()); o != "" {
			if msg == "" {
				msg = o
			} else {
				msg += "\n" + o
			}
		}
		if msg == "" {
			return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return out.String(), nil
}
