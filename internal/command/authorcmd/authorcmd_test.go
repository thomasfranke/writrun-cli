package authorcmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/queue"
)

func TestNewDeclaresTheCommand(t *testing.T) {
	c := New(Deps{})
	if c.Name != "author" {
		t.Errorf("name = %q, want author", c.Name)
	}
	if c.Need != command.NeedAdopted {
		t.Errorf("need = %v, want NeedAdopted", c.Need)
	}
	if c.Summary == "" {
		t.Error("the command carries no summary, and --help prints one line per command")
	}
	if c.Run == nil {
		t.Fatal("the command carries no Run")
	}
	h := newHarness(t)
	if err := New(h.deps()).Run(h.ctx, []string{"--title", title}); err != nil {
		t.Errorf("the wired command: %v", err)
	}
}

func TestStyleExampleFollowsTheDeclaredStyle(t *testing.T) {
	if got := styleExample("conventional"); !strings.HasPrefix(got, "docs(") {
		t.Errorf("styleExample(conventional) = %q", got)
	}
	if got := styleExample("bracketed"); !strings.HasPrefix(got, "[DOCS]") {
		t.Errorf("styleExample(bracketed) = %q", got)
	}
}

// A git that cannot answer is never read as an empty answer: each read
// names what it was asking.
func TestAGitThatCannotAnswerIsNamed(t *testing.T) {
	cases := []struct{ failing, want string }{
		{"status", "reading the working tree"},
		{"rev-parse --abbrev-ref", "reading the current branch"},
		{"diff --name-only origin", "reading the diff"},
		{"diff --name-only --diff-filter=A", "adds under"},
	}
	for _, c := range cases {
		t.Run(c.failing, func(t *testing.T) {
			h := newHarness(t)
			base := h.git.run
			deps := h.deps()
			deps.Git = func(dir string, args ...string) (string, error) {
				if strings.HasPrefix(strings.Join(args, " "), c.failing) {
					return "", errors.New("git said no")
				}
				return base(dir, args...)
			}
			err := run(h.ctx, deps, []string{"--title", title})
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err = %v, want it to say %q", err, c.want)
			}
		})
	}
}

// The green path: the three checks in their fixed order, the
// composition shown, and a pull request opened **ready** — an authoring
// pull request has no work to announce (spec-0009, acceptance
// criteria).
func TestTheConfirmedRunOpensTheReadyPullRequest(t *testing.T) {
	h := newHarness(t)
	if err := h.author(); err != nil {
		t.Fatalf("author: %v", err)
	}

	want := []string{frontMatterScript, docShapesScript, stateScript}
	var ran []string
	for _, c := range h.scripts.calls {
		for _, w := range want {
			if strings.HasPrefix(c, w) {
				ran = append(ran, w)
			}
		}
	}
	if strings.Join(ran, " ") != strings.Join(want, " ") {
		t.Errorf("checks ran %v, want %v in that order", ran, want)
	}
	if !strings.Contains(strings.Join(h.scripts.calls, "\n"), stateScript+" origin/main...HEAD") {
		t.Errorf("the state check was not given the range: %v", h.scripts.calls)
	}
	if !h.gh.reached("pr create") {
		t.Fatalf("no pull request was opened: %v", h.gh.calls)
	}
	if strings.Contains(h.gh.created, "--draft") {
		t.Errorf("the pull request was opened as a draft: %s", h.gh.created)
	}
	for _, want := range []string{"--base main", "--head " + authorBranch, "--title " + title} {
		if !strings.Contains(h.gh.created, want) {
			t.Errorf("gh pr create carries no %q: %s", want, h.gh.created)
		}
	}
	if !h.git.did("push -u origin " + authorBranch) {
		t.Errorf("the branch was never pushed: %v", h.git.calls)
	}
	if !strings.Contains(h.out.String(), "ready for review") {
		t.Errorf("the run never said the pull request is ready:\n%s", h.out.String())
	}
}

