package command

import (
	"strings"
	"testing"
)

// explains is one command carrying a long description; asks says
// whether it is a command that would ask something — a daily one, which
// does its work bare, is the case that explains nothing.
func explains(name string, asks bool) Command {
	return Command{
		Name:        name,
		Summary:     "a one-liner",
		AsksNothing: !asks,
		Daily:       !asks,
		Need:        NeedAny,
		About: About{
			Sentence: []string{"what it is for", "said on a second line"},
			Parts: []AboutPart{
				{Label: "why you would", Lines: []string{"a first reason", "continued"}},
				{Label: "what it does", Lines: []string{"the work itself"}},
			},
		},
		Run: func(*Ctx, []string) error { return nil },
	}
}

// theBlock is the long description as the drawings state it: the
// sentence beside the command's name, one blank line above each part,
// the label in its own column and the text in the next.
const theBlock = `amend — what it is for
said on a second line

 why you would    a first reason
                  continued

 what it does     the work itself
`

func TestCommandHelpPrintsTheLongDescription(t *testing.T) {
	f, out, _ := frame(t, []Command{explains("amend", true)}, true, nil)
	if code := Run(f, []string{"amend", "--help"}); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if out.String() != theBlock {
		t.Fatalf("amend --help =\n%q\nwant\n%q", out.String(), theBlock)
	}
}

func TestCommandHelpAnswersForACommandThatDoesItsWorkBare(t *testing.T) {
	// `--help` is every command's, not only the ones that explain
	// themselves when run bare.
	f, out, _ := frame(t, []Command{explains("list", false)}, true, nil)
	if code := Run(f, []string{"list", "--help"}); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "list — what it is for") {
		t.Fatalf("list --help = %q; want the long description", out.String())
	}
}

func TestCommandHelpNeitherRunsTheCommandNorPrintsTheGlobalHelp(t *testing.T) {
	ran := false
	cmd := explains("amend", true)
	cmd.Run = func(*Ctx, []string) error { ran = true; return nil }
	f, out, _ := frame(t, []Command{cmd}, true, nil)
	if code := Run(f, []string{"amend", "--help"}); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if ran {
		t.Fatal("the command ran; --help asks what it is for, not for the work")
	}
	if strings.Contains(out.String(), docsAddress) {
		t.Fatalf("amend --help = %q; the global help answered a command's own", out.String())
	}
}

func TestHelpForAnUnknownCommandIsRefusedAsTheCommandWouldBe(t *testing.T) {
	f, _, errb := frame(t, []Command{explains("amend", true)}, true, nil)
	if code := Run(f, []string{"bogus", "--help"}); code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(errb.String(), `unknown command "bogus"`) {
		t.Fatalf("stderr = %q; want the unknown command named", errb.String())
	}
}

func TestABareRunExplainsItselfBeforeItAsksAnything(t *testing.T) {
	f, out, _ := frame(t, []Command{explains("amend", true)}, true, nil)
	// What had been printed by the time the command itself began: the
	// description comes before the first question, never after it.
	var before string
	f.Commands[0].Run = func(*Ctx, []string) error { before = out.String(); return nil }
	f.Terminal = &FakeTerminal{In: true}
	if code := Run(f, []string{"amend"}); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if want := theBlock + "\n"; before != want {
		t.Fatalf("printed before the command ran =\n%q\nwant\n%q", before, want)
	}
}

func TestABareRunOfADailyCommandExplainsNothing(t *testing.T) {
	// `list`, `status`, `doctor` and `work` are run daily, and a daily
	// explanation is noise (spec-0039, step 4). The set is `Daily`'s to
	// declare and not `AsksNothing`'s: `doctor` reads the terminal now,
	// so the screen may not capture it, and it is still run every day
	// (report-0044).
	f, out, _ := frame(t, []Command{explains("list", false)}, true, nil)
	f.Terminal = &FakeTerminal{In: true}
	if code := Run(f, []string{"list"}); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if out.Len() != 0 {
		t.Fatalf("list = %q; want its work and no description", out.String())
	}
}

func TestAnArgumentSkipsTheExplanation(t *testing.T) {
	f, out, _ := frame(t, []Command{explains("amend", true)}, true, nil)
	f.Terminal = &FakeTerminal{In: true}
	if code := Run(f, []string{"amend", "spec-0011"}); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if out.Len() != 0 {
		t.Fatalf("amend spec-0011 = %q; a run that names what it needs explains nothing", out.String())
	}
}

func TestWithoutATerminalTheExplanationIsSkipped(t *testing.T) {
	// A script reading this output keeps reading what it read before.
	f, out, _ := frame(t, []Command{explains("amend", true)}, true, nil)
	if code := Run(f, []string{"amend"}); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if out.Len() != 0 {
		t.Fatalf("amend = %q; want nothing where there is no terminal", out.String())
	}
}
