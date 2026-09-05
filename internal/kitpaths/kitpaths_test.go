package kitpaths

import (
	"strings"
	"testing"
)

// The three sets answer one question each, and a path in the wrong one
// is a command reaching where it may not: every workflow the kit owns
// must be both refreshable and removable, and nothing under the
// untouchable prefixes may appear in either.
func TestTheWorkflowsAreBothRefreshedAndRemoved(t *testing.T) {
	refresh := set(RefreshFiles())
	remove := set(RemoveFiles())
	for _, wf := range Workflows {
		if !refresh[wf] {
			t.Errorf("%s is not refreshed", wf)
		}
		if !remove[wf] {
			t.Errorf("%s is not removed", wf)
		}
	}
	if !refresh[".writrun/VERSION"] {
		t.Error("the recorded tag is not refreshed")
	}
}

func TestNothingUntouchableIsEverWritten(t *testing.T) {
	// `.writrun/` is removed whole by uninstall, so the removal set is
	// not held to the refresh set's promise; this is update's.
	for _, rel := range append(RefreshFiles(), RefreshDirs...) {
		for _, forbidden := range Untouchable {
			if rel == forbidden || strings.HasPrefix(rel, forbidden+"/") {
				t.Errorf("%s is refreshed although %s is untouchable", rel, forbidden)
			}
		}
	}
}

func TestTheKeepSetIsNotRemoved(t *testing.T) {
	remove := set(RemoveFiles())
	for _, keep := range Keep {
		if remove[keep] {
			t.Errorf("%s is in the removal set although it is the project's", keep)
		}
		for _, dir := range RemoveDirs {
			if dir == keep {
				t.Errorf("%s is removed whole although it is the project's", keep)
			}
		}
	}
	// The one file under docs/ the kit does own.
	if !remove["docs/writrun-instructions.md"] {
		t.Error("the kit's own instructions doc is not removed")
	}
}

func set(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, i := range items {
		m[i] = true
	}
	return m
}
