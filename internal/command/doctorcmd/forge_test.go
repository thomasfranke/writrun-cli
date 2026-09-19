package doctorcmd

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestStageOneMakesNoForgeRead(t *testing.T) {
	f := newFixture(t, "1")
	if found := f.findings(); len(found) != 0 {
		t.Errorf("findings = %d, want none:\n%s", len(found), texts(found))
	}
	if len(f.forge.calls) != 0 {
		t.Errorf("forge calls = %v; want none at stage 1", f.forge.calls)
	}
}

func TestStageTwoMakesNoIssuesRead(t *testing.T) {
	f := newFixture(t, "2")
	if found := f.findings(); len(found) != 0 {
		t.Errorf("findings = %d, want none:\n%s", len(found), texts(found))
	}
	if f.forge.asked("api repos/{owner}/{repo} --jq .has_issues") {
		t.Errorf("Issues were read at stage 2: %v", f.forge.calls)
	}
}

func TestSquashMergingOffIsNamed(t *testing.T) {
	f := newFixture(t, "3")
	f.forge.replies["api repos/{owner}/{repo} --jq .allow_squash_merge"] = "false\n"
	only(t, f.findings(), 2, breaks, "squash merging is off")
}

// A repository default of read is the tighter arrangement, not a
// defect: the workflows that record raise `contents: write` for
// themselves and the ones that never push stay on read (spec-0019).
func TestAReadDefaultPassesWhereEveryPushingWorkflowRaisesTheRight(t *testing.T) {
	cases := []struct{ name, workflow string }{
		{"contents: write on the workflow", recordingWorkflow},
		{"write-all on the workflow", writeAllWorkflow},
		{"contents: write on the job", jobWriteWorkflow},
		{"a push to another branch", branchWorkflow},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFixture(t, "3")
			readDefault(f)
			write(t, f.root, workflowsDir+"/record.yml", c.workflow)
			if found := f.findings(); len(found) != 0 {
				t.Errorf("findings = %d, want none:\n%s", len(found), texts(found))
			}
		})
	}
}

// A repository with no workflow file has no recording push, so a
// default of read stops nothing.
func TestAReadDefaultWithNoWorkflowIsNoFinding(t *testing.T) {
	f := newFixture(t, "3")
	readDefault(f)
	if found := f.findings(); len(found) != 0 {
		t.Errorf("findings = %d, want none:\n%s", len(found), texts(found))
	}
}

func TestAPushingWorkflowThatRaisesNothingIsNamed(t *testing.T) {
	f := newFixture(t, "3")
	readDefault(f)
	write(t, f.root, workflowsDir+"/record.yml", silentWorkflow)
	found := f.findings()
	only(t, found, 2, breaks, ".github/workflows/record.yml pushes to main and raises no `contents: write` of its own")
	if breaking(found) != 1 {
		t.Errorf("breaking findings = %d, want 1:\n%s", breaking(found), texts(found))
	}
}

// Only a `permissions:` block grants the right. `contents: write`
// written under another key says nothing about what the workflow may
// do, and the block ends where the indentation returns to the key's
// own.
func TestContentsWriteOutsideAPermissionsBlockGrantsNothing(t *testing.T) {
	f := newFixture(t, "3")
	readDefault(f)
	write(t, f.root, workflowsDir+"/record.yml", strayWriteWorkflow)
	only(t, f.findings(), 2, breaks, ".github/workflows/record.yml pushes to main")
}

// An empty bypass list denies nothing where the ruleset enables no rule
// a fast-forward push meets — the finding report-0013 recorded against
// a repository whose recording push lands.
func TestAnEmptyBypassListWithNothingToBypassIsNoFinding(t *testing.T) {
	f := newFixture(t, "3")
	f.forge.replies["api repos/{owner}/{repo}/rulesets/42 --jq (.bypass_actors // [])[].actor_type"] = "\n"
	if found := f.findings(); len(found) != 0 {
		t.Errorf("findings = %d, want none:\n%s", len(found), texts(found))
	}
}

