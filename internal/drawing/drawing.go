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

	// framePad is the inset the frames transcribed from real runs draw
	// their content at.
	framePad = 8.0
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
	lines, err := framed(path, caption, framePad, true)
	if err != nil {
		return nil, err
	}
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

// Screen returns the lines a captioned frame draws, column for column,
// as a screen renders them.
//
// It takes nothing off. A screen's own output starts at the frame's
// first column and every space in front of a row is the screen's, so
// the padding is read from the frame rather than assumed: the frames
// are drawn at their own insets, and one assumed inset would read every
// line of half of them one column out.
func Screen(path, caption string) ([]string, error) {
	lines, err := framed(path, caption, -1, false)
	if err != nil {
		return nil, err
	}
	for len(lines) > 0 && lines[0] == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines, nil
}

// framed is both readings' common half: the frame found, its content
// bounded, and its texts placed on the terminal's grid. A pad below
// zero is read from the frame's own leftmost text.
//
// cutAtDivider ends the reading at the frame's divider. A transcript
// frame draws a second scene under one, and that scene is not this
// run's output. A screen frame draws its explanation pane under one,
// and that pane is the screen — so the two readings differ here and
// nowhere else.
func framed(path, caption string, pad float64, cutAtDivider bool) ([]string, error) {
	f, err := read(path)
	if err != nil {
		return nil, err
	}
	win, bar, ok := window(f, caption)
	if !ok {
		return nil, fmt.Errorf("%s: no frame captioned %q", path, caption)
	}
	top := bar.Y + bar.Height
	bottom := win.Y + win.Height
	if cutAtDivider {
		for _, e := range f.Elements {
			if e.Deleted || e.Type != "rectangle" || e.Background != dividerFill {
				continue
			}
			if inside(win, e.X, e.Y) && e.Y < bottom {
				bottom = e.Y
			}
		}
	}
	if pad < 0 {
		pad = contentPad(f, win, top, bottom)
	}
	return layout(f, win, top, bottom, pad), nil
}

// contentPad is the frame's own inset: how far its leftmost text sits
// from its left edge.
func contentPad(f *file, win element, top, bottom float64) float64 {
	pad := -1.0
	for _, e := range f.Elements {
		if e.Deleted || e.Type != "text" {
			continue
		}
		if e.X < win.X || e.X >= win.X+win.Width || e.Y < top || e.Y >= bottom {
			continue
		}
		if at := e.X - win.X; pad < 0 || at < pad {
			pad = at
		}
	}
	if pad < 0 {
		return framePad
	}
	return pad
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
func layout(f *file, win element, top, bottom, pad float64) []string {
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
		col := int((c.x-win.X-pad)/charWidth + 0.5)
		// Columns are characters, not bytes: a glyph and a dash are one
		// column each and several bytes, so the line is assembled as
		// runes or every element after the first `✓` lands short.
		r := []rune(line)
		for len(r) < col {
			r = append(r, ' ')
		}
		line = string(r[:col]) + c.text
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

// Compare answers how a rendered screen differs from the frame it is
// checked against, and "" where the two agree.
//
// It is a function rather than an assertion so that it can live beside
// the reader: seven packages compose a plan, and one comparison is what
// keeps them from being checked seven ways.
//
// What it compares exactly and what it compares as a sentence is the
// rule the doctor screen's cases settled. A row's columns are the
// screen's own answer and are compared column for column. Prose is
// compared as a sentence where the two wrapped it differently, because
// where a sentence breaks is the terminal's width and the canvas's —
// and column for column where they did not, so an alignment the screen
// computes is still held.
func Compare(got, want []string) string {
	if d := CompareRows(got, want); d != "" {
		return d
	}
	_, gotSays, _ := split(got)
	_, wantSays, _ := split(want)
	if gotSays != wantSays {
		return difference("the explanation differs", gotSays, wantSays, got)
	}
	return ""
}

// CompareRows is Compare without the explanation: the rows and the key
// line, and nothing said about the selected row.
//
// It is for a frame whose explanation names a fact the rows it is drawn
// from do not carry. A test that asserted such a sentence would be
// asserting the drawing against itself, and the difference belongs in
// the spec's Outcome where a person reads it.
func CompareRows(got, want []string) string {
	gotRows, _, gotKeys := split(got)
	wantRows, _, wantKeys := split(want)

	gotHead, gotRest := lead(gotRows)
	wantHead, wantRest := lead(wantRows)
	if len(gotHead) != len(wantHead) {
		if g, w := sentence(gotHead), sentence(wantHead); g != w {
			return difference("the opening differs", g, w, got)
		}
	} else if d := lines(gotHead, wantHead, got); d != "" {
		return d
	}
	if d := lines(gotRest, wantRest, got); d != "" {
		return d
	}
	if gotKeys != wantKeys {
		return difference("the keys differ", gotKeys, wantKeys, got)
	}
	return ""
}

// split cuts a frame into its rows, the explanation under them and the
// key line last.
func split(in []string) (rows []string, says, keys string) {
	if len(in) == 0 {
		return nil, "", ""
	}
	keys = strings.TrimSpace(in[len(in)-1])
	body := trimBlanks(in[:len(in)-1])
	var said []string
	for len(body) > 0 {
		last := body[len(body)-1]
		if strings.TrimSpace(last) == "" || strings.HasPrefix(strings.TrimSpace(last), "───") {
			break
		}
		said = append([]string{strings.TrimSpace(last)}, said...)
		body = body[:len(body)-1]
	}
	return trimBlanks(body), strings.Join(said, " "), keys
}

// lead peels the prose a frame opens with: the lines before its first
// row, which is the first line indented two columns or carrying the
// cursor.
func lead(in []string) (head, rest []string) {
	for i, l := range in {
		if strings.TrimSpace(l) == "" || isRow(l) {
			return in[:i], in[i:]
		}
	}
	return in, nil
}

func isRow(line string) bool {
	return strings.HasPrefix(line, "›") || strings.HasPrefix(line, "  ")
}

func trimBlanks(in []string) []string {
	for len(in) > 0 && strings.TrimSpace(in[len(in)-1]) == "" {
		in = in[:len(in)-1]
	}
	return in
}

func sentence(in []string) string {
	var said []string
	for _, l := range in {
		if s := strings.TrimSpace(l); s != "" {
			said = append(said, s)
		}
	}
	return strings.Join(said, " ")
}

// lines names the first line the two disagree on, which is the one a
// reader has to look at.
func lines(got, want, whole []string) string {
	for i := 0; i < len(got) || i < len(want); i++ {
		g, w := "", ""
		if i < len(got) {
			g = got[i]
		}
		if i < len(want) {
			w = want[i]
		}
		if g != w {
			return difference(fmt.Sprintf("line %d differs", i+1), g, w, whole)
		}
	}
	return ""
}

func difference(what, got, want string, whole []string) string {
	return fmt.Sprintf("%s\n  drawn:    %q\n  rendered: %q\n\nrendered whole:\n%s",
		what, want, got, strings.Join(whole, "\n"))
}
