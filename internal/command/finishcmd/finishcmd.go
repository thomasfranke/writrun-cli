// Package finishcmd is `writrun finish`: the completion sequence in the
// one order the methodology fixed — the promised deltas verified, the
// outcome recorded, the provenance appended, the completion gates run,
// and the pull request marked ready on the human's word. Every check is
// the adopted repository's own script; what this command adds is the
// order, the two writes those scripts do not make, the question before
// the forge, and the undo that keeps a refusal from leaving those two
// writes behind (docs/product/pull-requests/finish.md, spec-0010,
// spec-0017).
package finishcmd

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/kit"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// The three authorities this command runs, in the order it runs them.
// Not one of their judgements is repeated here.
const (
	deltasScript     = ".writrun/skills/writrun-check-spec-deltas/check_deltas.sh"
	provenanceScript = ".writrun/scripts/stage-1-tasks-and-specs/record_provenance.sh"
	preflightScript  = ".writrun/scripts/stage-1-tasks-and-specs/preflight.sh"
)

// Deps is the wiring finish needs beyond the frame's Ctx.
type Deps struct {
	// Scripts runs the adopted repository's own scripts.
	Scripts kit.Runner
	// Files is the filesystem the two completion writes go through.
	Files vfs.FS
	// Git answers which branch this is and which base the checks read
	// the change against.
	Git gitx.Runner
	// Gh is the forge: the pull request read, and marked ready.
	Gh func(args ...string) (string, error)
	// Now stamps the task's `completed` date.
	Now func() time.Time
}

// ledgerFlags is record_provenance.sh's own vocabulary, offered as
// flags and passed through unread. The script validates them — a second
// opinion on `by=`, on a model id or on a count is a second authority,
// and a project that keeps no ledger is asked for none of them.
var ledgerFlags = []struct{ flag, key string }{
	{"by", "by"},
	{"login", "login"},
	{"model", "model"},
	{"input", "input"},
	{"output", "output"},
	{"cache-read", "cache_read"},
	{"cache-write", "cache_write"},
}

// New returns the finish command wired with its dependencies.
func New(d Deps) command.Command {
	return command.Command{
		Name:    "finish",
		Summary: "finish a task: the deltas checked, the outcome recorded, the pull request marked ready",
		Need:    command.NeedAdopted,
		Run: func(ctx *command.Ctx, args []string) error {
			return run(ctx, d, args)
		},
	}
}

