package screen

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Every screen's frames, asserted against what the model renders. The
// drawing is not the thing under test: it is what the test is written
// against (docs/product/screens/README.md).
//
// The values in each fixture are ones this repository's own files
// produce: the summaries are the command table's, the marks are
// `doctor`'s, and the rows are the lister's own output.

// drawnIdentity is the line every frame opens with.
const drawnIdentity = "writrun-cli v0.0.2 · pins WritRun v0.0.08 · branch main"

// drawnWidth is the terminal the entry, queue and settings frames are
// drawn at: a 69-column rule and the column of padding either side.
const drawnWidth = 71

// atWidth drives one window width into a model and answers it back. The
// height is left unsaid, so every row renders and the frame is the
// whole screen rather than a window over it.
func atWidth[M tea.Model](t *testing.T, m M, width int) M {
	t.Helper()
	next, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: 0})
	return next.(M)
}

// ---------------------------------------------------------------- entry

// drawnEntry is the entry screen's own frame: the four groups in the
// order `cmd/writrun/main.go` lists them, each row carrying the summary
// the command table holds.
func drawnEntry() Entry {
	return Entry{
		Identity: drawnIdentity,
		Context:  "STAGE 3 · GitHub issues",
		Source:   "writrun/settings.json",
		Conduct:  []string{"  commit ask · push auto · pull request ask · titles bracketed"},
		Groups: []Group{
			{Name: "tasks", Rows: []Command{
				{Name: "list", Summary: "see what work is waiting, and what is blocked"},
				{Name: "take", Summary: "start work on a task, in one act"},
				{Name: "work", Summary: "hand the next task to your AI agent"},
				{Name: "status", Summary: "where does the work on this branch stand?"},
				{Name: "finish", Summary: "close the work and mark the pull request ready"},
			}},
			{Name: "authoring", Rows: []Command{
				{Name: "author", Summary: "publish a finished rule and the work it created"},
				{Name: "amend", Summary: "reopen a spec that turned out wrong"},
			}},
			{Name: "reports", Rows: []Command{
				{Name: "report", Summary: "write down something you noticed, without stopping"},
			}},
			{Name: "adoption", Rows: []Command{
				{Name: "doctor", Summary: "check this repository still satisfies WritRun"},
				{Name: "config", Summary: "read and change the settings WritRun reads"},
				{Name: "update", Summary: "bring WritRun up to the version this binary pins"},
				{Name: "uninstall", Summary: "remove WritRun, keep everything it helped you write"},
			}},
		},
	}
}

func TestTheEntryScreenFrame(t *testing.T) {
	m := atWidth(t, newEntry(drawnEntry()), drawnWidth)
	sameFrame(t, rendered(m.View()), frame(t, entryDrawing, "writrun — the entry screen"))
}

// ---------------------------------------------------------------- queue

// drawnQueue is the lister's output the queue frames are drawn from:
// one task in progress, one available, two held back, one report.
const drawnQueue = `In progress — resume before selecting anything new:
  0023       Two homes, caught up
             paused — spec-0022 is missing here; the work
             waits on the re-approval

Available — any of these may be taken:
  0024       low     Give the fetch a seam

Order is a suggestion for a person and binding for an agent.

Held back:
  0021       MISMATCH — stored ready but spec-0020 is missing
  0022       blocked: waiting on spec-0021

Open reports — waiting to be triaged, never selected:
  0031       The suite inherits the terminal`

// emptyListing is the lister answering that nothing may be taken, with
// the note that qualifies the whole answer.
const emptyListing = `Nothing is available.

Note: could not reach GitHub, so nothing above accounts for
work already in flight.`

// queueAt is the queue screen with the cursor moved onto a task.
func queueAt(t *testing.T, listing, id string) model {
	t.Helper()
	m := atWidth(t, newModel(drawnIdentity, Parse(listing)), drawnWidth)
	for i, r := range m.rows {
		if r.ID == id {
			m.cursor = i
			return m
		}
	}
	t.Fatalf("the listing holds no row for %q", id)
	return m
}

