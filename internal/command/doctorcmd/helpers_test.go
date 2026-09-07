package doctorcmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// agentsDoc is an AGENTS.md as the kit grafts it: the project's own
// prose, and WritRun's section pointing at the flow.
const agentsDoc = `# AGENTS.md — entry point for AI agents

A project.

## WritRun

This project tracks its work with WritRun. Before touching ` + "`work/`" + `,
read and follow [` + "`.writrun/AGENTS.md`" + `](.writrun/AGENTS.md).
`

// legacyAgents is what a v0.0.03 adoption left: the fenced section,
// which v0.0.04 duplicates in `.writrun/AGENTS.md`.
const legacyAgents = `# AGENTS.md — entry point for AI agents

A project.

## WritRun — working the queue

<!-- writrun:begin
     This section is WritRun's flow. -->

### Picking work

The flow's text.

<!-- writrun:end -->
`

// gatesDoc is `.writrun/gates.md` as a project answers it. A case that
// wants one unanswered rewrites the row it is about.
const gatesDoc = `# Human gates — per principle 7

| Transition | Who |
|---|---|
| Writing or changing anything under ` + "`docs/`" + ` | Thomas reviews before merge. |
| An authored rule is finished, so derivation may start | Thomas declares it. |
| Spec ` + "`draft → approved`" + ` | Thomas only, via the merged PR. |
| Task with empty ` + "`spec_ref`" + ` and insufficient brief | Stop and ask for a spec. |
| Derived work, before the PR opens | Present it in the session. |
| Changing repository/forge settings (Actions permissions, rulesets) | Thomas assents in session. |
| A report becomes a task (` + "`tracked`" + `) | The agent derives; the merge assents. |
| Everything else | Agent, autonomously. |
`

// The workflow files a case writes into `.github/workflows`, read only
// where the repository default is `read` — the arrangement in which
// each workflow's own `permissions:` block decides whether its push has
// the right to write.
const (
	// recordingWorkflow pushes to the branch a variable names, which
	// main is one value of, and raises the right at workflow level.
	recordingWorkflow = `name: record
on: [pull_request]
permissions:
  contents: write
  pull-requests: read
jobs:
  record:
    runs-on: ubuntu-latest
    steps:
      - run: git push origin "HEAD:${BASE_REF}"
`
	// writeAllWorkflow says the same grant in one word.
	writeAllWorkflow = `name: record
on: [pull_request]
permissions: write-all
jobs:
  record:
    runs-on: ubuntu-latest
    steps:
      - run: git push origin main
`
	// jobWriteWorkflow raises the right on the job rather than the
	// workflow, which grants it just the same.
	jobWriteWorkflow = `name: record
on: [pull_request]
jobs:
  record:
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - run: git push origin main
`
	// branchWorkflow pushes to a branch that is not main, so what it
	// may write is not stage 2's business.
	branchWorkflow = `name: pages
on: [push]
jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - run: git push origin HEAD:gh-pages
`
	// silentWorkflow pushes to main and raises nothing — the one
	// arrangement a repository default of read leaves unable to record.
	silentWorkflow = `name: record
on: [pull_request]
permissions:
  contents: read
jobs:
  record:
    runs-on: ubuntu-latest
    steps:
      - run: git push origin "HEAD:${BASE_REF}"
`
	// strayWriteWorkflow writes `contents: write` where no
	// `permissions:` key governs it, which grants nothing.
	strayWriteWorkflow = `name: record
on: [pull_request]
env:
  contents: write
jobs:
  record:
    runs-on: ubuntu-latest
    steps:
      - run: git push origin main
`
)

// readDefault sets the repository's Actions default to read — the
// tighter arrangement, in which the workflow files answer the question.
func readDefault(f *fixture) {
	f.forge.replies["api repos/{owner}/{repo}/actions/permissions/workflow --jq .default_workflow_permissions"] = "read\n"
}

// rulesOnMain sets what the forge answers about the rules governing
// main: one `rule@ruleset` argument per entry, in the order the forge
// lists them. The two reads are one array's two columns, so a case
// states them together or the columns drift apart.
func rulesOnMain(f *fixture, entries ...string) {
	var types, ids []string
	for _, entry := range entries {
		rule, id, _ := strings.Cut(entry, "@")
		types = append(types, rule)
		ids = append(ids, id)
	}
	f.forge.replies["api repos/{owner}/{repo}/rules/branches/main --jq .[].type"] = strings.Join(types, "\n") + "\n"
	f.forge.replies["api repos/{owner}/{repo}/rules/branches/main --jq .[].ruleset_id"] = strings.Join(ids, "\n") + "\n"
}

// bypass sets one ruleset's bypass list. No actor argument is a ruleset
// that names none.
func bypass(f *fixture, id string, actors ...string) {
	f.forge.replies["api repos/{owner}/{repo}/rulesets/"+id+" --jq (.bypass_actors // [])[].actor_type"] = strings.Join(actors, "\n") + "\n"
}

// ownedBy sets the repository's owner type, which is "User" or
// "Organization" — the two answers the forge gives.
func ownedBy(f *fixture, kind string) {
	f.forge.replies["api repos/{owner}/{repo} --jq .owner.type"] = kind + "\n"
}

// repoRoot is this repository, from the directory the package's tests
// run in: the live `.github/workflows` the report-0013 regression is
// about.
func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("the repository root could not be resolved: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(workflowsDir))); err != nil {
		t.Fatalf("%s holds no %s: %v", root, workflowsDir, err)
	}
	return root
}

