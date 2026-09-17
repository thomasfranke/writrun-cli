package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/drawing"
)

// screensDir is where the design lives. A frame in it is what the
// binary must render, and the binary is what changes where the two
// disagree (docs/product/screens/README.md).
const screensDir = "../../docs/product/screens"

// drawings names the drawing each command's screens are stated in — one
// file per row of the entry screen, under the folder its section names.
var drawings = map[string]string{
	"init":      "adoption/init.excalidraw",
	"update":    "adoption/update.excalidraw",
	"doctor":    "adoption/doctor.excalidraw",
	"uninstall": "adoption/uninstall.excalidraw",
	"config":    "adoption/config.excalidraw",
	"list":      "tasks/list.excalidraw",
	"take":      "tasks/take.excalidraw",
	"work":      "tasks/work.excalidraw",
	"status":    "tasks/status.excalidraw",
	"finish":    "tasks/finish.excalidraw",
	"author":    "authoring/author.excalidraw",
	"amend":     "authoring/amend.excalidraw",
	"report":    "reports/report.excalidraw",
}

// answer runs the production command table over the frame and returns
// what it printed. Nothing here reaches a repository: both answers are
// the binary's own, given anywhere.
func answer(t *testing.T, args ...string) string {
	t.Helper()
	var out, errb bytes.Buffer
	code := command.Run(command.Frame{
		Version:    "v0.0.2",
		WritRunTag: writrunTag,
		Commands:   commands(),
		Stdout:     &out,
		Stderr:     &errb,
		Terminal:   &command.FakeTerminal{},
		FindRepo:   func(string) (string, bool, error) { return "/repo", true, nil },
		Getenv:     func(string) string { return "" },
		Getwd:      func() (string, error) { return "/repo", nil },
	}, args)
	if code != 0 {
		t.Fatalf("writrun %s = exit %d (%s)", strings.Join(args, " "), code, errb.String())
	}
	return out.String()
}

// opensWithTheFrame fails unless what the binary printed opens with the
// frame the caption names, line for line.
func opensWithTheFrame(t *testing.T, got, file, caption string) []string {
	t.Helper()
	want, err := drawing.Frame(filepath.Join(screensDir, file), caption)
	if err != nil {
		t.Fatalf("%v", err)
	}
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) < len(want) {
		t.Fatalf("%s — %s:\nthe binary printed %d lines and the frame draws %d\n%s",
			file, caption, len(lines), len(want), got)
	}
	for i := range want {
		// A drawn line stops where its window does, so the frame's line
		// is what the rendered one opens with.
		if !strings.HasPrefix(lines[i], want[i]) {
			t.Errorf("%s — %s, line %d:\n  binary %q\n  frame  %q", file, caption, i+1, lines[i], want[i])
		}
	}
	return lines
}

// isTheFrame fails unless what the binary printed is the frame and
// nothing besides.
func isTheFrame(t *testing.T, got, file, caption string) {
	t.Helper()
	lines := opensWithTheFrame(t, got, file, caption)
	want, err := drawing.Frame(filepath.Join(screensDir, file), caption)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(lines) != len(want) {
		t.Fatalf("%s — %s:\nthe binary printed %d lines and the frame draws %d\n%s",
			file, caption, len(lines), len(want), got)
	}
}

// The help's own frame draws its last two lines below the frame's
// divider, where a drawing puts what is not this scene — so the rows
// are what is compared here, and the two lines under them are held by
// internal/command's own case.
func TestTheHelpIsWhatTheDrawingStates(t *testing.T) {
	opensWithTheFrame(t, answer(t, "--help"),
		"help.excalidraw", "writrun --help — grouped by what a person is doing")
}

func TestEveryCommandExplainsItselfAsItsDrawingStates(t *testing.T) {
	for _, c := range commands() {
		t.Run(c.Name, func(t *testing.T) {
			file, ok := drawings[c.Name]
			if !ok {
				t.Fatalf("%s: no drawing names this command's screens", c.Name)
			}
			isTheFrame(t, answer(t, c.Name, "--help"),
				file, "writrun "+c.Name+" — what it is for, in its own words")
		})
	}
}

// A command added without a long description is a command a newcomer
// cannot learn from the binary. The table is where that is caught.
func TestEveryCommandCarriesALongDescription(t *testing.T) {
	for _, c := range commands() {
		if c.About.Empty() {
			t.Errorf("%s: no long description — see docs/product/screens/%s", c.Name, drawings[c.Name])
		}
	}
}