// One fault is one finding. The rule and the bypass list that would
// clear it belong to the same ruleset, so naming the ruleset says what a
// second finding used to add — and the two used to contradict each
// other's remedy (spec-0024).
func TestARuleThatRefusesThePushWithNoBypassActorNamesTheRule(t *testing.T) {
	f := newFixture(t, "3")
	rulesOnMain(f, "deletion@42", "update@42")
	bypass(f, "42")
	found := f.findings()
	only(t, found, 2, breaks, "ruleset 42 governs main and enables update (restrict updates)")
	if breaking(found) != 1 {
		t.Errorf("breaking findings = %d, want 1:\n%s", breaking(found), texts(found))
	}
}

// A rule one ruleset enables is not the other ruleset's: the forge
// answers the rules on main as one array, and the entry's ruleset_id
// says whose bypass list would let the bot past it.
func TestTheBypassFindingNamesOnlyTheRulesetThatEnablesTheRule(t *testing.T) {
	f := newFixture(t, "3")
	rulesOnMain(f, "deletion@42", "update@43")
	bypass(f, "42")
	bypass(f, "43", "Integration")
	found := f.findings()
	for _, got := range found {
		if strings.Contains(said(got), "ruleset 42 governs main") {
			t.Errorf("ruleset 42 was named for a rule ruleset 43 enables:\n%s", texts(found))
		}
	}
	only(t, found, 2, breaks, "ruleset 43 governs main and enables update (restrict updates)")
}

// Two rulesets govern main, one of them bypassed: the finding names the
// one that stops the push and leaves the other alone (spec-0024, edge
// cases).
func TestTheFindingNamesTheRulesetThatStopsThePush(t *testing.T) {
	f := newFixture(t, "3")
	rulesOnMain(f, "update@42", "required_signatures@43")
	bypass(f, "42", "Integration")
	bypass(f, "43")
	ownedBy(f, "Organization")
	found := f.findings()
	only(t, found, 2, breaks, "ruleset 43 governs main, enables required_signatures (require signed commits) and names no bypass actor")
	if breaking(found) != 1 {
		t.Errorf("breaking findings = %d, want only the ruleset that stops the push:\n%s", breaking(found), texts(found))
	}
}

// Ownership and the ruleset's own bypass list together decide whether
// the bot is past a rule: an organization can put GitHub Actions on the
// list, and the forge offers a person no actor the bot is
// (product/adoption/doctor.md).
func TestABypassActorClearsARuleOnlyOnAnOrganization(t *testing.T) {
	cases := []struct {
		owner  string
		actors []string
		want   string
	}{
		{"Organization", []string{"Integration"}, ""},
		{"Organization", nil, "ruleset 42 governs main, enables update (restrict updates) and names no bypass actor"},
		{"User", []string{"Integration"}, "ruleset 42 governs main and enables update (restrict updates)"},
		{"User", nil, "ruleset 42 governs main and enables update (restrict updates)"},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%s with %d bypass actor(s)", c.owner, len(c.actors)), func(t *testing.T) {
			f := newFixture(t, "3")
			rulesOnMain(f, "update@42")
			bypass(f, "42", c.actors...)
			ownedBy(f, c.owner)
			found := f.findings()
			if c.want == "" {
				if len(found) != 0 {
					t.Errorf("findings = %d, want none where the bot is past the rule:\n%s", len(found), texts(found))
				}
				return
			}
			only(t, found, 2, breaks, c.want)
		})
	}
}

// An owner type the forge will not answer is read as a person's: that is
// the reading under which the finding stands, and a rule that does block
// must not pass in silence (spec-0024, edge cases).
func TestAnUnreadableOwnerTypeStillNamesTheRule(t *testing.T) {
	f := newFixture(t, "3")
	rulesOnMain(f, "update@42")
	bypass(f, "42", "Integration")
	f.forge.fails["api repos/{owner}/{repo} --jq .owner.type"] = errors.New("gh api: HTTP 403")
	only(t, f.findings(), 2, breaks, "ruleset 42 governs main and enables update (restrict updates)")
}

