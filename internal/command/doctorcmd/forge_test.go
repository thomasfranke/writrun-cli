package doctorcmd

import (
	"errors"
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

func TestTheForgeSettingsAreNamedWhenTheyDiffer(t *testing.T) {
	cases := []struct{ name, key, reply, expect string }{
		{"squash merging off",
			"api repos/{owner}/{repo} --jq .allow_squash_merge", "false\n",
			"squash merging is off"},
		{"workflow permissions read-only",
			"api repos/{owner}/{repo}/actions/permissions/workflow --jq .default_workflow_permissions", "read\n",
			"the Actions workflow permissions are read-only"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFixture(t, "3")
			f.forge.replies[c.key] = c.reply
			only(t, f.findings(), 2, breaks, c.expect)
		})
	}
}

func TestIssuesDisabledIsNamedAtStageThree(t *testing.T) {
	f := newFixture(t, "3")
	f.forge.replies["api repos/{owner}/{repo} --jq .has_issues"] = "false\n"
	only(t, f.findings(), 3, breaks, "Issues are disabled")
}

func TestTheFourBlockingRulesAreNamedWhenOn(t *testing.T) {
	cases := []struct{ rule, expect string }{
		{"update", "the rule update (restrict updates) is on for main"},
		{"required_signatures", "the rule required_signatures (require signed commits) is on for main"},
		{"required_status_checks", "the rule required_status_checks (require status checks to pass) is on for main"},
		{"pull_request", "the rule pull_request (require a pull request before merging) is on for main"},
	}
	for _, c := range cases {
		t.Run(c.rule, func(t *testing.T) {
			f := newFixture(t, "3")
			f.forge.replies["api repos/{owner}/{repo}/rules/branches/main --jq .[].type"] = "deletion\n" + c.rule + "\n"
			only(t, f.findings(), 2, breaks, c.expect)
		})
	}
}

// The pull-request rule is the one whose meaning depends on who owns
// the repository: an organization can let the Actions bot past it, a
// person cannot (product/adoption/doctor.md).
func TestThePullRequestRuleIsNamedOnlyOnAUserOwnedRepository(t *testing.T) {
	f := newFixture(t, "3")
	f.forge.replies["api repos/{owner}/{repo}/rules/branches/main --jq .[].type"] = "pull_request\n"
	f.forge.replies["api repos/{owner}/{repo} --jq .owner.type"] = "Organization\n"
	if found := f.findings(); len(found) != 0 {
		t.Errorf("findings = %d, want none on an organization:\n%s", len(found), texts(found))
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

func TestAnEmptyBypassListIsARecommendation(t *testing.T) {
	f := newFixture(t, "3")
	f.forge.replies["api repos/{owner}/{repo}/rulesets/42 --jq (.bypass_actors // [])[].actor_type"] = "\n"
	found := f.findings()
	only(t, found, 2, advises, "ruleset 42 governs main and names no bypass actor")
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
