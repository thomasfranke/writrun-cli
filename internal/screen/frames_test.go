package screen

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/drawing"
)

// The drawing is the assertion, and `internal/drawing` is how a frame
// is read back out of one.
//
// An `.excalidraw` file is JSON. A frame is a rectangle, its caption is
// the text element in its title bar, and its content is every text
// element inside it — each at a position the font's own metrics turn
// back into a line and a column. So a frame reconstructs to the lines a
// terminal would show, and a case can hold the binary to them
// (docs/product/screens/README.md — where the two disagree, the binary
// is what changes).
//
// The reading used to live here, in a copy of the package below it.
// One reader is what keeps two frames of one screen from being read by
// two rules.
func frame(t *testing.T, file, caption string) []string {
	t.Helper()
	lines, err := drawing.Screen(filepath.Join("..", "..", filepath.FromSlash(file)), caption)
	if err != nil {
		t.Fatalf("the drawing could not be read: %v", err)
	}
	return lines
}

// sameLines fails naming the first line the two disagree on, which is
// the one a reader has to look at.
func sameLines(t *testing.T, got, want []string) {
	t.Helper()
	for i := 0; i < len(got) || i < len(want); i++ {
		g, w := "", ""
		if i < len(got) {
			g = got[i]
		}
		if i < len(want) {
			w = want[i]
		}
		if g != w {
			t.Fatalf("line %d differs\n  drawn:    %q\n  rendered: %q\n\nrendered whole:\n%s",
				i+1, w, g, strings.Join(got, "\n"))
		}
	}
}

// rendered is output as the frames are: trailing spaces gone, and the
// blanks at either end with them.
func rendered(out string) []string {
	var trimmed []string
	for _, l := range strings.Split(out, "\n") {
		trimmed = append(trimmed, strings.TrimRight(l, " "))
	}
	for len(trimmed) > 0 && trimmed[0] == "" {
		trimmed = trimmed[1:]
	}
	for len(trimmed) > 0 && trimmed[len(trimmed)-1] == "" {
		trimmed = trimmed[:len(trimmed)-1]
	}
	return trimmed
}

// sameFrame holds a rendered screen to a drawn one.
//
// The rows are compared column for column: where a row sits is the
// screen's own answer. The explanation and the key line are compared as
// sentences, because where a sentence breaks is the terminal's width
// and the canvas's — the words are the binary's and the wrapping is
// not. The doctor screen's cases settled that, and this is the same
// rule applied to every screen.
func sameFrame(t *testing.T, got, want []string) {
	t.Helper()
	gotRows, gotSays, gotKeys := parts(got)
	wantRows, wantSays, wantKeys := parts(want)
	sameLines(t, gotRows, wantRows)
	if gotSays != wantSays {
		t.Errorf("the explanation differs\n  drawn:    %q\n  rendered: %q", wantSays, gotSays)
	}
	if gotKeys != wantKeys {
		t.Errorf("the keys differ\n  drawn:    %q\n  rendered: %q", wantKeys, gotKeys)
	}
}

// parts cuts a frame into the rows, the explanation and the key line.
//
// The key line is the last line; the explanation is the block of lines
// above it, up to the blank or the rule that separates it from the
// rows.
func parts(lines []string) (rows []string, says, keys string) {
	if len(lines) == 0 {
		return nil, "", ""
	}
	keys = strings.TrimSpace(lines[len(lines)-1])
	body := lines[:len(lines)-1]
	for len(body) > 0 && body[len(body)-1] == "" {
		body = body[:len(body)-1]
	}
	var said []string
	for len(body) > 0 {
		last := body[len(body)-1]
		if last == "" || isRule(last) {
			break
		}
		said = append([]string{strings.TrimSpace(last)}, said...)
		body = body[:len(body)-1]
	}
	for len(body) > 0 && body[len(body)-1] == "" {
		body = body[:len(body)-1]
	}
	return body, strings.Join(said, " "), keys
}

func isRule(line string) bool { return strings.HasPrefix(strings.TrimSpace(line), "───") }

// keyLine is the last line of a frame, which is the footer on every
// screen that draws one.
func keyLine(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.TrimSpace(lines[len(lines)-1])
}

const (
	doctorDrawing   = "docs/product/screens/adoption/doctor.excalidraw"
	entryDrawing    = "docs/product/screens/entry.excalidraw"
	firstRunDrawing = "docs/product/screens/first-run.excalidraw"
	listDrawing     = "docs/product/screens/tasks/list.excalidraw"
	configDrawing   = "docs/product/screens/adoption/config.excalidraw"
)
