package listcmd

import (
	"strings"
	"testing"
)

func TestAFilterSelectsSections(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		want    []string
		unwant  []string
		wantAll bool
	}{
		{
			name:   "available",
			args:   []string{"--available"},
			want:   []string{"In progress —", "Available —", "task-0006", "Order is a suggestion", "In flight —"},
			unwant: []string{"Held back:", "task-0009", "Open reports —", "report-0002"},
		},
		{
			name:   "held",
			args:   []string{"--held"},
			want:   []string{"Held back:", "task-0009"},
			unwant: []string{"Available —", "task-0006", "In flight —", "Open reports —"},
		},
		{
			name:   "reports",
			args:   []string{"--reports"},
			want:   []string{"Open reports —", "report-0002"},
			unwant: []string{"Available —", "Held back:", "task-0009"},
		},
		{
			name:   "two filters are the union",
			args:   []string{"--held", "--reports"},
			want:   []string{"Held back:", "Open reports —"},
			unwant: []string{"Available —", "In flight —"},
		},
		{
			name:    "no filter is everything",
			args:    nil,
			wantAll: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			stdout, _, err := runList(t, listing, nil, c.args...)
			if err != nil {
				t.Fatalf("run = %v", err)
			}
			if c.wantAll && stdout != listing {
				t.Fatalf("output was edited:\n%s", stdout)
			}
			for _, w := range c.want {
				if !strings.Contains(stdout, w) {
					t.Errorf("output does not carry %q:\n%s", w, stdout)
				}
			}
			for _, u := range c.unwant {
				if strings.Contains(stdout, u) {
					t.Errorf("output carries %q, which no filter selected:\n%s", u, stdout)
				}
			}
		})
	}
}

func TestTheListersNoteSurvivesEveryFilter(t *testing.T) {
	for _, flag := range []string{"--available", "--held", "--reports"} {
		stdout, _, err := runList(t, listing, nil, flag)
		if err != nil {
			t.Fatalf("run %s = %v", flag, err)
		}
		if !strings.Contains(stdout, "could not reach GitHub") {
			t.Errorf("%s dropped the lister's note:\n%s", flag, stdout)
		}
	}
}

// A filter chooses sections, never tasks: whatever it shows is the run
// it did not filter, minus the sections it dropped (spec-0006).
func TestAFilterChangesNoGroupAndNoOrder(t *testing.T) {
	for _, c := range []struct {
		flag    string
		heading string
	}{
		{"--available", "Available —"},
		{"--held", "Held back:"},
		{"--reports", "Open reports —"},
	} {
		stdout, _, err := runList(t, listing, nil, c.flag)
		if err != nil {
			t.Fatalf("run %s = %v", c.flag, err)
		}
		want := sectionOf(t, listing, c.heading)
		got := sectionOf(t, stdout, c.heading)
		if got != want {
			t.Errorf("%s rewrote its section:\nwant:\n%s\ngot:\n%s", c.flag, want, got)
		}
	}
}

// sectionOf is the block a heading opens, up to the next blank line.
func sectionOf(t *testing.T, out, heading string) string {
	t.Helper()
	var keep []string
	in := false
	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(line, heading):
			in = true
		case in && strings.TrimSpace(line) == "":
			in = false
		}
		if in {
			keep = append(keep, line)
		}
	}
	if len(keep) == 0 {
		t.Fatalf("no section %q in:\n%s", heading, out)
	}
	return strings.Join(keep, "\n")
}

func TestAnIndentedLineNeverOpensASection(t *testing.T) {
	out := "Held back:\n  task-0009  Available — a title that reads like a heading\n"
	stdout, _, err := runList(t, out, nil, "--held")
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	if stdout != out {
		t.Errorf("a task row was read as a heading:\n%s", stdout)
	}
}

func TestWhatPrecedesTheFirstHeadingIsAlwaysPrinted(t *testing.T) {
	out := "Something new upstream added:\n  task-0011  a row\n\nHeld back:\n  task-0009  spec-0009 is draft\n"
	stdout, _, err := runList(t, out, nil, "--reports")
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	if !strings.Contains(stdout, "Something new upstream added:") {
		t.Errorf("output no filter names was dropped:\n%s", stdout)
	}
	if strings.Contains(stdout, "Held back:") {
		t.Errorf("a recognised section was printed unselected:\n%s", stdout)
	}
}

func TestKeptSectionsAreOneBlankLineApart(t *testing.T) {
	stdout, _, err := runList(t, listing, nil, "--held", "--reports")
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	want := "Held back:\n  task-0009  spec-0009 is draft\n\nOpen reports — waiting to be triaged, never selected:\n  report-0002  A thing that was noticed\n\n"
	if !strings.HasPrefix(stdout, want) {
		t.Errorf("sections are not one blank line apart:\n%q", stdout)
	}
}

func TestAFilterSelectingNothingPrintsNothing(t *testing.T) {
	stdout, _, err := runList(t, "Nothing is available.\n", exitErr(1), "--held")
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	if stdout != "" {
		t.Errorf("output = %q; want silence where the section is empty", stdout)
	}
}

func TestSplitTrimsSeparatorsAndKeepsASectionsOwnSpacing(t *testing.T) {
	blocks := split(listing)
	if len(blocks) != 6 {
		t.Fatalf("blocks = %d; want one per section", len(blocks))
	}
	for _, b := range blocks {
		if strings.TrimSpace(b.lines[len(b.lines)-1]) == "" {
			t.Errorf("a separator survived into a section: %q", b.lines)
		}
	}
	available := blocks[1].lines
	if available[len(available)-1] != "Order is a suggestion for a person and binding for an agent." {
		t.Errorf("the available section lost its closing line: %q", available)
	}
	if available[len(available)-2] != "" {
		t.Errorf("the available section lost its own blank line: %q", available)
	}
}

func TestWantsAnswersEveryGroup(t *testing.T) {
	all := sections{}
	if !all.wants(groupAvailable) || !all.wants(groupHeld) || !all.wants(groupReports) || !all.wants(groupAlways) {
		t.Error("an unfiltered run dropped a group")
	}
	if all.filtering() {
		t.Error("no flag reads as filtering")
	}
	only := sections{held: true}
	if !only.filtering() {
		t.Error("a flag does not read as filtering")
	}
	if only.wants(groupAvailable) || only.wants(groupReports) {
		t.Error("--held selected another group")
	}
	if !only.wants(groupHeld) || !only.wants(groupAlways) {
		t.Error("--held dropped its own group or the lister's notes")
	}
}