func run(ctx *command.Ctx, d Deps, args []string) error {
	id, flags, err := split(args)
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("finish", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	rangeFlag := fs.String("range", "", "the diff range the checks read this change against")
	ledger := make([]*string, len(ledgerFlags))
	for i, f := range ledgerFlags {
		ledger[i] = fs.String(f.flag, "", "record_provenance.sh's "+f.key+"=")
	}
	if err := fs.Parse(flags); err != nil {
		return err
	}

	if id == "" {
		if id, err = taskOfBranch(d.Git, ctx.Root); err != nil {
			return err
		}
	}
	taskPath, err := queueFile(d.Files, ctx.Root, tasksDir, "task", id)
	if err != nil {
		return err
	}
	task, err := d.Files.ReadFile(path.Join(ctx.Root, taskPath))
	if err != nil {
		return fmt.Errorf("reading %s: %w", taskPath, err)
	}
	taskID := field(task, "id")
	if taskID == "" {
		return fmt.Errorf("%s carries no id", taskPath)
	}
	diffRange := *rangeFlag
	if diffRange == "" {
		if diffRange, err = baseRange(d.Git, ctx.Root); err != nil {
			return err
		}
	}

	// 1 — the promised deltas. A non-zero verdict stops the command
	// here: nothing is written, nothing else runs (spec-0010, step 1).
	specIDs := specRefs(task)
	if len(specIDs) == 0 {
		fmt.Fprintf(ctx.Stdout, "%s carries no spec — no deltas to check.\n", taskID)
	} else if err := d.Scripts(ctx.Root, ctx.Stdout, ctx.Stderr, deltasScript,
		strings.Join(specIDs, ","), diffRange); err != nil {
		// One call carrying every spec: MISSING is judged per spec and
		// UNDECLARED against the union, which is the script's own rule
		// for a multi-spec completion (spec-0010, edge cases).
		return passthrough(deltasScript, err)
	}

	// 2 — the outcome, then the two writes. Every Outcome is read
	// before any file is written, so a two-spec branch missing one is
	// refused whole rather than half-recorded.
	specs, err := readSpecs(d.Files, ctx.Root, specIDs)
	if err != nil {
		return err
	}
	for _, s := range specs {
		if !outcomeFilled(s.content) {
			return fmt.Errorf("%s's ## Outcome is empty — record what was built before finishing (%s)", s.id, s.path)
		}
	}
	// Every write from here on is remembered, so any end that is not a
	// success can put the tree back exactly as it was found
	// (spec-0017; product/pull-requests/shape.md).
	undo := &journal{}
	for _, s := range specs {
		if err := write(ctx, d, undo, s.path, s.content, "status", "implemented"); err != nil {
			return undo.restore(ctx, d, err)
		}
	}
	// The task's `completed` date, and nothing else on the task. Its
	// `status` line has one writer and it is not this command — not
	// here, and not anywhere in this package (spec-0010, scope). A date
	// already there is the worker's declaration of finishing, so a
	// second run reports it rather than restamping it.
	if done := strings.TrimSpace(field(task, "completed")); done != "" && done != "null" {
		fmt.Fprintf(ctx.Stdout, "unchanged: %s already carries completed: %s\n", taskPath, done)
	} else if err := write(ctx, d, undo, taskPath, task, "completed", d.Now().UTC().Format(time.RFC3339)); err != nil {
		return undo.restore(ctx, d, err)
	}

	// 3 — the ledger, unconditionally. The script reads the setting
	// itself and a project that declares none is asked for nothing
	// (spec-0010, step 3).
	prov := []string{taskID}
	for i, f := range ledgerFlags {
		if *ledger[i] != "" {
			prov = append(prov, f.key+"="+*ledger[i])
		}
	}
	if err := d.Scripts(ctx.Root, ctx.Stdout, ctx.Stderr, provenanceScript, prov...); err != nil {
		return undo.restore(ctx, d, passthrough(provenanceScript, err))
	}

	// 4 — the completion gates. Exit 0 is required before the forge is
	// asked for anything (spec-0010, step 4). They run after the writes
	// because preflight's first stage sweeps the queue as it stands on
	// disk and its completion warning reads the `completed` date off
	// the same tree — a run before the writes is the run that stage
	// tells you not to trust.
	if err := d.Scripts(ctx.Root, ctx.Stdout, ctx.Stderr, preflightScript, taskID, diffRange); err != nil {
		return undo.restore(ctx, d, passthrough(preflightScript, err))
	}

	// 5 — what will happen, then the question, then the forge. A no
	// here, or a forge that will not answer, ends with the two writes
	// undone: the order stands and the refusal still leaves nothing
	// behind (spec-0017, step 1).
	if err := markReady(ctx, d, taskID, specs); err != nil {
		return undo.restore(ctx, d, err)
	}
	return nil
}

// journal remembers what the completion writes replaced.
//
// The two writes sit at step 2, before `preflight.sh`, because the
// gates read the queue off the working tree — moving them after the
// question would ask the human to answer before the gates had spoken,
// and would leave the same two edits behind whenever a gate then said
// no. `shape.md` fixes that same order in its own sentence: checks,
// then the status, then the composition, then the confirmation. So the
// order stands, and the promise that a refused command leaves nothing
// behind is kept by putting the writes back (spec-0017).
type journal struct {
	entries []undone
}

// undone is one file and the bytes it carried before this command
// touched it.
type undone struct {
	rel    string
	before []byte
}

// record remembers a write that actually happened. A field already
// carrying its value was not written, so there is nothing to put back.
func (j *journal) record(rel string, before []byte) {
	j.entries = append(j.entries, undone{rel: rel, before: append([]byte(nil), before...)})
}

// restore puts every remembered write back and hands up the failure
// that caused it, unedited — so a declined finish is still a decline
// and a script's exit code is still that script's.
//
// A restore that fails is the one case that rewrites the verdict: the
// tree is left changed, and saying so out loud matters more than
// carrying a code up, because the frame passes an exit code through
// without printing a word (spec-0017, edge cases).
func (j *journal) restore(ctx *command.Ctx, d Deps, cause error) error {
	var failed []string
	for i := len(j.entries) - 1; i >= 0; i-- {
		e := j.entries[i]
		if err := d.Files.WriteFile(path.Join(ctx.Root, e.rel), e.before, 0o644); err != nil {
			failed = append(failed, fmt.Sprintf("%s (%v)", e.rel, err))
			continue
		}
		fmt.Fprintf(ctx.Stdout, "restored %s — the completion edit is undone\n", e.rel)
	}
	if len(failed) > 0 {
		return fmt.Errorf("%v — and the completion edits could not be undone: %s. The working tree is left changed; put those files back by hand",
			cause, strings.Join(failed, "; "))
	}
	return cause
}

// spec is one of the branch's specs: where it lives and what it says.
type spec struct {
	id      string
	path    string
	content []byte
}

func readSpecs(files vfs.FS, root string, ids []string) ([]spec, error) {
	var specs []spec
	for _, id := range ids {
		p, err := queueFile(files, root, specsDir, "spec", id)
		if err != nil {
			return nil, err
		}
		content, err := files.ReadFile(path.Join(root, p))
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", p, err)
		}
		specs = append(specs, spec{id: id, path: p, content: content})
	}
	return specs, nil
}

