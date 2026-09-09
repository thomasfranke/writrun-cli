package screen

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// settings is the shape the caller hands over: sections in the kit's
// order, each with the keys the file names and the values its reader
// answered.
func settings() Settings {
	return Settings{
		Header: "CONFIG · writrun/settings.json   checked by check_settings.sh",
		Groups: []SettingGroup{
			{Name: "THE STAGE", Rows: []Setting{
				{Name: "stage", Key: "stage", Value: "3"},
			}},
			{Name: "STAGE 2", Rows: []Setting{
				{Name: "auto_push", Key: "stage_2.auto_push", Value: "true"},
				{Name: "pr_title_style", Key: "stage_2.pr_title_style", Value: "bracketed"},
			}},
		},
	}
}

func newTestSettings(t *testing.T, change Change) settingsModel {
	t.Helper()
	if change == nil {
		change = func(string) { t.Error("a change ran in a case that chose none") }
	}
	load := func() (Settings, error) { return settings(), nil }
	m := newSettings(settings(), load, change, strings.NewReader(""), &strings.Builder{})
	out, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	return out.(settingsModel)
}

// The rows are the sections and keys given, and nothing else: no list
// of keys lives here, so a key the file stops naming stops being shown
// without a second edit.
func TestTheRowsAreTheGivenSettings(t *testing.T) {
	view := newTestSettings(t, nil).View()
	for _, want := range []string{"THE STAGE", "stage", "STAGE 2", "auto_push", "pr_title_style", "bracketed"} {
		if !strings.Contains(view, want) {
			t.Errorf("%q is missing from the screen:\n%s", want, view)
		}
	}
}

// Only keys are selectable — a heading is shown and skipped.
func TestOnlyKeysAreSelectable(t *testing.T) {
	m := newTestSettings(t, nil)
	if got := m.selected(); got != "stage" {
		t.Errorf("the cursor starts on %q, want the first key", got)
	}
	for _, want := range []string{"stage_2.auto_push", "stage_2.pr_title_style"} {
		out, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = out.(settingsModel)
		if got := m.selected(); got != want {
			t.Errorf("down reached %q, want %q", got, want)
		}
	}
}

// Movement stops at the ends rather than wrapping.
func TestMovementStopsAtBothEnds(t *testing.T) {
	m := newTestSettings(t, nil)
	for i := 0; i < 10; i++ {
		out, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = out.(settingsModel)
	}
	if got := m.selected(); got != "stage_2.pr_title_style" {
		t.Errorf("down ran past the end to %q", got)
	}
	for i := 0; i < 10; i++ {
		out, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
		m = out.(settingsModel)
	}
	if got := m.selected(); got != "stage" {
		t.Errorf("up ran past the start to %q", got)
	}
}

// The detail line names the selected key and its value.
func TestTheDetailLineNamesTheSelection(t *testing.T) {
	out, _ := newTestSettings(t, nil).Update(tea.KeyMsg{Type: tea.KeyDown})
	if view := out.(settingsModel).View(); !strings.Contains(view, "auto_push — true") {
		t.Errorf("the detail line does not name the selection:\n%s", view)
	}
}

// `enter` asks for the change and does not perform it here: the model
// decides, and the change runs on a terminal this screen released.
func TestEnterAsksForTheChange(t *testing.T) {
	_, cmd := newTestSettings(t, func(string) {}).Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("enter on a key asked for nothing")
	}
}

// `q` and `esc` both leave. This screen is what `writrun config` opens,
// so leaving it ends the command — there is nothing behind it.
func TestQAndEscBothLeave(t *testing.T) {
	for _, k := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'q'}},
		{Type: tea.KeyEsc},
	} {
		if _, cmd := newTestSettings(t, nil).Update(k); cmd == nil {
			t.Errorf("%v did not leave", k)
		}
	}
}

