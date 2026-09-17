package screen

import "strings"

// The furniture every screen carries, in one place: two lines of
// header, one line of footer in three groups, and one way out
// (docs/product/screens/README.md).
//
// Seven screens wrote their own and no two agreed — `q` quit under a
// footer that called it `back`, three screens named the binary and
// four did not, and the running state drew no footer at all
// (report-0035). A screen now states its facts and its keys; the lines
// they become are composed here.

// defaultWidth is how wide a screen draws its content before a
// terminal has said how wide it is: the rule, the right edge of the
// context line, and the column an explanation wraps at.
const defaultWidth = 71

// chrome is the two lines every screen opens with.
type chrome struct {
	// identity is the first line: the product, its version, the tag it
	// pins, the branch. The caller composes it, because every fact in
	// it is one the caller already holds.
	identity string
	// context is the second: what this screen reads, named.
	context string
	// source is drawn at the right of the context line, where a screen
	// names the file or the check behind what it shows. Empty is a
	// context line that needs no second half.
	source string
}

// lines are the header at a content width. The context line truncates
// rather than wraps: a header that wrapped would push into the rows.
func (c chrome) lines(width int) []string {
	line := c.context
	if c.source != "" {
		gap := width - columns(line) - columns(c.source)
		if gap < 1 {
			gap = 1
		}
		line += strings.Repeat(" ", gap) + c.source
	}
	return []string{" " + truncate(c.identity, width), " " + truncate(line, width)}
}

// truncate cuts a line to a width in columns. Width is columns, not
// bytes: a `·` is one column and two bytes, and counting bytes would
// cut a header short of the terminal it fits in.
func truncate(s string, width int) string {
	if width <= 0 || columns(s) <= width {
		return s
	}
	return string([]rune(s)[:width])
}

// wayOut is how a screen is left. There are three forms and no fourth:
// `q quit` where there is nowhere to go back to, `esc back · q quit`
// where there is, and `esc cancels` on a plan or a question — which is
// not a screen and has no `q`, because `q` there would be a value
// (docs/product/screens/README.md, spec-0042).
type wayOut int

const (
	// quitOnly is the entry screen and the first run: nothing is behind
	// them but the shell.
	quitOnly wayOut = iota
	// backOrQuit is every screen reached from another.
	backOrQuit
	// cancels is a plan or a question.
	cancels
	// silent is a screen with no key to press, which is a screen
	// waiting on a command it must not interrupt.
	silent
)

func (w wayOut) groups() []string {
	switch w {
	case backOrQuit:
		return []string{"esc back", "q quit"}
	case cancels:
		return []string{"esc cancels"}
	case silent:
		return nil
	default:
		return []string{"q quit"}
	}
}

// footer is the last line of every screen, in three groups and in one
// order: movement, the actions one key each, the way out.
type footer struct {
	// movement is the first group, or the sentence that replaces it
	// where nothing moves — `nothing to select`, as the queue already
	// says.
	movement string
	// actions are the keys that act, one group each.
	actions []string
	way     wayOut
}

// line is the footer rendered. A group with nothing in it is left out
// rather than drawn empty: a key that cannot act is not offered.
func (f footer) line() string {
	groups := make([]string, 0, len(f.actions)+3)
	if f.movement != "" {
		groups = append(groups, f.movement)
	}
	groups = append(groups, f.actions...)
	groups = append(groups, f.way.groups()...)
	return " " + strings.Join(groups, " · ")
}

// rule is the line a screen draws between its rows and the pane that
// explains the selected one.
func rule(width int) string { return " " + strings.Repeat("─", width) }

// contentWidth is how wide a screen draws its rule and wraps its
// explanation: the terminal's, less the column of padding every line
// carries, and defaultWidth until the terminal has said.
func contentWidth(terminal int) int {
	if terminal > 2 {
		return terminal - 2
	}
	return defaultWidth
}

// wrap breaks text on spaces to a width, the first line under one
// prefix and every line after it under another. Width is columns, not
// bytes: a glyph and a dash are one column each and several bytes, and
// counting bytes would wrap a line of them early.
func wrap(text string, width int, first, hanging string) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	line := first + words[0]
	for _, w := range words[1:] {
		if columns(line)+1+columns(w) > width {
			lines = append(lines, line)
			line = hanging + w
			continue
		}
		line += " " + w
	}
	return append(lines, line)
}

// columns is how wide a string is on a terminal, counting characters.
func columns(s string) int { return len([]rune(s)) }

// pad right-fills a name to a width, so a column of them lines up.
func pad(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// The cursor sits in the screen's gutter and never moves what a row
// says: a list does not shift sideways as the cursor travels it.
//
// Which column the gutter is depends on what the rows carry. Rows with
// a mark of their own — the doctor screen's, the first run's — keep the
// mark in column 3 and the cursor in column 1, so the two read as one
// gutter. Rows with no mark carry their own indent and take the cursor
// in front of them, in column 0 (the entry screen, the queue, the
// settings). Both are drawn that way, and neither is a second answer
// about one screen (docs/product/screens/).
const cursorGlyph = "›"

// cursorIn puts the glyph in column 1 of a row indented for it.
func cursorIn(line string) string {
	r := []rune(line)
	if len(r) < 2 {
		return cursorGlyph
	}
	return string(r[:1]) + cursorGlyph + string(r[2:])
}

// cursorBefore is the glyph, or the space that stands in its column, in
// front of a row that carries its own indent.
func cursorBefore(selected bool) string {
	if selected {
		return cursorGlyph
	}
	return " "
}

// window is the slice of rows a viewport of a height shows from a top.
func window(rows, top, height int) int {
	if height > 0 && top+height < rows {
		return top + height
	}
	return rows
}
