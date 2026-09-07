package finishcmd

import (
	"strings"
)

// The queue's two folders, relative to the repository root. They are
// the methodology's own layout, not this command's choice.
const ()

// unfilled is what the generator writes into a fresh spec's Outcome,
// plus the placeholder a hand-written one tends to carry. Either one is
// an Outcome nobody has written.
var unfilled = map[string]bool{
	"_(fill after execution)_": true,
	"TODO":                     true,
}

// outcomeFilled reports whether the spec's `## Outcome` says anything.
// The section is the lines under the heading up to the next one; blank
// lines and the placeholder do not count as an answer.
func outcomeFilled(content []byte) bool {
	inside := false
	for _, l := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(l, "#") {
			inside = trimmed == "## Outcome"
			continue
		}
		if !inside || trimmed == "" {
			continue
		}
		if !unfilled[trimmed] {
			return true
		}
	}
	return false
}
