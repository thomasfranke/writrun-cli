package main

import (
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/drawing"
)

// The drawings and the command table state one field, and this is what
// holds them to it.
//
// Two drawings went on stating the one-liners the table held before
// spec-0039 rewrote them, and one of them drew no `config` row at all
// — for a command the entry screen's own grouping lists. Nothing said
// so: the rows are prose on a canvas, and no case read them
// (report-0042).

// drawnRows are the rows a frame draws under its group headings, as
// name and text.
func drawnRows(t *testing.T, file, caption string, headings ...string) map[string]string {
	t.Helper()
	lines, err := drawing.Screen("../../docs/product/screens/"+file, caption)
	if err != nil {
		t.Fatalf("the drawing could not be read: %v", err)
	}
	under := false
	rows := map[string]string{}
	for _, line := range lines {
		// The cursor takes a column in the gutter, and which column is
		// the screen's own answer — so it is read as the space it
		// replaced rather than as the start of a row.
		line = strings.Replace(line, "›", " ", 1)
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(line, "  ") {
			under = false
			for _, h := range headings {
				if trimmed == h {
					under = true
				}
			}
			continue
		}
		if !under {
			continue
		}
		fields := strings.SplitN(trimmed, "   ", 2)
		if len(fields) == 2 {
			rows[strings.TrimSpace(fields[0])] = strings.TrimSpace(fields[1])
		}
	}
	return rows
}

// Every row the entry screen's drawing carries names a command the
// table holds, with that command's own summary — and every command the
// screen groups is drawn.
func TestTheEntryDrawingCarriesTheTablesSummaries(t *testing.T) {
	rows := drawnRows(t, "entry.excalidraw", "writrun — the entry screen",
		"TASKS", "AUTHORING", "REPORTS", "ADOPTION")
	summaries := map[string]string{}
	for _, c := range commands() {
		summaries[c.Name] = c.Summary
	}
	for name, drawn := range rows {
		want, there := summaries[name]
		if !there {
			t.Errorf("the drawing carries a row for %q, which the table does not name", name)
			continue
		}
		if drawn != want {
			t.Errorf("%s: the drawing says %q and the table says %q", name, drawn, want)
		}
	}
	for _, g := range entryGroups {
		for _, name := range g.names {
			if _, drawn := rows[name]; !drawn {
				t.Errorf("the screen lists %q under %s and the drawing draws no row for it", name, g.name)
			}
		}
	}
}

// The first run offers one command, and the words on its row are the
// table's too.
func TestTheFirstRunDrawingCarriesTheTablesSummary(t *testing.T) {
	rows := drawnRows(t, "first-run.excalidraw", "writrun — a repository with no kit in it", "ADOPTION")
	if got, want := rows["init"], summaryOf("init"); got != want {
		t.Errorf("the drawing says %q and the table says %q", got, want)
	}
}
