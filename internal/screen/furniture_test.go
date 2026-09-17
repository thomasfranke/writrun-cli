package screen

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Every screen opens with the same two lines and ends with the same
// one. Seven screens wrote their own and no two agreed, which is what
// report-0035 recorded and spec-0042 answers.

// screens are every screen this package renders, each with the two
// lines it must open with and the footer it must end on.
func screens(t *testing.T) []struct {
	name string
	view string
	back bool
} {
	t.Helper()
	entry := atWidth(t, newEntry(drawnEntry()), drawnWidth)
	queue := queueAt(t, drawnQueue, "task-0024")
	pager := atWidth(t, newPager(drawnIdentity, "doctor", "all clear\n"), drawnWidth)
	set := drawnSettings()
	settings := atWidth(t, newSettings(set, func() (Settings, error) { return set, nil },
		func(string) {}, strings.NewReader(""), &strings.Builder{}), drawnWidth)
	first := atWidth(t, newFirstRun(FirstRun{
		Identity: drawnIdentity,
		Context:  "NO KIT HERE · this repository has no .writrun/",
		Ready:    true,
		Groups: []FirstRunGroup{{Name: "ADOPTION", Rows: []FirstRunRow{
			{Name: "init", Text: "install WritRun into this repository",
				Runs: "init", Selects: true, Explain: initExplain},
		}}},
	}, nil), firstRunWidth)
	running := session{identity: drawnIdentity, width: drawnWidth, where: atRunning, running: "list"}
	unreadable := session{identity: drawnIdentity, width: drawnWidth, where: atQueue,
		err: errors.New("the script is not there")}

	return []struct {
		name string
		view string
		back bool
	}{
		{"the entry screen", entry.View(), false},
		{"the queue", queue.View(), true},
		{"the pager", pager.View(), true},
		{"the settings", settings.View(), true},
		{"the first run", first.View(), false},
		{"the queue that could not be read", unreadable.View(), true},
		{"a captured command running", running.View(), false},
	}
}

// The identity line and a context line, on every screen. A screen that
// named neither is the queue's old opening: the lister's first row
// (report-0035).
func TestEveryScreenOpensWithTheSameTwoLines(t *testing.T) {
	for _, s := range screens(t) {
		lines := rendered(s.view)
		if len(lines) < 2 {
			t.Errorf("%s renders fewer than two lines", s.name)
			continue
		}
		// The first run introduces the product above the header; every
		// other screen opens on it.
		if s.name == "the first run" {
			lines = lines[3:]
		}
		if lines[0] != " "+drawnIdentity {
			t.Errorf("%s does not open on the identity line: %q", s.name, lines[0])
		}
		if !strings.Contains(lines[1], " · ") {
			t.Errorf("%s has no context line naming what it reads: %q", s.name, lines[1])
		}
	}
}

// `esc` goes back and `q` quits. A screen with nowhere to go back to
// offers `q quit` alone, and no screen calls `q` a way back — the
// config screen did, over a key that quit (report-0035).
func TestEveryFooterEndsInOneOfTheTwoForms(t *testing.T) {
	for _, s := range screens(t) {
		last := keyLine(rendered(s.view))
		if s.name == "a captured command running" {
			// Nothing can be pressed while a command is running, so the
			// footer offers no way out to press.
			if strings.Contains(last, "quit") {
				t.Errorf("%s offers a key over a command it must not interrupt: %q", s.name, last)
			}
			continue
		}
		want := "q quit"
		if s.back {
			want = "esc back · q quit"
		}
		if !strings.HasSuffix(last, want) {
			t.Errorf("%s ends %q, want it to end %q", s.name, last, want)
		}
		if !s.back && strings.Contains(last, "esc") {
			t.Errorf("%s offers esc with nowhere to go back to: %q", s.name, last)
		}
		if strings.Contains(last, "q back") {
			t.Errorf("%s calls q a way back over a key that quits: %q", s.name, last)
		}
	}
}

