package amendcmd

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
)

func TestNewDeclaresTheCommand(t *testing.T) {
	c := New(Deps{})
	if c.Name != "amend" {
		t.Errorf("name = %q, want amend", c.Name)
	}
	if c.Need != command.NeedAdopted {
		t.Errorf("need = %v; want an adopted repository", c.Need)
	}
	if c.Summary == "" || c.Run == nil {
		t.Error("the command carries no summary or no work")
	}
}

// The command the frame dispatches is the work itself — the wiring New
// builds is what runs, not a second path beside it.
func TestTheDeclaredCommandRunsTheWork(t *testing.T) {
	h := newHarness(t)
	if err := New(h.deps()).Run(h.ctx, []string{"spec-0011", "--title", amendTitle}); err != nil {
		t.Fatalf("amend = %v", err)
	}
	if !h.gh.reached("pr create") {
		t.Error("the command the frame would dispatch opened nothing")
	}
}

// Step 1: a spec that is not approved is refused, and the refusal is
// the whole command — nothing is read from the forge, nothing is asked
// (spec-0011, acceptance criteria).
func TestANonApprovedSpecIsRefused(t *testing.T) {
	for _, status := range []string{"draft", "implemented", "superseded"} {
		t.Run(status, func(t *testing.T) {
			h := newHarness(t)
			h.seed(specPath, specFixture(status))
			err := h.amend()
			if err == nil {
				t.Fatal("amend accepted a spec that is not approved")
			}
			if !strings.Contains(err.Error(), "spec-0011 is '"+status+"'") {
				t.Errorf("error = %v; want it to name the status it found", err)
			}
			if got := h.read(t, specPath); got != specFixture(status) {
				t.Error("the spec file was written despite the refusal")
			}
			if len(h.term.Asked) != 0 {
				t.Errorf("asked %v; a refusal asks nothing", h.term.Asked)
			}
			if len(h.gh.calls) != 0 {
				t.Errorf("the forge was reached: %v", h.gh.calls)
			}
		})
	}
}

// The whole green path: the spec flipped, the branch cut, the commit
// made, the branch pushed, and the pull request opened ready.
func TestTheConfirmedAmendWritesPushesAndOpensReady(t *testing.T) {
	h := newHarness(t)
	if err := h.amend(); err != nil {
		t.Fatalf("amend = %v", err)
	}
	if got := field([]byte(h.read(t, specPath)), "status"); got != "draft" {
		t.Errorf("spec status = %q, want draft", got)
	}
	if !h.git.ran("switch -c docs/amend-command origin/main") {
		t.Errorf("the branch was not cut from the authority branch; git saw %v", h.git.calls)
	}
	if got := h.git.arg("commit"); !strings.Contains(got, "docs(specs): return spec-0011 to draft") {
		t.Errorf("commit = %q; want the conventional subject", got)
	}
	if !h.git.ran("push -u origin docs/amend-command") {
		t.Errorf("the branch was never pushed; git saw %v", h.git.calls)
	}
	created := h.gh.created()
	if created["--title"] != "[Docs][Specs] "+amendTitle {
		t.Errorf("title = %q; want the declared bracketed style with no task tag", created["--title"])
	}
	if _, drafted := created["--draft"]; drafted {
		t.Error("the amendment opened as a draft; an amendment announces no work and opens ready")
	}
	if created["--base"] != "main" {
		t.Errorf("base = %q, want main", created["--base"])
	}
}

// The order the effects land in is the order spec-0011 fixes: the queue
// edit before the push, the push before the pull request.
func TestTheEditPrecedesThePushAndThePushPrecedesTheOpening(t *testing.T) {
	h := newHarness(t)
	wroteBeforeSwitch := false
	h.git.onSwap = func() {
		if field([]byte(h.read(t, specPath)), "status") == "draft" {
			wroteBeforeSwitch = true
		}
	}
	if err := h.amend(); err != nil {
		t.Fatalf("amend = %v", err)
	}
	if wroteBeforeSwitch {
		t.Error("the queue edit landed before the branch existed — it belongs on the branch")
	}
	var order []string
	for _, c := range h.git.calls {
		if strings.HasPrefix(c, "commit") || strings.HasPrefix(c, "push") {
			order = append(order, strings.Fields(c)[0])
		}
	}
	if strings.Join(order, ",") != "commit,push" {
		t.Errorf("git order = %v; want the commit then the push", order)
	}
	if !h.gh.reached("pr create") {
		t.Error("no pull request was opened")
	}
}

