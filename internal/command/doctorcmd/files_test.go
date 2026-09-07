package doctorcmd

import (
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/requirements"
)

func TestEveryAssumptionHoldingFindsNothing(t *testing.T) {
	f := newFixture(t, "3")
	if found := f.findings(); len(found) != 0 {
		t.Errorf("findings = %d, want none:\n%s", len(found), texts(found))
	}
}

func TestAMissingRequirementIsNamed(t *testing.T) {
	for _, bin := range requirements.All() {
		t.Run(bin, func(t *testing.T) {
			f := newFixture(t, "1")
			f.path[bin] = false
			only(t, f.findings(), 0, breaks, bin+" is not on the PATH")
		})
	}
}

func TestEveryRequirementMissingNamesEveryOne(t *testing.T) {
	f := newFixture(t, "1")
	want := requirements.All()
	for _, bin := range want {
		f.path[bin] = false
	}
	found := at(f.findings(), 0)
	if len(found) != len(want) {
		t.Fatalf("stage 0 findings = %d, want %d:\n%s", len(found), len(want), texts(found))
	}
}

func TestTheFilesTheMethodologyRequiresAreNamedWhenAbsent(t *testing.T) {
	cases := []struct {
		name   string
		gone   string
		expect string
	}{
		{"the About file", "docs/about.md", "docs/about.md — an About file is required"},
		{"a real product chapter", "docs/product/rules.md", "docs/product/ — at least one real product doc"},
		{"a technical doc", "docs/technical/boundaries.md", "docs/technical/ — at least one real technical doc"},
		{"the tasks folder", "work/tasks", "work/tasks/ — the docs/ and work/ split"},
		{"the specs folder", "work/specs", "work/specs/ — the docs/ and work/ split"},
		{"the reports folder", "work/reports", "work/reports/ — the docs/ and work/ split"},
		{"AGENTS.md", "AGENTS.md", "AGENTS.md — the agents' entry point is missing"},
		{"the recorded tag", ".writrun/VERSION", ".writrun/VERSION — the kit's tag is not recorded"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFixture(t, "1")
			remove(t, f.root, c.gone)
			only(t, f.findings(), 1, breaks, c.expect)
		})
	}
}

func TestAReadmeAloneIsNotAChapter(t *testing.T) {
	f := newFixture(t, "1")
	remove(t, f.root, "docs/product/rules.md")
	only(t, f.findings(), 1, breaks, "docs/product/ — at least one real product doc")
}

// TestALegacyFencedSectionAdvises inverts TestDamagedMarkersAreNamed.
// The fence was what a refresh rewrote, so damaged markers broke a
// flow; from v0.0.04 the flow lives in `.writrun/AGENTS.md` and a
// leftover section is a stale duplicate, which advises.
func TestALegacyFencedSectionAdvises(t *testing.T) {
	f := newFixture(t, "1")
	write(t, f.root, "AGENTS.md", legacyAgents)
	only(t, f.findings(), 1, advises, "a writrun:begin/writrun:end section is still there")
}

func TestAnAgentsFileWithNoWritRunSectionIsNotAFinding(t *testing.T) {
	// AGENTS.md is the project's whole from v0.0.04 on. What it must
	// say is `init`'s question at adoption, not doctor's forever.
	f := newFixture(t, "1")
	write(t, f.root, "AGENTS.md", "# AGENTS.md\n\nA project.\n")
	if found := f.findings(); len(found) != 0 {
		t.Errorf("findings = %d, want none:\n%s", len(found), texts(found))
	}
}

func TestAMissingAgentsFileIsNamed(t *testing.T) {
	f := newFixture(t, "1")
	remove(t, f.root, "AGENTS.md")
	only(t, f.findings(), 1, breaks, "the agents' entry point is missing")
}

func TestAnUnansweredGateIsNamed(t *testing.T) {
	cases := []struct{ name, was, expect string }{
		{"docs changes", "Thomas reviews before merge.", "Writing or changing anything under docs/"},
		{"a rule declared finished", "Thomas declares it.", "An authored rule is finished, so derivation may start"},
		{"spec approval", "Thomas only, via the merged PR.", "Spec draft → approved"},
		{"a task with no spec", "Stop and ask for a spec.", "Task with empty spec_ref and insufficient brief"},
		{"derived work", "Present it in the session.", "Derived work, before the PR opens"},
		{"a report tracked", "The agent derives; the merge assents.", "A report becomes a task (tracked)"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFixture(t, "1")
			write(t, f.root, ".writrun/gates.md",
				strings.Replace(gatesDoc, c.was, "<!-- TODO — default: someone -->", 1))
			only(t, f.findings(), 1, breaks, "the gate for "+c.expect+" is unanswered")
		})
	}
}

