package main

import (
	"os"
	"path/filepath"
	"testing"
)

// menuRow is the width the entry screen's drawing gives a summary: its
// window is 80 columns and the name field takes the first 14
// (docs/product/screens/entry.excalidraw).
const menuRow = 66

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

// A command that ran and refused is not a spawn that failed.
//
// The two are told apart by the error's type, and the distinction is
// load-bearing: the caller falls back to running the command in this
// process when the port answers an error, so reading a refusal as a
// failure would run the command a second time, having already run it
// once. A `take` that opened a pull request would open two.
func TestARefusalIsNotASpawnFailure(t *testing.T) {
	if _, err := os.Stat("/bin/sh"); err != nil {
		t.Skip("no /bin/sh to exit with a code")
	}
	if err := spawn("/bin/sh", []string{"-c", "exit 1"}); err != nil {
		t.Errorf("a child that exited 1 answered %v, want nil — it ran", err)
	}
}

// A process that could not be started is reported, so the caller can
// run the command here instead of not at all.
func TestASpawnThatCannotStartIsAnError(t *testing.T) {
	if err := spawn(filepath.Join(t.TempDir(), "writrun"), nil); err == nil {
		t.Error("a binary that does not exist answered nil")
	}
}

// The four commands that explain nothing on a bare run are the four that
// do their work bare.
//
// `Daily` and `AsksNothing` named the same set until `doctor` became a
// screen, and the day they came apart nothing would have said so: a
// `doctor` left out of this set prints twenty lines into the normal
// buffer and then enters the alternate one, which wipes them — the
// reader meets the explanation on the way out, for a command they run
// every day (report-0044). So the set is asserted by name.
func TestTheDailyCommandsAreTheOnesRunBare(t *testing.T) {
	want := map[string]bool{"list": true, "status": true, "doctor": true, "work": true}
	for _, c := range commands() {
		if c.Daily != want[c.Name] {
			if c.Daily {
				t.Errorf("%s is declared Daily and is not one of the four run bare", c.Name)
			} else {
				t.Errorf("%s is run bare and does not declare Daily — a bare run would "+
					"explain itself to someone who meets it every day", c.Name)
			}
		}
	}
}