// After a change the rows are read again rather than edited here:
// whether the change was kept is the file's answer, and the kit's
// checker may have refused it and put the previous bytes back.
func TestTheSettingsAreReReadAfterAChange(t *testing.T) {
	reads := 0
	load := func() (Settings, error) {
		reads++
		s := settings()
		s.Groups[1].Rows[0].Value = "false" // what the file now says
		return s, nil
	}
	m := newSettings(settings(), load, func(string) {}, strings.NewReader(""), &strings.Builder{})
	out, _ := m.Update(changedMsg{})
	m = out.(settingsModel)
	if reads != 1 {
		t.Errorf("the settings were read %d times, want once after the change", reads)
	}
	if !strings.Contains(m.View(), "false") {
		t.Errorf("the screen shows a remembered value, not the file's:\n%s", m.View())
	}
}

// A reader's place survives a change, so a refusal does not move them
// off the key they were changing.
func TestThePlaceSurvivesAChange(t *testing.T) {
	load := func() (Settings, error) { return settings(), nil }
	m := newSettings(settings(), load, func(string) {}, strings.NewReader(""), &strings.Builder{})
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = out.(settingsModel)
	was := m.selected()

	out, _ = m.Update(changedMsg{})
	if got := out.(settingsModel).selected(); got != was {
		t.Errorf("the cursor moved from %q to %q across a change", was, got)
	}
}

// Settings that cannot be read are said so, and the screen stays open:
// a file that cannot be read is not a reason to close on somebody
// mid-change.
func TestSettingsThatCannotBeReadAreNamed(t *testing.T) {
	load := func() (Settings, error) { return Settings{}, errors.New("the reader refused") }
	m := newSettings(settings(), load, func(string) {}, strings.NewReader(""), &strings.Builder{})
	out, cmd := m.Update(changedMsg{})
	if cmd != nil {
		t.Error("a failed read ended the screen")
	}
	if !strings.Contains(out.(settingsModel).View(), "the reader refused") {
		t.Errorf("the cause was not shown:\n%s", out.(settingsModel).View())
	}
}

// OpenSettings drives the real program, and answers what load answers
// when it cannot read at all — there is no screen to open over nothing.
func TestOpenSettingsCarriesTheFirstReadsFailure(t *testing.T) {
	err := OpenSettings(
		func() (Settings, error) { return Settings{}, errors.New("no settings file") },
		func(string) { t.Error("a change ran though the screen never opened") },
		strings.NewReader("q"), &strings.Builder{},
	)
	if err == nil || !strings.Contains(err.Error(), "no settings file") {
		t.Errorf("OpenSettings = %v, want the reader's own failure", err)
	}
}

// And it runs, and leaves on q.
func TestOpenSettingsRunsAndLeaves(t *testing.T) {
	err := OpenSettings(
		func() (Settings, error) { return settings(), nil },
		func(string) { t.Error("a change ran in a case that chose none") },
		strings.NewReader("q"), &strings.Builder{},
	)
	if err != nil {
		t.Fatalf("OpenSettings = %v", err)
	}
}

// What the screen shows and what it hands back are not the same thing.
//
// The drawing shows `auto_push`; the kit knows it as
// `stage_2.auto_push`, and only the second identifies a key to change.
// A screen that handed back its label would name a key the kit cannot
// find, and the change would fail after the reader had already typed
// the value.
func TestTheChangeIsAskedForTheKeyAndNotTheLabel(t *testing.T) {
	var asked string
	m := newTestSettings(t, func(key string) { asked = key })
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = out.(settingsModel)

	if view := m.View(); !strings.Contains(view, "auto_push — true") {
		t.Errorf("the detail line does not show the label:\n%s", view)
	}
	if got := m.selected(); got != "stage_2.auto_push" {
		t.Fatalf("the selection is %q, want the key the kit knows", got)
	}

	// The change runs on a released terminal, so reach it directly.
	d := dispatch{label: "x", run: func() { m.change(m.selected()) },
		in: strings.NewReader("\n"), out: &strings.Builder{}}
	if err := d.Run(); err != nil {
		t.Fatalf("Run = %v", err)
	}
	if asked != "stage_2.auto_push" {
		t.Errorf("the change was asked for %q, want the key and not the label", asked)
	}
}