// Who owns the repository is one answer for the whole run, however many
// rulesets ask for it.
func TestOwnershipIsReadOnceForSeveralRulesets(t *testing.T) {
	f := newFixture(t, "3")
	rulesOnMain(f, "update@42", "required_signatures@43")
	bypass(f, "42")
	bypass(f, "43")
	f.findings()
	asked := 0
	for _, c := range f.forge.calls {
		if c == "api repos/{owner}/{repo} --jq .owner.type" {
			asked++
		}
	}
	if asked != 1 {
		t.Errorf("ownership was read %d times, want 1: %v", asked, f.forge.calls)
	}
}

// The shape report-0013 recorded against this repository: workflow
// permissions of read with every pushing workflow raising
// `contents: write`, and a protect-main ruleset with no bypass actor
// whose four rules a fast-forward push meets. It reported two findings
// and must report none.
func TestThisRepositoryHasNoStageTwoFinding(t *testing.T) {
	f := newFixture(t, "3")
	readDefault(f)
	f.forge.replies["api repos/{owner}/{repo}/rules/branches/main --jq .[].type"] = "deletion\nnon_fast_forward\ncreation\nrequired_linear_history\n"
	f.forge.replies["api repos/{owner}/{repo}/rules/branches/main --jq .[].ruleset_id"] = "22247734\n22247734\n22247734\n22247734\n"
	f.forge.replies["api repos/{owner}/{repo}/rulesets/22247734 --jq (.bypass_actors // [])[].actor_type"] = "\n"

	found, reachable := stage2(repoRoot(t), f.deps())
	if !reachable {
		t.Fatal("the forge was reported unreachable")
	}
	if got := counted(found, met); got != len(forgeChecks) {
		t.Errorf("met = %d of %d against this repository:\n%s", got, len(forgeChecks), texts(found))
	}
}

func TestIssuesDisabledIsNamedAtStageThree(t *testing.T) {
	f := newFixture(t, "3")
	f.forge.replies["api repos/{owner}/{repo} --jq .has_issues"] = "false\n"
	only(t, f.findings(), 3, breaks, "Issues are disabled")
}

// fourRules is every rule that refuses the recording push, with the
// words a finding names it in.
var fourRules = []struct{ rule, names string }{
	{"update", "restrict updates"},
	{"required_signatures", "require signed commits"},
	{"required_status_checks", "require status checks to pass"},
	{"pull_request", "require a pull request before merging"},
}

// A bypass actor on a user-owned repository is never the Actions bot, so
// all four rules stand however the list is filled.
func TestTheFourBlockingRulesAreNamedOnAUserOwnedRepository(t *testing.T) {
	for _, c := range fourRules {
		t.Run(c.rule, func(t *testing.T) {
			f := newFixture(t, "3")
			rulesOnMain(f, "deletion@42", c.rule+"@42")
			bypass(f, "42", "Integration")
			only(t, f.findings(), 2, breaks,
				fmt.Sprintf("ruleset 42 governs main and enables %s (%s)", c.rule, c.names))
		})
	}
}

// On an organization the forge offers GitHub Actions as a bypass actor,
// so a ruleset naming none stops the push — `pull_request` included,
// which used to be dropped there and reported all clear (spec-0024).
func TestTheFourBlockingRulesAreNamedOnAnOrganizationWithNoBypassActor(t *testing.T) {
	for _, c := range fourRules {
		t.Run(c.rule, func(t *testing.T) {
			f := newFixture(t, "3")
			rulesOnMain(f, "deletion@42", c.rule+"@42")
			bypass(f, "42")
			ownedBy(f, "Organization")
			only(t, f.findings(), 2, breaks,
				fmt.Sprintf("ruleset 42 governs main, enables %s (%s) and names no bypass actor", c.rule, c.names))
		})
	}
}

