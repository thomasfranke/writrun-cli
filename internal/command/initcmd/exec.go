package initcmd

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// ExecGit is the production git runner behind Deps.Git: one invocation,
// stdout returned, a failure carrying everything git said — stderr
// first, then stdout — so the refusal's reason reaches the user
// unedited.
func ExecGit(dir string, args ...string) (string, error) {
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