// The one difference from finish this package makes deliberately: a no
// leaves the working tree exactly as it was, which is what shape.md
// asks of a refused command (report-0015; the package comment).
func TestADeclineWritesNothingAndReachesNoForge(t *testing.T) {
	h := newHarness(t)
	h.term.ConfirmAnswer = false
	err := h.amend()
	if !errors.Is(err, command.ErrDeclined) {
		t.Fatalf("amend = %v; want the decline", err)
	}
	if got := h.read(t, specPath); got != specFixture("approved") {
		t.Error("the declined amend left the spec edited")
	}
	if h.git.ran("switch") || h.git.ran("commit") || h.git.ran("push") {
		t.Errorf("the declined amend acted on git: %v", h.git.calls)
	}
	if h.gh.reached("pr create") {
		t.Error("the declined amend opened a pull request")
	}
}

// The composition is shown before the question, so a yes answers
// something the human has read (product/pull-requests/shape.md).
func TestTheCompositionIsShownBeforeTheQuestion(t *testing.T) {
	h := newHarness(t)
	if err := h.amend(); err != nil {
		t.Fatalf("amend = %v", err)
	}
	shown := h.out.String()
	for _, want := range []string{
		"spec-0011 → draft",
		"branch     docs/amend-command",
		"commit     docs(specs): return spec-0011 to draft",
		"title      [Docs][Specs] " + amendTitle,
		"Suspends #42 — task-0012 waits on this amendment.",
		"ready for review",
	} {
		if !strings.Contains(shown, want) {
			t.Errorf("the composition never showed %q:\n%s", want, shown)
		}
	}
	if len(h.term.Asked) != 1 {
		t.Fatalf("asked %v; --title answered the first question, so one is left", h.term.Asked)
	}
	if !strings.Contains(h.term.Asked[0], "spec-0011") {
		t.Errorf("the question is %q; want it to name what it does", h.term.Asked[0])
	}
}

// Without --title the sentence is typed, and it is asked before the
// confirmation — a yes answers a composition that is already whole.
func TestTheTitleIsAskedWhenNoFlagAnsweredIt(t *testing.T) {
	h := newHarness(t)
	h.term.InputAnswer = "Reopen the gate by hand"
	if err := run(h.ctx, h.deps(), []string{"spec-0011"}); err != nil {
		t.Fatalf("amend = %v", err)
	}
	if len(h.term.Asked) != 2 {
		t.Fatalf("asked %v; want the sentence then the confirmation", h.term.Asked)
	}
	if got := h.gh.created()["--title"]; got != "[Docs][Specs] Reopen the gate by hand" {
		t.Errorf("title = %q; want what was typed", got)
	}
}

// A sentence nobody wrote is not a title.
func TestAnEmptyTitleIsRefused(t *testing.T) {
	h := newHarness(t)
	h.term.InputAnswer = "   "
	err := run(h.ctx, h.deps(), []string{"spec-0011"})
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("amend = %v; want a refusal", err)
	}
	if h.git.ran("switch") {
		t.Error("a branch was cut for a title nobody wrote")
	}
}

// Step 3: the body carries the line check_amendment_reference.sh
// accepts, naming the right pull request and the right task.
func TestTheBodyNamesTheSuspendedPullRequest(t *testing.T) {
	h := newHarness(t)
	if err := h.amend(); err != nil {
		t.Fatalf("amend = %v", err)
	}
	body := h.gh.created()["--body"]
	if !strings.Contains(body, "Suspends #42 — task-0012 waits on this amendment.") {
		t.Errorf("body carries no suspension line:\n%s", body)
	}
}

// Several tasks in flight on one spec: one Suspends line per pull
// request (spec-0011, edge cases).
func TestEveryInFlightTaskGetsItsOwnLine(t *testing.T) {
	h := newHarness(t)
	h.seed(otherTask, taskFixture("task-0013", "in-review", "spec-0011"))
	h.env["WRITRUN_PR_LIST"] = strings.Join([]string{
		"42\ttask/0012-amend-command\tsomeone\t[TASK-0012] Amend the thing",
		"43\tsome/other-branch\tsomeone\t[TASK-0013] Another thing",
	}, "\n")
	if err := h.amend(); err != nil {
		t.Fatalf("amend = %v", err)
	}
	body := h.gh.created()["--body"]
	for _, want := range []string{
		"Suspends #42 — task-0012 waits on this amendment.",
		"Suspends #43 — task-0013 waits on this amendment.",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body carries no %q:\n%s", want, body)
		}
	}
}

