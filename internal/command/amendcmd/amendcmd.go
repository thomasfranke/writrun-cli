// Package amendcmd is `writrun amend`: an approved spec returned to
// `draft` for re-approval, and the pull request that says why — naming
// the in-flight pull request the amendment suspends, so the kit's own
// `check_amendment_reference.sh` finds the reference it asks for
// (docs/product/pull-requests/amend.md, spec-0011).
//
// What it never does: approve, re-approve, or merge. Returning a spec
// to draft reopens the gate; walking back through it is the merge's,
// and the merge is the maintainer's (docs/product/rules.md). It touches
// no task either — flight belongs to the task's own pull request's
// events, and the only record of the pause is the relation between the
// two pull requests.
//
// # Where the one queue edit lands
//
// spec-0011's Steps put the write at step 2 and the confirmation at
// step 4. Here the edit is computed at step 2, shown at step 4, and
// written alongside the push, so a declined amend leaves the working
// tree exactly as it found it — which is what
// `product/pull-requests/shape.md` asks of a refused command ("no
// half-written status, no orphan branch").
//
// **The deferral is required, not a preference.** Two later steps read
// the spec's status, and a step-2 write to the working tree would make
// this command refuse itself:
//
//   - the dirty-tree refusal below reads `git status --porcelain`, which
//     a written spec makes non-empty; and
//   - act re-reads the spec after `git switch -c`, which carries an
//     uncommitted modification onto the new branch, and stops on a
//     status that is no longer `approved`.
//
// Both guards are this command's own additions rather than spec-0011's
// Steps, so this is not the claim that the order does not matter. It is
// the opposite: writing at step 2 is what the run cannot survive, and
// holding the edit to the confirmed path is what lets the rest of it
// happen at all.
package amendcmd

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/kit"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// settingScript is the adopted repository's own reader for
// `.writrun/settings.json` — the declared title style is read there and
// never parsed here.
const settingScript = ".writrun/scripts/stage-2-pull-requests/read_setting.sh"

// observanceScript is the door: the same check the forge runs over an
// open pull request, asked here about the composition instead. The
// vocabulary it judges against lives in that file and nowhere in this
// binary.
const observanceScript = ".writrun/scripts/stage-2-pull-requests/check_observance.sh"

// composedRange is what the door is handed while the amendment's own
// commit does not exist. `HEAD...HEAD` resolves to a merge base of HEAD
// and a tip of HEAD, so it names no commits and the credit half judges
// the composed body alone — which is the only thing there is to judge
// before the first write.
const composedRange = "HEAD...HEAD"

// Deps is the wiring amend needs beyond the frame's Ctx.
type Deps struct {
	// Scripts runs the adopted repository's own scripts.
	Scripts kit.Runner
	// Files is the filesystem: the queue read, and the one edit written.
	Files vfs.FS
	// Git cuts the branch, commits the edit and pushes it.
	Git gitx.Runner
	// Gh is the forge: the open pull requests listed, and the
	// amendment's own opened.
	Gh func(args ...string) (string, error)
	// Getenv reads WRITRUN_PR_LIST — the kit's own seam for a supplied
	// pull-request list, honoured here so a suite can answer the forge's
	// question without a forge (take_task.sh).
	Getenv func(string) string
}

