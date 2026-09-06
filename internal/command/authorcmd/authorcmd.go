// Package authorcmd is `writrun author`: the authoring pull request —
// the one that writes a rule and derives the work from it — composed
// from a diff that already exists and opened **ready**, never draft.
// The rule is written before this command runs; what it adds is the
// checks in their fixed order, the composition read off the diff and
// the queue, and the question before anything reaches the forge
// (docs/product/pull-requests/author.md, spec-0009).
package authorcmd

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/kit"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// The three authorities this command runs, in the order it runs them,
// and the two files it reads through the script port. Not one of their
// judgements is repeated here.
const (
	frontMatterScript = ".writrun/skills/writrun-check-front-matter/check_front_matter.sh"
	docShapesScript   = ".writrun/scripts/stage-2-pull-requests/check_doc_shapes.sh"
	stateScript       = ".writrun/skills/writrun-check-task-state/check_state.sh"
	settingScript     = ".writrun/scripts/stage-2-pull-requests/read_setting.sh"
)

// Deps is the wiring author needs beyond the frame's Ctx.
type Deps struct {
	// Scripts runs the adopted repository's own scripts.
	Scripts kit.Runner
	// Files reads the queue files the diff adds and the body template.
	Files vfs.FS
	// Git answers what the change is, and carries the branch to the
	// forge once the word is given.
	Git gitx.Runner
	// Gh is the forge: the pull request opened ready.
	Gh func(args ...string) (string, error)
}

// New returns the author command wired with its dependencies.
func New(d Deps) command.Command {
	return command.Command{
		Name:    "author",
		Summary: "author a rule: the checks run, the derived work listed, the pull request opened ready",
		Need:    command.NeedAdopted,
		Run: func(ctx *command.Ctx, args []string) error {
			return run(ctx, d, args)
		},
	}
}

func run(ctx *command.Ctx, d Deps, args []string) error {
	fs := flag.NewFlagSet("author", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	slugFlag := fs.String("slug", "", "the branch's subject words, after docs/")
	titleFlag := fs.String("title", "", "the pull request's title, in the declared style, with no task tag")
	rangeFlag := fs.String("range", "", "the diff range the checks read this change against")
	resumeFlag := fs.Bool("resume", false, "finish an authoring whose branch is pushed and whose pull request never opened")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("author takes no arguments (%q given) — what it opens is the diff already on this branch", fs.Arg(0))
	}

	// 1 — what the diff is. A change that is not an authoring one is
	// refused before any check runs, and before the forge is named
	// (spec-0009, step 1 and edge cases).
	ch, err := readChange(d.Git, ctx.Root, *rangeFlag)
	if err != nil {
		return err
	}
	if !*resumeFlag {
		if _, err := d.Git(ctx.Root, "rev-parse", "--verify", "--quiet", "refs/remotes/origin/"+ch.branch); err == nil {
			return fmt.Errorf("%s is already on the forge — authoring starts locally; --resume finishes an authoring whose pull request never opened", ch.branch)
		}
	}

	branch, err := branchName(ch, *slugFlag)
	if err != nil {
		return err
	}
	if branch != ch.branch {
		if *resumeFlag {
			return fmt.Errorf("--resume finishes the authoring this branch already pushed, but the composition names %s and this is %s", branch, ch.branch)
		}
		if _, err := d.Git(ctx.Root, "rev-parse", "--verify", "--quiet", "refs/heads/"+branch); err == nil {
			return fmt.Errorf("%s already exists locally and is not this branch — name another with --slug", branch)
		}
	}

	// 2 — the checks, in the order the methodology fixed. The first
	// non-zero verdict is the whole answer: no branch, no push, no
	// pull request (spec-0009, acceptance criteria).
	for _, c := range []struct {
		script string
		args   []string
	}{
		{frontMatterScript, nil},
		{docShapesScript, nil},
		{stateScript, []string{ch.rng}},
	} {
		if err := d.Scripts(ctx.Root, ctx.Stdout, ctx.Stderr, c.script, c.args...); err != nil {
			return passthrough(c.script, err)
		}
	}

	// 3 — the composition, filled from the diff and the queue rather
	// than typed. Only the summary is the human's (spec-0009, step 2).
	rows, err := derived(d, ctx.Root, ch.rng)
	if err != nil {
		return err
	}
	title, err := ctx.AskInput(titleQuestion(d, ctx.Root, *titleFlag), *titleFlag, "--title")
	if err != nil {
		return err
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return errors.New("the pull request's title is empty — the summary is yours to write")
	}
	// An authoring pull request's tasks are born in it, so it carries
	// no `[TASK-NNNN]` tag (conventions/prs.md). This command never
	// adds one, and refuses to carry one it was handed.
	if taskTag.MatchString(title) {
		return fmt.Errorf("%q carries a task tag — an authoring pull request carries none; its tasks are born in it", title)
	}
	c := composition{
		branch: branch,
		slug:   strings.TrimPrefix(branch, docsPrefix),
		title:  title,
		body:   body(d, ctx.Root, rows),
		files:  ch.files,
		rng:    *rangeFlag,
	}
	show(ctx.Stdout, c)

	// 4 — the question, then the forge, and nothing before it.
	if err := ctx.AskConfirm("Push the branch and open the pull request ready for review?"); err != nil {
		return err
	}
	return open(ctx, d, ch, c)
}