func TestOwnershipIsAskedOnlyWhereItChangesAFinding(t *testing.T) {
	f := newFixture(t, "3")
	f.findings()
	if f.forge.asked("api repos/{owner}/{repo} --jq .owner.type") {
		t.Errorf("ownership was read with no pull-request rule on: %v", f.forge.calls)
	}
}

func TestAnUnprotectedMainIsARecommendation(t *testing.T) {
	f := newFixture(t, "3")
	f.forge.replies["api repos/{owner}/{repo}/rules/branches/main --jq .[].type"] = "\n"
	f.forge.replies["api repos/{owner}/{repo}/rules/branches/main --jq .[].ruleset_id"] = "\n"
	found := f.findings()
	only(t, found, 2, advises, "main is governed by a protection rule — the methodology recommends protecting it")
	if breaking(found) != 0 {
		t.Errorf("a recommendation broke a flow:\n%s", texts(found))
	}
}

// Several rules from one ruleset are one ruleset to ask about.
func TestARulesetGoverningMainIsAskedAboutOnce(t *testing.T) {
	f := newFixture(t, "3")
	f.forge.replies["api repos/{owner}/{repo}/rules/branches/main --jq .[].ruleset_id"] = "42\n42\n42\n"
	f.findings()
	asked := 0
	for _, c := range f.forge.calls {
		if c == "api repos/{owner}/{repo}/rulesets/42 --jq (.bypass_actors // [])[].actor_type" {
			asked++
		}
	}
	if asked != 1 {
		t.Errorf("the bypass list was read %d times, want 1: %v", asked, f.forge.calls)
	}
}

func TestAnUnusableGhReportsWhatItCouldNotCheck(t *testing.T) {
	cases := []struct {
		name   string
		set    func(*fixture)
		expect string
		// unread is how many requirements the stand-down leaves unread:
		// the forge rows from the one that failed downwards, and Issues.
		unread int
	}{
		{"gh absent", func(f *fixture) { f.path["gh"] = false }, "gh on the PATH — not on the PATH", 7},
		{"gh unauthenticated", func(f *fixture) {
			f.forge.fails["auth status"] = errors.New("gh auth status: not logged in")
		}, "gh authenticated — run `gh auth login`", 6},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFixture(t, "3")
			c.set(f)
			found := f.findings()
			only(t, found, 2, unread, c.expect)
			only(t, found, 3, unread, "Issues are enabled — the forge did not answer")
			if breaking(found) != 0 {
				t.Errorf("an unreachable forge failed a check:\n%s", texts(found))
			}
			// Every requirement from the one that failed downwards is a
			// row a reader can count, not a silence (spec-0036).
			if len(found) != c.unread {
				t.Errorf("unread requirements = %d, want %d:\n%s", len(found), c.unread, texts(found))
			}
		})
	}
}

func TestAReadThatFailsIsNotAFailedCheck(t *testing.T) {
	cases := []struct{ name, key, expect string }{
		{"squash merging",
			"api repos/{owner}/{repo} --jq .allow_squash_merge",
			"squash merging is on — it could not be read"},
		{"workflow permissions",
			"api repos/{owner}/{repo}/actions/permissions/workflow --jq .default_workflow_permissions",
			"the Actions workflow permissions could not be read"},
		{"the rules on main",
			"api repos/{owner}/{repo}/rules/branches/main --jq .[].type",
			"the rules governing main could not be read"},
		{"the rulesets on main",
			"api repos/{owner}/{repo}/rules/branches/main --jq .[].ruleset_id",
			"the rulesets governing main could not be read"},
		{"a bypass list",
			"api repos/{owner}/{repo}/rulesets/42 --jq (.bypass_actors // [])[].actor_type",
			"no rule over main refuses the recording push — the bypass list of ruleset 42 could not be read"},
		{"Issues",
			"api repos/{owner}/{repo} --jq .has_issues",
			"Issues are enabled — it could not be read"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFixture(t, "3")
			f.forge.fails[c.key] = errors.New("gh api: HTTP 403\nyou lack permission")
			found := f.findings()
			only(t, found, stageOf(found, c.expect), unread, c.expect)
			if breaking(found) != 0 {
				t.Errorf("an unread check failed the run:\n%s", texts(found))
			}
		})
	}
}

