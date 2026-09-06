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

func TestBinariesAreTheScriptsOwn(t *testing.T) {
	want := []string{"git", "bash", "awk", "sed"}
	if !reflect.DeepEqual(Binaries, want) {
		t.Errorf("Binaries = %v, want %v", Binaries, want)
	}
	for _, bin := range Binaries {
		if bin == "gh" {
			t.Error("gh is a stage-2 requirement, not an environment one")
		}
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
	got := Missing(found(Binaries...))
	if !reflect.DeepEqual(got, Binaries) {
		t.Errorf("Missing = %v, want %v", got, Binaries)
	}
}
