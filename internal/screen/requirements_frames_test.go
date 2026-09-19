package screen

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// The two doctor-screen frames, asserted against what the model renders.
// The drawing is not the thing under test: it is what the test is
// written against (docs/product/screens/README.md).
//
// The rows are compared line for line. The footer is compared as one
// paragraph, because where it wraps is the terminal's width and the
// canvas's — the sentence is the binary's.

// drawnRows is the repository both screen frames are drawn from: stage
// 1 declared, one stale AGENTS.md, and a stage-2 preview with one row
// unmet and two unread. Every value is one `doctor` produces for that
// repository (internal/command/doctorcmd, requirementsScreen).
func drawnRows() Requirements {
	return Requirements{
		Header: "writrun-cli v0.0.2 · pins WritRun v0.0.09 · branch main",
		Line:   "STAGE 1 · files — stages 0–1 examined, stage 2 previewed",
		Groups: []RequirementGroup{
			{Name: "STAGE 0 · ENVIRONMENT — 4 of 4 met", Rows: []Requirement{
				{Mark: "✓", Text: "git", Name: "git"},
				{Mark: "✓", Text: "bash", Name: "bash"},
				{Mark: "✓", Text: "awk", Name: "awk"},
				{Mark: "✓", Text: "sed", Name: "sed"},
			}},
			{Name: "STAGE 1 · FILES — 8 of 9 met", Rows: []Requirement{
				{Mark: "✓", Text: "docs/about.md", Name: "docs/about.md"},
				{Mark: "✓", Text: "docs/product/ — a chapter beyond the README", Name: "docs/product/"},
				{Mark: "✓", Text: "docs/technical/ — a chapter beyond the README", Name: "docs/technical/"},
				{Mark: "✓", Text: "docs/, work/tasks/, work/specs/, work/reports/", Name: "docs/, work/tasks/, work/specs/, work/reports/"},
				{Mark: "!", Text: "AGENTS.md — a writrun:begin/writrun:end section is stale", Name: "AGENTS.md"},
				{Mark: "✓", Text: "writrun/gates.md — 8 gates, 8 answered", Name: "writrun/gates.md",
					Explain: "who operates each gate the methodology names. The rows are that file's own, so a gate a newer kit adds is named without this binary knowing it. Met: 8 gates, 8 answered."},
				{Mark: "✓", Text: ".writrun/VERSION — v0.0.09", Name: ".writrun/VERSION"},
				{Mark: "✓", Text: "check_front_matter.sh", Name: "check_front_matter.sh"},
				{Mark: "✓", Text: "check_settings.sh", Name: "check_settings.sh"},
			}},
			{Name: "STAGE 2 · THE FORGE, PREVIEWED — 3 of 6 met", Rows: []Requirement{
				{Mark: "✓", Text: "gh on the PATH", Name: "gh on the PATH"},
				{Mark: "✓", Text: "gh authenticated", Name: "gh authenticated"},
				{Mark: "✓", Text: "squash merging is on", Name: "squash merging is on"},
				{Mark: "✗", Text: "the recording push can write to main", Name: "the recording push can write to main",
					Explain: "from stage 2 the workflows record the queue's state by pushing to main. Not met: .github/workflows/record.yml pushes to main and raises no `contents: write` of its own. Clear it by setting the Actions workflow permissions to read-and-write, or raising `contents: write` in that file."},
				{Mark: "?", Text: "main is governed by a ruleset", Name: "main is governed by a ruleset"},
				{Mark: "?", Text: "no rule over main refuses the recording push", Name: "no rule over main refuses the recording push"},
			}},
			{Name: "STAGE 3 · ISSUES — not previewed, one rung at a time", Rows: nil},
		},
	}
}

// screenAt is the model with the cursor moved onto the row named.
func screenAt(t *testing.T, name string) requirementsModel {
	t.Helper()
	r := drawnRows()
	m := newRequirements(r, func() (Requirements, error) { return r, nil })
	// 65 columns puts the rule at the 63 the frames draw it at.
	next, _ := m.Update(tea.WindowSizeMsg{Width: 65, Height: 0})
	m = next.(requirementsModel)
	for i, row := range m.rows {
		if row.name == name {
			m.cursor = i
			return m
		}
	}
	t.Fatalf("the screen holds no row named %q", name)
	return m
}

func TestTheScreenFrameForARequirementThatIsMet(t *testing.T) {
	m := screenAt(t, "writrun/gates.md")
	drawn := frame(t, doctorDrawing, "the doctor screen — a requirement that is met")
	// The frame disagrees with itself on one word: its footer quotes the
	// selected row as "8 rows, 8 answered" where the row it points at is
	// drawn "8 gates, 8 answered". The footer quotes the row's own note,
	// so the row is the authority and the word is reconciled here rather
	// than left to read as the binary drifting.
	for i, l := range drawn {
		drawn[i] = strings.ReplaceAll(l, "8 rows,", "8 gates,")
	}
	sameScreen(t, rendered(m.View()), drawn)
}

func TestTheScreenFrameForARequirementThatIsNot(t *testing.T) {
	m := screenAt(t, "the recording push can write to main")
	drawn := frame(t, doctorDrawing, "the doctor screen — a requirement that is not")
	sameScreen(t, rendered(m.View()), drawn)
}

// sameScreen compares the rows line for line and the footer as one
// paragraph.
func sameScreen(t *testing.T, got, want []string) {
	t.Helper()
	gotRows, gotFoot, gotKeys := split(t, got)
	wantRows, wantFoot, wantKeys := split(t, want)
	sameLines(t, gotRows, wantRows)
	if gotFoot != wantFoot {
		t.Errorf("the footer differs\n  drawn:    %q\n  rendered: %q", wantFoot, gotFoot)
	}
	if gotKeys != wantKeys {
		t.Errorf("the keys differ\n  drawn:    %q\n  rendered: %q", wantKeys, gotKeys)
	}
}

// split cuts a frame at its rule: the rows above it, the footer under
// it as one paragraph, and the keys last.
func split(t *testing.T, lines []string) (rows []string, footer, keys string) {
	t.Helper()
	at := -1
	for i, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), "───") {
			at = i
			break
		}
	}
	if at < 0 {
		t.Fatalf("the frame draws no rule under its rows:\n%s", strings.Join(lines, "\n"))
	}
	rows = lines[:at]
	for len(rows) > 0 && rows[len(rows)-1] == "" {
		rows = rows[:len(rows)-1]
	}
	var paragraph []string
	for _, l := range lines[at+1:] {
		if strings.TrimSpace(l) == "" {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(l), "↑↓") {
			keys = strings.TrimSpace(l)
			continue
		}
		paragraph = append(paragraph, strings.TrimSpace(l))
	}
	return rows, strings.Join(paragraph, " "), keys
}
