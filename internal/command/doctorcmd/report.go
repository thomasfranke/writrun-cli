package doctorcmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/palette"
)

// mark is the state one requirement is in. Four states, four glyphs, and
// only one of them reaches the exit status: a recommended setting
// missing is a recommendation, and a check the forge would not answer is
// not a failed check (product/adoption/doctor.md).
type mark int

const (
	// met says the requirement holds.
	met mark = iota
	// breaks says the requirement is unmet and stops a flow the
	// methodology runs.
	breaks
	// advises says the methodology recommends it and nothing breaks
	// without it.
	advises
	// unread says the check could not be made at all — the forge did not
	// answer. It is reported so the reader knows what went unexamined,
	// and it never fails the run.
	unread
)

// glyphs are the four states in the report's left column. The glyph is
// what a reader without colour reads, so it carries the state and the
// colour repeats it (docs/product/rules.md).
var glyphs = map[mark]string{met: "✓", breaks: "✗", advises: "!", unread: "?"}

// words name each state for a reader who is being told about one rather
// than looking at a row: the summary's counts and the screen's footer.
var words = map[mark]string{met: "met", breaks: "breaks", advises: "advises", unread: "unread"}

// requirement is one thing a stage asks of this repository: the stage
// that asks it, the stable name the document explains it under, the
// state it is in, what the row adds after the name, and — where a
// wrapped script or the forge spoke — those words, printed under it
// unedited.
//
// **A requirement that holds is a row.** The report names every one of
// them, so a repository satisfying nine checks reads differently from
// one this binary never examined (spec-0036).
type requirement struct {
	stage  int
	name   string
	note   string
	detail string
	mark   mark
}

// text is the row after the glyph: the name, and what this run has to
// add about it.
func (r requirement) text() string {
	if r.note == "" {
		return r.name
	}
	return r.name + " — " + r.note
}

// stageNames title each group in the report. They are the subjects the
// groups examine, not the stages' own names, which are `init`'s
// (product/adoption/doctor.md).
var stageNames = [4]string{"environment", "files", "the forge", "Issues"}

// rowIndent, detailIndent and the glyph's column are the drawing's
// (docs/product/screens/adoption/doctor.excalidraw).
const (
	lineIndent   = " "
	rowIndent    = "   "
	detailIndent = "        "
)