// A task out of flight is the ordinary pre-implementation amendment,
// which owes no reference and costs nothing (spec-0011, edge cases).
func TestNoInFlightTaskOwesNoReference(t *testing.T) {
	h := newHarness(t)
	h.seed(taskPath, taskFixture("task-0012", "ready", "spec-0011"))
	if err := h.amend(); err != nil {
		t.Fatalf("amend = %v", err)
	}
	body := h.gh.created()["--body"]
	if strings.Contains(body, "Suspends") {
		t.Errorf("body claims a suspension nothing is waiting on:\n%s", body)
	}
	if !strings.Contains(body, "suspends nothing") {
		t.Errorf("body never says the amendment suspends nothing:\n%s", body)
	}
	if !strings.Contains(h.out.String(), "suspends   nothing") {
		t.Errorf("the composition never said so:\n%s", h.out.String())
	}
}

// When the forge cannot be read, the reference is still composed from
// the queue and the command says it must be checked by hand
// (spec-0011, acceptance criteria).
func TestAnUnreadableForgeStillComposesFromTheQueue(t *testing.T) {
	h := newHarness(t)
	delete(h.env, "WRITRUN_PR_LIST")
	h.gh.listErr = errors.New("gh: could not reach the forge")
	if err := h.amend(); err != nil {
		t.Fatalf("amend = %v", err)
	}
	if !strings.Contains(h.errb.String(), "check by hand") {
		t.Errorf("the narrow view was never named:\n%s", h.errb.String())
	}
	body := h.gh.created()["--body"]
	if !strings.Contains(body, "task-0012") {
		t.Errorf("the reference was not composed from the queue:\n%s", body)
	}
}

// The branch carries no task id, whatever the suspended task is: an
// amendment records that an approval is in question, it does not work
// the task (conventions/branches.md).
func TestTheBranchCarriesNoTaskId(t *testing.T) {
	h := newHarness(t)
	if err := h.amend(); err != nil {
		t.Fatalf("amend = %v", err)
	}
	branch := "docs/amend-command"
	if !h.git.ran("switch -c " + branch) {
		t.Fatalf("branch was not %s; git saw %v", branch, h.git.calls)
	}
	if strings.Contains(branch, "task/") || strings.Contains(branch, "0012") {
		t.Errorf("branch %q carries a task id", branch)
	}
	if title := h.gh.created()["--title"]; strings.Contains(title, "[TASK-") {
		t.Errorf("title %q carries a task tag", title)
	}
}

// No task's status or taken_by moves — flight belongs to the task's own
// pull request's events (spec-0011, acceptance criteria).
func TestNoTaskIsTouched(t *testing.T) {
	h := newHarness(t)
	before := h.read(t, taskPath)
	if err := h.amend(); err != nil {
		t.Fatalf("amend = %v", err)
	}
	if after := h.read(t, taskPath); after != before {
		t.Error("the task file changed; amend touches no task")
	}
}

// A dirty tree would ride into the branch this cuts, so it is refused
// before anything is composed or asked.
func TestADirtyTreeIsRefused(t *testing.T) {
	h := newHarness(t)
	h.git.dirty = " M docs/about.md\n"
	err := h.amend()
	if err == nil || !strings.Contains(err.Error(), "dirty") {
		t.Fatalf("amend = %v; want a refusal naming the dirty tree", err)
	}
	if len(h.term.Asked) != 0 {
		t.Errorf("asked %v; a refusal asks nothing", h.term.Asked)
	}
}

// A branch already there is two amendments on one pull request.
func TestAnExistingBranchIsRefused(t *testing.T) {
	for _, ref := range []string{"refs/heads/docs/amend-command", "refs/remotes/origin/docs/amend-command"} {
		t.Run(ref, func(t *testing.T) {
			h := newHarness(t)
			h.git.refs[ref] = true
			err := h.amend()
			if err == nil || !strings.Contains(err.Error(), "already exists") {
				t.Fatalf("amend = %v; want a refusal naming the branch", err)
			}
			if h.git.ran("switch") {
				t.Error("a branch was cut anyway")
			}
		})
	}
}

