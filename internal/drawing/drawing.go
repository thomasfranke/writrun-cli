// Package drawing reads one frame out of a screen drawing.
//
// `docs/product/screens/` is the design, and the binary is what changes
// where the two disagree (docs/product/screens/README.md). A test can
// only hold the binary to that if it can read a frame, so this reads
// one: the caption names it, and what comes back is the terminal lines
// the frame draws, as the binary would print them.
//
// It is read at test time rather than transcribed into a fixture,
// because a transcription is a second copy of the drawing and would go
// on passing after the drawing changed.
package drawing

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// The drawings' terminal frames are one shape: a dark window, a lighter
// title bar carrying the caption, and 14px monospace text inside.
const (
	windowFill  = "#1c1c1c"
	barFill     = "#2a2a2a"
	dividerFill = "#3a3a3a"

	// charWidth and lineHeight are that text's advance and leading, in
	// the canvas's pixels. They turn a text's position into a row and a
	// column, which is what a terminal line is.
	charWidth  = 8.4
	lineHeight = 17.5
)

type element struct {
	Type         string  `json:"type"`
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
	Width        float64 `json:"width"`
	Height       float64 `json:"height"`
	Background   string  `json:"backgroundColor"`
	Text         string  `json:"text"`
	OriginalText string  `json:"originalText"`
	Deleted      bool    `json:"isDeleted"`
}

type file struct {
	Elements []element `json:"elements"`
}

// Frame returns the lines the captioned frame draws, as the binary
// prints them.
//
// Three things are taken off on the way. The shell prompt the frame
// opens with is the reader typing, not the binary answering. Everything
// below the frame's divider is a second scene or the annotation beside
// it, and neither is this run's output. And every drawn output line
// carries one column of frame padding, which the prompt line does not —
// the frames transcribed from real runs are what says so, `Branch` and
// `writrun-cli v0.0.2 (pins WritRun v0.0.08)` among them.
func Frame(path, caption string) ([]string, error) {
	f, err := read(path)
	if err != nil {
		return nil, err
	}
	win, bar, ok := window(f, caption)
	if !ok {
		return nil, fmt.Errorf("%s: no frame captioned %q", path, caption)
	}
	bottom := win.Y + win.Height
	for _, e := range f.Elements {
		if e.Deleted || e.Type != "rectangle" || e.Background != dividerFill {
			continue
		}
		if inside(win, e.X, e.Y) && e.Y < bottom {
			bottom = e.Y
		}
	}
	lines := layout(f, win, bar.Y+bar.Height, bottom)
	if len(lines) > 0 && strings.HasPrefix(lines[0], "$ ") {
		lines = lines[1:]
	}
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		out = append(out, strings.TrimPrefix(l, " "))
	}
	for len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}
	return out, nil
}

// read parses the drawing.
func read(path string) (*file, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f file
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &f, nil
}

// window finds the frame the caption names: the title bar carrying it
// and the window under it.
func window(f *file, caption string) (element, element, bool) {
	for _, bar := range f.Elements {
		if bar.Deleted || bar.Type != "rectangle" || bar.Background != barFill {
			continue
		}
		if !captioned(f, bar, caption) {
			continue
		}
		for _, win := range f.Elements {
			if win.Deleted || win.Type != "rectangle" || win.Background != windowFill {
				continue
			}
			if near(win.X, bar.X) && near(win.Y, bar.Y) {
				return win, bar, true
			}
		}
	}
	return element{}, element{}, false
}

// captioned says whether the bar carries this caption.
func captioned(f *file, bar element, caption string) bool {
	for _, e := range f.Elements {
		if e.Deleted || e.Type != "text" || !inside(bar, e.X, e.Y) {
			continue
		}
		if strings.TrimSpace(text(e)) == caption {
			return true
		}
	}
	return false
}

// layout places every text between two heights on the grid the terminal
// draws on, and answers the lines.
func layout(f *file, win element, top, bottom float64) []string {
	type placed struct {
		y, x float64
		text string
	}
	var cells []placed
	for _, e := range f.Elements {
		if e.Deleted || e.Type != "text" {
			continue
		}
		if e.X < win.X || e.X >= win.X+win.Width || e.Y < top || e.Y >= bottom {
			continue
		}
		for i, l := range strings.Split(text(e), "\n") {
			cells = append(cells, placed{y: e.Y + float64(i)*lineHeight, x: e.X, text: l})
		}
	}
	sort.SliceStable(cells, func(i, j int) bool {
		if !near(cells[i].y, cells[j].y) {
			return cells[i].y < cells[j].y
		}
		return cells[i].x < cells[j].x
	})

	var lines []string
	var line string
	first := true
	var y float64
	for _, c := range cells {
		if first || c.y-y > lineHeight*0.6 {
			if !first {
				lines = append(lines, strings.TrimRight(line, " "))
				// The gap between two drawn lines is the blank lines
				// between them.
				for blank := int((c.y-y)/lineHeight + 0.5); blank > 1; blank-- {
					lines = append(lines, "")
				}
			}
			first, y, line = false, c.y, ""
		}
		col := int((c.x-win.X-8)/charWidth + 0.5)
		if col > len(line) {
			line += strings.Repeat(" ", col-len(line))
		}
		line = line[:col] + c.text
	}
	if !first {
		lines = append(lines, strings.TrimRight(line, " "))
	}
	return lines
}

// text is what the element says, as it was typed: excalidraw keeps the
// author's own line breaks in originalText and its wrapped copy in
// text.
func text(e element) string {
	if e.OriginalText != "" {
		return e.OriginalText
	}
	return e.Text
}

func inside(r element, x, y float64) bool {
	return x >= r.X && x <= r.X+r.Width && y >= r.Y && y <= r.Y+r.Height
}

func near(a, b float64) bool { return a-b < 2 && b-a < 2 }
