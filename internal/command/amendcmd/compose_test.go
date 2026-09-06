package amendcmd

import (
	"strings"
	"testing"
)

func TestSlugify(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Reopen The Gate", "reopen-the-gate"},
		{"  spaced  out  ", "spaced-out"},
		{"already-slugged", "already-slugged"},
		{"Ünïcode & symbols!", "n-code-symbols"},
		{"", ""},
		{"---", ""},
	}
	for _, c := range cases {
		if got := slugify(c.in); got != c.want {
			t.Errorf("slugify(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSpecSlug(t *testing.T) {
	cases := []struct{ in, want string }{
		{"work/specs/spec-0011-amend-command.md", "amend-command"},
		{"spec-0011-amend-command.md", "amend-command"},
		{"work/specs/spec-0011.md", "0011"},
	}
	for _, c := range cases {
		if got := specSlug(c.in); got != c.want {
			t.Errorf("specSlug(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// The title is composed in the declared style and carries no task tag —
// an amendment is not work on a task (conventions/prs.md).
func TestTitleFollowsTheStyle(t *testing.T) {
	cases := []struct{ style, kind, want string }{
		{"bracketed", "docs", "[Docs][Specs] A sentence"},
		{"conventional", "docs", "docs(specs): A sentence"},
		{"fix", "fix", "[Fix][Specs] A sentence"},
		{"", "docs", "[Docs][Specs] A sentence"},
	}
	for _, c := range cases {
		got := title(c.style, c.kind, "A sentence")
		if got != c.want {
			t.Errorf("title(%q, %q) = %q, want %q", c.style, c.kind, got, c.want)
		}
		if strings.Contains(got, "[TASK-") {
			t.Errorf("title %q carries a task tag", got)
		}
	}
}

// The commit subject is Conventional Commits whatever the title style
// says (conventions/commits.md).
func TestSubjectIsConventionalWhateverTheStyle(t *testing.T) {
	if got := subject("docs", "spec-0011"); got != "docs(specs): return spec-0011 to draft" {
		t.Errorf("subject = %q", got)
	}
}

func TestBranchNameCarriesNoId(t *testing.T) {
	got := branchName("docs", "amend-command")
	if got != "docs/amend-command" {
		t.Errorf("branchName = %q", got)
	}
	if strings.HasPrefix(got, "task/") {
		t.Error("the branch reads as flight")
	}
}

// The line is the one check_amendment_reference.sh asks for, character
// for character — the script prints it as the fix, and this composes it.
func TestTheSuspensionLineIsTheOneTheCheckAsksFor(t *testing.T) {
	got := suspension{task: "task-0012", number: 42}.line()
	want := "Suspends #42 — task-0012 waits on this amendment."
	if got != want {
		t.Errorf("line = %q, want %q", got, want)
	}
}

// A pull request the forge could not number is still named, by its
// task, with what a person must finish (spec-0011, acceptance criteria).
func TestAnUnnumberedSuspensionNamesTheTask(t *testing.T) {
	got := suspension{task: "task-0012"}.line()
	if !strings.Contains(got, "task-0012") || !strings.Contains(got, "by hand") {
		t.Errorf("line = %q", got)
	}
}

// The template is this repository's own, transformed: no guidance
// comments, no authoring half, the fence kept, the three sections
// filled, and the suspension where the check reads it.
func TestBodyFromTheProjectTemplate(t *testing.T) {
	tmpl := `<!--
Shipped by WritRun — guidance a filler reads and a reviewer does not.
-->

## What

## Why

<!-- writrun:begin -->

## Derived work

<!-- AUTHORING PRs ONLY. -->

| Task | Spec | What it implements |
|---|---|---|
| task-NNNN | spec-NNNN | |

## Spec

<!-- IMPLEMENTATION PRs ONLY. -->

Implements spec-NNNN.

## How to verify

<!-- Implementation PRs: the writrun-check-spec-deltas result. -->

<!-- writrun:end -->

## Notes`

	got := body(tmpl, "spec-0011", "The contract was wrong.",
		[]suspension{{task: "task-0012", number: 42}})

	for _, want := range []string{
		"## What",
		"Returns spec-0011 to `draft` so its approval can be reconsidered.",
		"## Why",
		"The contract was wrong.",
		"<!-- writrun:begin -->",
		"<!-- writrun:end -->",
		"Returns spec-0011 to `draft`. Re-approval is the merge's, and the maintainer's.",
		"Suspends #42 — task-0012 waits on this amendment.",
		"## How to verify",
		"check_amendment_reference.sh",
		"## Notes",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("body carries no %q:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{
		"## Derived work",
		"Shipped by WritRun",
		"AUTHORING PRs ONLY",
		"Implements spec-NNNN.",
		"| task-NNNN |",
	} {
		if strings.Contains(got, unwanted) {
			t.Errorf("body still carries %q:\n%s", unwanted, got)
		}
	}
	if strings.Contains(got, "\n\n\n") {
		t.Errorf("body carries a run of blank lines:\n%s", got)
	}
}

// A repository whose template is gone still gets a body the checks can
// read.
func TestBodyWithoutATemplate(t *testing.T) {
	got := body("", "spec-0011", "Because.", []suspension{{task: "task-0012", number: 7}})
	for _, want := range []string{"## What", "## Why", "Suspends #7 — task-0012 waits on this amendment.", "## Notes"} {
		if !strings.Contains(got, want) {
			t.Errorf("body carries no %q:\n%s", want, got)
		}
	}
}

// No task in flight: the body says so rather than claiming a suspension
// nothing is waiting on.
func TestBodySaysWhenNothingIsSuspended(t *testing.T) {
	got := body("", "spec-0011", "Because.", nil)
	if strings.Contains(got, "Suspends") {
		t.Errorf("body claims a suspension:\n%s", got)
	}
	if !strings.Contains(got, "suspends nothing") {
		t.Errorf("body never says nothing waits:\n%s", got)
	}
}

// A template that carries no placeholder sentence still gets the
// statement, under the section that exists — the reference the gate
// reads is never the part that goes missing.
func TestBodyKeepsTheStatementWhenThePlaceholderIsGone(t *testing.T) {
	got := body("## What\n\n## Spec\n\n## Notes\n", "spec-0011", "Because.",
		[]suspension{{task: "task-0012", number: 42}})
	if !strings.Contains(got, "Suspends #42") {
		t.Errorf("the statement went missing:\n%s", got)
	}
}

// A template with neither placeholder nor `## Spec` keeps everything it
// has and takes the statement at the end.
func TestBodyAppendsTheStatementWhenThereIsNowhereToPutIt(t *testing.T) {
	got := body("## Notes\n\nSomething the project wrote.\n", "spec-0011", "Because.",
		[]suspension{{task: "task-0012", number: 42}})
	if !strings.Contains(got, "Something the project wrote.") {
		t.Errorf("the project's own body was dropped:\n%s", got)
	}
	if !strings.Contains(got, "Suspends #42") {
		t.Errorf("the statement went missing:\n%s", got)
	}
}

// A section somebody already wrote is left exactly as they wrote it.
func TestFillLeavesAWrittenSectionAlone(t *testing.T) {
	got := body("## What\n\nAlready said.\n\n## Notes\n", "spec-0011", "Because.", nil)
	if !strings.Contains(got, "Already said.") {
		t.Errorf("the written section was replaced:\n%s", got)
	}
	if strings.Contains(got, "so its approval can be reconsidered") {
		t.Errorf("the written section was added to:\n%s", got)
	}
}