// The style is the adopter's, read by the adopter's own reader.
func TestTheTitleFollowsTheDeclaredStyle(t *testing.T) {
	h := newHarness(t)
	h.scripts.replies["stage_2.pr_title_style"] = "conventional\n"
	if err := h.amend(); err != nil {
		t.Fatalf("amend = %v", err)
	}
	if got := h.gh.created()["--title"]; got != "docs(specs): "+amendTitle {
		t.Errorf("title = %q; want the conventional style", got)
	}
	if len(h.scripts.calls) == 0 || !strings.Contains(h.scripts.calls[0], "stage_2.pr_title_style") {
		t.Errorf("the style was not read from the repository's own reader: %v", h.scripts.calls)
	}
}

// --slug names the branch's subject; --type names its prefix.
func TestTheBranchTakesTheGivenSlugAndType(t *testing.T) {
	h := newHarness(t)
	if err := h.amend("--slug", "Reopen The Gate", "--type", "fix"); err != nil {
		t.Fatalf("amend = %v", err)
	}
	if !h.git.ran("switch -c fix/reopen-the-gate") {
		t.Errorf("git saw %v", h.git.calls)
	}
	if got := h.gh.created()["--title"]; got != "[Fix][Specs] "+amendTitle {
		t.Errorf("title = %q", got)
	}
}

// A --type that is not a branch prefix at all is refused here; whether
// it is in the project's vocabulary is the door's judgement.
func TestAnUnusableTypeIsRefused(t *testing.T) {
	h := newHarness(t)
	err := h.amend("--type", "Docs/Specs")
	if err == nil || !strings.Contains(err.Error(), "--type") {
		t.Fatalf("amend = %v; want a refusal naming the flag", err)
	}
}

// Two ids, no id, and an id resolving to nothing are all refusals that
// change nothing.
func TestTheArgumentsAreRefusedBeforeAnythingHappens(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"no id", []string{"--title", amendTitle}, "names the spec"},
		{"two ids", []string{"spec-0011", "spec-0012", "--title", amendTitle}, "amend takes one"},
		{"no such spec", []string{"spec-0099", "--title", amendTitle}, "resolves to no file"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newHarness(t)
			err := run(h.ctx, h.deps(), c.args)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("amend = %v; want %q", err, c.want)
			}
			if len(h.git.calls) != 0 || len(h.gh.calls) != 0 {
				t.Error("a refusal reached git or the forge")
			}
		})
	}
}

// A failure after the branch exists names the state it left behind
// (product/rules.md).
func TestAFailureAfterTheBranchNamesTheResume(t *testing.T) {
	h := newHarness(t)
	h.git.fail["push"] = errors.New("remote rejected")
	err := h.amend()
	if err == nil || !strings.Contains(err.Error(), "git push -u origin docs/amend-command") {
		t.Fatalf("amend = %v; want the resume named", err)
	}
	if h.gh.reached("pr create") {
		t.Error("the pull request was opened after the push failed")
	}
}

// A forge that refuses the opening leaves a pushed branch, which is the
// one state the act must not leave silently.
func TestAFailedOpeningNamesThePushedBranch(t *testing.T) {
	h := newHarness(t)
	h.gh.createErr = errors.New("gh: validation failed")
	err := h.amend()
	if err == nil || !strings.Contains(err.Error(), "is pushed and has no pull request") {
		t.Fatalf("amend = %v; want the half-done act named", err)
	}
}

// The spec is re-read on the branch: a base carrying a newer spec is
// amended, never overwritten with the old checkout's bytes.
func TestTheSpecIsReReadOnTheBranch(t *testing.T) {
	h := newHarness(t)
	newer := strings.Replace(specFixture("approved"), "A body paragraph.", "A newer paragraph.", 1)
	h.git.onSwap = func() { h.seed(specPath, newer) }
	if err := h.amend(); err != nil {
		t.Fatalf("amend = %v", err)
	}
	got := h.read(t, specPath)
	if !strings.Contains(got, "A newer paragraph.") {
		t.Error("the branch's own version of the spec was overwritten")
	}
	if field([]byte(got), "status") != "draft" {
		t.Error("the re-read spec was not returned to draft")
	}
}