// write sets one front-matter field and says what it did. A field
// already carrying the value is left alone and reported as such: the
// command is rerunnable, and a second run must not restamp a date the
// first one declared.
func write(ctx *command.Ctx, d Deps, undo *journal, rel string, content []byte, name, value string) error {
	if strings.TrimSpace(field(content, name)) == value {
		fmt.Fprintf(ctx.Stdout, "unchanged: %s already carries %s: %s\n", rel, name, value)
		return nil
	}
	next, changed, err := setField(content, name, value)
	if err != nil {
		return fmt.Errorf("%s: %w", rel, err)
	}
	if !changed {
		fmt.Fprintf(ctx.Stdout, "unchanged: %s already carries %s: %s\n", rel, name, value)
		return nil
	}
	if err := d.Files.WriteFile(path.Join(ctx.Root, rel), next, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", rel, err)
	}
	undo.record(rel, content)
	fmt.Fprintf(ctx.Stdout, "wrote %s: %s on %s\n", name, value, rel)
	return nil
}

// pullRequest is the part of `gh pr view` this command reads.
type pullRequest struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	State   string `json:"state"`
	IsDraft bool   `json:"isDraft"`
}

// markReady shows the composition and, on the human's word, marks the
// pull request ready for review — the one act of this command that
// reaches the forge, and the only one behind a question.
func markReady(ctx *command.Ctx, d Deps, taskID string, specs []spec) error {
	out, err := d.Gh("pr", "view", "--json", "number,title,state,isDraft")
	if err != nil {
		return fmt.Errorf("reading this branch's pull request: %w", err)
	}
	var pr pullRequest
	if err := json.Unmarshal([]byte(out), &pr); err != nil {
		return fmt.Errorf("reading this branch's pull request: %w", err)
	}
	if pr.State != "" && pr.State != "OPEN" {
		return fmt.Errorf("pull request #%d is %s — finish marks an open draft ready for review", pr.Number, strings.ToLower(pr.State))
	}
	if !pr.IsDraft {
		fmt.Fprintf(ctx.Stdout, "Pull request #%d is already ready for review — nothing to mark.\n", pr.Number)
		return nil
	}

	ids := make([]string, len(specs))
	for i, s := range specs {
		ids[i] = s.id
	}
	shown := strings.Join(ids, ", ")
	if shown == "" {
		shown = "none"
	}
	fmt.Fprintf(ctx.Stdout, "\nready for review:\n  task           %s\n  specs          %s\n  pull request   #%d %s\n",
		taskID, shown, pr.Number, pr.Title)

	if err := ctx.AskConfirm(fmt.Sprintf("Mark pull request #%d ready for review?", pr.Number)); err != nil {
		return err
	}
	if _, err := d.Gh("pr", "ready", strconv.Itoa(pr.Number)); err != nil {
		return fmt.Errorf("marking pull request #%d ready: %w", pr.Number, err)
	}
	fmt.Fprintf(ctx.Stdout, "Pull request #%d is ready for review.\n", pr.Number)
	return nil
}

