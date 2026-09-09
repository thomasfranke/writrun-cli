package screen

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// sample is the shape the caller hands over: groups of commands, each
// carrying the summary the command table already holds.
func sample() Entry {
	return Entry{
		Header: "writrun-cli v0.1.0 · pins WritRun v0.0.07 · branch main",
		Stage:  []string{"STAGE 3", "  commit ask · push auto · pull request ask"},
		Groups: []Group{
			{Name: "tasks", Rows: []Command{
				{"list", "the queue: available, held back, untriaged"},
				{"take", "begin a task: branch pushed, draft PR opened"},
			}},
			{Name: "adoption", Rows: []Command{
				{"doctor", "what the methodology assumes, checked"},
			}},
		},
	}
}

func keys(m tea.Model, ks ...tea.KeyMsg) tea.Model {
	for _, k := range ks {
		m, _ = m.Update(k)
	}
	return m
}

var (
	down  = tea.KeyMsg{Type: tea.KeyDown}
	enter = tea.KeyMsg{Type: tea.KeyEnter}
)

// Only command rows are selectable: the header, the stage block and the
// group headings are shown and skipped, the way the queue's own
// headings are.
func TestOnlyCommandRowsAreSelectable(t *testing.T) {
	m := newEntry(sample())
	if got := m.selected(); got != "list" {
		t.Errorf("the screen opens on %q, want the first command", got)
	}
	m = keys(m, down).(entryModel)
	if got := m.selected(); got != "take" {
		t.Errorf("down reached %q, want the next command", got)
	}
	// Two downs cross a blank and a heading to reach the next group.
	m = keys(m, down).(entryModel)
	if got := m.selected(); got != "doctor" {
		t.Errorf("down reached %q, want the first command of the next group", got)
	}
}

// The screen lists what the caller gives it and nothing else — a
// command added to the table appears without a second list being
// edited, and one the caller withholds does not.
func TestTheRowsAreTheGivenCommands(t *testing.T) {
	e := sample()
	e.Groups[0].Rows = append(e.Groups[0].Rows, Command{"finish", "complete a task: deltas checked, PR ready"})
	view := newEntry(e).View()
	if !strings.Contains(view, "finish") {
		t.Error("a command the caller listed is missing from the screen")
	}
	if strings.Contains(view, "init") {
		t.Error("the screen shows a command the caller did not list")
	}
}

// The row shows what fits; the detail line shows what the command table
// says, whole.
func TestTheDetailLineIsTheWholeSummary(t *testing.T) {
	m := newEntry(sample())
	if !strings.Contains(m.View(), "list — the queue: available, held back, untriaged") {
		t.Errorf("the detail line does not carry the selected summary:\n%s", m.View())
	}
	m = keys(m, down).(entryModel)
	if !strings.Contains(m.View(), "take — begin a task") {
		t.Errorf("the detail line did not follow the cursor:\n%s", m.View())
	}
}

// `list` opens a screen; every other row runs a command. The
// difference is the whole of what the entry screen decides.
func TestEnterOnListOpensTheQueueRatherThanRunning(t *testing.T) {
	m := keys(newEntry(sample()), enter).(entryModel)
	if !m.queue {
		t.Error("enter on list did not open the queue")
	}
	if m.action != (Action{}) {
		t.Errorf("enter on list dispatched %v — it opens a screen, it runs nothing", m.action)
	}
}

func TestEnterOnACommandDispatchesIt(t *testing.T) {
	m := keys(newEntry(sample()), down, down, enter).(entryModel)
	if m.queue {
		t.Error("enter on a command opened the queue")
	}
	if m.action != (Action{Command: "doctor"}) {
		t.Errorf("action = %v, want doctor", m.action)
	}
}

func TestQuitLeavesTheEntryScreenWithNothing(t *testing.T) {
	m := keys(newEntry(sample()), tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}).(entryModel)
	if m.action != (Action{}) || m.queue {
		t.Errorf("q resolved to %v/queue=%v, want neither", m.action, m.queue)
	}
}

// esc on the queue is the way back: it is not an action and not a
// departure, so the caller shows the entry screen again.
func TestEscOnTheQueueGoesBack(t *testing.T) {
	m := newModel(Parse("Available — any of these may be taken:\n  task-0020  A thing\n"))
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	q := out.(model)
	if !q.back {
		t.Error("esc did not ask to go back")
	}
	if q.action != (Action{}) {
		t.Errorf("esc dispatched %v — going back runs nothing", q.action)
	}
}