// The composition is shown whole — branch, title, body, files — before
// the question that acts on it (spec-0009, step 3).
func TestTheCompositionIsShownBeforeTheQuestion(t *testing.T) {
	h := newHarness(t)
	if err := h.author(); err != nil {
		t.Fatalf("author: %v", err)
	}
	out := h.out.String()
	for _, want := range []string{
		"branch: " + authorBranch,
		"title:  " + title,
		"| " + derivedHeading,
		"| | task-0016 | spec-0014 | Declare the derived work |",
		"docs/product/pull-requests/author.md",
		newTaskPath,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the composition never showed %q:\n%s", want, out)
		}
	}
	if len(h.term.Asked) == 0 || !strings.HasPrefix(h.term.Asked[len(h.term.Asked)-1], "Push the branch") {
		t.Errorf("the last question was not the one about the forge: %v", h.term.Asked)
	}
}

// The body is the template's authoring half: the Derived-work table
// filled from what the diff adds, and no `## Spec` (spec-0009, step 2).
func TestTheBodyCarriesTheFilledTableAndNoSpecHalf(t *testing.T) {
	h := newHarness(t)
	if err := h.author(); err != nil {
		t.Fatalf("author: %v", err)
	}
	body := h.gh.created
	if !strings.Contains(body, "| task-0016 | spec-0014 | Declare the derived work |") {
		t.Errorf("the table was not filled from the diff:\n%s", body)
	}
	if strings.Contains(body, "task-NNNN") {
		t.Errorf("the template's placeholder row survived:\n%s", body)
	}
	if strings.Contains(body, specHeading+"\n") || strings.Contains(body, "Implements spec-NNNN") {
		t.Errorf("the implementing half survived:\n%s", body)
	}
	if !strings.Contains(body, derivedHeading) {
		t.Errorf("the contract marker heading is gone:\n%s", body)
	}
}

// A rule that derives nothing declares it: an empty section and a
// forgotten one look identical (spec-0009, acceptance criteria).
func TestARuleThatDerivesNothingDeclaresNone(t *testing.T) {
	h := newHarness(t)
	h.git.files = []string{"docs/product/rules.md"}
	h.git.added = map[string][]string{}
	if err := h.author(); err != nil {
		t.Fatalf("author: %v", err)
	}
	if !strings.Contains(h.gh.created, derivedNone) {
		t.Errorf("the body declares no derived work at all:\n%s", h.gh.created)
	}
}

// A spec derived for a task that already existed is derived work too,
// and the table says so rather than dropping it.
func TestASpecWithNoAddedTaskStillGetsARow(t *testing.T) {
	h := newHarness(t)
	h.git.added = map[string][]string{queue.SpecsDir + "/spec-*.md": {newSpecPath}}
	if err := h.author(); err != nil {
		t.Fatalf("author: %v", err)
	}
	if !strings.Contains(h.gh.created, "| task-0016 | spec-0014 | The declaration is the section |") {
		t.Errorf("the orphan spec has no row:\n%s", h.gh.created)
	}
}

// A cell the queue leaves empty is shown as empty, never invented.
func TestASpecNamingNoTaskLeavesTheColumnEmpty(t *testing.T) {
	h := newHarness(t)
	h.git.added = map[string][]string{queue.SpecsDir + "/spec-*.md": {newSpecPath}}
	h.seed(newSpecPath, specFixture("spec-0014", "null", "The declaration is the section"))
	if err := h.author(); err != nil {
		t.Fatalf("author: %v", err)
	}
	if !strings.Contains(h.gh.created, "| — | spec-0014 |") {
		t.Errorf("the empty task column was filled in:\n%s", h.gh.created)
	}
}

// A queue file the diff adds but whose front matter names no id is
// named by its filename rather than dropped from the table.
func TestAQueueFileWithNoIdIsNamedByItsFile(t *testing.T) {
	h := newHarness(t)
	h.seed(newTaskPath, "# Declare the derived work\n\nNo front matter at all.\n")
	h.seed(newSpecPath, "# spec — the contract\n")
	if err := h.author(); err != nil {
		t.Fatalf("author: %v", err)
	}
	if !strings.Contains(h.gh.created, "| task-0016-declare-derived-work |") ||
		!strings.Contains(h.gh.created, "spec-0014-declare-derived-work") {
		t.Errorf("the table lost a derived file:\n%s", h.gh.created)
	}
}

