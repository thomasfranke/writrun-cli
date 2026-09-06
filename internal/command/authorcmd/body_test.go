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
	got, filled := authoringHalf(template, table([]row{{task: "task-0016", spec: "spec-0014", what: "A thing"}}))
	if !filled {
		t.Fatal("the section was never filled")
	}
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
	got, _ := authoringHalf(template, table(nil))
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
// kept, and nothing is invented around it — but nothing was filled
// either, and the caller is told so rather than left to grep.
func TestATemplateWithoutTheHeadingsIsKeptAsItIs(t *testing.T) {
	got, filled := authoringHalf("## What\n\n## Notes\n", table(nil))
	if got != "## What\n\n## Notes\n" {
		t.Errorf("authoringHalf = %q", got)
	}
	if filled {
		t.Error("a template with no `## Derived work` reported a filled section")
	}
}

// A section that runs to the end of the file ends there.
func TestASectionAtTheEndIsDroppedWhole(t *testing.T) {
	got, _ := authoringHalf("## What\n\n"+specHeading+"\n\nImplements spec-NNNN.\n", table(nil))
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
	// half-read. What is left carries no `## Derived work`, so the fill
	// reports that it did not happen and the caller answers with the
	// fallback — the template is never shipped half-read either way.
	if got := skipLeadingComment([]string{"<!--", "still open"}); got < 2 {
		t.Errorf("skipLeadingComment = %d, want the whole file consumed", got)
	}
	if _, filled := authoringHalf("<!--\nstill open\n"+derivedHeading+"\n\n| x |\n", table(nil)); filled {
		t.Error("a template consumed by its own comment reported a filled section")
	}
}

// The template is the adopter's to edit, and two edits leave it unable
// to carry the declaration: the contract-marker heading renamed — which
// the template itself warns blinds the check — and a leading comment
// left unterminated. Both are silent in the fill, so both are answered
// by the fallback rather than by shipping the placeholder row and the
// `none` comment the door would then read.
func TestATemplateThatCannotCarryTheDeclarationFallsBack(t *testing.T) {
	cases := []struct{ name, tmpl string }{
		{"the heading renamed", strings.Replace(template, derivedHeading, "## Derived Work", 1)},
		{"a comment that never closes", "<!--\nstill open, and nothing below closes it\n\n## What\n\n" +
			derivedHeading + "\n\n| task-NNNN | spec-NNNN | |\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newHarness(t)
			h.seed(templatePath, c.tmpl)
			if err := h.author(); err != nil {
				t.Fatalf("author: %v", err)
			}
			body := h.gh.created
			if strings.Contains(body, "task-NNNN") {
				t.Errorf("the template's placeholder row reached the forge:\n%s", body)
			}
			if strings.Contains(body, "AUTHORING PRs ONLY") {
				t.Errorf("the instruction comment the door reads as `none` reached the forge:\n%s", body)
			}
			if !strings.Contains(body, derivedHeading) {
				t.Errorf("the contract marker is gone, so the door goes blind:\n%s", body)
			}
			if !strings.Contains(body, "| task-0016 | spec-0014 | Declare the derived work |") {
				t.Errorf("the body does not declare what this change derived:\n%s", body)
			}
		})
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
