package doctorcmd

import (
	"fmt"
	"strings"
)

// The forge reads doctor makes, written down once. `rules/branches/main`
// is the forge's own answer to which rules govern main — every active
// ruleset's, at whatever level it was configured — so asking it is one
// read where enumerating the rulesets and matching their conditions
// would be several and a second opinion on the forge's own matching.
const (
	repoAPI      = "repos/{owner}/{repo}"
	workflowAPI  = "repos/{owner}/{repo}/actions/permissions/workflow"
	mainRulesAPI = "repos/{owner}/{repo}/rules/branches/main"
)

// blocker is one ruleset rule that stops the recording push the
// workflows make, and how the finding names it.
type blocker struct {
	rule     string
	names    string
	userOnly bool
}

// blockers are the four rules the methodology calls out, named when
// they are on. `pull_request` is named only on a user-owned repository:
// there is no way to let the Actions bot past it, where an organization
// can (product/adoption/doctor.md).
var blockers = []blocker{
	{rule: "update", names: "restrict updates"},
	{rule: "required_signatures", names: "require signed commits"},
	{rule: "required_status_checks", names: "require status checks to pass"},
	{rule: "pull_request", names: "require a pull request before merging", userOnly: true},
}

// stage2 is the forge. It reports whether the settings the recording
// machinery depends on are as the methodology assumes, and says which
// checks it could not make rather than failing them: a forge that does
// not answer has told doctor nothing about the repository (spec-0004,
// acceptance criteria).
//
// An unusable gh is the one finding that stands alone. Every read after
// it would restate the same fault, so the reads stop and the bool says
// the forge never answered, which is also stage 3's answer.
func stage2(d Deps) ([]finding, bool) {
	if _, err := d.LookPath("gh"); err != nil {
		return []finding{{stage: 2, level: unread,
			text: "gh is not on the PATH — from stage 2 the flows ask the forge through it, so no forge check was made"}}, false
	}
	if _, err := d.Gh("auth", "status"); err != nil {
		return []finding{{stage: 2, level: unread,
			text: "gh is not authenticated — run `gh auth login`; no forge check was made"}}, false
	}

	var found []finding
	found = append(found, want(d, "whether squash merging is on", repoAPI, ".allow_squash_merge", "true",
		"squash merging is off — the methodology lands every pull request as one commit")...)
	found = append(found, want(d, "the Actions workflow permissions", workflowAPI, ".default_workflow_permissions", "write",
		"the Actions workflow permissions are read-only — the recording bot needs read-and-write to push to main")...)
	found = append(found, mainReachable(d)...)
	return found, true
}

// stage3 is Issues: the mirror has to have somewhere to land. A forge
// that never answered leaves it unread, not failed.
func stage3(d Deps, reachable bool) []finding {
	if !reachable {
		return []finding{{stage: 3, level: unread,
			text: "whether Issues are enabled was not read — the forge did not answer"}}
	}
	found := want(d, "whether Issues are enabled", repoAPI, ".has_issues", "true",
		"Issues are disabled — the mirror the flows keep needs somewhere to land")
	for i := range found {
		found[i].stage = 3
	}
	return found
}

// want reads one forge setting and compares it with what the
// methodology expects. A read that fails is a check that could not be
// made; a value that differs is the finding, named with the setting.
func want(d Deps, what, path, jq, expected, breakage string) []finding {
	got, err := d.Gh("api", path, "--jq", jq)
	if err != nil {
		return []finding{{stage: 2, level: unread,
			text: what + " could not be read: " + firstLine(err.Error())}}
	}
	if strings.TrimSpace(got) != expected {
		return []finding{{stage: 2, level: breaks, text: breakage}}
	}
	return nil
}

// mainReachable answers the question the recording push depends on: can
// the Actions bot write to main.
//
// Two of its three answers are recommendations, and the third is not
// here. No ruleset over main leaves the branch unprotected with nothing
// in the bot's way; a ruleset that lists no bypass actor lets a plain
// fast-forward push through all the same, and only leaves the bot
// without a way past a rule that does block it. What actually refuses
// the push is one of the four rules, and named() reports those. A
// ruleset that names bypass actors passes: which of them the forge
// resolves the Actions token to is the forge's answer, and deciding it
// here would be a second authority on it.
func mainReachable(d Deps) []finding {
	types, err := lines(d, mainRulesAPI, ".[].type")
	if err != nil {
		return []finding{{stage: 2, level: unread,
			text: "the rules governing main could not be read: " + firstLine(err.Error())}}
	}
	ids, err := lines(d, mainRulesAPI, ".[].ruleset_id")
	if err != nil {
		return []finding{{stage: 2, level: unread,
			text: "the rulesets governing main could not be read: " + firstLine(err.Error())}}
	}
	if len(ids) == 0 {
		return []finding{{stage: 2, level: advises,
			text: "main is governed by no ruleset — the methodology recommends protecting it; nothing blocks the recording push meanwhile"}}
	}

	var found []finding
	for _, id := range distinct(ids) {
		actors, err := lines(d, "repos/{owner}/{repo}/rulesets/"+id, "(.bypass_actors // [])[].actor_type")
		switch {
		case err != nil:
			found = append(found, finding{stage: 2, level: unread,
				text: fmt.Sprintf("the bypass list of ruleset %s could not be read: %s", id, firstLine(err.Error()))})
		case len(actors) == 0:
			found = append(found, finding{stage: 2, level: advises,
				text: fmt.Sprintf("ruleset %s governs main and names no bypass actor — the Actions bot has no way past a rule that blocks the recording push; put it on the ruleset's bypass list", id)})
		}
	}
	found = append(found, named(d, types)...)
	return found
}

// named reports the four rules that block the recording push, one
// finding each, where they are on for main.
func named(d Deps, types []string) []finding {
	var found []finding
	for _, b := range blockers {
		if !contains(types, b.rule) {
			continue
		}
		if b.userOnly && !userOwned(d) {
			continue
		}
		found = append(found, finding{stage: 2, level: breaks,
			text: fmt.Sprintf("the rule %s (%s) is on for main — it refuses the recording push the workflows make", b.rule, b.names)})
	}
	return found
}

// userOwned reports whether the repository belongs to a person rather
// than an organization. It is asked only where the answer changes a
// finding, so a repository with no pull-request rule costs no read.
func userOwned(d Deps) bool {
	out, err := d.Gh("api", repoAPI, "--jq", ".owner.type")
	if err != nil {
		// Unknown ownership is read as user-owned: naming a rule that
		// may not block is a finding a reader can dismiss, where
		// staying silent about one that does is a broken flow nobody
		// was told about.
		return true
	}
	return strings.TrimSpace(out) == "User"
}

// lines reads one forge query as the list of values it printed, blanks
// dropped — the shape every `--jq` iteration answers in.
func lines(d Deps, path, jq string) ([]string, error) {
	out, err := d.Gh("api", path, "--jq", jq)
	if err != nil {
		return nil, err
	}
	var values []string
	for _, line := range strings.Split(out, "\n") {
		if v := strings.TrimSpace(line); v != "" {
			values = append(values, v)
		}
	}
	return values, nil
}

// distinct keeps the first appearance of each value, in order: one
// ruleset governing main through several rules is one ruleset to ask
// about.
func distinct(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, v := range values {
		if seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}

func contains(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}