// A base whose spec is no longer approved is somebody else's amendment:
// nothing is written over it.
func TestASpecAmendedElsewhereStopsBeforeTheWrite(t *testing.T) {
	h := newHarness(t)
	h.git.onSwap = func() { h.seed(specPath, specFixture("draft")) }
	err := h.amend()
	if err == nil || !strings.Contains(err.Error(), "amended elsewhere") {
		t.Fatalf("amend = %v; want the collision named", err)
	}
	// statusOrNone carries its own quotes, so the sentence must not add
	// a second pair around them.
	if !strings.Contains(err.Error(), "reads 'draft' on") {
		t.Errorf("error = %v; want the status quoted exactly once", err)
	}
	if h.git.ran("commit") {
		t.Error("a commit was made over somebody else's amendment")
	}
}

// The blocker this package shipped with: `amend task-0012` resolved to
// spec-0012 — a real file, about different work — and under --yes,
// which is the path the project's own rules make first-class, nothing
// showed the composition in time for a person to catch it. The refusal
// is before anything is composed, so neither git nor the forge is
// touched.
func TestATaskIdIsNotResolvedToASpec(t *testing.T) {
	h := newHarness(t)
	otherSpec := "work/specs/spec-0012-release-distribution.md"
	h.seed(otherSpec, strings.ReplaceAll(specFixture("approved"), "spec-0011", "spec-0012"))
	h.ctx.Yes = true

	err := run(h.ctx, h.deps(), []string{"task-0012", "--title", amendTitle})
	if err == nil {
		t.Fatal("amend accepted a task id and amended a spec")
	}
	if !strings.Contains(err.Error(), "task-0012") || !strings.Contains(err.Error(), "different") {
		t.Errorf("error = %v; want it to name the id and say it is another file", err)
	}
	if got := h.read(t, otherSpec); field([]byte(got), "status") != "approved" {
		t.Error("spec-0012 was returned to draft by an id naming a task")
	}
	if got := h.read(t, specPath); field([]byte(got), "status") != "approved" {
		t.Error("spec-0011 was written")
	}
	if len(h.git.calls) != 0 || len(h.gh.calls) != 0 {
		t.Errorf("a refusal reached git (%v) or the forge (%v)", h.git.calls, h.gh.calls)
	}
}

// The forge answered and named no pull request working the task. The
// terminal already said so; the body used to say "the forge did not
// answer" and claim a suspension that does not exist — a false sentence
// the kit's own gate passes over, because it reads the same state as a
// stale flight and asks for no reference.
func TestAStaleFlightStateIsNotBlamedOnTheForge(t *testing.T) {
	h := newHarness(t)
	h.env["WRITRUN_PR_LIST"] = "77\tdocs/unrelated\tsomeone\t[Docs] Something else"
	if err := h.amend(); err != nil {
		t.Fatalf("amend = %v", err)
	}
	body := h.gh.created()["--body"]
	if strings.Contains(body, "the forge did not answer") {
		t.Errorf("the body blames a forge that answered:\n%s", body)
	}
	if strings.Contains(body, "Suspends") {
		t.Errorf("the body claims a suspension no pull request carries:\n%s", body)
	}
	if !strings.Contains(body, "task-0012") || !strings.Contains(body, "stale") {
		t.Errorf("the body never names the stale flight state:\n%s", body)
	}
	// The terminal was already right, and stays right.
	if !strings.Contains(h.out.String(), "no open pull request works it") {
		t.Errorf("the composition never said so:\n%s", h.out.String())
	}
	if strings.Contains(h.errb.String(), "check by hand") {
		t.Errorf("a forge that answered was reported as unreachable:\n%s", h.errb.String())
	}
}

