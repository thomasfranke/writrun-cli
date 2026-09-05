package workcmd

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestANamedAvailableTaskIsTheOneLaunched(t *testing.T) {
	got := work(t, configuredAgent("claude"), scripts(), nil, "task-0011")
	if got.err != nil {
		t.Fatalf("run = %v", got.err)
	}
	if !strings.Contains(got.prompted(t), "Work task-0011") {
		t.Errorf("the named task was not the one launched:\n%s", got.prompted(t))
	}
}

func TestATaskIdIsResolvedHoweverItIsSpelled(t *testing.T) {
	for _, spelling := range []string{"task-0011", "task-11", "TASK-0011", "0011", "11"} {
		s := scripts()
		got := work(t, configuredAgent("claude"), s, nil, spelling)
		if got.err != nil {
			t.Fatalf("%s: run = %v", spelling, got.err)
		}
		last := s.runs[len(s.runs)-1]
		if want := []string{"task-0011"}; !reflect.DeepEqual(last.args, want) {
			t.Errorf("%s: brief args = %v; want %v", spelling, last.args, want)
		}
	}
}

func TestANamedHeldBackTaskIsRefusedInTheListersOwnWords(t *testing.T) {
	got := work(t, configuredAgent("claude"), scripts(), nil, "task-0009")
	got.wantsNoLaunch(t)
	got.wantsError(t, "Held back:")
	got.wantsError(t, "  task-0009  spec-0009 is draft")
}

func TestANamedInFlightTaskIsRefusedUnderTheSectionThatHoldsIt(t *testing.T) {
	got := work(t, configuredAgent("claude"), scripts(), nil, "task-0003")
	got.wantsNoLaunch(t)
	got.wantsError(t, "In flight — an open pull request already exists:")
	got.wantsError(t, "#12 by @someone")
}

func TestAPauseTravelsWithTheTaskItQualifies(t *testing.T) {
	got := work(t, configuredAgent("claude"), scripts(), nil, "task-0005")
	got.wantsNoLaunch(t)
	got.wantsError(t, "paused — spec-0005 is amended by #21")
}

func TestANamedTaskAlreadyInProgressIsRefusedUnderItsOwnHeading(t *testing.T) {
	got := work(t, configuredAgent("claude"), scripts(), nil, "task-0004")
	got.wantsNoLaunch(t)
	got.wantsError(t, "In progress — resume before selecting anything new:")
}

func TestATaskInNoSectionOfTheAnswerIsRefusedAndTheAnswerShown(t *testing.T) {
	got := work(t, configuredAgent("claude"), scripts(), nil, "task-0099")
	got.wantsNoLaunch(t)
	got.wantsError(t, "task-0099 is in no section")
	if !strings.Contains(got.stdout, "Held back:") {
		t.Errorf("the lister's answer was not shown:\n%s", got.stdout)
	}
}

func TestSomethingThatIsNotATaskIdIsRefused(t *testing.T) {
	got := work(t, configuredAgent("claude"), scripts(), nil, "report-0002")
	got.wantsNoLaunch(t)
	got.wantsError(t, "does not name a task")
}

func TestNothingAvailableIsRefusedAndTheListersAnswerShown(t *testing.T) {
	s := scripts()
	s.listing = nothing
	s.listErr = exitErr(1)

	got := work(t, configuredAgent("claude"), s, nil)
	got.wantsNoLaunch(t)
	got.wantsError(t, "nothing is available")
	if !strings.Contains(got.stdout, "Nothing is available.") {
		t.Errorf("output = %q; want the lister's own message", got.stdout)
	}
	if want := []string{listScript}; !reflect.DeepEqual(s.ran(), want) {
		t.Errorf("ran %v; want no brief assembled for a task nobody selected", s.ran())
	}
}

func TestAListerThatCouldNotAnswerAtAllIsAnError(t *testing.T) {
	s := scripts()
	s.listing = ""
	s.listErr = exitErr(3)

	got := work(t, configuredAgent("claude"), s, nil)
	got.wantsNoLaunch(t)
	got.wantsError(t, listScript)
	if exitCode(got.err) != 3 {
		t.Errorf("exit = %d; want the lister's own 3", exitCode(got.err))
	}
}

func TestAListerErrorCarryingNoCodeTravelsUp(t *testing.T) {
	s := scripts()
	s.listErr = errors.New("bash is not installed")
	got := work(t, configuredAgent("claude"), s, nil)
	if !errors.Is(got.err, s.listErr) {
		t.Errorf("err = %v; want the cause preserved", got.err)
	}
}

func TestTheAvailableSectionEndsWhereTheListerEndsIt(t *testing.T) {
	// "Order is a suggestion…" closes the section, and the groups
	// below it are other sections — none of them is available.
	l := parse(answer)
	if want := []string{"task-0007", "task-0011"}; !reflect.DeepEqual(l.available(), want) {
		t.Errorf("available = %v; want %v", l.available(), want)
	}
}

func TestAnEmptyAnswerHoldsNoSections(t *testing.T) {
	if l := parse(""); len(l) != 0 {
		t.Errorf("parse(\"\") = %v; want no sections", l)
	}
	if l := parse("  task-0007  orphaned row\n"); len(l) != 0 {
		t.Errorf("a row under no heading became a section: %v", l)
	}
}