// branchTask reads the task number out of a `task/NNNN-...` branch.
var branchTask = regexp.MustCompile(`^task/0*([0-9]+)`)

// taskOfBranch infers the task from the branch name — preflight.sh's
// own inference, so the two agree about which task a branch is for. A
// branch naming none is an error, not a guess.
func taskOfBranch(git gitx.Runner, root string) (string, error) {
	out, err := git(root, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("reading the current branch: %w", err)
	}
	branch := strings.TrimSpace(out)
	m := branchTask.FindStringSubmatch(branch)
	if m == nil {
		return "", fmt.Errorf("branch %q names no task — give the id: writrun finish task-NNNN", branch)
	}
	return "task-" + m[1], nil
}

// baseRange resolves the range the checks read the change against, the
// way preflight.sh resolves its own: the pushed main, else the local
// one, else nothing to read against and the caller is asked to say.
func baseRange(git gitx.Runner, root string) (string, error) {
	bases := []struct{ ref, rng string }{
		{"refs/remotes/origin/main", "origin/main...HEAD"},
		{"refs/heads/main", "main...HEAD"},
	}
	for _, b := range bases {
		if _, err := git(root, "rev-parse", "--verify", "--quiet", b.ref); err == nil {
			return b.rng, nil
		}
	}
	return "", errors.New("no origin/main and no main to read this change against — name a range with --range")
}

// split separates the task id from the flags. Go's flag package stops
// at the first operand, so `finish task-0011 --by human` would leave
// the rest unparsed; the id is lifted out first and the flags parsed
// whole — takecmd's rule, kept here rather than shared, because the two
// commands' flag sets are their own.
func split(args []string) (string, []string, error) {
	takesValue := map[string]bool{"range": true}
	for _, f := range ledgerFlags {
		takesValue[f.flag] = true
	}
	id := ""
	var flags []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case strings.HasPrefix(a, "-"):
			flags = append(flags, a)
			name := strings.TrimLeft(a, "-")
			if strings.Contains(name, "=") {
				continue
			}
			if takesValue[name] && i+1 < len(args) {
				flags = append(flags, args[i+1])
				i++
			}
		case id == "":
			id = a
		default:
			return "", nil, fmt.Errorf("two task ids given (%q and %q) — finish takes one", id, a)
		}
	}
	return id, flags, nil
}

// exitCode reads the script's own verdict off the error the runner
// returned; -1 says the runner failed before the script spoke, which is
// not a verdict to map.
func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var verdict interface{ ExitCode() int }
	if errors.As(err, &verdict) && verdict.ExitCode() > 0 {
		return verdict.ExitCode()
	}
	return -1
}

// passthrough hands the script's verdict up unedited: the frame turns
// an error carrying an exit code into that exit code, having reported
// nothing over what the script already said on its own stream.
func passthrough(script string, err error) error {
	if err == nil {
		return nil
	}
	if exitCode(err) < 0 {
		return fmt.Errorf("running %s: %w", script, err)
	}
	return err
}