// `--body-file <file>` named nothing: the composed body existed only as
// the indented block `show` printed, and product/rules.md asks a failure
// after the first write to name the exact command that resumes the flow.
func TestTheFailedOpeningNamesABodyFileThatExists(t *testing.T) {
	h := newHarness(t)
	h.gh.createErr = errors.New("gh: validation failed")
	err := run(h.ctx, h.deps(), []string{"spec-0011", "--title", "Reopen $HOME and `id`"})
	if err == nil {
		t.Fatal("the failed opening was not reported")
	}
	if strings.Contains(err.Error(), "<file>") {
		t.Errorf("the resume command still names no file:\n%v", err)
	}

	// The named file exists and holds exactly the body that was handed
	// to the forge.
	file := ""
	for _, f := range strings.Fields(err.Error()) {
		if strings.Contains(f, "writrun-amend-") {
			file = strings.Trim(f, "'")
		}
	}
	if file == "" {
		t.Fatalf("no body file was named:\n%v", err)
	}
	saved, readErr := h.files.ReadFile(file)
	if readErr != nil {
		t.Fatalf("the named body file is not there: %v", readErr)
	}
	if string(saved) != h.gh.created()["--body"] {
		t.Errorf("the saved body is not the one that was sent:\n%s", saved)
	}

	// And the title is quoted for a shell, not for Go: `$HOME` and the
	// backticks must survive the paste.
	if !strings.Contains(err.Error(), "--title '[Docs][Specs] Reopen $HOME and `id`'") {
		t.Errorf("the title is not shell-quoted:\n%v", err)
	}
}

// The ordinary run leaves no body file behind — it is written on the one
// failure that needs it.
func TestTheOrdinaryRunWritesNoBodyFile(t *testing.T) {
	h := newHarness(t)
	if err := h.amend(); err != nil {
		t.Fatalf("amend = %v", err)
	}
	for _, p := range h.files.Paths() {
		if strings.Contains(p, "writrun-amend-") {
			t.Errorf("the green path left a body file behind: %s", p)
		}
	}
}

// A temp directory that cannot be made is not worth a second failure:
// the resume line then says where the body actually is.
func TestABodyThatCannotBeSavedStillNamesWhereItIs(t *testing.T) {
	h := newHarness(t)
	h.gh.createErr = errors.New("gh: validation failed")
	h.files.FailOp("mkdirtemp", "/tmp", errors.New("read-only file system"))
	err := h.amend()
	if err == nil {
		t.Fatal("the failed opening was not reported")
	}
	if !strings.Contains(err.Error(), "the body printed above") {
		t.Errorf("the fallback never says where the body is:\n%v", err)
	}
	if !strings.Contains(err.Error(), "read-only file system") {
		t.Errorf("the reason the body could not be saved went missing:\n%v", err)
	}
}

// A branch prefix has to be a word. `-` passed the older rule and
// composed the branch `-/amend-command`, which no convention describes.
func TestATypeThatIsNotAWordIsRefused(t *testing.T) {
	for _, kind := range []string{"-", "---", "-fix", "fix-", "Docs/Specs", "d0cs", ""} {
		t.Run(kind, func(t *testing.T) {
			h := newHarness(t)
			err := h.amend("--type", kind)
			if err == nil || !strings.Contains(err.Error(), "--type") {
				t.Fatalf("amend --type %q = %v; want a refusal naming the flag", kind, err)
			}
			if h.git.ran("switch") {
				t.Error("a branch was cut for a prefix that is not a word")
			}
		})
	}
	// A word still passes, dashes inside it included.
	h := newHarness(t)
	if err := h.amend("--type", "fix-up"); err != nil {
		t.Fatalf("amend --type fix-up = %v", err)
	}
	if !h.git.ran("switch -c fix-up/amend-command") {
		t.Errorf("git saw %v", h.git.calls)
	}
}

// A quoted status was echoed raw as `spec-0011 is '"approved"'`, which
// reads as a fault in the reader rather than in the file. Which files
// are accepted does not change — only what the refusal says.
func TestAQuotedStatusIsRefusedLegibly(t *testing.T) {
	h := newHarness(t)
	h.seed(specPath, strings.Replace(specFixture("approved"), "status: approved", `status: "approved"`, 1))
	err := h.amend()
	if err == nil {
		t.Fatal("a quoted status was accepted")
	}
	if strings.Contains(err.Error(), `'"approved"'`) {
		t.Errorf("the raw echo is still there:\n%v", err)
	}
	for _, want := range []string{"quoted value", "front matter quotes nothing"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %v; want it to name %q", err, want)
		}
	}

	// A spec carrying no status at all is still said in words.
	h2 := newHarness(t)
	h2.seed(specPath, strings.Replace(specFixture("approved"), "status: approved\n", "", 1))
	if err := h2.amend(); err == nil || !strings.Contains(err.Error(), "no status at all") {
		t.Fatalf("amend = %v; want the absent status named", err)
	}
}

