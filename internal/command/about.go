package command

import (
	"fmt"
	"io"
)

// About is the long description a command carries beside its one-liner:
// the sentence it opens with, then the parts a reader who has not read
// the methodology needs — why you would reach for it, what it does,
// what it leaves alone, what comes next (spec-0039).
//
// It is lines rather than paragraphs because the drawing states the
// lines. `docs/product/screens/` is the reference this rendering is
// checked against, and a wrapper of this package's own would answer a
// width no drawing chose.
type About struct {
	// Sentence follows the command's name on the first line; a further
	// line continues it at the left margin.
	Sentence []string
	// Parts are the labelled parts, printed in order, one blank line
	// above each.
	Parts []AboutPart
}

// AboutPart is one labelled part of a long description.
type AboutPart struct {
	// Label is the part's own words — `why you would`, `what it does`,
	// `what it leaves`, `what it never`, `what comes next`.
	Label string
	// Lines are the part's text, one per rendered line.
	Lines []string
}

// Empty says the command carries no long description. Every command
// carries one, and this is what the check over the table reads.
func (a About) Empty() bool { return len(a.Sentence) == 0 || len(a.Parts) == 0 }

// aboutLabel is the width of the label column; the text starts one
// column past it.
const aboutLabel = 17

// writeAbout prints a long description as the drawings state it: the
// sentence beside the command's name, then each labelled part under one
// blank line.
func writeAbout(w io.Writer, name string, a About) {
	if a.Empty() {
		return
	}
	for i, l := range a.Sentence {
		if i == 0 {
			fmt.Fprintf(w, "%s — %s\n", name, l)
			continue
		}
		fmt.Fprintln(w, l)
	}
	for _, p := range a.Parts {
		fmt.Fprintln(w)
		for i, l := range p.Lines {
			if i == 0 {
				fmt.Fprintf(w, " %-*s%s\n", aboutLabel, p.Label, l)
				continue
			}
			fmt.Fprintf(w, "%*s%s\n", aboutLabel+1, "", l)
		}
	}
}
