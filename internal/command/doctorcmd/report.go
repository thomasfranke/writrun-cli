package doctorcmd

import (
	"fmt"
	"io"
	"strings"
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
func render(w io.Writer, stage int, found []finding) {
	fmt.Fprintf(w, "Stage %d is declared — stages 0–%d examined. doctor reports; it repairs nothing.\n", stage, stage)
	for s := 0; s <= 3; s++ {
		group := at(found, s)
		fmt.Fprintf(w, "\nStage %d — %s: ", s, stageNames[s])
		switch {
		case s > stage:
			fmt.Fprintf(w, "not examined — the repository declares stage %d.\n", stage)
		case len(group) == 0:
			fmt.Fprintln(w, "all clear.")
		default:
			fmt.Fprintf(w, "%d finding(s).\n", len(group))
			for _, f := range group {
				fmt.Fprintf(w, "  %-7s  %s\n", labels[f.level], f.text)
				for _, line := range detailLines(f.detail) {
					fmt.Fprintf(w, "           | %s\n", line)
				}
			}
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, summary(stage, found))
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

// at is the findings one stage made, in the order they were found.
func at(found []finding, stage int) []finding {
	var group []finding
	for _, f := range found {
		if f.stage == stage {
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
