package screen

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// The drawing is the assertion, and this is how a frame is read back out
// of one.
//
// An `.excalidraw` file is JSON. A frame is a rectangle, its caption is
// the text element inside its title bar, and its content is every text
// element inside it — each at a position the font's own metrics turn
// back into a line and a column. So a frame reconstructs to the lines a
// terminal would show, and a case can hold the binary to them
// (docs/product/screens/README.md — where the two disagree, the binary
// is what changes).
const (
	// charWidth and lineHeight are the drawing's monospace metrics at
	// font size 14: every text element in these frames is at that size,
	// and every one of them measures exactly this.
	charWidth  = 8.4
	lineHeight = 17.5
	// framePad is the inset the frames draw their content at, and
	// titleBar is the height of the strip the caption sits in.
	framePad = 8.0
	titleBar = 30.0
)

type drawing struct {
	Elements []element `json:"elements"`
}

type element struct {
	Type         string  `json:"type"`
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
	Width        float64 `json:"width"`
	Height       float64 `json:"height"`
	Text         string  `json:"text"`
	OriginalText string  `json:"originalText"`
}

func (e element) content() string {
	if e.OriginalText != "" {
		return e.OriginalText
	}
	return e.Text
}

// frame is the lines one captioned frame of a drawing shows, top and
// bottom blanks trimmed.
func frame(t *testing.T, file, caption string) []string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(file)))
	if err != nil {
		t.Fatalf("the drawing could not be read: %v", err)
	}
	var d drawing
	if err := json.Unmarshal(raw, &d); err != nil {
		t.Fatalf("%s is not readable JSON: %v", file, err)
	}

	var cap *element
	for i, e := range d.Elements {
		if e.Type == "text" && strings.TrimSpace(e.content()) == caption {
			cap = &d.Elements[i]
			break
		}
	}
	if cap == nil {
		t.Fatalf("%s draws no frame captioned %q", file, caption)
	}
	var rect *element
	for i, e := range d.Elements {
		if e.Type != "rectangle" {
			continue
		}
		if e.X <= cap.X && cap.X <= e.X+e.Width && e.Y <= cap.Y && cap.Y <= e.Y+e.Height {
			if rect == nil || e.Height > rect.Height {
				rect = &d.Elements[i]
			}
		}
	}
	if rect == nil {
		t.Fatalf("the caption %q sits in no frame", caption)
	}

	top := rect.Y + titleBar
	rows := map[int][]struct {
		col  int
		text string
	}{}
	for _, e := range d.Elements {
		if e.Type != "text" {
			continue
		}
		if e.X < rect.X || e.X >= rect.X+rect.Width || e.Y < top || e.Y >= rect.Y+rect.Height {
			continue
		}
		col := int(math.Round((e.X - rect.X - framePad) / charWidth))
		for i, line := range strings.Split(e.content(), "\n") {
			row := int(math.Round((e.Y + float64(i)*lineHeight - top) / lineHeight))
			rows[row] = append(rows[row], struct {
				col  int
				text string
			}{col, line})
		}
	}

	last := 0
	for row := range rows {
		if row > last {
			last = row
		}
	}
	var out []string
	for row := 0; row <= last; row++ {
		pieces := rows[row]
		sort.Slice(pieces, func(i, j int) bool { return pieces[i].col < pieces[j].col })
		// Columns are characters, not bytes: a glyph and a dash are one
		// column each and several bytes, so the line is assembled as
		// runes or every element after the first `✓` lands short.
		var line []rune
		for _, p := range pieces {
			for len(line) < p.col {
				line = append(line, ' ')
			}
			line = append(line[:p.col], []rune(p.text)...)
		}
		out = append(out, strings.TrimRight(string(line), " "))
	}
	for len(out) > 0 && out[0] == "" {
		out = out[1:]
	}
	for len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}
	return out
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

// lines is rendered output as the frames are: trailing spaces gone, and
// the blanks at either end with them.
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

const doctorDrawing = "docs/product/screens/adoption/doctor.excalidraw"