// A truncated list reports a suspended task's pull request as
// nonexistent — which, since the forge did answer, now composes a body
// saying the flight state is stale. The same warning
// check_amendment_reference.sh prints, for the same reason.
func TestATruncatedPullRequestListIsNamed(t *testing.T) {
	h := newHarness(t)
	delete(h.env, "WRITRUN_PR_LIST")
	lines := make([]string, prFetchLimit)
	for i := range lines {
		lines[i] = fmt.Sprintf("%d\tdocs/other-%d\tsomeone\t[Docs] Something", i+1, i)
	}
	h.gh.list = strings.Join(lines, "\n")
	if err := h.amend(); err != nil {
		t.Fatalf("amend = %v", err)
	}
	if !strings.Contains(h.errb.String(), "hit its 200 limit") {
		t.Errorf("the truncated list was never named:\n%s", h.errb.String())
	}
}

// A list that fits says nothing.
func TestAListThatFitsIsNotCalledTruncated(t *testing.T) {
	h := newHarness(t)
	delete(h.env, "WRITRUN_PR_LIST")
	h.gh.list = "42\ttask/0012-amend-command\tsomeone\t[TASK-0012] Amend the thing"
	if err := h.amend(); err != nil {
		t.Fatalf("amend = %v", err)
	}
	if strings.Contains(h.errb.String(), "limit") {
		t.Errorf("a list that fits was called truncated:\n%s", h.errb.String())
	}
}

// With no origin/main and no main, the amendment is cut from HEAD and
// whatever this checkout carries rides into the pull request. That is a
// fallback, not the intent, and it used to happen in silence.
func TestTheHeadFallbackIsNamed(t *testing.T) {
	h := newHarness(t)
	h.git.refs = map[string]bool{}
	if err := h.amend(); err != nil {
		t.Fatalf("amend = %v", err)
	}
	if !h.git.ran("switch -c docs/amend-command HEAD") {
		t.Errorf("git saw %v", h.git.calls)
	}
	if !strings.Contains(h.errb.String(), "cut from HEAD") {
		t.Errorf("the fallback was silent:\n%s", h.errb.String())
	}
}

// The local main is still a base, and taking it says nothing.
func TestTheLocalMainIsTakenQuietly(t *testing.T) {
	h := newHarness(t)
	h.git.refs = map[string]bool{"refs/heads/main": true}
	if err := h.amend(); err != nil {
		t.Fatalf("amend = %v", err)
	}
	if !h.git.ran("switch -c docs/amend-command main") {
		t.Errorf("git saw %v", h.git.calls)
	}
	if strings.Contains(h.errb.String(), "HEAD") {
		t.Errorf("a base that resolved was reported as a fallback:\n%s", h.errb.String())
	}
}

// A successful amend leaves the checkout on the amendment branch, which
// both error paths already said and the success path did not.
func TestTheSuccessSaysWhereItLeftYou(t *testing.T) {
	h := newHarness(t)
	if err := h.amend(); err != nil {
		t.Fatalf("amend = %v", err)
	}
	if !strings.Contains(h.out.String(), "git switch -") {
		t.Errorf("the success never said how to get back:\n%s", h.out.String())
	}
}

// The forge is asked the question in the shape `gh` answers. No
// integration case reaches this — WRITRUN_PR_LIST short-circuits before
// d.Gh — so the argument list is held here.
func TestTheForgeIsAskedInTheShapeGhAnswers(t *testing.T) {
	h := newHarness(t)
	delete(h.env, "WRITRUN_PR_LIST")
	if err := h.amend(); err != nil {
		t.Fatalf("amend = %v", err)
	}
	var listed []string
	for _, c := range h.gh.calls {
		if len(c) > 1 && c[0] == "pr" && c[1] == "list" {
			listed = c
		}
	}
	if listed == nil {
		t.Fatalf("the forge was never listed: %v", h.gh.calls)
	}
	joined := strings.Join(listed, " ")
	for _, want := range []string{
		"--state open",
		"--limit 200",
		"--json number,headRefName,author,title",
		`.[] | "\(.number)\t\(.headRefName)\t\(.author.login)\t\(.title)"`,
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("gh pr list carried no %q: %v", want, listed)
		}
	}
}