// A question carries no header and no `q`: it is not a screen a reader
// arrived at, and `q` over a list of options would be a value.
func TestAQuestionCarriesNoHeaderAndNoQuit(t *testing.T) {
	lines := Draw(Choice{
		Title: "Which stage?",
		Rows:  []Line{{Text: "  1   files", Detail: "1 — files.", Selects: true}},
		Verb:  "choose",
	}, drawnWidth, 0)
	if strings.Contains(lines[0], "writrun-cli") {
		t.Errorf("a question opened with a header: %q", lines[0])
	}
	last := keyLine(lines)
	if !strings.HasSuffix(last, "esc cancels") {
		t.Errorf("a question ends %q, want it to end \"esc cancels\"", last)
	}
	if strings.Contains(last, "q ") {
		t.Errorf("a question offers q, which over a list of options is a value: %q", last)
	}
}

// The three groups, in one order: movement, the actions one key each,
// the way out.
func TestTheFooterIsThreeGroupsInOneOrder(t *testing.T) {
	got := footer{movement: "↑↓ move", actions: []string{"enter take", "w work"}, way: backOrQuit}.line()
	want := " ↑↓ move · enter take · w work · esc back · q quit"
	if got != want {
		t.Errorf("footer = %q, want %q", got, want)
	}
	if got := (footer{way: silent}).line(); got != " " {
		t.Errorf("a screen with no key to press drew %q, want nothing", got)
	}
}

// The context line names a source at the right edge, and the header
// never wraps into the rows.
func TestTheContextLineRightAlignsItsSource(t *testing.T) {
	lines := chrome{identity: "writrun-cli", context: "CONFIG · a.json", source: "checked by b.sh"}.lines(40)
	if columns(lines[1]) != 41 {
		t.Errorf("the context line is %d columns, want 41: %q", columns(lines[1]), lines[1])
	}
	if !strings.HasSuffix(lines[1], "checked by b.sh") {
		t.Errorf("the source is not at the right edge: %q", lines[1])
	}
	narrow := chrome{identity: strings.Repeat("x", 50), context: "c"}.lines(10)
	if columns(narrow[0]) != 11 {
		t.Errorf("a header wider than the terminal was not truncated: %q", narrow[0])
	}
}

// A dispatched command's terminal is a screen too: the same two lines
// above it, and the reason it was handed the terminal at all.
func TestTheHandedTerminalCarriesTheSameFurniture(t *testing.T) {
	var out strings.Builder
	d := dispatch{
		label:    "take",
		identity: drawnIdentity,
		run:      func() {},
		in:       strings.NewReader("\n"),
		out:      &out,
	}
	if err := d.Run(); err != nil {
		t.Fatalf("Run = %v", err)
	}
	got := rendered(strings.ReplaceAll(strings.ReplaceAll(out.String(), altOn, ""), altOff, ""))
	drawn := frame(t, entryDrawing, "a command that asks, taking the terminal")
	sameLines(t, got[:4], drawn[:4])
	// The frame's closing paragraph is its note about what the pause is,
	// not a line the terminal carries; the sentence above it is.
	if g, w := sentence(until(got[4:])), sentence(until(drawn[5:])); g != w {
		t.Errorf("the reason differs\n  drawn:    %q\n  rendered: %q", w, g)
	}
}

// until is the block of lines up to the next blank.
func until(lines []string) []string {
	for i, l := range lines {
		if strings.TrimSpace(l) == "" && i > 0 {
			return lines[:i]
		}
	}
	return lines
}

// A window size is not needed for a screen to render every row: a
// terminal that never said how tall it is gets all of them.
func TestAScreenRendersWithoutAWindowSize(t *testing.T) {
	m := newEntry(drawnEntry())
	if _, ok := any(m).(tea.Model); !ok {
		t.Fatal("the entry screen is not a model")
	}
	if !strings.Contains(m.View(), "uninstall") {
		t.Error("a screen that was never sized hid its last row")
	}
}