// Each check stops the sequence where it failed: nothing later runs and
// nothing reaches the forge (spec-0009, acceptance criteria).
func TestAFailingCheckStopsTheSequence(t *testing.T) {
	order := []string{frontMatterScript, docShapesScript, stateScript}
	for i, failing := range order {
		t.Run(failing, func(t *testing.T) {
			h := newHarness(t)
			h.scripts.replies[failing] = reply{errOut: "REJECTED: something\n", err: scriptExit(1)}
			err := h.author()
			if exitOf(err) != 1 {
				t.Fatalf("exit = %d, want the script's own 1 (err %v)", exitOf(err), err)
			}
			for _, later := range order[i+1:] {
				if h.scripts.ran(later) {
					t.Errorf("%s ran after %s failed", later, failing)
				}
			}
			if h.gh.reached("pr create") || h.git.did("push") || h.git.did("switch") {
				t.Errorf("the forge or the branch was reached after %s failed: gh %v git %v", failing, h.gh.calls, h.git.calls)
			}
			if !strings.Contains(h.errb.String(), "REJECTED") {
				t.Errorf("the script's own words never reached stderr: %q", h.errb.String())
			}
		})
	}
}

// A runner that failed before the script spoke is not a verdict to map,
// so it is named instead of passed through.
func TestARunnerFailureNamesTheScript(t *testing.T) {
	h := newHarness(t)
	h.scripts.replies[docShapesScript] = reply{err: errors.New("bash: not found")}
	err := h.author()
	if err == nil || !strings.Contains(err.Error(), docShapesScript) {
		t.Fatalf("err = %v, want it to name %s", err, docShapesScript)
	}
}

// The refusals that cost no check and no forge call: they are read off
// the diff and the branch alone (spec-0009, step 1 and edge cases).
func TestTheChangeIsRefusedBeforeAnyCheckRuns(t *testing.T) {
	cases := []struct {
		name  string
		spoil func(*harness)
		want  string
	}{
		{
			name:  "no docs change",
			spoil: func(h *harness) { h.git.files = []string{newTaskPath, newSpecPath} },
			want:  "touches no docs/ path",
		},
		{
			name:  "a mixed diff",
			spoil: func(h *harness) { h.git.files = append(h.git.files, "internal/command/run.go") },
			want:  "one kind per change",
		},
		{
			name:  "an empty diff",
			spoil: func(h *harness) { h.git.files = nil },
			want:  "is empty",
		},
		{
			name:  "a dirty tree",
			spoil: func(h *harness) { h.git.dirty = " M docs/product/rules.md\n" },
			want:  "working tree is dirty",
		},
		{
			name:  "a detached HEAD",
			spoil: func(h *harness) { h.git.branch = "HEAD" },
			want:  "detached HEAD",
		},
		{
			name:  "no base to read against",
			spoil: func(h *harness) { h.git.refs = map[string]bool{} },
			want:  "--range",
		},
		{
			name: "a branch already on the forge",
			spoil: func(h *harness) {
				h.git.refs["refs/remotes/origin/"+authorBranch] = true
			},
			want: "authoring starts locally",
		},
		{
			name: "a composed branch that already exists",
			spoil: func(h *harness) {
				h.git.branch = "scratch"
				h.git.refs["refs/heads/docs/author"] = true
			},
			want: "already exists locally",
		},
		{
			// The branch HEAD carries is local, so the refusal above
			// never fires — and the act would cut the composed name
			// fresh and push it over whatever the forge already has
			// under it, which is somebody's open pull request.
			name: "a composed branch already on the forge",
			spoil: func(h *harness) {
				h.git.branch = "scratch"
				h.git.refs["refs/remotes/origin/docs/author"] = true
			},
			want: "already on the forge",
		},
		{
			name: "a fetch that could not answer",
			spoil: func(h *harness) {
				h.git.fetchErr = errors.New("could not read from remote repository")
			},
			want: "stale origin",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newHarness(t)
			c.spoil(h)
			err := h.author()
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err = %v, want it to say %q", err, c.want)
			}
			if h.scripts.ran(frontMatterScript) {
				t.Error("a check ran on a change that was refused")
			}
			if len(h.gh.calls) > 0 {
				t.Errorf("the forge was reached: %v", h.gh.calls)
			}
		})
	}
}