// A gh error runs to several lines; the finding keeps the first, which
// is the one that names the cause.
func TestAFailedReadKeepsOnlyTheLineThatNamesTheCause(t *testing.T) {
	f := newFixture(t, "3")
	f.forge.fails["api repos/{owner}/{repo} --jq .has_issues"] = errors.New("gh api: HTTP 403\nyou lack permission")
	found := f.findings()
	only(t, found, 3, unread, "Issues are enabled — it could not be read gh api: HTTP 403")
	for _, got := range found {
		if strings.Contains(said(got), "you lack permission") {
			t.Errorf("text = %q; want only the first line of the error", said(got))
		}
	}
}

// stageOf is the stage the finding matching want was reported at — the
// forge reads all sit at stage 2 except the Issues one.
func stageOf(found []requirement, want string) int {
	for _, r := range found {
		if strings.Contains(said(r), want) {
			return r.stage
		}
	}
	return 2
}

// A classic branch protection rule is not a ruleset rule, and
// `rules/branches/main` reports only the latter. Both rows used to be
// inverted at once on a branch one of these governs: ungoverned, and
// nothing refusing the push (report-0048).
func TestAClassicRuleRequiringAPullRequestIsNamedOnEitherOwner(t *testing.T) {
	for _, owner := range []string{"User", "Organization"} {
		t.Run(owner, func(t *testing.T) {
			f := newFixture(t, "3")
			noRulesets(f)
			f.forge.replies["api repos/{owner}/{repo} --jq .owner.type"] = owner + "\n"
			classicOn(f, `{"required_pull_request_reviews":{"required_approving_review_count":0},"enforce_admins":{"enabled":true}}`)
			found := f.findings()
			only(t, found, 2, breaks, "the branch protection rule over main enables pull_request (require a pull request before merging)")
			if breaking(found) != 1 {
				t.Errorf("breaking findings = %d, want 1:\n%s", breaking(found), texts(found))
			}
		})
	}
}

// A classic rule governs the branch, so the recommendation is not owed.
func TestAClassicRuleAloneGovernsMain(t *testing.T) {
	f := newFixture(t, "3")
	noRulesets(f)
	classicOn(f, `{"required_linear_history":{"enabled":true}}`)
	for _, r := range f.all() {
		if r.name == mainGoverned && r.mark != met {
			t.Errorf("a protected main is %s:\n%s", words[r.mark], texts(f.all()))
		}
	}
}

// A plain fast-forward push meets none of these, so a classic rule
// carrying only them stops nothing (spec-0047, acceptance criteria).
func TestAClassicRuleAFastForwardMeetsIsNoFinding(t *testing.T) {
	f := newFixture(t, "3")
	noRulesets(f)
	classicOn(f, `{"required_linear_history":{"enabled":true},"required_conversation_resolution":{"enabled":true},"allow_force_pushes":{"enabled":false},"allow_deletions":{"enabled":false},"enforce_admins":{"enabled":false}}`)
	if found := f.findings(); len(found) != 0 {
		t.Errorf("a rule a fast-forward meets was reported:\n%s", texts(found))
	}
}