// composition is what the command will do, shown whole before it is
// asked for.
type composition struct {
	branch string
	slug   string
	title  string
	body   string
	files  []string
	// rng is the range the run was given, empty when it was inferred —
	// carried so a resume names the same one.
	rng string
}

// show prints the branch, the title, the body and the files, in that
// order: everything the question is about, before the question.
func show(w io.Writer, c composition) {
	fmt.Fprintf(w, "\nbranch: %s\n", c.branch)
	fmt.Fprintf(w, "title:  %s\n", c.title)
	fmt.Fprintf(w, "body:\n")
	for _, l := range strings.Split(strings.TrimRight(c.body, "\n"), "\n") {
		fmt.Fprintf(w, "  | %s\n", l)
	}
	fmt.Fprintf(w, "files:\n")
	for _, f := range c.files {
		fmt.Fprintf(w, "  %s\n", f)
	}
	fmt.Fprintln(w)
}

// open performs exactly the act that was shown. The forge is verified
// before the branch is cut, so a repository left carrying a branch
// nobody can push is a state this ordering never reaches — take_task.sh's
// own ordering, kept because the failure it avoids is the same one.
func open(ctx *command.Ctx, d Deps, ch change, c composition) error {
	if _, err := d.Gh("auth", "status"); err != nil {
		return fmt.Errorf("gh cannot reach the forge: %w\nNothing was pushed and no pull request was opened", err)
	}
	if c.branch != ch.branch {
		if _, err := d.Git(ctx.Root, "switch", "-c", c.branch); err != nil {
			return fmt.Errorf("cutting %s: %w", c.branch, err)
		}
	}
	if _, err := d.Git(ctx.Root, "push", "-u", "origin", c.branch); err != nil {
		return fmt.Errorf("%w\n%s is kept local. Finish the act with:\n  %s", err, c.branch, resumeCommand(c))
	}
	out, err := d.Gh("pr", "create", "--base", "main", "--head", c.branch,
		"--title", c.title, "--body", c.body)
	if err != nil {
		return fmt.Errorf("%w\n%s is pushed but has no pull request, which is the one state this act must not leave behind.\nFinish it with:\n  %s",
			err, c.branch, resumeCommand(c))
	}
	if s := strings.TrimSpace(out); s != "" {
		fmt.Fprintln(ctx.Stdout, s)
	}
	fmt.Fprintf(ctx.Stdout, "Authored: %s pushed, pull request open and ready for review.\n", c.branch)
	fmt.Fprintln(ctx.Stdout, "An authoring pull request has no work to announce, so it never opens as a draft.")
	return nil
}

// resumeCommand is the rerun that finishes a half-done act, written in
// one place so the two failure paths cannot name different acts. It
// carries every argument that decided the branch, the title and the
// range: a rerun composing something else would not be the resume it
// claims to be.
func resumeCommand(c composition) string {
	s := fmt.Sprintf("writrun author --slug %s --title %q", c.slug, c.title)
	if c.rng != "" {
		s += " --range " + c.rng
	}
	return s + " --resume"
}

// titleQuestion names the style the project declared, so the summary is
// written in it rather than looked up. The style is read through the
// repository's own `read_setting.sh` — this command has no opinion about
// what the file says, and none about whether the answer obeys it: the
// title is judged at the door, by check_observance.sh, and a second
// judge here would be a second authority.
func titleQuestion(d Deps, root, preset string) string {
	if preset != "" {
		return "The pull request's title:"
	}
	style := setting(d, root, "stage_2.pr_title_style")
	if style == "" {
		style = "declared"
	}
	return fmt.Sprintf("The pull request's title — no task tag, in the %s style (e.g. %s):", style, styleExample(style))
}

// styleExample is one title the named style accepts, printed with the
// question so the shape needs no second document.
func styleExample(style string) string {
	if style == "conventional" {
		return "docs(product): the merge is the assenting act"
	}
	return "[DOCS] The merge is the assenting act"
}

// setting reads one value from the adopted repository's settings the
// way every other reader does — through the kit's own script. An
// unreadable setting is an empty answer, never a guess.
func setting(d Deps, root, address string) string {
	var out bytes.Buffer
	if err := d.Scripts(root, &out, io.Discard, settingScript, address); err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
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
