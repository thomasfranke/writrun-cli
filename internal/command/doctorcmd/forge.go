package doctorcmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// The forge reads doctor makes, written down once. `rules/branches/main`
// is the forge's own answer to which rules govern main — every active
// ruleset's, at whatever level it was configured — so asking it is one
// read where enumerating the rulesets and matching their conditions
// would be several and a second opinion on the forge's own matching.
//
// It answers for rulesets and for nothing else. The forge keeps a second,
// older mechanism over the same branch — a branch protection rule — and
// a branch one of those governs is reported by that endpoint as governed
// by nothing (report-0048). `branches/main` says whether one is on, and
// `branches/main/protection` is the rule itself.
const (
	repoAPI           = "repos/{owner}/{repo}"
	workflowAPI       = "repos/{owner}/{repo}/actions/permissions/workflow"
	mainRulesAPI      = "repos/{owner}/{repo}/rules/branches/main"
	mainBranchAPI     = "repos/{owner}/{repo}/branches/main"
	mainProtectionAPI = "repos/{owner}/{repo}/branches/main/protection"
)

// workflowsDir holds the workflow files whose own `permissions:` block
// is the second way the recording push gets the right to write.
const workflowsDir = ".github/workflows"

// The six requirements stage 2 makes, named once each. A run that never
// reaches the forge still reports all of them, unread: a check that
// could not be made is not a check that failed, and a reader has to be
// told which is which (spec-0036).
const (
	ghOnPath      = "gh on the PATH"
	ghAuthed      = "gh authenticated"
	squashOn      = "squash merging is on"
	pushCanWrite  = "the recording push can write to main"
	mainGoverned  = "main is governed by a protection rule"
	noRuleRefuses = "no rule over main refuses the recording push"
	rulesJudged   = "every rule over main is one this binary judges"
)

// issuesOn is stage 3's one requirement.
const issuesOn = "Issues are enabled"

// blocker is one ruleset rule that stops the recording push the
// workflows make, and how the row names it.
type blocker struct {
	rule  string
	names string
}

// blockers are the rules this binary judges to refuse the recording
// push, in the order a row prefers them. Whether the Actions bot is
// past one is the enabling ruleset's bypass list and the repository's
// owner to answer, never the rule's (product/adoption/doctor.md).
//
// `lock_branch` leads: a read-only branch takes no push at all, and it
// is a branch protection rule's alone — no ruleset rule locks a branch.
var blockers = []blocker{
	{rule: "lock_branch", names: "make the branch read-only"},
	{rule: "update", names: "restrict updates"},
	{rule: "required_signatures", names: "require signed commits"},
	{rule: "required_status_checks", names: "require status checks to pass"},
	{rule: "pull_request", names: "require a pull request before merging"},
}

// metByAFastForward are the rules the recording push satisfies by being
// what it is: one commit appended to main. They are the only rules that
// pass in silence.
//
// The forge's ruleset vocabulary is twenty-five types and grows; this
// binary judges the five above and these four, and says so about the
// rest rather than reporting them as refusing nothing (spec-0048).
var metByAFastForward = []string{"deletion", "creation", "non_fast_forward", "required_linear_history"}

// judged reports whether this binary has an opinion about a rule —
// either that it refuses the recording push, or that the push meets it.
func judged(rule string) bool {
	return blocking(rule) || contains(metByAFastForward, rule)
}

// blocking reports whether the rule is one of the blockers.
func blocking(rule string) bool {
	for _, b := range blockers {
		if b.rule == rule {
			return true
		}
	}
	return false
}

// stage2 is the forge. It reports whether the settings the recording
// machinery depends on are as the methodology assumes, and says which
// checks it could not make rather than failing them: a forge that does
// not answer has told doctor nothing about the repository (spec-0004,
// acceptance criteria).
//
// An unusable gh is the one fault that stands alone. Every read after it
// would restate it, so the reads stop and every requirement below is
// unread — which is also stage 3's answer.
func stage2(root string, d Deps) ([]requirement, bool) {
	if _, err := d.LookPath("gh"); err != nil {
		return silent(ghOnPath, "not on the PATH — from stage 2 the flows ask the forge through it"), false
	}
	if _, err := d.Gh("auth", "status"); err != nil {
		return silent(ghAuthed, "run `gh auth login`"), false
	}

	found := []requirement{{stage: 2, name: ghOnPath}, {stage: 2, name: ghAuthed}}
	found = append(found, want(d, squashOn, repoAPI, ".allow_squash_merge", "true",
		"squash merging is off — the methodology lands every pull request as one commit"))
	found = append(found, canWrite(root, d))
	found = append(found, mainReachable(d)...)
	return found, true
}

