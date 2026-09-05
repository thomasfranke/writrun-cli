package takecmd

import (
	"strings"
	"testing"
)

// listing is the lister's shape: the Available group, then the groups
// that are named and never offered.
const listing = `Available — any of these may be taken:
  task-0004  medium  Report repository health with writrun doctor
  task-0008  medium  Record an observation with writrun report

Order is a suggestion for a person and binding for an agent.

In flight — an open pull request already exists:
  task-0006  #37 by @someone Show the queue with writrun list

Held back:
  task-0007  waiting on task-0006 (ready)

Open reports — waiting to be triaged, never selected:
  report-0005  something was noticed
`

func TestParseAvailableOffersOnlyTheAvailableGroup(t *testing.T) {
	ids, labels := parseAvailable(listing)
	want := []string{"task-0004", "task-0008"}
	if strings.Join(ids, " ") != strings.Join(want, " ") {
		t.Fatalf("ids = %v, want %v", ids, want)
	}
	if len(labels) != 2 || !strings.Contains(labels[0], "Report repository health") {
		t.Fatalf("labels = %v; want the lister's own lines", labels)
	}
}

func TestParseAvailableFindsNothingWithoutTheGroup(t *testing.T) {
	ids, _ := parseAvailable("Held back:\n  task-0007  waiting on task-0006 (ready)\n")
	if len(ids) != 0 {
		t.Fatalf("ids = %v; want none where nothing is available", ids)
	}
}

func TestTheTaskIsArrowSelectedWhenNoneIsGiven(t *testing.T) {
	h := newHarness(t, reply{out: listing}, reply{})
	h.term.In = true
	h.term.SelectIndex = 1
	h.term.InputAnswer = title

	if err := h.take(); err != nil {
		t.Fatalf("take = %v", err)
	}
	if len(h.scripts.calls) != 2 {
		t.Fatalf("%d calls, want the lister and the take", len(h.scripts.calls))
	}
	if h.scripts.calls[0].script != listScript {
		t.Errorf("first script = %q, want the lister", h.scripts.calls[0].script)
	}
	if got := h.argsOf(t, 1); got != "task-0008 --title "+title {
		t.Errorf("args = %q; want the selected task", got)
	}
	if len(h.term.Asked) != 2 || !strings.Contains(h.term.Asked[0], "which task") {
		t.Errorf("asked %v; want the selection then the title", h.term.Asked)
	}
}

// The listing is the lister's answer to what is held back, so it is
// shown rather than summarised — and nothing is taken.
func TestNothingAvailableStopsWithTheListersOwnListing(t *testing.T) {
	held := "Held back:\n  task-0007  waiting on task-0006 (ready)\n"
	h := newHarness(t, reply{out: held, err: scriptExit(1)})
	h.term.In = true

	err := h.take()
	if err == nil || !strings.Contains(err.Error(), "no task is available") {
		t.Fatalf("err = %v; want the empty group named", err)
	}
	if !strings.Contains(h.out.String(), "waiting on task-0006") {
		t.Errorf("stdout = %q; want the lister's listing shown", h.out.String())
	}
	if len(h.scripts.calls) != 1 {
		t.Errorf("%d calls; nothing is taken when nothing is available", len(h.scripts.calls))
	}
}

func TestAListerFailureNamesTheScript(t *testing.T) {
	h := newHarness(t, reply{err: scriptExit(3)})
	h.term.In = true
	err := h.take()
	if err == nil || !strings.Contains(err.Error(), listScript) {
		t.Fatalf("err = %v; want the lister named", err)
	}
	if len(h.scripts.calls) != 1 {
		t.Errorf("%d calls; the take never ran", len(h.scripts.calls))
	}
}

func TestWithoutATerminalTheTaskIdIsRequired(t *testing.T) {
	h := newHarness(t, reply{out: listing})
	err := h.take("--title", title)
	if err == nil || !strings.Contains(err.Error(), "the task id as an argument") {
		t.Fatalf("err = %v; want the argument named", err)
	}
	if len(h.scripts.calls) != 1 {
		t.Errorf("%d calls; only the lister ran", len(h.scripts.calls))
	}
}
