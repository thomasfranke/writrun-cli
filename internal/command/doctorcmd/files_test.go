package doctorcmd

import (
	"strings"
	"testing"
)

func TestEveryAssumptionHoldingFindsNothing(t *testing.T) {
	f := newFixture(t, "3")
	if found := f.findings(); len(found) != 0 {
		t.Errorf("findings = %d, want none:\n%s", len(found), texts(found))
	}
}

func TestAMissingRequirementIsNamed(t *testing.T) {
	for _, bin := range requirements {
		t.Run(bin, func(t *testing.T) {
			f := newFixture(t, "1")
			f.path[bin] = false
			only(t, f.findings(), 0, breaks, bin+" is not on the PATH")
		})
	}
}

func TestEveryRequirementMissingNamesEveryOne(t *testing.T) {
	f := newFixture(t, "1")
	for _, bin := range requirements {
		f.path[bin] = false
	}
	found := at(f.findings(), 0)
	if len(found) != len(requirements) {
		t.Fatalf("stage 0 findings = %d, want %d:\n%s", len(found), len(requirements), texts(found))
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

func TestDamagedMarkersAreNamed(t *testing.T) {
	cases := []struct{ name, doc string }{
		{"no markers at all", "# AGENTS.md\n\nA project.\n"},
		{"the closing marker cut", strings.Replace(agentsDoc, "<!-- writrun:end -->", "", 1)},
		{"the markers reversed", "<!-- writrun:end -->\n<!-- writrun:begin\n -->\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFixture(t, "1")
			write(t, f.root, "AGENTS.md", c.doc)
			only(t, f.findings(), 1, breaks, "the fenced writrun:begin/writrun:end markers are damaged")
		})
	}
}

func TestAnUnansweredGateIsNamed(t *testing.T) {
	cases := []struct{ name, was, expect string }{
		{"docs changes", "Thomas reviews before merge.", "who writes or reviews a change under docs/"},
		{"a rule declared finished", "Thomas declares it.", "who declares an authored rule finished"},
		{"spec approval", "Thomas only, via the merged PR.", "who moves a spec from draft to approved"},
		{"a task with no spec", "Stop and ask for a spec.", "who acts on a task carrying no spec"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFixture(t, "1")
			write(t, f.root, "AGENTS.md",
				strings.Replace(agentsDoc, c.was, "<!-- TODO — default: someone -->", 1))
			only(t, f.findings(), 1, breaks, "the gate for "+c.expect+" is still a placeholder")
		})
	}
}

func TestAGateStatedNowhereIsNamed(t *testing.T) {
	f := newFixture(t, "1")
	write(t, f.root, "AGENTS.md",
		strings.Replace(agentsDoc, "| Task with empty `spec_ref` and insufficient brief | Stop and ask for a spec. |\n", "", 1))
	only(t, f.findings(), 1, breaks, "states no row for who acts on a task carrying no spec")
}

// The skills trigger table this repository's own AGENTS.md carries
// mentions `docs/` too. The gates are read from the Human gates section
// where one is headed, so a neighbouring table cannot answer for them.
func TestAnotherTableDoesNotAnswerAGate(t *testing.T) {
	f := newFixture(t, "1")
	doc := strings.Replace(agentsDoc,
		"### Human gates",
		"| Writing or editing markdown under `docs/` | The docs skill. |\n\n### Human gates", 1)
	doc = strings.Replace(doc, "| Writing or changing anything under `docs/` | Thomas reviews before merge. |",
		"| Writing or changing anything under `docs/` | <!-- TODO --> |", 1)
	write(t, f.root, "AGENTS.md", doc)
	only(t, f.findings(), 1, breaks, "the gate for who writes or reviews a change under docs/ is still a placeholder")
}

// Not every project keeps the kit's heading. Where none is found the
// whole document is read, so a gate answered under another title is
// answered.
func TestGatesAreFoundWithoutTheKitsHeading(t *testing.T) {
	f := newFixture(t, "1")
	write(t, f.root, "AGENTS.md", strings.Replace(agentsDoc, "### Human gates", "### Who decides what", 1))
	if found := f.findings(); len(found) != 0 {
		t.Errorf("findings = %d, want none:\n%s", len(found), texts(found))
	}
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

func TestATagWithALeadingZeroReads(t *testing.T) {
	for _, tag := range []string{"v0.0.03", "v0.0.3", "v1.10", "v10.0.100"} {
		if !parseableTag(tag) {
			t.Errorf("parseableTag(%q) = false, want true", tag)
		}
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
