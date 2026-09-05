package listcmd

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// group is the filter a section answers to. A section is printed
// whole or not at all: a filter chooses sections, never tasks, so
// every task shown keeps the group and the order the lister gave it
// (spec-0006, acceptance criteria).
type group int

const (
	// groupAlways is printed by every run: the lister's notes, which
	// qualify the whole answer, and anything it printed before the
	// first heading this binary recognises — no filter names either, and
	// dropping a line nobody selected is worse than printing it.
	groupAlways group = iota
	groupAvailable
	groupHeld
	groupReports
)

// headers names each of the lister's sections and the group it belongs
// to. The match is the opening phrase rather than the whole line: the
// wording is the skill's, and a release that rewords a heading must
// not silently drop the section it heads.
//
// In progress and in flight belong to the available group because they
// are what taking work now depends on — one says resume before
// selecting anything new, the other says why an eligible task is not
// offered.
var headers = []struct {
	prefix string
	group  group
}{
	{"In progress", groupAvailable},
	{"Available", groupAvailable},
	{"Nothing is available", groupAvailable},
	{"In flight", groupAvailable},
	{"Held back", groupHeld},
	{"Open reports", groupReports},
	{"Note:", groupAlways},
}

// sections is what a run prints. No flag prints everything.
type sections struct {
	available bool
	held      bool
	reports   bool
}

// filtering reports whether any filter was given.
func (s sections) filtering() bool { return s.available || s.held || s.reports }

// wants reports whether a group is printed under these filters. No
// filter prints every group.
func (s sections) wants(g group) bool {
	if !s.filtering() {
		return true
	}
	switch g {
	case groupAvailable:
		return s.available
	case groupHeld:
		return s.held
	case groupReports:
		return s.reports
	default:
		return true
	}
}

// block is one of the lister's sections: its group and its lines,
// unedited.
type block struct {
	group group
	lines []string
}

// split cuts the lister's output into sections. Everything that is not
// a heading continues the section it follows — an indented row, a
// blank line the section spaces itself with, and the unindented line
// the available section closes with. The blank line before the next
// heading is a separator and nothing else: it is trimmed here and
// re-emitted by render, so a selected section is printed exactly as
// the lister wrote it.
func split(out string) []block {
	var blocks []block
	sc := bufio.NewScanner(strings.NewReader(out))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if g, ok := header(line); ok {
			blocks = append(blocks, block{group: g, lines: []string{line}})
			continue
		}
		if len(blocks) == 0 {
			if strings.TrimSpace(line) == "" {
				continue
			}
			blocks = append(blocks, block{group: groupAlways})
		}
		last := &blocks[len(blocks)-1]
		last.lines = append(last.lines, line)
	}
	for i := range blocks {
		blocks[i].lines = trimBlank(blocks[i].lines)
	}
	return blocks
}

// trimBlank drops the trailing blank lines of a section.
func trimBlank(lines []string) []string {
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// header reports the group a line opens a section for. An indented
// line is a row of the section above it and never a heading.
func header(line string) (group, bool) {
	if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
		return groupAlways, false
	}
	for _, h := range headers {
		if strings.HasPrefix(line, h.prefix) {
			return h.group, true
		}
	}
	return groupAlways, false
}

// render prints the selected sections in the order the lister printed
// them, one blank line apart.
func (s sections) render(w io.Writer, out string) {
	first := true
	for _, b := range split(out) {
		if !s.wants(b.group) {
			continue
		}
		if !first {
			fmt.Fprintln(w)
		}
		first = false
		for _, line := range b.lines {
			fmt.Fprintln(w, line)
		}
	}
}
