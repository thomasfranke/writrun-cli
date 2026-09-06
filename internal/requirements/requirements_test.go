package requirements

import (
	"errors"
	"reflect"
	"testing"
)

// found is a PATH where every binary but the named ones answers.
func found(gone ...string) func(string) (string, error) {
	missing := map[string]bool{}
	for _, name := range gone {
		missing[name] = true
	}
	return func(name string) (string, error) {
		if missing[name] {
			return "", errors.New("not found")
		}
		return "/usr/bin/" + name, nil
	}
}

func TestAllIsTheScriptsOwn(t *testing.T) {
	want := []string{"git", "bash", "awk", "sed"}
	if !reflect.DeepEqual(All(), want) {
		t.Errorf("All() = %v, want %v", All(), want)
	}
	for _, bin := range All() {
		if bin == "gh" {
			t.Error("gh is a stage-2 requirement, not an environment one")
		}
	}
}

// The list is not state a caller can write: All hands out a copy, so a
// caller that sorts or truncates what it got changes nothing here
// (technical/engineering/boundaries.md).
func TestAllHandsOutACopy(t *testing.T) {
	got := All()
	if len(got) == 0 {
		t.Fatal("All() returned nothing")
	}
	got[0] = "tampered"
	if All()[0] == "tampered" {
		t.Error("writing the returned slice changed the list")
	}
	if Missing(found("tampered")) != nil {
		t.Error("the tampered name reached Missing")
	}
}

func TestMissingIsEmptyWhenEveryBinaryAnswers(t *testing.T) {
	if got := Missing(found()); len(got) != 0 {
		t.Errorf("Missing = %v, want none", got)
	}
}

func TestMissingNamesOnlyWhatIsGone(t *testing.T) {
	if got := Missing(found("awk")); !reflect.DeepEqual(got, []string{"awk"}) {
		t.Errorf("Missing = %v, want [awk]", got)
	}
}

// Every one is named in one pass, in the list's order, so a reader
// installs all of them at once.
func TestMissingNamesEveryOneInOrder(t *testing.T) {
	want := All()
	got := Missing(found(want...))
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Missing = %v, want %v", got, want)
	}
}
