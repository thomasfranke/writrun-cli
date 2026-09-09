package doctorcmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/palette"
)

// level is what a finding costs. Only one of the three reaches the exit
// status: a recommended setting missing is a recommendation, and a
// check the forge would not answer is not a failed check
// (product/adoption/doctor.md; spec-0004, acceptance criteria).
type level int

const (
	// breaks says the finding stops a flow the methodology runs.
	breaks level = iota
	// advises says the methodology recommends it and nothing breaks
	// without it.
	advises
	// unread says the check could not be made at all — the forge did
	// not answer. It is reported so the reader knows what went
	// unexamined, and it never fails the run.
	unread
)

// labels are how each level appears in the left column, one word wide
// enough that the findings line up under it.
var labels = map[level]string{
	breaks:  "breaks",
	advises: "advises",
	unread:  "unread",
}

// finding is one answer about one assumption: the stage that makes it,
// what the answer costs, the sentence naming the file or setting and
// what is expected of it, and — where a wrapped script spoke — that
// script's own words, printed under it unedited.
type finding struct {
	stage  int
	level  level
	text   string
	detail string
}

// stageNames titles each group in the report.
var stageNames = [4]string{"environment", "files", "the forge", "Issues"}

// render prints the findings grouped by stage, every group named even
// when it holds nothing: a stage that was not examined has to say so,
// or a clean report and an unexamined one read alike (spec-0004, edge
// cases).
func render(w io.Writer, p palette.Palette, declared, examined int, found []finding) {
	if examined == declared {
		fmt.Fprintf(w, "Stage %d is declared — stages 0–%d examined. doctor reports; it repairs nothing.\n", declared, declared)
	} else {
		// Both numbers, always: a run reaching past the declaration is
		// an answer about a stage this repository has not taken on, and
		// reading it as the verdict would be reading someone else's.
		fmt.Fprintf(w, "Stage %d is declared, stages 0–%d examined at your asking — what stage %d costs, not what this repository owes. doctor reports; it repairs nothing.\n", declared, examined, examined)
	}
	for s := 0; s <= 3; s++ {
		group := at(found, s)
		fmt.Fprintf(w, "\n%s: ", p.Heading(fmt.Sprintf("Stage %d — %s", s, stageNames[s])))
		switch {
		case s > examined:
			fmt.Fprintf(w, "not examined — the repository declares stage %d.\n", declared)
		case len(group) == 0:
			fmt.Fprintln(w, "all clear.")
		default:
			fmt.Fprintf(w, "%d finding(s).\n", len(group))
			for _, f := range group {
				fmt.Fprintf(w, "  %s  %s\n", paintLevel(p, f.level), f.text)
				for _, line := range detailLines(f.detail) {
					fmt.Fprintf(w, "           | %s\n", line)
				}
			}
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, summary(examined, found))
}

// summary is the one line a reader can stop at, and it agrees with the
// exit status: a non-zero run always has something that breaks a flow
// in it.
func summary(stage int, found []finding) string {
	if len(found) == 0 {
		return fmt.Sprintf("Every assumption up to stage %d holds.", stage)
	}
	b, a, u := counts(found)
	if b == 0 {
		return fmt.Sprintf("%d finding(s), none breaking a flow: %d recommended, %d unread.", len(found), a, u)
	}
	return fmt.Sprintf("%d finding(s): %d breaking a flow, %d recommended, %d unread.", len(found), b, a, u)
}

// level paints a finding's word without changing it: the column is
// what a reader without colour reads, so the word keeps its width and
// its place (docs/product/rules.md — colour never carries meaning
// alone).
func paintLevel(p palette.Palette, l level) string {
	word := fmt.Sprintf("%-7s", labels[l])
	switch l {
	case breaks:
		return p.Breaks(word)
	case advises:
		return p.Advises(word)
	default:
		return p.Unread(word)
	}
}

// at is the findings the named stages made, in the order they were
// found. One stage renders a group; the declaration's whole range is
// what the exit status answers for.
func at(found []finding, stages ...int) []finding {
	want := map[int]bool{}
	for _, s := range stages {
		want[s] = true
	}
	var group []finding
	for _, f := range found {
		if want[f.stage] {
			group = append(group, f)
		}
	}
	return group
}

// counts tallies the findings by level.
func counts(found []finding) (b, a, u int) {
	for _, f := range found {
		switch f.level {
		case breaks:
			b++
		case advises:
			a++
		case unread:
			u++
		}
	}
	return b, a, u
}

// breaking is how many findings break a flow — the only number the exit
// status reads.
func breaking(found []finding) int {
	b, _, _ := counts(found)
	return b
}

// detailLines is a wrapped script's own reporting, split for the
// indented block under the finding. Blank lines at either end are the
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