// New returns the amend command wired with its dependencies.
func New(d Deps) command.Command {
	return command.Command{
		Name:    "amend",
		Summary: "amend a spec: returned to draft, the suspended pull request named, the amendment opened",
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
	fs := flag.NewFlagSet("amend", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	summaryFlag := fs.String("title", "", "the sentence the pull-request title carries")
	slugFlag := fs.String("slug", "", "the branch's subject words")
	kindFlag := fs.String("type", "docs", "the branch prefix and the title's type")
	if err := fs.Parse(flags); err != nil {
		return err
	}
	if id == "" {
		return errors.New("amend names the spec it returns to draft: writrun amend spec-NNNN")
	}

	// 1 — a spec that is not approved. Nothing else runs, nothing is
	// read from the forge, nothing is asked (spec-0011, step 1).
	rel, err := queueFile(d.Files, ctx.Root, specsDir, "spec", id)
	if err != nil {
		return err
	}
	content, err := d.Files.ReadFile(path.Join(ctx.Root, rel))
	if err != nil {
		return fmt.Errorf("reading %s: %w", rel, err)
	}
	specID := field(content, "id")
	if specID == "" {
		return fmt.Errorf("%s carries no id", rel)
	}
	if status := field(content, "status"); status != "approved" {
		return fmt.Errorf("%s is %s — amend returns an approved spec to draft; "+
			"a draft is already open for change and an implemented one is history (%s)",
			specID, statusOrNone(status), rel)
	}

	// 2 — the one queue edit, proved here and written on the confirmed
	// path only. See the package comment for why the write cannot land
	// here; proving it now is what keeps a spec whose front matter
	// cannot be written from being discovered after the branch is cut.
	// The status was just read as `approved`, so a proof that returns no
	// error always changed the file: only the error is a verdict.
	if _, _, err := setField(content, "status", "draft"); err != nil {
		return fmt.Errorf("%s: %w", rel, err)
	}

	kind := strings.ToLower(strings.TrimSpace(*kindFlag))
	if !plainWord(kind) {
		return fmt.Errorf("--type %q is not a branch prefix — one lowercase word, e.g. docs", *kindFlag)
	}
	// A dirty tree would ride into the branch this cuts, and an
	// amendment carrying somebody's unrelated work is not an amendment.
	if dirty, err := d.Git(ctx.Root, "status", "--porcelain"); err != nil {
		return fmt.Errorf("reading the working tree: %w", err)
	} else if strings.TrimSpace(dirty) != "" {
		return fmt.Errorf("the working tree is dirty — commit or stash first:\n%s", indent(dirty, 5))
	}

	slug := slugify(*slugFlag)
	if slug == "" {
		slug = slugify(specSlug(rel))
	}
	if slug == "" {
		return fmt.Errorf("%s gives the branch no subject words — name them with --slug", rel)
	}
	branch := branchName(kind, slug)
	if err := branchIsFree(d, ctx.Root, branch); err != nil {
		return err
	}

	// 3 — who is suspended, and which pull request each of them rides.
	tasks, err := suspended(d.Files, ctx.Root, specID)
	if err != nil {
		return err
	}
	pulls, forgeRead := openPulls(d, ctx.Stderr)
	susp := match(tasks, pulls)
	if len(tasks) > 0 && !forgeRead {
		fmt.Fprintf(ctx.Stderr,
			"The forge did not answer, so the pull request this amendment suspends could not\n"+
				"be numbered. The reference below is composed from the queue — check by hand that\n"+
				"the body names the right pull request before this is opened.\n")
	}

	// 4 — the composition, shown whole, then the question.
	style := readSetting(ctx, d, "stage_2.pr_title_style")
	summary, err := ctx.AskInput("The sentence the pull-request title carries:", *summaryFlag, "--title")
	if err != nil {
		return err
	}
	if strings.TrimSpace(summary) == "" {
		return errors.New("the title's sentence is empty — say what the amendment changes")
	}
	plan := plan{
		specID:  specID,
		relPath: rel,
		branch:  branch,
		subject: subject(kind, specID),
		title:   title(style, kind, summary),
		body:    body(readTemplate(d, ctx.Root), specID, summary, susp, forgeRead),
	}
	if err := observe(ctx, d, plan.title, plan.body); err != nil {
		return err
	}
	show(ctx, plan, tasks, susp, forgeRead)
	if err := ctx.AskConfirm(fmt.Sprintf(
		"Return %s to draft, push %s and open the amendment pull request?", specID, branch)); err != nil {
		return err
	}
	return act(ctx, d, plan)
}

// plan is everything the confirmed path performs, composed before the
// question and unchanged by it.
type plan struct {
	specID  string
	relPath string
	branch  string
	subject string
	title   string
	body    string
}

// show prints the whole act: what is edited, who waits, and the branch,
// the title and the body exactly as they will be opened
// (product/pull-requests/shape.md).
func show(ctx *command.Ctx, p plan, tasks []string, susp []suspension, forgeRead bool) {
	fmt.Fprintf(ctx.Stdout, "\namendment:\n")
	fmt.Fprintf(ctx.Stdout, "  spec       %s → draft (%s)\n", p.specID, p.relPath)
	switch {
	case len(tasks) == 0:
		fmt.Fprintf(ctx.Stdout, "  suspends   nothing — no task referencing %s is in flight\n", p.specID)
	default:
		for _, s := range susp {
			if s.number > 0 {
				fmt.Fprintf(ctx.Stdout, "  suspends   %s on #%d\n", s.task, s.number)
			} else if forgeRead {
				fmt.Fprintf(ctx.Stdout, "  suspends   %s — no open pull request works it\n", s.task)
			} else {
				fmt.Fprintf(ctx.Stdout, "  suspends   %s — the forge did not name its pull request\n", s.task)
			}
		}
	}
	fmt.Fprintf(ctx.Stdout, "  branch     %s\n", p.branch)
	fmt.Fprintf(ctx.Stdout, "  commit     %s\n", p.subject)
	fmt.Fprintf(ctx.Stdout, "  title      %s\n", p.title)
	fmt.Fprintf(ctx.Stdout, "  body:\n%s\n", indent(p.body, 4))
	fmt.Fprintf(ctx.Stdout, "  The pull request opens ready for review: an amendment announces no work.\n\n")
}

// act performs exactly what show printed: the branch cut from a fresh
// authority branch, the one queue edit written on it, committed, pushed,
// and the pull request opened ready. A failure after the branch exists
// names the state it left and how to finish it (product/rules.md).
func act(ctx *command.Ctx, d Deps, p plan) error {
	base := authority(ctx, d)
	if _, err := d.Git(ctx.Root, "switch", "-c", p.branch, base); err != nil {
		return fmt.Errorf("cutting %s from %s: %w", p.branch, base, err)
	}

	// The file is re-read on the branch it is about to change: the base
	// is a tree this checkout may not have been standing on, and an
	// amendment that pasted the old checkout's bytes over it would
	// revert whatever landed in between.
	full := path.Join(ctx.Root, p.relPath)
	onBranch, err := d.Files.ReadFile(full)
	if err != nil {
		return fmt.Errorf("reading %s on %s: %w — the branch is cut; `git switch -` leaves it behind", p.relPath, p.branch, err)
	}
	if status := field(onBranch, "status"); status != "approved" {
		return fmt.Errorf("%s reads %s on %s — it was amended elsewhere; nothing was written, and `git switch -` leaves the branch behind",
			p.specID, statusOrNone(status), base)
	}
	next, _, err := setField(onBranch, "status", "draft")
	if err != nil {
		return fmt.Errorf("%s: %w", p.relPath, err)
	}
	if err := d.Files.WriteFile(full, next, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", p.relPath, err)
	}
	fmt.Fprintf(ctx.Stdout, "wrote status: draft on %s\n", p.relPath)

	if _, err := d.Git(ctx.Root, "add", "--", p.relPath); err != nil {
		return fmt.Errorf("staging %s: %w", p.relPath, err)
	}
	if _, err := d.Git(ctx.Root, "commit", "-m", p.subject); err != nil {
		return fmt.Errorf("committing %s: %w", p.subject, err)
	}
	if _, err := d.Git(ctx.Root, "push", "-u", "origin", p.branch); err != nil {
		return fmt.Errorf("pushing %s: %w\nThe amendment is committed locally on %s; finish it with:\n  git push -u origin %s",
			p.branch, err, p.branch, p.branch)
	}
	out, err := d.Gh("pr", "create", "--base", "main", "--head", p.branch,
		"--title", p.title, "--body", p.body)
	if err != nil {
		return fmt.Errorf("opening the amendment pull request: %w\n"+
			"%s is pushed and has no pull request, which is the one state this act must not leave behind.\n"+
			"Finish it with:\n  gh pr create --base main --head %s --title %s --body-file %s",
			err, p.branch, p.branch, shellQuote(p.title), bodyFileArg(d, p.body))
	}
	if said := strings.TrimSpace(out); said != "" {
		fmt.Fprintln(ctx.Stdout, said)
	}
	fmt.Fprintf(ctx.Stdout, "Amended %s: %s pushed, pull request open and ready for review.\n", p.specID, p.branch)
	fmt.Fprintf(ctx.Stdout, "You are on %s now; `git switch -` returns to where you were.\n", p.branch)
	fmt.Fprintf(ctx.Stdout, "Re-approval is the merge's — no command walks back through that gate.\n")
	return nil
}

// bodyFileArg is the `--body-file` argument of the resume command, and
// it names a file that exists: the body is the one part of this act a
// person cannot retype, and a placeholder `<file>` left them
// reconstructing it out of the indented block `show` printed. It is
// written only here, on the failure that needs it, so the ordinary run
// leaves no litter. A temp directory that cannot be made is not worth a
// second failure — the argument then says where the body actually is.
func bodyFileArg(d Deps, body string) string {
	dir, err := d.Files.MkdirTemp("", "writrun-amend-")
	if err == nil {
		name := path.Join(dir, "amendment-body.md")
		if err = d.Files.WriteFile(name, []byte(body), 0o644); err == nil {
			return shellQuote(name)
		}
	}
	return "<the body printed above, saved to a file by hand: " + err.Error() + ">"
}

// shellQuote wraps s for a POSIX shell. The resume line is a command a
// person pastes, and Go's %q is Go's quoting — a title carrying `$`, a
// backtick or a backslash would run as something other than itself.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// authority is the branch the amendment is cut from, resolved the way
// preflight.sh and take_task.sh resolve their own: the pushed main,
// else the local one, else wherever this checkout stands. The fetch is
// best-effort — a stale base is worth naming, not worth refusing over,
// and it happens on the confirmed path so a declined amend reaches
// nothing at all.
func authority(ctx *command.Ctx, d Deps) string {
	_, _ = d.Git(ctx.Root, "fetch", "origin", "main")
	for _, b := range []struct{ ref, name string }{
		{"refs/remotes/origin/main", "origin/main"},
		{"refs/heads/main", "main"},
	} {
		if _, err := d.Git(ctx.Root, "rev-parse", "--verify", "--quiet", b.ref); err == nil {
			return b.name
		}
	}
	// Standing where the amendment is cut from is the fallback, not the
	// intent: whatever this checkout carries rides into the pull request
	// alongside the one queue edit, and against `--base main` that reads
	// as the amendment's own work.
	fmt.Fprintf(ctx.Stderr,
		"Neither origin/main nor main resolves here, so the amendment is cut from HEAD.\n"+
			"Whatever this checkout carries beyond main will ride into the pull request.\n")
	return "HEAD"
}

// branchIsFree refuses a branch that already exists on either side. An
// amendment is a new change; reusing a name is how two of them end up
// on one pull request.
func branchIsFree(d Deps, root, branch string) error {
	for _, ref := range []string{"refs/heads/" + branch, "refs/remotes/origin/" + branch} {
		if _, err := d.Git(root, "rev-parse", "--verify", "--quiet", ref); err == nil {
			return fmt.Errorf("%s already exists (%s) — name another with --slug", branch, ref)
		}
	}
	return nil
}

// pull is the part of an open pull request this command reads.
type pull struct {
	number int
	branch string
	title  string
}

// openPulls answers which pull requests are open, and whether the
// question could be asked at all. WRITRUN_PR_LIST is the kit's own
// seam, honoured first for the same reason take_task.sh honours it: a
// suite must be able to answer the forge's question without a forge.
func openPulls(d Deps, w io.Writer) ([]pull, bool) {
	if d.Getenv != nil {
		if raw := d.Getenv("WRITRUN_PR_LIST"); raw != "" {
			return parsePulls(raw), true
		}
	}
	if d.Gh == nil {
		return nil, false
	}
	out, err := d.Gh("pr", "list", "--state", "open", "--limit", prFetchLimitArg,
		"--json", "number,headRefName,author,title",
		"--jq", `.[] | "\(.number)\t\(.headRefName)\t\(.author.login)\t\(.title)"`)
	if err != nil {
		// Best-effort, deliberately — the same contract
		// check_amendment_reference.sh states: without an answer the
		// number cannot be known, and a command that failed here would
		// refuse an amendment over a question it could not ask.
		return nil, false
	}
	pulls := parsePulls(out)
	// The same warning check_amendment_reference.sh prints for the same
	// reason: a silently truncated list reports a suspended task's pull
	// request as nonexistent, which here would compose a body saying
	// nothing works the task when something does.
	if len(pulls) >= prFetchLimit && w != nil {
		fmt.Fprintf(w, "The open-pull-request list hit its %d limit, so it may be incomplete —\n"+
			"a suspension named below as unworked may simply have fallen off it.\n", prFetchLimit)
	}
	return pulls, true
}

// parsePulls reads the tab-separated listing both the seam and `gh`
// produce: number, head branch, author, title. A three-field line is
// read as number, branch, title — the shape
// check_amendment_reference.sh asks `gh` for — because the title is the
// field that carries meaning here and the author carries none.
func parsePulls(raw string) []pull {
	var out []pull
	for _, line := range strings.Split(raw, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		f := strings.Split(line, "\t")
		if len(f) < 2 {
			continue
		}
		num := 0
		if _, err := fmt.Sscanf(strings.TrimSpace(f[0]), "%d", &num); err != nil || num <= 0 {
			continue
		}
		p := pull{number: num, branch: strings.TrimSpace(f[1])}
		switch {
		case len(f) >= 4:
			p.title = f[3]
		case len(f) == 3:
			p.title = f[2]
		}
		out = append(out, p)
	}
	return out
}

// match pairs each suspended task with the open pull request working
// it. A task no open pull request works keeps its place in the list
// with no number: check_amendment_reference.sh calls that a stale
// flight state and not this change's business, and the composition
// says as much rather than dropping the task in silence.
func match(tasks []string, pulls []pull) []suspension {
	out := make([]suspension, 0, len(tasks))
	for _, task := range tasks {
		s := suspension{task: task}
		want := numOf(task)
		for _, p := range pulls {
			for _, carried := range carriedOf(p.branch, p.title) {
				if carried == want {
					s.number = p.number
					break
				}
			}
			if s.number > 0 {
				break
			}
		}
		out = append(out, s)
	}
	return out
}

// readSetting asks the adopted repository's own reader. An unreadable
// answer is the empty string, and composing then falls to the declared
// default rather than to a second parser of somebody else's file.
func readSetting(ctx *command.Ctx, d Deps, address string) string {
	if d.Scripts == nil {
		return ""
	}
	var out bytes.Buffer
	if err := d.Scripts(ctx.Root, &out, io.Discard, nil, settingScript, address); err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}

// observe hands the composed title and body to the door and carries
// its verdict back unchanged — the same script the forge runs, asked
// where a refusal still costs nothing. Whether the type is in the
// project's vocabulary is that script's judgement and stays there; this
// command holds no copy of the list.
//
// **The text goes through the environment, never argv.** The script's
// own header names inline interpolation as the way its title and body
// must not arrive, and `kit.Runner` carries an environment for exactly
// this call.
//
// The command's own streams are handed over, so a green run says "OK —
// the title observes …" and a refusal says which word is outside the
// list, both in the script's words rather than this command's.
func observe(ctx *command.Ctx, d Deps, title, body string) error {
	if d.Scripts == nil {
		return errors.New("no script runner is wired, so the composed title cannot be judged — " +
			"amend does not compose past a check it cannot run")
	}
	env := []string{"PR_TITLE=" + title, "PR_BODY=" + body}
	return passthrough(observanceScript,
		d.Scripts(ctx.Root, ctx.Stdout, ctx.Stderr, env, observanceScript, composedRange))
}

// passthrough hands the script's verdict up unedited: the frame turns
// an error carrying an exit code into that exit code, having reported
// nothing over what the script already said on its own stream. An error
// with no exit code is the runner failing before the script spoke —
// a missing or unreadable check among them — and that is named.
func passthrough(script string, err error) error {
	if err == nil {
		return nil
	}
	var verdict interface{ ExitCode() int }
	if errors.As(err, &verdict) && verdict.ExitCode() > 0 {
		return err
	}
	return fmt.Errorf("running %s: %w", script, err)
}

// readTemplate reads the adopter's pull-request body template. Absent,
// the composition falls back to the same headings.
func readTemplate(d Deps, root string) string {
	b, err := d.Files.ReadFile(path.Join(root, template))
	if err != nil {
		return ""
	}
	return string(b)
}

// split separates the spec id from the flags. Go's flag package stops
// at the first operand, so `amend spec-0011 --title x` would leave the
// title unparsed; the id is lifted out first and the flags parsed whole
// — takecmd's rule, kept here rather than shared, because the two
// commands' flag sets are their own.
func split(args []string) (string, []string, error) {
	takesValue := map[string]bool{"title": true, "slug": true, "type": true}
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
			return "", nil, fmt.Errorf("two spec ids given (%q and %q) — amend takes one", id, a)
		}
	}
	return id, flags, nil
}

