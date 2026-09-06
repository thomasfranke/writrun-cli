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
		if strings.Contains(got.text, "ruleset 42 governs main") {
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
	if len(found) != 0 {
		t.Errorf("findings = %d, want none against this repository:\n%s", len(found), texts(found))
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
	only(t, found, 2, advises, "main is governed by no ruleset")
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
	}{
		{"gh absent", func(f *fixture) { f.path["gh"] = false }, "gh is not on the PATH"},
		{"gh unauthenticated", func(f *fixture) {
			f.forge.fails["auth status"] = errors.New("gh auth status: not logged in")
		}, "gh is not authenticated"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFixture(t, "3")
			c.set(f)
			found := f.findings()
			only(t, found, 2, unread, c.expect)
			only(t, found, 3, unread, "whether Issues are enabled was not read")
			if breaking(found) != 0 {
				t.Errorf("an unreachable forge failed a check:\n%s", texts(found))
			}
			if len(found) != 2 {
				t.Errorf("findings = %d, want the two stand-downs and no piled-on reads:\n%s", len(found), texts(found))
			}
		})
	}
}

func TestAReadThatFailsIsNotAFailedCheck(t *testing.T) {
	cases := []struct{ name, key, expect string }{
		{"squash merging",
			"api repos/{owner}/{repo} --jq .allow_squash_merge",
			"whether squash merging is on could not be read"},
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
			"the bypass list of ruleset 42 could not be read"},
		{"Issues",
			"api repos/{owner}/{repo} --jq .has_issues",
			"whether Issues are enabled could not be read"},
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
	only(t, found, 3, unread, "whether Issues are enabled could not be read: gh api: HTTP 403")
	for _, got := range found {
		if strings.Contains(got.text, "you lack permission") {
			t.Errorf("text = %q; want only the first line of the error", got.text)
		}
	}
}

// stageOf is the stage the finding matching want was reported at — the
// forge reads all sit at stage 2 except the Issues one.
func stageOf(found []finding, want string) int {
	for _, f := range found {
		if strings.Contains(f.text, want) {
			return f.stage
		}
	}
	return 2
}
