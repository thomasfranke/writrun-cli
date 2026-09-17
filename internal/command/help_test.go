package command

import (
	"strings"
	"testing"
)

// named is the command table down to what --help reads from it.
func named(names ...string) []Command {
	var cmds []Command
	for _, n := range names {
		cmds = append(cmds, Command{Name: n, Summary: n + "'s one-liner"})
	}
	return cmds
}

func TestHelpGroupsItsRowsByWhatAPersonIsDoing(t *testing.T) {
	// The production table's own names, so the rows fall in the column
	// the drawing gives them.
	f, out, _ := frame(t, named("init", "doctor", "config", "list", "take",
		"work", "status", "finish", "author", "amend", "report", "update",
		"uninstall"), true, nil)
	if code := Run(f, []string{"--help"}); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	got := out.String()
	want := []string{
		"GETTING STARTED",
		"  init        init's one-liner",
		"  doctor      doctor's one-liner",
		"  config      config's one-liner",
		"DOING THE WORK",
		"  list        list's one-liner",
		"  take        take's one-liner",
		"  work        work's one-liner",
		"  status      status's one-liner",
		"  finish      finish's one-liner",
		"WRITING THE RULES",
		"  author      author's one-liner",
		"  amend       amend's one-liner",
		"  report      report's one-liner",
		"KEEPING IT CURRENT",
		"  update      update's one-liner",
		"  uninstall   uninstall's one-liner",
	}
	at := 0
	for _, w := range want {
		i := strings.Index(got[at:], w+"\n")
		if i < 0 {
			t.Fatalf("help =\n%s\nwant %q after what precedes it", got, w)
		}
		at += i
	}
}

func TestEveryRowsTextIsTheCommandTablesOwnSummary(t *testing.T) {
	// One string, two surfaces: the entry screen builds its rows from
	// the same field, so neither surface can word a command its own way
	// (spec-0039, step 5).
	cmds := named("status")
	cmds[0].Summary = "where does the work on this branch stand?"
	f, out, _ := frame(t, cmds, true, nil)
	if code := Run(f, []string{"--help"}); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "  status   where does the work on this branch stand?") {
		t.Fatalf("help =\n%s\nwant the table's own summary", out.String())
	}
}

func TestACommandNoGroupNamesIsListedAnyway(t *testing.T) {
	f, out, _ := frame(t, named("init", "nogroup"), true, nil)
	if code := Run(f, []string{"--help"}); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "nogroup") {
		t.Fatalf("help =\n%s\nwant the ungrouped command listed", out.String())
	}
}

func TestHelpOpensWithTheProductAndEndsWithTheDocs(t *testing.T) {
	f, out, _ := frame(t, named("init"), true, nil)
	if code := Run(f, []string{"--help"}); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	got := out.String()
	if !strings.HasPrefix(got, Product+" — run a project by WritRun, from the command line.\nWhat is written, runs.\n") {
		t.Fatalf("help =\n%s\nwant the product and what it is", got)
	}
	if !strings.HasSuffix(got, "Run any command with no arguments and it explains itself first.\nDocs: "+docsAddress+"\n") {
		t.Fatalf("help =\n%s\nwant the explanation offered and the docs' address", got)
	}
}
