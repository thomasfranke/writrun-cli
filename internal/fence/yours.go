package fence

import (
	"bytes"
	"fmt"
	"strings"
)

// block is the run of lines a `yours` marker governs, as a half-open
// line range over the section it was found in.
type block struct {
	marker int // index of the marker line
	from   int // first governed line
	to     int // one past the last governed line
}

// carryYours rewrites the incoming section so that every block the
// document marked `yours` arrives unchanged.
//
// The marker's position relative to what it governs is not fixed: the
// gates table follows its marker, and the deriving default precedes
// one. So a marker adopts the block after it, and the block before it
// when nothing follows — bounded either way by the next heading, which
// is where one subject ends.
func carryYours(current, section []byte) ([]byte, error) {
	cur := splitLines(current)
	inc := splitLines(section)

	curBlocks := yoursBlocks(cur)
	incBlocks := yoursBlocks(inc)

	if len(incBlocks) < len(curBlocks) {
		return nil, fmt.Errorf(
			"the refreshed section carries %d `yours` marker(s) and this AGENTS.md holds %d — refreshing would drop an answer only this project can give",
			len(incBlocks), len(curBlocks))
	}
	if len(curBlocks) == 0 {
		return section, nil
	}

	// Paired by order of appearance: the markers name subjects in a
	// fixed sequence, and pairing by their prose would break on the
	// first reworded comment.
	var out []string
	prev := 0
	for i, b := range incBlocks {
		if i >= len(curBlocks) {
			break
		}
		// A marker whose governed block the previous one already
		// consumed would slice backwards; leave the incoming block
		// standing rather than panic on a shape nobody has shipped.
		if b.from < prev {
			continue
		}
		out = append(out, inc[prev:b.from]...)
		out = append(out, cur[curBlocks[i].from:curBlocks[i].to]...)
		prev = b.to
	}
	out = append(out, inc[prev:]...)
	return []byte(strings.Join(out, "\n")), nil
}

// yoursBlocks finds every `yours` marker and the lines it governs.
func yoursBlocks(lines []string) []block {
	var blocks []block
	for i, l := range lines {
		if !strings.HasPrefix(strings.TrimSpace(l), yoursPrefix) {
			continue
		}
		b := block{marker: i}
		if from, to, ok := runAfter(lines, i); ok {
			b.from, b.to = from, to
		} else if from, to, ok := runBefore(lines, i); ok {
			b.from, b.to = from, to
		} else {
			// A marker governing nothing still pairs, so the sequence
			// does not shift; it simply carries no lines.
			b.from, b.to = i+1, i+1
		}
		blocks = append(blocks, b)
	}
	return blocks
}

// runAfter is the block of content following the marker: blank lines
// skipped, then every line up to the next blank line or heading.
func runAfter(lines []string, marker int) (from, to int, ok bool) {
	i := marker + 1
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}
	if i >= len(lines) || isBoundary(lines[i]) {
		return 0, 0, false
	}
	from = i
	for i < len(lines) && strings.TrimSpace(lines[i]) != "" && !isBoundary(lines[i]) {
		i++
	}
	return from, i, true
}

// runBefore is the block of content preceding the marker, read the same
// way and in the same bounds.
func runBefore(lines []string, marker int) (from, to int, ok bool) {
	i := marker - 1
	for i >= 0 && strings.TrimSpace(lines[i]) == "" {
		i--
	}
	if i < 0 || isBoundary(lines[i]) {
		return 0, 0, false
	}
	to = i + 1
	for i >= 0 && strings.TrimSpace(lines[i]) != "" && !isBoundary(lines[i]) {
		i--
	}
	return i + 1, to, true
}

// isBoundary reports the lines a governed block never crosses: a
// heading, which is where one subject ends, and the fence's own
// markers, which are the section's edges and nobody's answer.
func isBoundary(line string) bool {
	l := strings.TrimSpace(line)
	return strings.HasPrefix(l, "#") ||
		strings.HasPrefix(l, Begin) ||
		strings.HasPrefix(l, End) ||
		strings.HasPrefix(l, yoursPrefix)
}

func splitLines(b []byte) []string {
	return strings.Split(string(bytes.TrimSuffix(b, []byte("\n"))), "\n")
}
