package authorcmd

import (
	"fmt"
	"path"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/kit"
	"github.com/thomasfranke/writrun-cli/internal/queue"
)

// templatePath is the kit's one home for the pull-request body
// (conventions/prs.md). Everything in it is the project's to edit
// except one marker.
const templatePath = kit.PullRequestTemplate

// The two headings the template's two halves open with. `## Derived
// work` is a **contract marker** — `writrun check` finds the
// declaration by that exact heading — so it is matched literally and
// never rewritten.
const (
	derivedHeading = "## Derived work"
	specHeading    = "## Spec"
	reportHeading  = "## Report"
)

// derivedNone is the declaration a rule that derives no work carries.
// An empty section and a forgotten one look identical, which is why the
// word is written rather than the section left blank
// (check_derived_work.sh).
const derivedNone = "none — this rule derives no work."

// row is one line of the Derived-work table: what was derived, and what
// it is for.
type row struct{ task, spec, what string }

// derived reads the table off the diff and the queue. The tasks and the
// specs the change *adds* are the derivation — the same read
// check_derived_work.sh makes at the door, so the body and the check
// cannot disagree about what this change derived.
func derived(d Deps, root, rng string) ([]row, error) {
	tasks, err := added(d, root, rng, queue.TasksDir+"/task-*.md")
	if err != nil {
		return nil, err
	}
	specs, err := added(d, root, rng, queue.SpecsDir+"/spec-*.md")
	if err != nil {
		return nil, err
	}

	// The specs are read first so a task's row can name the ones it
	// carries, and so the ones no added task claims still get a line:
	// a spec derived for a task that already existed is derived work
	// too, and a table that dropped it would under-report the change.
	type queueFile struct{ id, taskRef, heading string }
	bySpec := map[string]queueFile{}
	var specOrder []string
	for _, p := range specs {
		content, err := d.Files.ReadFile(path.Join(root, p))
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", p, err)
		}
		id := queue.Field(content, "id")
		if id == "" {
			id = strings.TrimSuffix(path.Base(p), ".md")
		}
		bySpec[id] = queueFile{id: id, taskRef: queue.Field(content, "task_ref"), heading: subject(content)}
		specOrder = append(specOrder, id)
	}

	claimed := map[string]bool{}
	var rows []row
	for _, p := range tasks {
		content, err := d.Files.ReadFile(path.Join(root, p))
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", p, err)
		}
		id := queue.Field(content, "id")
		if id == "" {
			id = strings.TrimSuffix(path.Base(p), ".md")
		}
		refs := queue.List(content, "spec_ref")
		for _, r := range refs {
			claimed[r] = true
		}
		rows = append(rows, row{task: id, spec: joined(refs), what: subject(content)})
	}
	for _, id := range specOrder {
		if claimed[id] {
			continue
		}
		s := bySpec[id]
		rows = append(rows, row{task: nonEmpty(s.taskRef), spec: id, what: s.heading})
	}
	return rows, nil
}

// added lists the files the change creates under one path spec. The
// filter is `A` because deriving work is adding it: a task the change
// merely edits was derived by some earlier rule.
func added(d Deps, root, rng, pathspec string) ([]string, error) {
	out, err := d.Git(root, "diff", "--name-only", "--diff-filter=A", rng, "--", pathspec)
	if err != nil {
		return nil, fmt.Errorf("reading what %s adds under %s: %w", rng, pathspec, err)
	}
	return trimmedLines(out), nil
}

// subject is what the table's third column says a derived task or spec
// is for: the file's title, less the id a spec's title opens with. The
// id is already in the column beside it.
func subject(content []byte) string {
	title := queue.Heading(content)
	if _, after, split := strings.Cut(title, " — "); split {
		return strings.TrimSpace(after)
	}
	return title
}

// table renders the rows, or the declaration that there are none.
func table(rows []row) string {
	if len(rows) == 0 {
		return derivedNone
	}
	out := []string{"| Task | Spec | What it implements |", "|---|---|---|"}
	for _, r := range rows {
		out = append(out, fmt.Sprintf("| %s | %s | %s |", cell(r.task), cell(r.spec), cell(r.what)))
	}
	return strings.Join(out, "\n")
}

