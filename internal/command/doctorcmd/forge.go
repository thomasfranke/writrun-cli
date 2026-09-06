package doctorcmd

import (
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
const (
	repoAPI      = "repos/{owner}/{repo}"
	workflowAPI  = "repos/{owner}/{repo}/actions/permissions/workflow"
	mainRulesAPI = "repos/{owner}/{repo}/rules/branches/main"
)

// workflowsDir holds the workflow files whose own `permissions:` block
// is the second way the recording push gets the right to write.
const workflowsDir = ".github/workflows"

// blocker is one ruleset rule that stops the recording push the
// workflows make, and how the finding names it.
type blocker struct {
	rule  string
	names string
}

// blockers are the four rules that refuse the recording push, in the
// order a finding prefers them. Whether the Actions bot is past one is
// the enabling ruleset's bypass list and the repository's owner to
// answer, never the rule's (product/adoption/doctor.md).
var blockers = []blocker{
	{rule: "update", names: "restrict updates"},
	{rule: "required_signatures", names: "require signed commits"},
	{rule: "required_status_checks", names: "require status checks to pass"},
	{rule: "pull_request", names: "require a pull request before merging"},
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
func stage2(root string, d Deps) ([]finding, bool) {
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
	found = append(found, canWrite(root, d)...)
	found = append(found, mainReachable(d)...)
	return found, true
}

// canWrite answers whether the recording push has the right to write,
// which two arrangements grant. A repository default of read-and-write
// grants it to every workflow at once. A default of read grants it to
// each workflow that raises `contents: write` for itself, and leaves it
// off the workflows that never push — the tighter of the two, so it
// passes (spec-0019). A workflow that pushes and raises nothing is the
// one arrangement in which the push cannot land.
func canWrite(root string, d Deps) []finding {
	got, err := d.Gh("api", workflowAPI, "--jq", ".default_workflow_permissions")
	if err != nil {
		return []finding{{stage: 2, level: unread,
			text: "the Actions workflow permissions could not be read: " + firstLine(err.Error())}}
	}
	if strings.TrimSpace(got) == "write" {
		return nil
	}
	var found []finding
	for _, rel := range silentPushers(root, d) {
		found = append(found, finding{stage: 2, level: breaks,
			text: rel + " pushes to main and raises no `contents: write` of its own — the Actions workflow permissions default to read, so that push has no right to write"})
	}
	return found
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
// the remote. A destination holding `$` is a variable this file does
// not resolve, and main is one of the values it takes; a push with no
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
// No ruleset over main leaves the branch unprotected with nothing in the
// bot's way, which is a recommendation. Every ruleset that does govern
// main is judged on its own, because a rule and the bypass list that
// would clear it belong to the same ruleset: the rules it contributes
// say whether it refuses the push, and its own bypass list says whether
// the bot is past them. A plain fast-forward push meets none of
// deletion, creation, non_fast_forward or required_linear_history, so a
// ruleset enabling only those is no finding whatever its bypass list
// holds (spec-0024).
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

	owner := &ownership{d: d}
	var found []finding
	for _, id := range distinct(ids) {
		actors, err := lines(d, "repos/{owner}/{repo}/rulesets/"+id, "(.bypass_actors // [])[].actor_type")
		if err != nil {
			found = append(found, finding{stage: 2, level: unread,
				text: fmt.Sprintf("the bypass list of ruleset %s could not be read: %s", id, firstLine(err.Error()))})
			continue
		}
		b, refuses := firstOf(rulesOf(types, ids, id))
		if !refuses {
			continue
		}
		// A bypass actor clears the rule only where the forge offers
		// one the Actions token resolves to. Which of the actors it
		// resolves to stays the forge's answer, and deciding it here
		// would be a second authority on it.
		if len(actors) > 0 && !owner.userOwned() {
			continue
		}
		found = append(found, finding{stage: 2, level: breaks,
			text: refusal(id, b, owner.userOwned())})
	}
	return found
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
// in the order blockers states them — the one a finding names where
// several are on at once.
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
// changes a finding, so a repository no ruleset refuses costs no read.
type ownership struct {
	d      Deps
	asked  bool
	byUser bool
}

// userOwned reports whether the repository belongs to a person rather
// than an organization. An owner type the forge will not answer is read
// as a person's: that is the reading under which the finding stands, and
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