// forgeChecks are stage 2's requirements in the order the report lists
// them, which is also the order a forge read would have made them.
var forgeChecks = []string{ghOnPath, ghAuthed, squashOn, pushCanWrite, mainGoverned, noRuleRefuses, rulesJudged}

// silent is the whole of stage 2 when the forge never answered: the
// requirement that says why, and every one below it unread because no
// read was attempted. A stage nobody could examine is six rows a reader
// can count, not a silence.
func silent(cause, why string) []requirement {
	var found []requirement
	past := false
	for _, name := range forgeChecks {
		r := requirement{stage: 2, name: name}
		switch {
		case name == cause:
			past = true
			r.mark, r.note = unread, why
		case past:
			r.mark, r.note = unread, "no forge check was made"
		}
		found = append(found, r)
	}
	return found
}

// canWrite answers whether the recording push has the right to write,
// which two arrangements grant. A repository default of read-and-write
// grants it to every workflow at once. A default of read grants it to
// each workflow that raises `contents: write` for itself, and leaves it
// off the workflows that never push — the tighter of the two, so it
// passes (spec-0019). A workflow that pushes and raises nothing is the
// one arrangement in which the push cannot land.
func canWrite(root string, d Deps) requirement {
	r := requirement{stage: 2, name: pushCanWrite}
	got, err := d.Gh("api", workflowAPI, "--jq", ".default_workflow_permissions")
	if err != nil {
		r.mark = unread
		r.note = "the Actions workflow permissions could not be read"
		r.detail = firstLine(err.Error())
		return r
	}
	if strings.TrimSpace(got) == "write" {
		return r
	}
	silent := silentPushers(root, d)
	if len(silent) == 0 {
		return r
	}
	r.mark = breaks
	var lines []string
	for _, rel := range silent {
		lines = append(lines, rel+" pushes to main and raises no `contents: write` of its own")
	}
	r.detail = strings.Join(lines, "\n")
	return r
}

// silentPushers are the workflow files that push to main without
// granting themselves the right to, named as paths from the repository
// root. A directory that is not there holds no workflow, so it yields
// none: nothing pushes, and nothing is blocked.
func silentPushers(root string, d Deps) []string {
	dir := filepath.Join(root, filepath.FromSlash(workflowsDir))
	var silent []string
	_ = d.Files.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !isYAML(entry.Name()) {
			return nil
		}
		raw, err := d.Files.ReadFile(path)
		if err != nil {
			return nil
		}
		doc := string(raw)
		if pushesToMain(doc) && !raisesContentsWrite(doc) {
			silent = append(silent, workflowsDir+"/"+entry.Name())
		}
		return nil
	})
	return silent
}

func isYAML(name string) bool {
	return strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml")
}

// pushesToMain reports whether a workflow runs a `git push` that main
// can be the destination of. A push naming another branch outright is
// not the recording push, so the permission it needs is not stage 2's
// business (spec-0019, edge cases). Comment lines are skipped: these
// files explain the push they make.
func pushesToMain(doc string) bool {
	for _, line := range strings.Split(doc, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		at := strings.Index(trimmed, "git push")
		if at < 0 {
			continue
		}
		if toMain(trimmed[at+len("git push"):]) {
			return true
		}
	}
	return false
}

// toMain reads a `git push` argument list and reports whether main can
// be where it lands: the refspec's right-hand side, or the ref after
// the remote. A destination holding `$` is a variable this file does not
// resolve, and main is one of the values it takes; a push with no
// destination at all lands wherever the branch tracks.
func toMain(args string) bool {
	var refs []string
	for _, arg := range strings.Fields(args) {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		refs = append(refs, strings.Trim(arg, `"'`))
	}
	if len(refs) < 2 {
		return true
	}
	dest := refs[len(refs)-1]
	if i := strings.LastIndex(dest, ":"); i >= 0 {
		dest = dest[i+1:]
	}
	return strings.Contains(dest, "$") || strings.TrimPrefix(dest, "refs/heads/") == "main"
}

