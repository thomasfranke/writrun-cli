package authorcmd

import (
	"strings"
	"testing"
)

func TestTableRendersTheRows(t *testing.T) {
	got := table([]row{
		{task: "task-0016", spec: "spec-0014", what: "Declare the derived work"},
		{task: "task-0017", spec: "", what: "A task with no spec"},
	})
	want := "| Task | Spec | What it implements |\n" +
		"|---|---|---|\n" +
		"| task-0016 | spec-0014 | Declare the derived work |\n" +
		"| task-0017 | — | A task with no spec |"
	if got != want {
		t.Errorf("table =\n%s\nwant\n%s", got, want)
	}
}

func TestTableDeclaresNoneRatherThanNothing(t *testing.T) {
	got := table(nil)
	if !strings.Contains(got, "none") {
		t.Errorf("table = %q, want the word check_derived_work.sh reads", got)
	}
}

// A heading carrying a pipe would end the column it sits in.
func TestACellKeepsItsPipe(t *testing.T) {
	got := table([]row{{task: "task-0016", spec: "spec-0014", what: "A | B"}})
	if !strings.Contains(got, `A \| B`) {
		t.Errorf("table = %q, want the pipe escaped", got)
	}
}

// The instruction comment goes with the section it instructs: it
// carries the word `none`, and check_derived_work.sh reads the section
// by grepping for exactly that word — a body that kept it would satisfy
// the check while declaring nothing.
func TestTheDerivedSectionsInstructionCommentIsDropped(t *testing.T) {
	got := authoringHalf(template, table([]row{{task: "task-0016", spec: "spec-0014", what: "A thing"}}))
	if strings.Contains(got, "AUTHORING PRs ONLY") {
		t.Errorf("the instruction comment survived:\n%s", got)
	}
	section := derivedSection(got)
	if strings.Contains(section, "none") {
		t.Errorf("the filled section still reads as a `none` declaration:\n%s", section)
	}
	if !strings.Contains(section, "| task-0016 | spec-0014 | A thing |") {
		t.Errorf("the section carries no table:\n%s", section)
	}
}

func TestTheImplementingHalfIsDropped(t *testing.T) {
	got := authoringHalf(template, table(nil))
	for _, gone := range []string{specHeading, "Implements spec-NNNN", "IMPLEMENTATION PRs ONLY", "Shipped by WritRun"} {
		if strings.Contains(got, gone) {
			t.Errorf("%q survived:\n%s", gone, got)
		}
	}
	for _, kept := range []string{"## What", "## Why", "<!-- writrun:begin -->", derivedHeading, "## How to verify", "<!-- writrun:end -->", "## Notes"} {
		if !strings.Contains(got, kept) {
			t.Errorf("%q was dropped:\n%s", kept, got)
		}
	}
	if !strings.HasSuffix(got, "\n") || strings.HasSuffix(got, "\n\n") {
		t.Errorf("the body does not end in exactly one newline: %q", got[len(got)-4:])
	}
}

// A template with neither half is still a template: what it has is
// kept, and nothing is invented around it.
func TestATemplateWithoutTheHeadingsIsKeptAsItIs(t *testing.T) {
	got := authoringHalf("## What\n\n## Notes\n", table(nil))
	if got != "## What\n\n## Notes\n" {
		t.Errorf("authoringHalf = %q", got)
	}
}

// A section that runs to the end of the file ends there.
func TestASectionAtTheEndIsDroppedWhole(t *testing.T) {
	got := authoringHalf("## What\n\n"+specHeading+"\n\nImplements spec-NNNN.\n", table(nil))
	if strings.Contains(got, "Implements") {
		t.Errorf("the trailing section survived: %q", got)
	}
}

func TestFallbackBodyCarriesTheContractMarker(t *testing.T) {
	got := fallbackBody(derivedNone)
	if !strings.Contains(got, derivedHeading) || !strings.Contains(got, derivedNone) {
		t.Errorf("fallbackBody = %q", got)
	}
	if strings.Contains(got, specHeading+"\n") {
		t.Errorf("the fallback carries an implementing half: %q", got)
	}
}

func TestSkipLeadingCommentOnlyStepsOverALeadingOne(t *testing.T) {
	if got := skipLeadingComment([]string{"## What", "<!--", "-->"}); got != 0 {
		t.Errorf("skipLeadingComment = %d, want 0 for a body that opens with prose", got)
	}
	if got := skipLeadingComment(nil); got != 0 {
		t.Errorf("skipLeadingComment(nil) = %d, want 0", got)
	}
	// An unterminated comment eats the file rather than emitting it
	// half-read — a template that shape is a fault to see, not to ship.
	if got := skipLeadingComment([]string{"<!--", "still open"}); got < 2 {
		t.Errorf("skipLeadingComment = %d, want the whole file consumed", got)
	}
}

func TestFieldAndListReadTheFrontMatterAlone(t *testing.T) {
	content := []byte(taskFixture("task-0016", "spec-0014, spec-0015", "A thing"))
	if got := field(content, "id"); got != "task-0016" {
		t.Errorf("id = %q", got)
	}
	if got := list(content, "spec_ref"); strings.Join(got, ",") != "spec-0014,spec-0015" {
		t.Errorf("spec_ref = %v", got)
	}
	if got := list(content, "depends_on"); len(got) != 0 {
		t.Errorf("depends_on = %v, want nothing", got)
	}
	if got := field(content, "nothing"); got != "" {
		t.Errorf("a field the schema has not = %q", got)
	}
	if got := field([]byte("# no front matter\n"), "id"); got != "" {
		t.Errorf("id of a file with no front matter = %q", got)
	}
	if got := field([]byte("---\nid: task-1\n"), "id"); got != "" {
		t.Errorf("id of an unclosed front matter = %q", got)
	}
	if got := list([]byte("---\nid: x\nspec_ref: null\n---\n"), "spec_ref"); len(got) != 0 {
		t.Errorf("a null list = %v", got)
	}
}

func TestHeadingIsTheSentenceAfterTheId(t *testing.T) {
	if got := heading([]byte(specFixture("spec-0014", "task-0016", "The declaration is the section"))); got != "The declaration is the section" {
		t.Errorf("heading = %q", got)
	}
	if got := heading([]byte(taskFixture("task-0016", "spec-0014", "Declare the derived work"))); got != "Declare the derived work" {
		t.Errorf("heading = %q", got)
	}
	if got := heading([]byte("no heading at all\n")); got != "" {
		t.Errorf("heading = %q, want nothing", got)
	}
}

// derivedSection is what check_derived_work.sh reads: the lines under
// the contract marker, up to the next heading.
func derivedSection(body string) string {
	var out []string
	inside := false
	for _, l := range strings.Split(body, "\n") {
		if l == derivedHeading {
			inside = true
			continue
		}
		if inside && strings.HasPrefix(l, "## ") {
			break
		}
		if inside {
			out = append(out, l)
		}
	}
	return strings.Join(out, "\n")
}
