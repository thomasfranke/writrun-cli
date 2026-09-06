// Package pointer edits WritRun's section of an AGENTS.md: the heading
// whose body links `.writrun/AGENTS.md`, where the kit's flow now
// lives. init grafts it, uninstall cuts it — two verbs over one shape,
// so the shape is read in exactly one place.
//
// This is the one file whose shape the binary knows without calling it
// (docs/technical/engineering/coupling.md, rule 3). AGENTS.md is the
// adopter's, so no file the kit ships can describe an edit to it.
package pointer

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

const (
	// Target is the path the pointer links, and the whole of how the
	// section is recognised.
	Target = ".writrun/AGENTS.md"

	// legacyBegin and legacyEnd are the markers a kit before v0.0.04
	// fenced its section with. Nothing writes them; they are read so a
	// refresh can name what it left behind and uninstall can cut it.
	legacyBegin = "<!-- writrun:begin"
	legacyEnd   = "<!-- writrun:end -->"
)

// link matches a markdown link whose target is the kit's AGENTS.md.
var link = regexp.MustCompile(`\]\([^)]*\.writrun/AGENTS\.md\)`)

// ErrNoSection is returned where no WritRun section is there to read.
// The commands that carry it stop and change nothing.
var ErrNoSection = fmt.Errorf("no heading links " + Target)

// Section cuts WritRun's part out of the template's AGENTS.md: the
// heading that introduces the link, through the line before the next
// heading of the same or higher level.
func Section(agents []byte) ([]byte, error) {
	start, end, ok := bounds(agents)
	if !ok {
		return nil, fmt.Errorf("the template's AGENTS.md carries no WritRun section: %w", ErrNoSection)
	}
	return agents[start:end], nil
}

// Has reports whether a document already carries the pointer, which is
// what makes a graft unnecessary.
func Has(agents []byte) bool {
	_, _, ok := bounds(agents)
	return ok
}

// Legacy reports the fenced section a kit before v0.0.04 left behind.
// A refresh names it; it never rewrites it, because at v0.0.04 the
// whole of AGENTS.md is the project's.
func Legacy(agents []byte) bool {
	begin := bytes.Index(agents, []byte(legacyBegin))
	end := bytes.Index(agents, []byte(legacyEnd))
	return begin >= 0 && end > begin
}

// Graft appends the section to an existing AGENTS.md, touching no byte
// outside it: everything already there survives verbatim, with exactly
// one blank line before the grafted part.
func Graft(existing, section []byte) []byte {
	out := make([]byte, 0, len(existing)+len(section)+2)
	out = append(out, existing...)
	if len(out) > 0 && !bytes.HasSuffix(out, []byte("\n")) {
		out = append(out, '\n')
	}
	out = append(out, '\n')
	out = append(out, section...)
	out = append(out, '\n')
	return out
}

// Remove cuts WritRun's section out, leaving every byte outside it as
// the project wrote it — including the blank line Graft added before
// it, so a graft followed by a removal is a round trip. The legacy
// fenced section is cut in preference to the pointer: it is the larger
// of the two and the one holding the kit's own prose.
//
// It reports whether anything but the section was there at all: a
// document that is nothing but the kit's skeleton is the caller's to
// delete whole.
func Remove(existing []byte) (out []byte, onlySection bool, err error) {
	start, end, ok := legacyBounds(existing)
	if !ok {
		start, end, ok = bounds(existing)
	}
	if !ok {
		return nil, false, ErrNoSection
	}
	prefix := existing[:start]
	suffix := existing[end:]
	// The newline closing the section's last line belongs to it.
	suffix = bytes.TrimPrefix(suffix, []byte("\n"))
	// Graft separates with one blank line; give it back.
	if bytes.HasSuffix(prefix, []byte("\n\n")) {
		prefix = prefix[:len(prefix)-1]
	}
	if len(bytes.TrimSpace(prefix)) == 0 && len(bytes.TrimSpace(suffix)) == 0 {
		return nil, true, nil
	}
	out = make([]byte, 0, len(prefix)+len(suffix))
	out = append(out, prefix...)
	out = append(out, suffix...)
	return out, false, nil
}

// bounds locates the pointer section as a byte range: the heading
// introducing the link, through the line before the next heading of the
// same or higher level, trailing blank lines dropped.
func bounds(doc []byte) (start, end int, ok bool) {
	lines, offsets := splitLines(doc)
	at := -1
	for i, l := range lines {
		if link.MatchString(l) {
			at = i
			break
		}
	}
	if at < 0 {
		return 0, 0, false
	}
	head := -1
	for i := at; i >= 0; i-- {
		if level(lines[i]) > 0 {
			head = i
			break
		}
	}
	if head < 0 {
		return 0, 0, false
	}
	last := len(lines)
	for i := head + 1; i < len(lines); i++ {
		if l := level(lines[i]); l > 0 && l <= level(lines[head]) {
			last = i
			break
		}
	}
	for last > head+1 && strings.TrimSpace(lines[last-1]) == "" {
		last--
	}
	return offsets[head], offsets[head] + lineSpan(lines[head:last]), true
}

// legacyBounds locates the fenced section: from the opening marker,
// or from the heading above it when there is one, through the closing
// marker.
func legacyBounds(doc []byte) (start, end int, ok bool) {
	begin := bytes.Index(doc, []byte(legacyBegin))
	closing := bytes.Index(doc, []byte(legacyEnd))
	if begin < 0 || closing < begin {
		return 0, 0, false
	}
	start = begin
	if nl := bytes.LastIndexByte(doc[:begin], '\n'); nl >= 0 {
		lines, offsets := splitLines(doc[:begin])
		for i := len(lines) - 1; i >= 0; i-- {
			if level(lines[i]) > 0 {
				start = offsets[i]
				break
			}
		}
	}
	return start, closing + len(legacyEnd), true
}

// level is a markdown heading's depth, or zero where the line is not
// a heading.
func level(line string) int {
	n := 0
	for n < len(line) && line[n] == '#' {
		n++
	}
	if n == 0 || n >= len(line) || line[n] != ' ' {
		return 0
	}
	return n
}

// lineSpan is the byte length of a run of lines, newlines included
// between them but not after the last.
func lineSpan(lines []string) int {
	n := 0
	for i, l := range lines {
		if i > 0 {
			n++
		}
		n += len(l)
	}
	return n
}

// splitLines is the document's lines and the byte offset each starts at.
func splitLines(doc []byte) ([]string, []int) {
	lines := strings.Split(string(doc), "\n")
	offsets := make([]int, len(lines))
	at := 0
	for i, l := range lines {
		offsets[i] = at
		at += len(l) + 1
	}
	return lines, offsets
}