// plainWord reports whether s is one lowercase word — enough to be a
// branch prefix. It must carry a letter and may not lean on a dash at
// either end: `-` passed the older rule and composed the branch
// `-/slug`, which is a name no convention describes.
//
// Whether the word is in the project's vocabulary is
// check_observance.sh's judgement, and a second copy of that list here
// would be a second authority — `product/rules.md` says no command
// reimplements a check, and `conventions/commits.md` is prose an agent
// reads rather than a format this can parse. That script is asked
// directly, over the composed title, before the branch is cut
// (observe).
func plainWord(s string) bool {
	if s == "" || strings.HasPrefix(s, "-") || strings.HasSuffix(s, "-") {
		return false
	}
	letter := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			letter = true
		case r == '-':
		default:
			return false
		}
	}
	return letter
}

// statusOrNone names what the spec's status line reads as, carrying its
// own quotes so the sentence around it does not add a second pair.
//
// An absent status is said in words, and a value that quotes itself is
// named as that: `status: "approved"` used to be echoed raw as
// `spec-0011 is '"approved"'`, which reads as a fault in this reader
// rather than in the file. The quotes really are the difference —
// ql_fm_field reads them as part of the value, so the queue's own
// reader does not see `approved` either.
func statusOrNone(s string) string {
	t := strings.TrimSpace(s)
	if t == "" {
		return "no status at all"
	}
	if strings.Trim(t, `"'`) != t {
		return fmt.Sprintf("%s — a quoted value, and queue front matter quotes nothing, "+
			"so the queue's own reader does not see the bare word either", t)
	}
	return "'" + t + "'"
}

// prFetchLimit is `gh pr list --limit`, the number
// check_amendment_reference.sh settled on for the same question.
const (
	prFetchLimit    = 200
	prFetchLimitArg = "200"
)

// indent shifts a block right so it reads as quoted rather than said.
func indent(s string, n int) string {
	pad := strings.Repeat(" ", n)
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i, l := range lines {
		lines[i] = pad + l
	}
	return strings.Join(lines, "\n")
}
