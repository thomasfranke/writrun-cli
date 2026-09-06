// Package kit runs the adopted repository's own `.writrun/` scripts as
// child processes — the execution authority the binary wraps and never
// reimplements.
package kit

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// Runner is one script invocation: the type every consumer names, so
// the wiring hands Run over without converting between identical
// declarations of it. The streams belong to the caller, because a
// command that shows a script's reporting and one that reads it back
// need the same runner.
//
// # Why the environment is a parameter and not a second port
//
// env carries `KEY=value` entries for the child. It sits in this
// signature rather than in a sibling `EnvRunner` because some of the
// kit's scripts read their whole input there and nowhere else:
// check_observance.sh takes the pull-request title and body through
// `$PR_TITLE` and `$PR_BODY`, "through the environment, never inline
// interpolation", and names argv as the way it must not arrive. A
// consumer holding a runner without this parameter has argv as its only
// way to hand a script a string, so the narrower type would be the
// shape that invites the one call the script forbids.
//
// A nil env is the ordinary case and says the script reads nothing from
// its caller's environment.
type Runner func(root string, stdout, stderr io.Writer, env []string, script string, args ...string) error

// Run executes one script with bash, from the repository root, and
// returns the script's own verdict — an *exec.ExitError carries its
// exit code, which is the whole answer a wrapping command maps.
//
// env is layered on this process's environment rather than replacing
// it: the kit's scripts read PATH, HOME and TMPDIR, and the suite
// reaches them through WRITRUN_PR_LIST the same way. An entry given
// here wins over an inherited one of the same name, because os/exec
// keeps the last of each key.
func Run(root string, stdout, stderr io.Writer, env []string, script string, args ...string) error {
	cmd := exec.Command("bash", append([]string{filepath.Join(root, script)}, args...)...)
	cmd.Dir = root
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	return cmd.Run()
}