func TestTheQueueFrameWithATaskSelected(t *testing.T) {
	m := queueAt(t, drawnQueue, "task-0024")
	drawn := frame(t, listDrawing, "the queue screen — a task selected")
	sameRowsAndKeys(t, rendered(m.View()), drawn)
}

func TestTheQueueFrameWithNothingInIt(t *testing.T) {
	m := atWidth(t, newModel(drawnIdentity, Parse(emptyListing)), drawnWidth)
	sameFrame(t, rendered(m.View()), frame(t, listDrawing, "the queue with nothing in it"))
}

// The worked example is the furniture itself, annotated under the
// footer with the three groups it is in. Its rows illustrate; its
// header and its footer are what it states, and those are asserted.
func TestTheFurnitureFrameIsTheHeaderAndTheFooter(t *testing.T) {
	m := queueAt(t, drawnQueue, "task-0024")
	got := rendered(m.View())
	drawn := frame(t, entryDrawing, "the furniture, on the queue screen as the worked example")
	for len(drawn) > 0 && strings.HasPrefix(strings.TrimSpace(drawn[len(drawn)-1]), "└") {
		drawn = drawn[:len(drawn)-1]
	}
	if got[0] != drawn[0] {
		t.Errorf("the identity line differs\n  drawn:    %q\n  rendered: %q", drawn[0], got[0])
	}
	if !strings.HasPrefix(got[1], " QUEUE · work/tasks/ — ") {
		t.Errorf("the context line does not name what the screen reads: %q", got[1])
	}
	if keyLine(got) != keyLine(drawn) {
		t.Errorf("the keys differ\n  drawn:    %q\n  rendered: %q", keyLine(drawn), keyLine(got))
	}
}

// A queue that could not be read is a screen, not an error: the failure
// is marked as `doctor` marks one, and the commands that answer it are
// named (spec-0042).
func TestTheQueueThatCouldNotBeRead(t *testing.T) {
	s := session{
		identity: drawnIdentity,
		width:    drawnWidth,
		where:    atQueue,
		err: errors.New(".writrun/skills/writrun-select-next-task/list_tasks.sh: " +
			"No such file or directory"),
	}
	sameUnreadable(t, rendered(s.View()), frame(t, entryDrawing, "the queue could not be read"))
}

// ------------------------------------------------------------- settings

// drawnSettings is the settings file the config frames are drawn from,
// in the sections the kit gives and with the values its reader answered.
func drawnSettings() Settings {
	return Settings{
		Identity: drawnIdentity,
		Context:  "CONFIG · writrun/settings.json",
		Source:   "checked by check_settings.sh",
		Checker:  "check_settings.sh",
		Groups: []SettingGroup{
			{Name: "THE STAGE", Rows: []Setting{
				{Name: "stage", Key: "stage", Value: "3"},
			}},
			{Name: "STAGE 1 — declared, and read by agents alone", Rows: []Setting{
				{Name: "decisions_style", Key: "stage_1.decisions_style", Value: "chronological"},
				{Name: "product_layout", Key: "stage_1.product_layout", Value: "by-feature"},
				{Name: "provenance_ledger", Key: "stage_1.provenance_ledger", Value: "false"},
				{Name: "spec_required", Key: "stage_1.spec_required", Value: "when-warranted"},
			}},
			{Name: "STAGE 2 — the conduct flags, and the title style", Rows: []Setting{
				{Name: "agent_coauthor", Key: "stage_2.agent_coauthor", Value: "false"},
				{Name: "auto_commit", Key: "stage_2.auto_commit", Value: "false"},
				{Name: "auto_pr", Key: "stage_2.auto_pr", Value: "false"},
				{Name: "auto_push", Key: "stage_2.auto_push", Value: "true"},
				{Name: "pr_title_style", Key: "stage_2.pr_title_style", Value: "bracketed"},
			}},
		},
	}
}

func TestTheSettingsFrame(t *testing.T) {
	s := drawnSettings()
	m := newSettings(s, func() (Settings, error) { return s, nil }, func(string) {},
		strings.NewReader(""), &strings.Builder{})
	m = atWidth(t, m, drawnWidth)
	sameRowsAndKeys(t, rendered(m.View()), frame(t, configDrawing, "writrun config"))
}