// Push restrictions are the classic form of restricting updates, and
// the one field that clears itself: the list names what the recording
// push runs as, or it does not.
func TestPushRestrictionsAreClearedOnlyByTheActionsApp(t *testing.T) {
	cases := []struct {
		name    string
		apps    string
		refused bool
	}{
		{"no app named", `{"restrictions":{"users":[],"teams":[],"apps":[]}}`, true},
		{"another app named", `{"restrictions":{"users":[],"teams":[],"apps":[{"slug":"dependabot"}]}}`, true},
		{"the Actions app named", `{"restrictions":{"users":[],"teams":[],"apps":[{"slug":"github-actions"}]}}`, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFixture(t, "3")
			noRulesets(f)
			classicOn(f, c.apps)
			found := f.findings()
			if c.refused {
				only(t, found, 2, breaks, "the branch protection rule over main enables update (restrict updates)")
				return
			}
			if len(found) != 0 {
				t.Errorf("an allow list naming the Actions app was reported:\n%s", texts(found))
			}
		})
	}
}

// Both mechanisms over one branch: each source that refuses the push
// earns its own line, and the branch is governed once.
func TestARulesetAndAClassicRuleAreBothNamed(t *testing.T) {
	f := newFixture(t, "3")
	rulesOnMain(f, "required_signatures@42")
	bypass(f, "42")
	classicOn(f, `{"required_pull_request_reviews":{}}`)
	found := f.findings()
	only(t, found, 2, breaks, "the branch protection rule over main enables pull_request")
	only(t, found, 2, breaks, "ruleset 42 governs main and enables required_signatures")
	if breaking(found) != 1 {
		t.Errorf("two sources refusing one push are %d findings, want 1 row:\n%s", breaking(found), texts(found))
	}
}

// A branch no classic rule protects costs one read and not two: the
// payload is asked for only where there is a rule to read.
func TestAnUnprotectedBranchIsNotAskedForItsRule(t *testing.T) {
	f := newFixture(t, "3")
	f.findings()
	if f.forge.asked("api repos/{owner}/{repo}/branches/main/protection") {
		t.Errorf("the protection payload was read with no classic rule on: %v", f.forge.calls)
	}
}

// A read that fails is not a check that failed, and here it is one row
// and not the other: what governs main answered, what it enables did
// not.
func TestAnUnreadableRuleLeavesTheRefusalUnreadAndTheBranchGoverned(t *testing.T) {
	f := newFixture(t, "3")
	noRulesets(f)
	classicOn(f, "")
	f.forge.fails["api repos/{owner}/{repo}/branches/main/protection"] = errors.New("gh: HTTP 403 Forbidden")
	found := f.all()
	for _, r := range found {
		switch r.name {
		case mainGoverned:
			if r.mark != met {
				t.Errorf("a protected main is %s:\n%s", words[r.mark], texts(found))
			}
		case noRuleRefuses:
			if r.mark != unread {
				t.Errorf("an unreadable rule is %s, want unread:\n%s", words[r.mark], texts(found))
			}
		}
	}
}

// The branch object is the one read both rows depend on, so a forge
// that will not answer it leaves both unread rather than either met.
func TestAnUnreadableBranchLeavesBothRowsUnread(t *testing.T) {
	f := newFixture(t, "3")
	f.forge.fails["api repos/{owner}/{repo}/branches/main --jq .protection.enabled"] = errors.New("gh: HTTP 500")
	found := f.all()
	for _, r := range found {
		if (r.name == mainGoverned || r.name == noRuleRefuses) && r.mark != unread {
			t.Errorf("%q is %s, want unread:\n%s", r.name, words[r.mark], texts(found))
		}
	}
}

// The row's note advises; the glyph denies. It used to do both at once
// — `main is governed by a ruleset — no ruleset governs it` — which
// asserts and denies one fact before reaching its advice (report-0048).
func TestTheRecommendationDoesNotDenyItsOwnRow(t *testing.T) {
	f := newFixture(t, "3")
	noRulesets(f)
	for _, r := range f.all() {
		if r.name != mainGoverned {
			continue
		}
		if r.mark != advises {
			t.Fatalf("an ungoverned main is %s, want advises", words[r.mark])
		}
		if strings.Contains(r.note, "governs it") || strings.Contains(r.note, "no ruleset") {
			t.Errorf("the note denies the row it sits on: %q", said(r))
		}
	}
}