// author opens the diff already on the branch, so it takes no operand
// to name something else.
func TestAnArgumentIsRefused(t *testing.T) {
	h := newHarness(t)
	err := run(h.ctx, h.deps(), []string{"task-0016", "--title", title})
	if err == nil || !strings.Contains(err.Error(), "takes no arguments") {
		t.Fatalf("err = %v, want a refusal naming the operand", err)
	}
}

func TestAnUnknownFlagIsRefused(t *testing.T) {
	h := newHarness(t)
	if err := run(h.ctx, h.deps(), []string{"--draft"}); err == nil {
		t.Fatal("an unknown flag was accepted")
	}
}

// The tasks of an authoring pull request are born in it, so its title
// carries no task tag — this command never adds one and refuses to
// carry one (conventions/prs.md).
func TestATitleCarryingATaskTagIsRefused(t *testing.T) {
	h := newHarness(t)
	err := run(h.ctx, h.deps(), []string{"--title", "[TASK-0016][DOCS] The merge is the assenting act"})
	if err == nil || !strings.Contains(err.Error(), "carries a task tag") {
		t.Fatalf("err = %v, want a refusal about the tag", err)
	}
	if len(h.gh.calls) > 0 {
		t.Errorf("the forge was reached: %v", h.gh.calls)
	}
}

func TestAnEmptyTitleIsRefused(t *testing.T) {
	h := newHarness(t)
	h.term.InputAnswer = "   "
	err := run(h.ctx, h.deps(), nil)
	if err == nil || !strings.Contains(err.Error(), "title is empty") {
		t.Fatalf("err = %v, want a refusal about the empty title", err)
	}
}

// The title is asked in the style the project declared, read through
// the repository's own settings script.
func TestTheTitleQuestionNamesTheDeclaredStyle(t *testing.T) {
	h := newHarness(t)
	h.term.InputAnswer = title
	if err := run(h.ctx, h.deps(), nil); err != nil {
		t.Fatalf("author: %v", err)
	}
	asked := strings.Join(h.term.Asked, "\n")
	if !strings.Contains(asked, "bracketed") || !strings.Contains(asked, "[DOCS]") {
		t.Errorf("the question never named the declared style: %v", h.term.Asked)
	}
}

func TestAnUnreadableStyleStillAsks(t *testing.T) {
	h := newHarness(t)
	h.scripts.replies[settingScript] = reply{err: scriptExit(3)}
	h.term.InputAnswer = title
	if err := run(h.ctx, h.deps(), nil); err != nil {
		t.Fatalf("author: %v", err)
	}
	if !strings.Contains(strings.Join(h.term.Asked, "\n"), "declared style") {
		t.Errorf("the question was not asked at all: %v", h.term.Asked)
	}
}

// A no leaves nothing behind: no branch, no push, no pull request
// (product/pull-requests/shape.md).
func TestADeclineReachesNothing(t *testing.T) {
	h := newHarness(t)
	h.term.ConfirmAnswer = false
	err := h.author()
	if !errors.Is(err, command.ErrDeclined) {
		t.Fatalf("err = %v, want ErrDeclined", err)
	}
	if len(h.gh.calls) > 0 {
		t.Errorf("the forge was reached: %v", h.gh.calls)
	}
	if h.git.did("switch") || h.git.did("push") {
		t.Errorf("the branch was touched: %v", h.git.calls)
	}
}

