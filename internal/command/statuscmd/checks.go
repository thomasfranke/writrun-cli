package statuscmd

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/kit"
)

// preflightScript is the completion gates in the order the methodology
// fixed — the authority this command wraps and never reimplements. It
// is the same script `finish` runs, so the stage it stops at here is
// the stage `finish` would stop at (docs/product/queue/status.md).
const preflightScript = ".writrun/scripts/stage-1-tasks-and-specs/preflight.sh"

// stopped opens the sentence preflight refuses with. It names the
// stage and the code, so the answer to step 3 is a line the script
// already wrote rather than a summary of one.
const stopped = "PREFLIGHT STOPPED"

// checkRun is one preflight invocation: what it printed and what it
// decided.
type checkRun struct {
	out  string
	err  error
	code int
}

// preflight runs the completion checks with the arguments the script
// prefers — none, so it infers the task from the branch and the range
// from the checkout exactly as a completion run would. Its reporting is
// captured rather than streamed: status answers in lines, and the sweep
// prints pages.
func preflight(scripts kit.Runner, root string) checkRun {
	var out bytes.Buffer
	err := scripts(root, &out, &out, nil, preflightScript)
	return checkRun{out: out.String(), err: err, code: exitCode(err)}
}

// verdict is step 3 in one line, in preflight's own words: the stage it
// stopped at, or that every stage passed. Nothing here re-reads the
// queue — a second opinion on a check is a second authority.
func verdict(r checkRun) string {
	if r.err == nil {
		return "all pass"
	}
	if line := find(r.out, stopped); line != "" {
		return line
	}
	// Preflight's own failures — a malformed argument, an id resolving
	// to no file — exit 4 and say so without naming a stage. Anything
	// else is a script that could not run at all.
	if last := lastLine(r.out); last != "" {
		return fmt.Sprintf("%s (%s exited %d)", last, preflightScript, r.code)
	}
	return fmt.Sprintf("%s could not be run: %v", preflightScript, r.err)
}

// find returns the first line opening with prefix.
func find(out, prefix string) string {
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(line)
		}
	}
	return ""
}

// lastLine is the last thing the script said — what a refusal carrying
// no stage still leaves behind.
func lastLine(out string) string {
	lines := strings.Split(out, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if line := strings.TrimSpace(lines[i]); line != "" {
			return line
		}
	}
	return ""
}

// exitCode reads the script's own verdict off the error the runner
// returned; -1 says the runner failed before the script spoke.
func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var code interface{ ExitCode() int }
	if errors.As(err, &code) && code.ExitCode() > 0 {
		return code.ExitCode()
	}
	return -1
}
