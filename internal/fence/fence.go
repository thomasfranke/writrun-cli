// Package fence edits the WritRun section of an AGENTS.md: the block
// between `writrun:begin` and `writrun:end`. init grafts it, update
// replaces it, uninstall cuts it out — three verbs over one fence, so
// the markers are read in exactly one place.
package fence

import (
	"bytes"
	"fmt"
)

const (
	// Heading is the section title that makes the fenced block findable
	// by a reader; it rides with the section although the markers alone
	// bound what a refresh rewrites.
	Heading = "## WritRun — working the queue"
	// Begin and End are the contract markers. Begin is a prefix: the
	// comment carries prose after it.
	Begin = "<!-- writrun:begin"
	End   = "<!-- writrun:end -->"

	// yoursPrefix marks a block as the project's own answer, which
	// survives every refresh (docs/product/adoption/update.md).
	yoursPrefix = "<!-- yours:"
)

// ErrNoFence is returned where the markers are missing or damaged. The
// commands that carry it stop and change nothing.
var ErrNoFence = fmt.Errorf("no intact writrun:begin/writrun:end fence")

// Section cuts WritRun's part out of the template's AGENTS.md: the
// section heading through the closing marker. The heading rides along
// because it is what makes the fenced block a section a reader can
// find; a refresh still rewrites only what the markers fence.
func Section(templateAgents []byte) ([]byte, error) {
	start, end, err := bounds(templateAgents)
	if err != nil {
		return nil, fmt.Errorf("the template's AGENTS.md carries no fenced WritRun section")
	}
	return templateAgents[start:end], nil
}

// Graft appends the fenced section to an existing AGENTS.md, touching
// no byte outside it: everything already there survives verbatim, with
// exactly one blank line before the grafted part (spec-0002).
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

// Replace swaps the document's fenced section for a fresh one, carrying
// every block the project marked `yours` across unchanged. Nothing
// outside the fence is read or written.
//
// A refresh that carried fewer `yours` markers than the document holds
// would silently drop an answer only the project can give — the gates
// table among them — so it is refused rather than applied (spec-0003:
// the lines marked `yours` shall survive).
func Replace(existing, section []byte) ([]byte, error) {
	start, end, err := bounds(existing)
	if err != nil {
		return nil, err
	}
	merged, err := carryYours(existing[start:end], section)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0, start+len(merged)+len(existing)-end)
	out = append(out, existing[:start]...)
	out = append(out, merged...)
	out = append(out, existing[end:]...)
	return out, nil
}

// Remove cuts the fenced section out, leaving every byte outside it as
// the project wrote it — including the blank line Graft added before
// the section, so a graft followed by a removal is a round trip. It
// reports whether anything but the section was there at all: a document
// that is nothing but the kit's skeleton is the caller's to delete
// whole (spec-0005, edge cases).
func Remove(existing []byte) (out []byte, onlySection bool, err error) {
	start, end, err := bounds(existing)
	if err != nil {
		return nil, false, err
	}
	prefix := existing[:start]
	suffix := existing[end:]
	// The newline closing the marker's own line belongs to the section.
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

// bounds locates the fenced section: from the heading when it is there,
// from the opening marker otherwise, through the closing marker.
func bounds(doc []byte) (start, end int, err error) {
	begin := bytes.Index(doc, []byte(Begin))
	if begin < 0 {
		return 0, 0, ErrNoFence
	}
	closing := bytes.Index(doc, []byte(End))
	if closing < 0 || closing < begin {
		return 0, 0, ErrNoFence
	}
	start = begin
	if h := bytes.LastIndex(doc[:begin], []byte(Heading)); h >= 0 {
		start = h
	}
	return start, closing + len(End), nil
}