// raisesContentsWrite reports whether a workflow grants itself the
// right to write the repository: `contents: write` inside a
// `permissions:` block, at workflow level or on a job, or the
// `write-all` that says the same in one word. The block ends where the
// indentation returns to the key's own, which is what makes a
// `contents: write` written elsewhere in the file not count.
func raisesContentsWrite(doc string) bool {
	inBlock := false
	indent := 0
	for _, line := range strings.Split(doc, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		at := len(line) - len(strings.TrimLeft(line, " \t"))
		if inBlock && at <= indent {
			inBlock = false
		}
		if rest, is := strings.CutPrefix(trimmed, "permissions:"); is {
			if value(rest) == "write-all" {
				return true
			}
			inBlock, indent = true, at
			continue
		}
		if rest, is := strings.CutPrefix(trimmed, "contents:"); is && inBlock && value(rest) == "write" {
			return true
		}
	}
	return false
}

// value is what a YAML scalar says, without the quoting that says
// nothing.
func value(s string) string {
	return strings.Trim(strings.TrimSpace(s), `"'`)
}

// stage3 is Issues: the mirror has to have somewhere to land. A forge
// that never answered leaves it unread, not failed.
func stage3(d Deps, reachable bool) []requirement {
	if !reachable {
		return []requirement{{stage: 3, name: issuesOn, mark: unread,
			note: "the forge did not answer"}}
	}
	r := want(d, issuesOn, repoAPI, ".has_issues", "true",
		"Issues are disabled — the mirror the flows keep needs somewhere to land")
	r.stage = 3
	return []requirement{r}
}

// want reads one forge setting and compares it with what the
// methodology expects. A read that fails is a check that could not be
// made; a value that differs is the fault, named with the setting.
func want(d Deps, name, path, jq, expected, breakage string) requirement {
	r := requirement{stage: 2, name: name}
	got, err := d.Gh("api", path, "--jq", jq)
	if err != nil {
		r.mark = unread
		r.note = "it could not be read"
		r.detail = firstLine(err.Error())
		return r
	}
	if strings.TrimSpace(got) != expected {
		r.mark = breaks
		r.detail = breakage
	}
	return r
}