// The forge is verified before the branch is cut, so a repository left
// carrying a branch nobody can push is a state this ordering never
// reaches.
func TestAForgeThatCannotBeReachedCutsNoBranch(t *testing.T) {
	h := newHarness(t)
	h.gh.authErr = errors.New("not logged in")
	err := h.author()
	if err == nil || !strings.Contains(err.Error(), "Nothing was pushed") {
		t.Fatalf("err = %v, want it to say nothing was pushed", err)
	}
	if h.git.did("switch") || h.git.did("push") {
		t.Errorf("the branch was touched: %v", h.git.calls)
	}
}

// A failure after the first write names the exact command that resumes
// the flow (product/rules.md).
func TestAFailureAfterTheFirstWriteNamesTheResume(t *testing.T) {
	cases := []struct {
		name  string
		spoil func(*harness)
		want  string
	}{
		{"a push that failed", func(h *harness) { h.git.pushErr = errors.New("permission denied") }, "is kept local"},
		{"a pull request that never opened", func(h *harness) { h.gh.createErr = errors.New("api error") }, "has no pull request"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newHarness(t)
			c.spoil(h)
			err := h.author()
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err = %v, want it to say %q", err, c.want)
			}
			if !strings.Contains(err.Error(), "writrun author --slug ") ||
				!strings.Contains(err.Error(), "--resume") {
				t.Errorf("the resume was not named: %v", err)
			}
		})
	}
}

// resumeLine is the command the failure printed, lifted off the error
// the way a person lifts it off their terminal.
func resumeLine(t *testing.T, err error) string {
	t.Helper()
	if err == nil {
		t.Fatal("the act did not fail, so nothing was printed to resume it")
	}
	for _, l := range strings.Split(err.Error(), "\n") {
		if l = strings.TrimSpace(l); strings.HasPrefix(l, "writrun author") {
			return l
		}
	}
	t.Fatalf("no resume was printed:\n%v", err)
	return ""
}

// **The printed resume is run, not read.** A failure after the first
// write names the exact command that resumes the flow, and "exact"
// means the line finishes the act when it is pasted back into the
// context that produced it — a branch of any name, a title of any
// shape, and no terminal to answer the question with
// (product/rules.md).
func TestThePrintedResumeFinishesTheAct(t *testing.T) {
	cases := []struct{ name, branch, title string }{
		{
			// More subject words than the branch alphabet keeps: a
			// slug put back through the composition would name
			// docs/one-two-three and refuse the branch it was printed
			// for.
			name:   "a branch of more words than the alphabet keeps",
			branch: "docs/one-two-three-four",
			title:  title,
		},
		{
			name:   "a branch the alphabet would fold",
			branch: "docs/Rules_V2",
			title:  title,
		},
		{
			// Pasted into a shell, an unquoted title of this shape
			// runs `x` and expands $HOME before author ever sees it.
			name:   "a title carrying what a shell would read",
			branch: authorBranch,
			title:  "[DOCS] The $HOME rule `x` costs $5 — it's declared",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newHarness(t)
			h.git.branch = c.branch
			h.ctx.Yes = true
			h.term.In = false
			h.gh.createErr = errors.New("api error")
			line := resumeLine(t, run(h.ctx, h.deps(), []string{"--title", c.title}))

			// The act is half done: the branch is pushed, and the one
			// state it must not be left in is this one.
			r := newHarness(t)
			r.git.branch = c.branch
			r.git.refs["refs/remotes/origin/"+c.branch] = true
			if err := replay(t, r, line); err != nil {
				t.Fatalf("the printed resume did not finish the act:\n  %s\n%v", line, err)
			}
			if !r.gh.reached("pr create") {
				t.Errorf("the resume opened no pull request: %v", r.gh.calls)
			}
			if !strings.Contains(r.gh.created, "--head "+c.branch) {
				t.Errorf("the resume opened another branch's pull request: %s", r.gh.created)
			}
			if !strings.Contains(r.gh.created, "--title "+c.title) {
				t.Errorf("the title did not survive the round trip: %s", r.gh.created)
			}
		})
	}
}