func TestTheSettingsThatCouldNotBeRead(t *testing.T) {
	s := drawnSettings()
	m := newSettings(s, func() (Settings, error) { return s, nil }, func(string) {},
		strings.NewReader(""), &strings.Builder{})
	m = atWidth(t, m, drawnWidth)
	m.err = errors.New("writrun/settings.json: No such file or directory")
	sameUnreadable(t, rendered(m.View()), frame(t, configDrawing, "the settings could not be read"))
}

// ------------------------------------------------------------ first run

// firstRunWidth is the terminal the first-run frames are drawn at: a
// 62-column rule and the column of padding either side.
const firstRunWidth = 64

// initSummary is the command table's own, which is what the row shows —
// one field, three surfaces (report-0042).
const initSummary = "install WritRun into this repository"

const initExplain = "init — installs the WritRun kit into this repository. It asks " +
	"the stage, extracts this repository's own conventions, grafts an existing " +
	"AGENTS.md, and leaves the queue empty."

const versionExplain = "--version — the product, the version of this binary, and the " +
	"WritRun tag it pins. It runs anywhere, adopted or not."

const helpExplain = "--help — every command with the one line that says what it is " +
	"for, and the address of the docs. It runs anywhere."

func binaryGroup() FirstRunGroup {
	return FirstRunGroup{Name: "THIS BINARY", Rows: []FirstRunRow{
		{Name: "--version", Text: "the product, its version, and the tag it pins", Explain: versionExplain},
		{Name: "--help", Text: "one line per command, and where the docs live", Explain: helpExplain},
	}}
}

func TestTheFirstRunFrameWithTheEnvironmentMet(t *testing.T) {
	f := FirstRun{
		Identity: drawnIdentity,
		Context:  "NO KIT HERE · this repository has no .writrun/",
		Ready:    true,
		Groups: []FirstRunGroup{
			{Name: "ENVIRONMENT — 4 of 4 met", Rows: []FirstRunRow{
				{Mark: "✓", Name: "git", Explain: "git — reason.", Selects: true},
				{Mark: "✓", Name: "bash", Explain: "bash — reason.", Selects: true},
				{Mark: "✓", Name: "awk", Explain: "awk — reason.", Selects: true},
				{Mark: "✓", Name: "sed", Explain: "sed — reason.", Selects: true},
			}},
			{Name: "ADOPTION", Rows: []FirstRunRow{
				{Name: "init", Text: initSummary, Runs: "init", Selects: true, Explain: initExplain},
			}},
			binaryGroup(),
		},
	}
	m := atWidth(t, newFirstRun(f, nil), firstRunWidth)
	// The cursor rests on `init`, which is the row the frame selects.
	for i, r := range m.rows {
		if r.runs == "init" {
			m.cursor = i
		}
	}
	sameFrame(t, rendered(m.View()), frame(t, firstRunDrawing, "writrun — a repository with no kit in it"))
}

func TestTheFirstRunFrameWithTheEnvironmentShort(t *testing.T) {
	const awkExplain = "awk — the kit's scripts read the queue's front matter with " +
		"POSIX awk, and it is not on the PATH. Install it, then press `r` to read " +
		"the PATH again. `init` is out of reach until all four are met: every flow " +
		"it installs runs through these four, so adopting without them installs a " +
		"kit that cannot run."
	f := FirstRun{
		Identity: "writrun-cli v0.0.2 · pins WritRun v0.0.08",
		Context:  "NO KIT HERE · this repository has no .writrun/",
		Ready:    false,
		Groups: []FirstRunGroup{
			{Name: "ENVIRONMENT — 3 of 4 met", Rows: []FirstRunRow{
				{Mark: "✓", Name: "git", Explain: "git — reason.", Selects: true},
				{Mark: "✓", Name: "bash", Explain: "bash — reason.", Selects: true},
				{Mark: "✗", Name: "awk", Text: "the kit's scripts require it",
					Explain: awkExplain, Selects: true},
				{Mark: "✓", Name: "sed", Explain: "sed — reason.", Selects: true},
			}},
			{Name: "ADOPTION", Rows: []FirstRunRow{
				{Name: "init", Text: "out of reach — 1 environment requirement unmet",
					Selects: true, Explain: initExplain},
			}},
			binaryGroup(),
		},
	}
	m := atWidth(t, newFirstRun(f, nil), firstRunWidth)
	for i, r := range m.rows {
		if strings.Contains(r.text, "awk") {
			m.cursor = i
		}
	}
	sameFrame(t, rendered(m.View()), frame(t, firstRunDrawing, "writrun — the environment short, init out of reach"))
}