// healthyForge is every forge read answered the way the methodology
// assumes: squash on, workflow permissions read-and-write, main
// governed by a ruleset that names a bypass actor and none of the four
// blocking rules, Issues on.
func healthyForge() map[string]string {
	return map[string]string{
		"auth status": "",
		"api repos/{owner}/{repo} --jq .allow_squash_merge":                                        "true\n",
		"api repos/{owner}/{repo} --jq .has_issues":                                                "true\n",
		"api repos/{owner}/{repo} --jq .owner.type":                                                "User\n",
		"api repos/{owner}/{repo}/actions/permissions/workflow --jq .default_workflow_permissions": "write\n",
		"api repos/{owner}/{repo}/rules/branches/main --jq .[].type":                               "deletion\nnon_fast_forward\n",
		"api repos/{owner}/{repo}/rules/branches/main --jq .[].ruleset_id":                         "42\n42\n",
		"api repos/{owner}/{repo}/rulesets/42 --jq (.bypass_actors // [])[].actor_type":            "Integration\n",
	}
}

// forge is the gh port faked from a script of canned replies, keyed on
// the whole invocation, with a record of every call made — which is how
// a case proves a read was never attempted.
type forge struct {
	replies map[string]string
	fails   map[string]error
	calls   []string
}

func (f *forge) run(args ...string) (string, error) {
	key := strings.Join(args, " ")
	f.calls = append(f.calls, key)
	if err, there := f.fails[key]; there {
		return "", err
	}
	out, there := f.replies[key]
	if !there {
		return "", fmt.Errorf("gh %s: no canned reply", key)
	}
	return out, nil
}

func (f *forge) asked(key string) bool {
	for _, c := range f.calls {
		if c == key {
			return true
		}
	}
	return false
}

// scripts is the exec port faked: one canned verdict and one canned
// reporting per script, and a record of what was run.
type scripts struct {
	said    map[string]string
	verdict map[string]error
	ran     []string
}

func (s *scripts) run(_ string, stdout, _ io.Writer, _ []string, name string, args ...string) error {
	s.ran = append(s.ran, strings.TrimSpace(name+" "+strings.Join(args, " ")))
	fmt.Fprint(stdout, s.said[name])
	return s.verdict[name]
}

// exitErr is a script's own verdict as the runner hands it up.
type exitErr int

func (e exitErr) Error() string { return fmt.Sprintf("exit status %d", int(e)) }
func (e exitErr) ExitCode() int { return int(e) }

// fixture is a repository every stage-1 check passes on, plus the
// scripts and the forge answering as the methodology assumes. A case
// breaks exactly the one thing it is about.
type fixture struct {
	root    string
	scripts *scripts
	forge   *forge
	path    map[string]bool
}

func newFixture(t *testing.T, stage string) *fixture {
	t.Helper()
	root := t.TempDir()
	write(t, root, "docs/about.md", "# About\n")
	write(t, root, "docs/product/README.md", "# Product\n")
	write(t, root, "docs/product/rules.md", "# Rules\n")
	write(t, root, "docs/technical/README.md", "# Technical\n")
	write(t, root, "docs/technical/boundaries.md", "# Boundaries\n")
	write(t, root, "work/tasks/README.md", "# Tasks\n")
	write(t, root, "work/specs/README.md", "# Specs\n")
	write(t, root, "work/reports/README.md", "# Reports\n")
	write(t, root, "AGENTS.md", agentsDoc)
	write(t, root, ".writrun/gates.md", gatesDoc)
	write(t, root, ".writrun/VERSION", "v0.0.03\n")

	return &fixture{
		root: root,
		scripts: &scripts{
			said:    map[string]string{settingsReader: stage + "\n"},
			verdict: map[string]error{},
		},
		forge: &forge{replies: healthyForge(), fails: map[string]error{}},
		path:  map[string]bool{"git": true, "bash": true, "awk": true, "sed": true, "gh": true},
	}
}

func (f *fixture) deps() Deps {
	return Deps{
		Scripts: f.scripts.run,
		Gh:      f.forge.run,
		Files:   vfs.OS{},
		LookPath: func(name string) (string, error) {
			if !f.path[name] {
				return "", errors.New("executable file not found in $PATH")
			}
			return "/usr/bin/" + name, nil
		},
	}
}

// findings is the whole examination this fixture produces, the declared
// stage read the way the command reads it.
func (f *fixture) findings() []finding {
	stage, unreadable := declaredStage(f.root, f.deps())
	return examine(f.root, stage, f.deps(), unreadable)
}

// texts is every finding's sentence, joined — what a case greps.
func texts(found []finding) string {
	var b strings.Builder
	for _, f := range found {
		fmt.Fprintf(&b, "%d %s %s\n", f.stage, labels[f.level], f.text)
	}
	return b.String()
}

// only fails unless exactly one finding matches want, at the level and
// stage named — the shape every failing fixture asserts.
func only(t *testing.T, found []finding, stage int, lvl level, want string) {
	t.Helper()
	hits := 0
	for _, f := range found {
		if strings.Contains(f.text, want) {
			hits++
			if f.stage != stage {
				t.Errorf("stage = %d, want %d for %q", f.stage, stage, want)
			}
			if f.level != lvl {
				t.Errorf("level = %s, want %s for %q", labels[f.level], labels[lvl], want)
			}
		}
	}
	if hits != 1 {
		t.Errorf("findings matching %q = %d, want 1:\n%s", want, hits, texts(found))
	}
}

func write(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func remove(t *testing.T, root, rel string) {
	t.Helper()
	if err := os.RemoveAll(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
		t.Fatalf("remove %s: %v", rel, err)
	}
}