// render prints the requirements grouped by stage, every group counted,
// and closes with what the declaration answers for and whether the rung
// above it is within reach.
func render(w io.Writer, p palette.Palette, declared, examined, preview int, found []requirement) {
	fmt.Fprintln(w, lineIndent+header(declared, examined, preview))
	fmt.Fprintln(w, lineIndent+"doctor reports; it repairs nothing.")
	for s := 0; s <= 3; s++ {
		fmt.Fprintln(w)
		fmt.Fprintln(w, lineIndent+p.Heading(groupHeading(s, examined, preview, at(found, s))))
		for _, r := range at(found, s) {
			fmt.Fprintf(w, "%s%s  %s\n", rowIndent, paint(p, r.mark), r.text())
			for _, line := range detailLines(r.detail) {
				fmt.Fprintln(w, detailIndent+line)
			}
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, lineIndent+summary(declared, at(found, upTo(declared)...)))
	fmt.Fprintln(w, lineIndent+reach(preview, at(found, preview)))
}

// header is the first line: which stage this repository declares, how
// far the run examined, and which rung it previewed.
func header(declared, examined, preview int) string {
	if examined != declared {
		// Both numbers, always: a run reaching past the declaration is an
		// answer about a stage this repository has not taken on, and
		// reading it as the verdict would be reading someone else's.
		line := fmt.Sprintf("Stage %d is declared, stages 0–%d examined at your asking — what stage %d costs, not what this repository owes", declared, examined, examined)
		if preview == 0 {
			return line + "."
		}
		return fmt.Sprintf("%s, with stage %d previewed.", line, preview)
	}
	if preview == 0 {
		return fmt.Sprintf("Stage %d is declared — stages 0–%d examined; there is no rung above it.", declared, declared)
	}
	return fmt.Sprintf("Stage %d is declared — stages 0–%d examined, stage %d previewed.", declared, declared, preview)
}

// groupHeading names one stage and counts what it required. A stage
// neither examined nor previewed says so: the preview reaches one rung,
// and a reader is told that rather than left to read silence.
func groupHeading(s, examined, preview int, group []requirement) string {
	switch {
	case preview > 0 && s == preview:
		return fmt.Sprintf("Stage %d — %s, previewed: %d of %d met.", s, stageNames[s], counted(group, met), len(group))
	case s > examined:
		return fmt.Sprintf("Stage %d — %s: not previewed — one rung at a time.", s, stageNames[s])
	default:
		return fmt.Sprintf("Stage %d — %s: %d of %d met.", s, stageNames[s], counted(group, met), len(group))
	}
}

// summary is the one line a reader can stop at, and it agrees with the
// exit status: a non-zero run always has something that breaks a flow in
// it. It answers for the declaration's own range and no further.
func summary(declared int, found []requirement) string {
	unmet := len(found) - counted(found, met)
	if unmet == 0 {
		return fmt.Sprintf("Every assumption up to stage %d holds.", declared)
	}
	b, a, u := counted(found, breaks), counted(found, advises), counted(found, unread)
	if b == 0 {
		return fmt.Sprintf("%s at stage %d, none breaking a flow: %s.",
			plural(unmet, "finding", "findings"), declared,
			list(count(a, "recommended"), count(u, "unread")))
	}
	return fmt.Sprintf("%s at stage %d: %s.",
		plural(unmet, "finding", "findings"), declared,
		list(count(b, "breaking a flow"), count(a, "recommended"), count(u, "unread")))
}

// reach is the question the reader actually came with — can I move up
// yet — answered in one line from the preview's own rows.
func reach(preview int, group []requirement) string {
	if preview == 0 {
		return "Stage 3 is the top rung — there is nothing above it to preview."
	}
	u := counted(group, unread)
	unmet := len(group) - counted(group, met) - u
	if unmet == 0 && u == 0 {
		return fmt.Sprintf("Stage %d is within reach: its %d requirements are met.", preview, len(group))
	}
	return fmt.Sprintf("Stage %d is not within reach: %s.", preview,
		list(count(unmet, "requirement unmet", "requirements unmet"), count(u, "unread")))
}

// paint colours a glyph without changing it: the glyph is the state, and
// the colour repeats it (docs/product/rules.md — colour never carries
// meaning alone).
func paint(p palette.Palette, m mark) string {
	switch m {
	case met:
		return p.Met(glyphs[m])
	case breaks:
		return p.Breaks(glyphs[m])
	case advises:
		return p.Advises(glyphs[m])
	default:
		return p.Unread(glyphs[m])
	}
}

// at is the requirements the named stages made, in the order they were
// found. One stage renders a group; the declaration's whole range is
// what the exit status answers for.
func at(found []requirement, stages ...int) []requirement {
	want := map[int]bool{}
	for _, s := range stages {
		want[s] = true
	}
	var group []requirement
	for _, r := range found {
		if want[r.stage] {
			group = append(group, r)
		}
	}
	return group
}

// upTo is every stage a declaration reaches, which is what the exit
// status answers for.
func upTo(stage int) []int {
	out := make([]int, 0, stage+1)
	for s := 0; s <= stage; s++ {
		out = append(out, s)
	}
	return out
}

// counted is how many requirements are in one state.
func counted(found []requirement, m mark) int {
	n := 0
	for _, r := range found {
		if r.mark == m {
			n++
		}
	}
	return n
}

// breaking is how many requirements break a flow — the only number the
// exit status reads.
func breaking(found []requirement) int { return counted(found, breaks) }

// count is one tally as a reader reads it, or nothing where there is
// nothing to say. The plural is given where the noun takes one.
func count(n int, one string, many ...string) string {
	if n == 0 {
		return ""
	}
	word := one
	if n != 1 && len(many) > 0 {
		word = many[0]
	}
	return fmt.Sprintf("%d %s", n, word)
}

// list joins the tallies that have something to say, dropping the ones
// that do not: a zero is not news.
func list(parts ...string) string {
	var kept []string
	for _, p := range parts {
		if p != "" {
			kept = append(kept, p)
		}
	}
	if len(kept) == 0 {
		return "none"
	}
	return strings.Join(kept, ", ")
}

// plural is a count and its noun.
func plural(n int, one, many string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, one)
	}
	return fmt.Sprintf("%d %s", n, many)
}

// detailLines is a wrapped script's own reporting, split for the
// indented block under the row. Blank lines at either end are the
// script's spacing, not its message, and are dropped.
func detailLines(detail string) []string {
	trimmed := strings.Trim(detail, "\n")
	if strings.TrimSpace(trimmed) == "" {
		return nil
	}
	return strings.Split(trimmed, "\n")
}

// firstLine keeps an error to the sentence that names the cause: gh
// prints the request and the response, and the report has room for the
// first of them.
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
