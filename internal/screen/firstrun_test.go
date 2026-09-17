package screen

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// shortEnvironment is the first run where one requirement is unmet:
// `init` is out of reach, the row says why, and the cursor rests on the
// requirement that put it there (spec-0038).
func shortEnvironment() FirstRun {
	return FirstRun{
		Identity: drawnIdentity,
		Context:  "NO KIT HERE · this repository has no .writrun/",
		Ready:    false,
		Groups: []FirstRunGroup{
			{Name: "ENVIRONMENT — 3 of 4 met", Rows: []FirstRunRow{
				{Mark: "✓", Name: "git", Explain: "git — reason.", Selects: true},
				{Mark: "✗", Name: "awk", Text: "the kit's scripts require it",
					Explain: "awk — it is not on the PATH.", Selects: true},
			}},
			{Name: "ADOPTION", Rows: []FirstRunRow{
				// The row still names the command it would run: what
				// holds it back is the screen's gate on the
				// environment, not a row emptied by the caller.
				{Name: "init", Text: "out of reach — 1 environment requirement unmet",
					Runs: "init", Selects: true, Explain: initExplain},
			}},
		},
	}
}

func metEnvironment() FirstRun {
	f := shortEnvironment()
	f.Ready = true
	f.Groups[0].Rows[1] = FirstRunRow{Mark: "✓", Name: "awk", Explain: "awk — reason.", Selects: true}
	f.Groups[1].Rows[0] = FirstRunRow{
		Name: "init", Text: "install WritRun into this repository",
		Runs: "init", Selects: true, Explain: initExplain,
	}
	return f
}

// `enter` on `init` runs `init` where the environment is met, and
// nothing where it is not. A row held out of reach is the whole of what
// this screen judges, and it judges it on the probe's answer.
func TestInitIsOutOfReachWhileARequirementIsUnmet(t *testing.T) {
	short := newFirstRun(shortEnvironment(), nil)
	for i, r := range short.rows {
		if strings.Contains(r.text, "init") {
			short.cursor = i
		}
	}
	out, _ := short.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if got := out.(firstRunModel).chosen; got != "" {
		t.Errorf("enter ran %q with a requirement unmet", got)
	}

	met := newFirstRun(metEnvironment(), nil)
	for i, r := range met.rows {
		if strings.Contains(r.text, "init") {
			met.cursor = i
		}
	}
	out, _ = met.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if got := out.(firstRunModel).chosen; got != "init" {
		t.Errorf("enter chose %q with every requirement met, want init", got)
	}
}

// The row says why it cannot run, rather than failing when pressed.
func TestTheUnreachableRowSaysWhy(t *testing.T) {
	view := newFirstRun(shortEnvironment(), nil).View()
	if !strings.Contains(view, "out of reach — 1 environment requirement unmet") {
		t.Errorf("the row does not say why init cannot run:\n%s", view)
	}
}

// The footer drops `enter` while nothing can run, and offers the key
// that reads the PATH again instead.
func TestTheFooterDropsEnterWhileNothingCanRun(t *testing.T) {
	short := keyLine(rendered(newFirstRun(shortEnvironment(), nil).View()))
	if strings.Contains(short, "enter") {
		t.Errorf("the footer offers enter with nothing to run: %q", short)
	}
	if !strings.Contains(short, "r re-check") {
		t.Errorf("the footer does not offer the re-check: %q", short)
	}
	met := keyLine(rendered(newFirstRun(metEnvironment(), nil).View()))
	if !strings.Contains(met, "enter run") {
		t.Errorf("the footer does not offer enter with everything met: %q", met)
	}
}

// `r` reads the environment again, and a requirement installed while
// the screen is open is met the moment it is asked for.
func TestTheRecheckReadsTheEnvironmentAgain(t *testing.T) {
	// The screen opens on an environment one short; by the time `r` is
	// pressed the requirement has been installed.
	reads := 0
	load := func() (FirstRun, error) {
		reads++
		return metEnvironment(), nil
	}
	m := newFirstRun(shortEnvironment(), load)
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = out.(firstRunModel)
	if reads != 1 {
		t.Errorf("the environment was read %d times, want once", reads)
	}
	if !m.ready {
		t.Error("the re-check did not read the environment again")
	}
	if !strings.Contains(keyLine(rendered(m.View())), "enter run") {
		t.Error("the footer did not follow the re-check")
	}
}

// A probe that could not be made is said, not swallowed.
func TestAFailedRecheckIsSaid(t *testing.T) {
	m := newFirstRun(shortEnvironment(), func() (FirstRun, error) {
		return FirstRun{}, errors.New("the PATH could not be read")
	})
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if !strings.Contains(out.(firstRunModel).View(), "the PATH could not be read") {
		t.Error("a failed re-check said nothing")
	}
}

// The wordmark is this screen's and no other's: it is the one screen a
// reader reaches without having adopted anything.
func TestTheWordmarkIsDrawnHereAndNowhereElse(t *testing.T) {
	if !strings.Contains(newFirstRun(metEnvironment(), nil).View(), wordmark) {
		t.Error("the first run does not introduce the product")
	}
	if strings.Contains(atWidth(t, newEntry(drawnEntry()), drawnWidth).View(), wordmark) {
		t.Error("the entry screen draws the wordmark, which is the first run's alone")
	}
}
