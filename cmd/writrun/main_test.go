package main

import "testing"

// menuRow is the width the entry screen's drawing gives a summary,
// after the name column (docs/product/screens/entry.excalidraw).
const menuRow = 44

// The screen prints a command's Summary and no string written for the
// screen alone, so one summary serves the menu and `--help` alike. A
// summary that outgrows the row is a defect in the summary — the screen
// truncates nothing, because a half-sentence is worse than a short one.
func TestEverySummaryFitsAMenuRow(t *testing.T) {
	for _, c := range commands() {
		if len(c.Summary) > menuRow {
			t.Errorf("%s: summary is %d characters and the row holds %d:\n  %s",
				c.Name, len(c.Summary), menuRow, c.Summary)
		}
	}
}