// mainReachable answers the two questions the recording push depends on:
// whether main is governed at all, and whether any rule over it refuses
// the push.
//
// Both mechanisms are read. Nothing governing main leaves the branch
// unprotected with nothing in the bot's way, which is a recommendation;
// a rule of either kind answers the first question, and each kind
// answers the second in its own terms.
//
// Every ruleset that governs main is judged on its own, because a rule
// and the bypass list that would clear it belong to the same ruleset:
// the rules it contributes say whether it refuses the push, and its own
// bypass list says whether the bot is past them. A plain fast-forward
// push meets none of deletion, creation, non_fast_forward or
// required_linear_history, so a ruleset enabling only those is no
// finding whatever its bypass list holds (spec-0024).
func mainReachable(d Deps) []requirement {
	governed := requirement{stage: 2, name: mainGoverned}
	refuses := requirement{stage: 2, name: noRuleRefuses}
	judges := requirement{stage: 2, name: rulesJudged}

	types, err := lines(d, mainRulesAPI, ".[].type")
	if err != nil {
		return unread3(governed, refuses, judges, "the rules governing main could not be read")
	}
	ids, err := lines(d, mainRulesAPI, ".[].ruleset_id")
	if err != nil {
		return unread3(governed, refuses, judges, "the rulesets governing main could not be read")
	}
	c, err := classicOver(d)
	if err != nil && !c.on {
		return unread3(governed, refuses, judges, "the branch protection over main could not be read")
	}
	if len(ids) == 0 && !c.on {
		governed.mark = advises
		governed.note = "the methodology recommends protecting it; nothing blocks the recording push meanwhile"
		return []requirement{governed, refuses, judges}
	}

	owner := &ownership{d: d}
	var refusals, unjudged []string
	if err != nil {
		// The branch is governed and the rule's own fields are not
		// readable, which is one row unread and not the other: what
		// governs main answered, what it enables did not.
		refuses.mark = unread
		refuses.note = "the branch protection rule over main could not be read"
		refuses.detail = firstLine(err.Error())
	}
	if b, refused := firstOf(c.rules); refused {
		refusals = append(refusals, classicRefusal(b))
	}
	for _, id := range distinct(ids) {
		for _, rule := range rulesOf(types, ids, id) {
			if !judged(rule) {
				unjudged = append(unjudged, fmt.Sprintf("ruleset %s enables %s — this binary does not judge it", id, rule))
			}
		}
		actors, err := lines(d, "repos/{owner}/{repo}/rulesets/"+id, "(.bypass_actors // [])[].actor_type")
		if err != nil {
			refuses.mark = unread
			refuses.note = fmt.Sprintf("the bypass list of ruleset %s could not be read", id)
			refuses.detail = firstLine(err.Error())
			continue
		}
		b, refused := firstOf(rulesOf(types, ids, id))
		if !refused {
			continue
		}
		// A bypass actor clears the rule only where the forge offers one
		// the Actions token resolves to. Which of the actors it resolves
		// to stays the forge's answer, and deciding it here would be a
		// second authority on it.
		if len(actors) > 0 && !owner.userOwned() {
			continue
		}
		refusals = append(refusals, refusal(id, b, owner.userOwned()))
	}
	if len(refusals) > 0 {
		refuses.mark = breaks
		refuses.note = ""
		refuses.detail = strings.Join(refusals, "\n")
	}
	if len(unjudged) > 0 {
		judges.mark = advises
		judges.detail = strings.Join(unjudged, "\n")
	}
	return []requirement{governed, refuses, judges}
}

// unread3 is all three rules requirements where the one read they share
// did not answer. The reason sits under the last of them, once: it is
// the three's, and printing it three times would read as three faults.
func unread3(governed, refuses, judges requirement, why string) []requirement {
	governed.mark = unread
	refuses.mark = unread
	judges.mark, judges.detail = unread, why
	return []requirement{governed, refuses, judges}
}

// classicRule is the branch protection rule over main: whether one is
// on at all, and which of the four blocking rules it enables.
type classicRule struct {
	on    bool
	rules []string
}

// actionsApp is what the forge calls the app the recording push runs
// as. A classic rule's push restrictions are the one list that clears
// it, and they clear it by naming this.
const actionsApp = "github-actions"

// protection is the part of a branch protection rule doctor reads. The
// payload is asked for whole and read here rather than narrowed by a
// `--jq` the forge would evaluate: which of the four blocking rules a
// classic rule enables is this binary's judgement, and a judgement made
// inside a query string is one no case can reach
// (technical/engineering/coupling.md).
type protection struct {
	PullRequest  *struct{} `json:"required_pull_request_reviews"`
	StatusChecks *struct{} `json:"required_status_checks"`
	Signatures   *struct {
		Enabled bool `json:"enabled"`
	} `json:"required_signatures"`
	Restrictions *restrictions `json:"restrictions"`
	Lock         *struct {
		Enabled bool `json:"enabled"`
	} `json:"lock_branch"`
}

// restrictions is a classic rule's push allow list, which the forge
// keeps in three: users, teams and apps. Only the apps can be what the
// recording push runs as.
type restrictions struct {
	Apps []struct {
		Slug string `json:"slug"`
	} `json:"apps"`
}

// admits reports whether the allow list names an app by slug.
func (r restrictions) admits(slug string) bool {
	for _, a := range r.Apps {
		if a.Slug == slug {
			return true
		}
	}
	return false
}

// blocking is the rules a classic rule enables, named as `blockers`
// names them. A field the payload omits is the rule not carrying it.
// `restrictions` is the classic form of restricting updates, and it is
// the one field that clears itself: a list naming the Actions app lets
// the recording push through.
func (p protection) blocking() []string {
	var rules []string
	if p.Lock != nil && p.Lock.Enabled {
		rules = append(rules, "lock_branch")
	}
	if p.Restrictions != nil && !p.Restrictions.admits(actionsApp) {
		rules = append(rules, "update")
	}
	if p.Signatures != nil && p.Signatures.Enabled {
		rules = append(rules, "required_signatures")
	}
	if p.StatusChecks != nil {
		rules = append(rules, "required_status_checks")
	}
	if p.PullRequest != nil {
		rules = append(rules, "pull_request")
	}
	return rules
}