// -------------------------------------------------- a command running

// A captured command answers into a screen of its own, and while it
// works the screen says so and offers no key. The prose the frame draws
// under the spinner is the drawing's note about why the spinner exists,
// not a line the screen writes.
func TestTheFrameWhileACapturedCommandRuns(t *testing.T) {
	s := session{
		identity: drawnIdentity,
		width:    drawnWidth,
		where:    atRunning,
		running:  "list",
		frame:    2,
	}
	got := rendered(s.View())
	drawn := frame(t, entryDrawing, "a command that asks nothing, while it runs")
	sameLines(t, got[:4], drawn[:4])
	if keyLine(got) != keyLine(drawn) {
		t.Errorf("the keys differ\n  drawn:    %q\n  rendered: %q", keyLine(drawn), keyLine(got))
	}
}

// The pager is a screen: the same two lines above the command's output
// and the same way out below it. What the frame draws between them is
// the command's own, elided on the canvas.
func TestThePagerFrameCarriesTheSameFurniture(t *testing.T) {
	// The frame draws an output taller than its window — it says so,
	// eliding seven lines — so the footer offers the scroll.
	var out strings.Builder
	for i := 0; i < 20; i++ {
		out.WriteString("  remove       a line of the plan\n")
	}
	p := newPager(drawnIdentity, "uninstall", out.String())
	next, _ := p.Update(tea.WindowSizeMsg{Width: drawnWidth, Height: 14})
	got := rendered(next.(pager).View())
	drawn := frame(t, entryDrawing, "the same two lines around a command's output")
	sameLines(t, got[:2], drawn[:2])
	if keyLine(got) != keyLine(drawn) {
		t.Errorf("the keys differ\n  drawn:    %q\n  rendered: %q", keyLine(drawn), keyLine(got))
	}
}

// ------------------------------------------------------------- helpers

// sameRowsAndKeys holds the rows and the footer's keys to the drawing
// and leaves the explanation alone. It is for a frame whose explanation
// states a fact the rows it is drawn from do not carry — the difference
// is recorded rather than asserted, because a test that asserted it
// would be asserting the drawing against itself.
func sameRowsAndKeys(t *testing.T, got, want []string) {
	t.Helper()
	gotRows, _, gotKeys := parts(got)
	wantRows, _, wantKeys := parts(want)
	sameLines(t, gotRows, wantRows)
	if gotKeys != wantKeys {
		t.Errorf("the keys differ\n  drawn:    %q\n  rendered: %q", wantKeys, gotKeys)
	}
}

// sameUnreadable holds a screen that could not read what it shows: the
// header, the marked failure, the message and the explanation — the
// last two as sentences, because where they break is the terminal's
// width and the canvas's.
func sameUnreadable(t *testing.T, got, want []string) {
	t.Helper()
	gotRows, gotSays, gotKeys := parts(got)
	wantRows, wantSays, wantKeys := parts(want)
	sameLines(t, gotRows[:3], wantRows[:3])
	if g, w := sentence(gotRows[3:]), sentence(wantRows[3:]); g != w {
		t.Errorf("the message differs\n  drawn:    %q\n  rendered: %q", w, g)
	}
	if gotSays != wantSays {
		t.Errorf("the explanation differs\n  drawn:    %q\n  rendered: %q", wantSays, gotSays)
	}
	if gotKeys != wantKeys {
		t.Errorf("the keys differ\n  drawn:    %q\n  rendered: %q", wantKeys, gotKeys)
	}
}

// sentence joins wrapped lines back into the one sentence they are.
func sentence(lines []string) string {
	var said []string
	for _, l := range lines {
		if s := strings.TrimSpace(l); s != "" {
			said = append(said, s)
		}
	}
	return strings.Join(said, " ")
}
