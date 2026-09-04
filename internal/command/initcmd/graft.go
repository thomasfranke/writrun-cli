package initcmd

import (
	"bytes"
	"fmt"
)

const (
	sectionHeading = "## WritRun — working the queue"
	markerBegin    = "<!-- writrun:begin"
	markerEnd      = "<!-- writrun:end -->"
)

// graftSection cuts WritRun's part out of the template's AGENTS.md:
// the section heading through the closing marker. The heading rides
// along because it is what makes the fenced block a section a reader
// can find; `writ update` still refreshes only what the markers fence.
func graftSection(templateAgents []byte) ([]byte, error) {
	start := bytes.Index(templateAgents, []byte(sectionHeading))
	if start < 0 {
		start = bytes.Index(templateAgents, []byte(markerBegin))
	}
	end := bytes.Index(templateAgents, []byte(markerEnd))
	if start < 0 || end < start {
		return nil, fmt.Errorf("the template's AGENTS.md carries no fenced WritRun section")
	}
	return templateAgents[start : end+len(markerEnd)], nil
}

// graft appends the fenced section to an existing AGENTS.md, touching
// no byte outside it: everything already there survives verbatim, with
// exactly one blank line before the grafted part (spec-0002).
func graft(existing, section []byte) []byte {
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