// classicOver reads that rule in two steps, and asks the second only
// where the first says there is something to read.
//
// The branch object answers whether a classic rule is on. It is asked
// rather than the rule itself because it answers for a branch nothing
// protects too, so the absence of a rule is read from an answer and
// never from a 404. `.protected` cannot carry the question — a branch a
// ruleset alone governs answers it `true` — and `.protection.enabled`
// can.
func classicOver(d Deps) (classicRule, error) {
	on, err := d.Gh("api", mainBranchAPI, "--jq", ".protection.enabled")
	if err != nil {
		return classicRule{}, err
	}
	if strings.TrimSpace(on) != "true" {
		return classicRule{}, nil
	}
	raw, err := d.Gh("api", mainProtectionAPI)
	if err != nil {
		return classicRule{on: true}, err
	}
	var p protection
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		return classicRule{on: true}, err
	}
	return classicRule{on: true, rules: p.blocking()}, nil
}

// classicRefusal is the sentence a branch protection rule that stops the
// recording push earns. A classic rule carries no bypass list the forge
// resolves the Actions token against — `restrictions` is the one that
// admits an app, and a rule reaching this sentence did not name the
// Actions app on it — so the remedy is the same on both owner types and
// the rule itself is what comes off (product/adoption/doctor.md).
func classicRefusal(b blocker) string {
	return fmt.Sprintf("the branch protection rule over main enables %s (%s) — the Actions bot has no way past a classic rule, so take the rule off main", b.rule, b.names)
}

// refusal is the one sentence a ruleset that stops the recording push
// earns: the rule it enables, and what closes the gap. The owner decides
// the remedy — an organization can put the Actions bot on the ruleset's
// bypass list, and the forge offers a person no bypass actor the bot is,
// so the rule itself is all there is to take off
// (product/adoption/doctor.md).
func refusal(id string, b blocker, userOwned bool) string {
	if userOwned {
		return fmt.Sprintf("ruleset %s governs main and enables %s (%s) — the forge offers the Actions bot no bypass actor on a user-owned repository, so take the rule off main", id, b.rule, b.names)
	}
	return fmt.Sprintf("ruleset %s governs main, enables %s (%s) and names no bypass actor — the Actions bot has no way past it; put the bot on the ruleset's bypass list", id, b.rule, b.names)
}

// rulesOf is the rules one ruleset contributes to main. The forge
// answers `rules/branches/main` as one array, so the type at an index
// and the ruleset_id at that index are the same entry's; a length the
// two answers disagree on drops the unpaired tail rather than
// attributing a rule to a ruleset that may not hold it.
func rulesOf(types, ids []string, id string) []string {
	var of []string
	for i, got := range ids {
		if got == id && i < len(types) {
			of = append(of, types[i])
		}
	}
	return of
}

// firstOf is the first rule among rules that refuses the recording push,
// in the order blockers states them — the one a row names where several
// are on at once.
func firstOf(rules []string) (blocker, bool) {
	for _, b := range blockers {
		if contains(rules, b.rule) {
			return b, true
		}
	}
	return blocker{}, false
}

// ownership is who owns the repository, read from the forge at most once
// per run: the answer is the same for every ruleset, and it is asked
// once per ruleset that refuses the push. It is asked only where it
// changes a row, so a repository no ruleset refuses costs no read.
type ownership struct {
	d      Deps
	asked  bool
	byUser bool
}

// userOwned reports whether the repository belongs to a person rather
// than an organization. An owner type the forge will not answer is read
// as a person's: that is the reading under which the refusal stands, and
// staying silent about a rule that does block is a broken flow nobody
// was told about.
func (o *ownership) userOwned() bool {
	if !o.asked {
		o.asked = true
		out, err := o.d.Gh("api", repoAPI, "--jq", ".owner.type")
		o.byUser = err != nil || strings.TrimSpace(out) == "User"
	}
	return o.byUser
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