// cell keeps a heading that carries a pipe from ending the column it
// sits in.
func cell(s string) string {
	s = strings.ReplaceAll(strings.TrimSpace(s), "|", `\|`)
	if s == "" {
		return "—"
	}
	return s
}

func nonEmpty(s string) string {
	if t := strings.TrimSpace(s); t != "" && t != "null" {
		return t
	}
	return ""
}

func joined(ids []string) string { return strings.Join(ids, ", ") }

// body is the template's authoring half, the table filled in. A
// template that cannot carry the declaration gets the fallback instead
// — the same headings, the contract marker included — because no state
// of the adopter's template is a reason to open a pull request the door
// cannot read.
//
// Unreadable is one such state. The other is a template that *is* read
// and does not carry `## Derived work` where this can fill it: the
// heading renamed (which the template itself warns blinds the check),
// or a leading comment left unterminated, which consumes the file. The
// edit is silent in both — the shipped placeholder row and its `none`
// comment would survive onto the forge, declaring `task-NNNN` — so the
// fill reports whether it happened rather than being assumed.
func body(d Deps, root string, rows []row) string {
	t := table(rows)
	content, err := d.Files.ReadFile(path.Join(root, templatePath))
	if err != nil {
		return fallbackBody(t)
	}
	half, filled := authoringHalf(string(content), t)
	if !filled {
		return fallbackBody(t)
	}
	return half
}

// authoringHalf keeps `## Derived work` and drops every other kind's
// section, the way take_task.sh keeps `## Spec` and drops this one.
// Three edits, and no more: the leading instruction comment goes, the
// Derived-work section becomes the table, and the sections belonging to
// the other kinds — `## Spec` and `## Report` — go whole.
//
// The list is the headings the template marks for one kind only, so a
// tag that adds a fourth kind adds a heading here. Nothing else does:
// `## How to verify` and `## How to test` are asked of every kind, and
// an authoring PR answers them like any other.
//
// The section's own instruction comment goes with it, and that is
// deliberate rather than tidy: it contains the word `none`, and
// check_derived_work.sh reads the section by grepping for exactly that
// word — a body that kept the comment would satisfy the check while
// declaring nothing, which is the blindness the declaration exists to
// end.
// The second return says whether the Derived-work section was found and
// filled. False is not a smaller success: it is a body carrying
// whatever the template said instead of what this change derived, and
// the caller answers it with the fallback.
func authoringHalf(tmpl, t string) (string, bool) {
	lines := strings.Split(tmpl, "\n")
	i := skipLeadingComment(lines)
	filled := false
	var out []string
	for ; i < len(lines); i++ {
		switch strings.TrimRight(lines[i], " \t") {
		case derivedHeading:
			out = append(out, derivedHeading, "", t, "")
			i = endOfSection(lines, i)
			filled = true
		case specHeading, reportHeading:
			i = endOfSection(lines, i)
		default:
			out = append(out, lines[i])
		}
	}
	return strings.TrimRight(strings.Join(out, "\n"), "\n") + "\n", filled
}

// skipLeadingComment steps over the template's own instructions to its
// filler — an HTML comment opening at line 1 — and the blank lines
// under it. A comment that never closes consumes the file, which leaves
// no `## Derived work` to fill and so answers itself through the
// fallback rather than emitting a template read half way.
func skipLeadingComment(lines []string) int {
	i := 0
	if len(lines) == 0 || !strings.HasPrefix(strings.TrimSpace(lines[0]), "<!--") {
		return 0
	}
	for i < len(lines) && !strings.Contains(lines[i], "-->") {
		i++
	}
	i++
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}
	return i
}

// endOfSection returns the index of the last line of the section
// opening at i, so the caller's loop lands on the next heading.
func endOfSection(lines []string, i int) int {
	for j := i + 1; j < len(lines); j++ {
		if strings.HasPrefix(lines[j], "## ") {
			return j - 1
		}
	}
	return len(lines) - 1
}

// fallbackBody is the same shape the template ships, written out where
// the template cannot be read.
func fallbackBody(t string) string {
	return "## What\n\n## Why\n\n<!-- writrun:begin -->\n\n" +
		derivedHeading + "\n\n" + t + "\n\n" +
		"## How to verify\n\n<!-- writrun:end -->\n\n## Notes\n"
}