// A resume finishes the branch under HEAD, and an authoring act never
// left a branch that is not one.
func TestResumeRefusesABranchNoActLeftBehind(t *testing.T) {
	h := newHarness(t)
	h.git.branch = "scratch"
	err := h.author("--resume")
	if err == nil || !strings.Contains(err.Error(), "is not a docs/ branch") {
		t.Fatalf("err = %v, want a refusal about the branch it is on", err)
	}
	if len(h.gh.calls) > 0 {
		t.Errorf("the forge was reached: %v", h.gh.calls)
	}
}

// The remote-tracking refs are a cache of the forge, and every answer
// read off them — the base, and whether a branch is public — is a
// confident wrong one until it is refreshed (take_task.sh, same read).
func TestTheForgeIsRefreshedBeforeItIsRead(t *testing.T) {
	h := newHarness(t)
	if err := h.author(); err != nil {
		t.Fatalf("author: %v", err)
	}
	fetched, read := -1, -1
	for i, c := range h.git.calls {
		if strings.HasPrefix(c, "fetch") && fetched < 0 {
			fetched = i
		}
		if strings.Contains(c, "refs/remotes/origin/") && read < 0 {
			read = i
		}
	}
	if fetched < 0 {
		t.Fatalf("nothing was fetched: %v", h.git.calls)
	}
	if read >= 0 && fetched > read {
		t.Errorf("origin was read at %d before the fetch at %d: %v", read, fetched, h.git.calls)
	}
}

// A repository with no forge is not stale, it is local: nothing is
// fetched, and the local main is the whole answer.
func TestARepositoryWithNoForgeFetchesNothing(t *testing.T) {
	h := newHarness(t)
	h.git.remotes = nil
	h.git.refs = map[string]bool{"refs/heads/main": true}
	if err := h.author(); err != nil {
		t.Fatalf("author: %v", err)
	}
	if h.git.did("fetch") {
		t.Errorf("a repository with no origin fetched anyway: %v", h.git.calls)
	}
	if !strings.Contains(strings.Join(h.scripts.calls, "\n"), stateScript+" main...HEAD") {
		t.Errorf("the local base was not used: %v", h.scripts.calls)
	}
}

// --resume finishes an authoring whose branch is pushed and whose pull
// request never opened — the one state the act must not leave behind.
func TestResumeFinishesAPushedBranch(t *testing.T) {
	h := newHarness(t)
	h.git.refs["refs/remotes/origin/"+authorBranch] = true
	if err := h.author("--resume"); err != nil {
		t.Fatalf("author --resume: %v", err)
	}
	if !h.gh.reached("pr create") {
		t.Errorf("the pull request was never opened: %v", h.gh.calls)
	}
	if h.git.did("switch") {
		t.Errorf("a resume cut a branch: %v", h.git.calls)
	}
}

// A resume that would compose a different branch is not the resume it
// claims to be.
func TestResumeRefusesADifferentBranch(t *testing.T) {
	h := newHarness(t)
	h.git.refs["refs/remotes/origin/"+authorBranch] = true
	err := h.author("--resume", "--slug", "something-else-entirely")
	if err == nil || !strings.Contains(err.Error(), "--resume finishes") {
		t.Fatalf("err = %v, want a refusal about the branch it names", err)
	}
	if len(h.gh.calls) > 0 {
		t.Errorf("the forge was reached: %v", h.gh.calls)
	}
}

// A branch that is not an authoring one is cut into `docs/<short-name>`
// on the word, and never before it.
func TestABranchIsCutWhenTheChangeIsNotOnAnAuthoringOne(t *testing.T) {
	h := newHarness(t)
	h.git.branch = "scratch"
	if err := h.author("--slug", "the merge is the assenting act"); err != nil {
		t.Fatalf("author: %v", err)
	}
	if !h.git.did("switch -c docs/the-merge-is") {
		t.Errorf("the branch was not cut as composed: %v", h.git.calls)
	}
	if !strings.Contains(h.gh.created, "--head docs/the-merge-is") {
		t.Errorf("the pull request was opened off another branch: %s", h.gh.created)
	}
}