// A gate this binary has never seen is judged by the same rule and
// named by the words the file states it in — the whole point of reading
// the rows rather than holding a list of gates
// (docs/technical/engineering/coupling.md, rule 2).
func TestAGateThisBinaryHasNeverSeenIsJudgedTheSameWay(t *testing.T) {
	f := newFixture(t, "1")
	write(t, f.root, ".writrun/gates.md",
		gatesDoc+"| Something no tag has shipped yet | <!-- TODO --> |\n")
	only(t, f.findings(), 1, breaks, "the gate for Something no tag has shipped yet is unanswered")
}

func TestAMissingGatesFileIsNamed(t *testing.T) {
	f := newFixture(t, "1")
	remove(t, f.root, ".writrun/gates.md")
	only(t, f.findings(), 1, breaks, ".writrun/gates.md — the project's gate answers are missing")
}

func TestAGatesFileWithNoTableIsNamed(t *testing.T) {
	f := newFixture(t, "1")
	write(t, f.root, ".writrun/gates.md", "# Human gates\n\nWe decide as we go.\n")
	only(t, f.findings(), 1, breaks, "no table of gates is readable in it")
}

// The gates are read from their own file, so a table anywhere else —
// the skills trigger table this repository's AGENTS.md carries, for
// one — can neither answer a gate nor invent one.
func TestATableInAgentsDoesNotAnswerAGate(t *testing.T) {
	f := newFixture(t, "1")
	write(t, f.root, "AGENTS.md", agentsDoc+
		"\n## Skills\n\n| Trigger | Skill |\n|---|---|\n| Writing markdown under `docs/` | The docs skill. |\n")
	write(t, f.root, ".writrun/gates.md",
		strings.Replace(gatesDoc, "Thomas reviews before merge.", "<!-- TODO -->", 1))
	only(t, f.findings(), 1, breaks, "the gate for Writing or changing anything under docs/ is unanswered")
}

func TestAnUnreadableTagIsNamed(t *testing.T) {
	cases := []struct{ name, tag string }{
		{"empty", "\n"},
		{"no v", "0.0.3\n"},
		{"one component", "v1\n"},
		{"not a number", "v0.0.x\n"},
		{"a branch name", "vmain\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFixture(t, "1")
			write(t, f.root, ".writrun/VERSION", c.tag)
			only(t, f.findings(), 1, breaks, "is not a readable tag")
		})
	}
}

func TestARefusingCheckIsNamedWithItsOwnWords(t *testing.T) {
	cases := []struct{ name, script, expect string }{
		{"front matter", frontMatterScript, "the queue's front matter is not canonical"},
		{"settings", settingsScript, "does not hold the shape the line-based readers can see"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFixture(t, "1")
			f.scripts.verdict[c.script] = exitErr(1)
			f.scripts.said[c.script] = "REJECTED: work/tasks/task-0001.md line 4\n"
			found := f.findings()
			only(t, found, 1, breaks, c.expect)
			for _, got := range found {
				if strings.Contains(got.text, c.expect) && !strings.Contains(got.detail, "REJECTED: work/tasks/task-0001.md line 4") {
					t.Errorf("detail = %q; want the script's own reporting", got.detail)
				}
			}
		})
	}
}

func TestTheDeclaredStageComesFromTheRepositorysOwnReader(t *testing.T) {
	f := newFixture(t, "2")
	stage, unreadable := declaredStage(f.root, f.deps())
	if stage != 2 {
		t.Errorf("stage = %d, want 2", stage)
	}
	if len(unreadable) != 0 {
		t.Errorf("findings = %v, want none", unreadable)
	}
	if len(f.scripts.ran) != 1 || f.scripts.ran[0] != settingsReader+" stage" {
		t.Errorf("ran = %v; want the settings reader asked for the stage", f.scripts.ran)
	}
}

func TestAnUnreadableStageIsNamedAndStandsDownToStageOne(t *testing.T) {
	cases := []struct {
		name   string
		set    func(*fixture)
		expect string
	}{
		{"the reader refuses", func(f *fixture) {
			f.scripts.verdict[settingsReader] = exitErr(3)
		}, "the declared stage could not be read"},
		{"the value is not a stage", func(f *fixture) {
			f.scripts.said[settingsReader] = "four\n"
		}, `the declared stage reads as "four"`},
		{"the value is out of range", func(f *fixture) {
			f.scripts.said[settingsReader] = "9\n"
		}, `the declared stage reads as "9"`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFixture(t, "3")
			c.set(f)
			found := f.findings()
			only(t, found, 1, breaks, c.expect)
			if len(f.forge.calls) != 0 {
				t.Errorf("forge calls = %v; want none once the stage stands down to 1", f.forge.calls)
			}
		})
	}
}
