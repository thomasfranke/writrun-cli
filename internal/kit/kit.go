// Package kit runs the adopted repository's own `.writrun/` scripts as
// child processes — the execution authority the binary wraps and never
// reimplements — and names them, once each.
//
// # Why the paths live here
//
// A path the binary calls is the kit's API, and it is declared in the
// package that owns the act (docs/technical/engineering/coupling.md,
// rule 1). Running a script is this package's act, so every script
// path is here; the tag is `internal/kittag`'s, the queue's folders
// are `internal/queue`'s, and the one file whose shape the binary
// knows is `internal/pointer`'s.
//
// Ten of these were declared in two to five packages each, under as
// many as four different names for one file. A rename that updated
// three of the four copies compiled, and the fourth command failed at
// run time against a repository that was correct (task-0027).
package kit

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// The scripts this binary runs, named after the act rather than the
// file. Each is relative to the repository root, slash-separated, the
// form Runner takes.
const (
	// Preflight is the three checks in the one order CI runs them.
	Preflight = ".writrun/scripts/stage-1-tasks-and-specs/preflight.sh"
	// RecordProvenance appends to the provenance ledger, where a
	// project keeps one.
	RecordProvenance = ".writrun/scripts/stage-1-tasks-and-specs/record_provenance.sh"

	// ReadSetting answers one key of `settings.json`, in the shape the
	// kit's own line-based readers see it.
	ReadSetting = ".writrun/scripts/stage-2-pull-requests/read_setting.sh"
	// CheckSettings judges whether that file holds the shape those
	// readers can see at all.
	CheckSettings = ".writrun/scripts/stage-2-pull-requests/check_settings.sh"
	// CheckObservance is the door a pull-request title passes.
	CheckObservance = ".writrun/scripts/stage-2-pull-requests/check_observance.sh"
	// CheckDocShapes judges the shape of the documents a change writes.
	CheckDocShapes = ".writrun/scripts/stage-2-pull-requests/check_doc_shapes.sh"
	// TakeTask cuts the branch, commits, pushes and opens the draft.
	TakeTask = ".writrun/scripts/stage-2-pull-requests/take_task.sh"

	// CheckFrontMatter judges every queue file against the schema.
	CheckFrontMatter = ".writrun/skills/writrun-check-front-matter/check_front_matter.sh"
	// CheckDeltas judges a diff against a spec's Proposed changes.
	CheckDeltas = ".writrun/skills/writrun-check-spec-deltas/check_deltas.sh"
	// CheckState judges the lifecycle transitions a range makes.
	CheckState = ".writrun/skills/writrun-check-task-state/check_state.sh"
	// NewQueueFile scaffolds a task, a spec or a report.
	NewQueueFile = ".writrun/skills/writrun-create-task-and-spec/new.sh"
	// ListTasks is the queue as the methodology groups it.
	ListTasks = ".writrun/skills/writrun-select-next-task/list_tasks.sh"
	// Brief is one task's whole brief, for an agent to work from.
	Brief = ".writrun/skills/writrun-select-next-task/brief.sh"
	// ResolveDoc answers what text is in force at an address in the
	// project's home: the project's own file, or the kit's default where
	// that file defers. Nothing here computes a default's address —
	// an address in Go freezes at the tag that wrote it
	// (docs/technical/engineering/coupling.md, rule 3).
	ResolveDoc = ".writrun/scripts/stage-1-tasks-and-specs/resolve_doc.sh"
)

// The kit's files this binary reads or composes from, as opposed to
// runs. Same rule, same reason.
const (
	// PullRequestTemplate is the body `take` and `author` compose.
	PullRequestTemplate = ".writrun/templates/pull_request_template.md"
)

// The adopter's files. They live in the project's home, which no
// refresh writes into, and every one of them is named here and nowhere
// else — so a tag that moves one is a constant to change rather than a
// hunt (docs/technical/engineering/coupling.md, rule 3).
const (
	// Settings is the adopter's file: the stage, the conduct flags and
	// the commit vocabulary.
	Settings = "writrun/settings.json"
	// Gates is the adopter's answers, one row per transition — or the
	// marker saying the kit's default answers them.
	Gates = "writrun/gates.md"
	// Conventions is the folder holding the project's taste, each file
	// deferring or answering whole.
	Conventions = "writrun/conventions"
	// Home is the project's home entire. A refresh never writes into it.
	Home = "writrun"
)

// The addresses the adopter's files used before WritRun v0.0.05 moved
// them out of the kit's home. `update` migrates a repository off them
// once and names them nowhere else.
const (
	LegacySettings    = ".writrun/settings.json"
	LegacyGates       = ".writrun/gates.md"
	LegacyConventions = ".writrun/conventions"
)

// Resolve answers which file is in force at an address in the project's
// home: the project's own file, or the kit's default where that file
// defers. The path it returns is relative to the repository root.
//
// The kit's own resolver decides. This function passes the address it
// was given and reads back a path — it never inspects the file for the
// marker, and never joins a default's address, because both are shapes
// a tag can change (docs/technical/engineering/coupling.md, rule 3).
func Resolve(run Runner, root, path string) (string, error) {
	var out, errb bytes.Buffer
	if err := run(root, &out, &errb, nil, ResolveDoc, path); err != nil {
		if msg := strings.TrimSpace(errb.String()); msg != "" {
			return "", fmt.Errorf("resolving %s: %s", path, msg)
		}
		return "", fmt.Errorf("resolving %s: %w", path, err)
	}
	answer := strings.TrimSpace(out.String())
	if answer == "" {
		return "", fmt.Errorf("resolving %s: the resolver named no file", path)
	}
	return answer, nil
}