func TestACutThatFailsIsNamed(t *testing.T) {
	h := newHarness(t)
	h.git.branch = "scratch"
	h.git.refs["refs/heads/docs/author"] = false
	h.gh.authErr = nil
	h.git.pushErr = nil
	// The cut fails by the ref already existing on the forge side of
	// the fake: the runner answers every switch with an error.
	h.git.calls = nil
	deps := h.deps()
	deps.Git = func(dir string, args ...string) (string, error) {
		if args[0] == "switch" {
			return "", errors.New("fatal: cannot switch")
		}
		return h.git.run(dir, args...)
	}
	err := run(h.ctx, deps, []string{"--title", title})
	if err == nil || !strings.Contains(err.Error(), "cutting docs/author") {
		t.Fatalf("err = %v, want it to name the cut", err)
	}
}

// A queue file the diff adds but the tree cannot read is a table this
// command must not guess at.
func TestAnUnreadableQueueFileStopsTheComposition(t *testing.T) {
	h := newHarness(t)
	h.files.FailOp("read", root+"/"+newTaskPath, errors.New("permission denied"))
	err := h.author()
	if err == nil || !strings.Contains(err.Error(), newTaskPath) {
		t.Fatalf("err = %v, want it to name the file it could not read", err)
	}
	if len(h.gh.calls) > 0 {
		t.Errorf("the forge was reached: %v", h.gh.calls)
	}
}

// A repository whose template cannot be read still gets a correct body:
// the contract marker is what the door reads, and it is never dropped.
func TestAMissingTemplateStillCarriesTheContractMarker(t *testing.T) {
	h := newHarness(t)
	h.files.FailOp("read", root+"/"+templatePath, errors.New("no such file"))
	if err := h.author(); err != nil {
		t.Fatalf("author: %v", err)
	}
	if !strings.Contains(h.gh.created, derivedHeading) ||
		!strings.Contains(h.gh.created, "| task-0016 | spec-0014 |") {
		t.Errorf("the fallback body lost the declaration:\n%s", h.gh.created)
	}
}

// The range the run was given is the range every reader gets, and the
// resume names it back.
func TestAGivenRangeReachesTheChecksAndTheResume(t *testing.T) {
	h := newHarness(t)
	h.git.pushErr = errors.New("permission denied")
	err := h.author("--range", "upstream/main...HEAD")
	if err == nil || !strings.Contains(err.Error(), "--range 'upstream/main...HEAD'") {
		t.Fatalf("err = %v, want the resume to name the range", err)
	}
	// The question was answered at a terminal, so the resume asks it
	// again there: --yes is carried where it was given, and nowhere
	// else, or a hand-confirmed act would resume unattended.
	if strings.Contains(resumeLine(t, err), "--yes") {
		t.Errorf("the resume carries --yes for a run that was asked: %s", resumeLine(t, err))
	}
	if !strings.Contains(strings.Join(h.scripts.calls, "\n"), stateScript+" upstream/main...HEAD") {
		t.Errorf("the state check read another range: %v", h.scripts.calls)
	}
}

// Without a terminal a question aborts rather than hanging, and --yes
// answers the one that is a confirmation.
func TestWithoutATerminalTheQuestionsAbort(t *testing.T) {
	h := newHarness(t)
	h.term.In = false
	if err := run(h.ctx, h.deps(), nil); err == nil || !strings.Contains(err.Error(), "--title") {
		t.Fatalf("err = %v, want it to name --title", err)
	}

	h = newHarness(t)
	h.term.In = false
	if err := h.author(); err == nil || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("err = %v, want it to name --yes", err)
	}

	h = newHarness(t)
	h.term.In = false
	h.ctx.Yes = true
	if err := h.author(); err != nil {
		t.Fatalf("author --yes: %v", err)
	}
	if !h.gh.reached("pr create") {
		t.Errorf("--yes did not answer the question: %v", h.gh.calls)
	}
}
