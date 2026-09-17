package command

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// Frame is the production wiring the binary hands the frame: identity,
// streams, ports, and the command table.
type Frame struct {
	Version    string
	WritRunTag string
	Commands   []Command

	Stdout io.Writer
	Stderr io.Writer
	// Stdin is handed on to a command that opens a screen; questions go
	// through Terminal, which holds its own reader.
	Stdin io.Reader

	Terminal Terminal
	// FindRepo walks up from a directory to the git toplevel; adopted
	// says whether `.writrun/` is there.
	FindRepo func(dir string) (root string, adopted bool, err error)
	Getenv   func(string) string
	Getwd    func() (string, error)
	// Screen opens the no-command screen and runs, through run, every
	// command chosen in it, returning when the reader leaves. It is a
	// field rather than a call so the frame keeps no dependency on the
	// screen's engine, and so a suite can drive the routing without one
	// (screens/README.md, spec-0020).
	//
	// The screen runs the commands rather than naming one back because a
	// session outlives them: a command that ended the screen would make
	// reading two things two runs of `writrun`.
	//
	// nil is a binary built without a screen: the no-command path then
	// prints the help, which is what it printed before there was one.
	Screen func(ctx *Ctx, run func(name, arg string, out io.Writer)) error
	// Again runs this binary a second time, as a process of its own, on
	// the terminal this one is holding — and answers when that process
	// has exited.
	//
	// It is what a command chosen on a screen goes through when the
	// command asks a question. A question is a terminal program, and a
	// terminal program's input reader can outlive it: cancelled while
	// already inside its own wait, it goes on holding a read on the
	// terminal and takes the next key for a program that has ended
	// (report-0040, decision 0015). Releasing the terminal does not end
	// that reader; exiting does.
	//
	// It answers an error only when the process could not be started. A
	// command that ran and refused is not this port's failure — a
	// refusal inside a session is read on the terminal, and the session
	// has never read a command's exit code.
	//
	// FirstRun opens the screen `writrun` shows where `.writrun/` is
	// absent, and answers the command a key chose there — empty where
	// the reader left without choosing one. It is a field for the same
	// reason Screen is: the frame keeps no dependency on the screen's
	// engine (screens/README.md, spec-0038).
	//
	// nil is a binary built without that screen: the path then prints
	// the help, which is what it printed before there was one.
	FirstRun func(ctx *Ctx) (string, error)

	// nil is a binary that cannot spawn itself, and so is an error back:
	// the command then runs in this process, which is what every command
	// did before this port existed. A screen that cannot open a question
	// is worse than one that may swallow a key.
	Again func(args []string) error
}

const docsAddress = "https://github.com/thomasfranke/writrun-cli/tree/main/docs"

// Run is the whole frame: global flags, --version and --help anywhere,
// dispatch, the need enforced, the exit status honest. It returns the
// process exit code.
func Run(f Frame, args []string) int {
	var (
		yes     bool
		noColor bool
		rest    []string
		name    string
	)

	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			// End of the frame's flags: everything after is the
			// command's verbatim, reserved names included.
			tail := args[i+1:]
			if name == "" && len(tail) > 0 {
				name, tail = tail[0], tail[1:]
			}
			rest = append(rest, tail...)
			break
		}
		switch {
		case a == "--version":
			fmt.Fprintf(f.Stdout, "%s %s (pins WritRun %s)\n", Product, f.Version, f.WritRunTag)
			return 0
		case a == "--help" || a == "-h":
			// Before the command's name the flag is the frame's own
			// question; after it, it is that command's, and the answer
			// is what the command is for (spec-0039).
			if name == "" {
				help(f)
				return 0
			}
			return commandHelp(f, name)
		case a == "--yes":
			yes = true
		case a == "--no-color":
			noColor = true
		case name == "" && strings.HasPrefix(a, "-"):
			fmt.Fprintf(f.Stderr, "writrun: unknown flag %s\n", a)
			usage(f.Stderr)
			return 2
		case name == "":
			name = a
		default:
			rest = append(rest, a)
		}
	}

	// The screen is a session: it runs what is chosen in it and comes
	// back, until the reader leaves. A command still owns the terminal
	// alone while it asks its questions — the screen is paused and the
	// terminal released, so the two never read a keyboard at once
	// (docs/product/screens/README.md, internal/screen/session.go).
	if name == "" {
		return openScreen(f, noColor, yes)
	}
	return dispatch(f, noColor, yes, name, rest)
}

func dispatch(f Frame, noColor, yes bool, name string, rest []string) int {
	cmd, ok := lookup(f.Commands, name)
	if !ok {
		fmt.Fprintf(f.Stderr, "writrun: unknown command %q\n", name)
		usage(f.Stderr)
		return 2
	}

	ctx := &Ctx{
		Stdout:   f.Stdout,
		Stderr:   f.Stderr,
		Stdin:    f.Stdin,
		Terminal: f.Terminal,
		Yes:      yes,
		Color:    colorEnabled(f.Terminal.InteractiveOut(), noColor, f.Getenv),
		Version:  f.Version,
		Again:    spawner(f, noColor, yes),
	}

	if code, failed := resolveNeed(f, cmd.Need, ctx); failed {
		return code
	}

	// A command run with nothing to go on says what it is for before it
	// asks its first question (spec-0039). A command that does its work
	// bare — `list`, `status`, `doctor`, `work` — is run daily, and a
	// daily explanation is noise; `Daily` is where that set is declared.
	//
	// It was `AsksNothing` until `doctor` became a screen, and the two
	// sets came apart there: `doctor` reads the terminal, so the screen
	// may not capture it, and it is still run every day. Printing into
	// the normal buffer before entering the alternate one would not even
	// be noise — the screen wipes it, and the reader meets it on the way
	// out (report-0044).
	//
	// The terminal that decides it is stdin's, because the description
	// precedes a question and a question is asked there. Without one
	// nothing is printed: a script reading this output keeps reading
	// what it read before.
	if len(rest) == 0 && !cmd.Daily && !cmd.About.Empty() && f.Terminal.InteractiveIn() {
		writeAbout(f.Stdout, cmd.Name, cmd.About)
		fmt.Fprintln(f.Stdout)
	}

	if err := cmd.Run(ctx, rest); err != nil {
		if errors.Is(err, ErrDeclined) {
			fmt.Fprintf(f.Stderr, "writrun %s: declined — nothing changed\n", cmd.Name)
			return 1
		}
		// A wrapped script's exit code is its own verdict, already
		// reported on stderr — pass it through instead of restating it.
		var verdict interface{ ExitCode() int }
		if errors.As(err, &verdict) && verdict.ExitCode() > 0 {
			return verdict.ExitCode()
		}
		fmt.Fprintf(f.Stderr, "writrun %s: %v\n", cmd.Name, err)
		return 1
	}
	return 0
}

// resolveNeed enforces the command's declared relationship to the
// repository; a refusal names the cause and changes nothing.
func resolveNeed(f Frame, need Need, ctx *Ctx) (int, bool) {
	wd, err := f.Getwd()
	if err != nil {
		fmt.Fprintf(f.Stderr, "writrun: %v\n", err)
		return 1, true
	}
	root, adopted, err := f.FindRepo(wd)
	switch need {
	case NeedAny:
		if err == nil {
			ctx.Root, ctx.Adopted = root, adopted
		}
		return 0, false
	case NeedAdopted:
		if err != nil {
			fmt.Fprintf(f.Stderr, "writrun: %v\n", err)
			return 1, true
		}
		if !adopted {
			fmt.Fprintf(f.Stderr, "writrun: not an adopted repository — no .writrun/ at %s\n", root)
			return 1, true
		}
	case NeedAbsent:
		if err != nil {
			fmt.Fprintf(f.Stderr, "writrun: %v\n", err)
			return 1, true
		}
		if adopted {
			fmt.Fprintf(f.Stderr, "writrun: already adopted — .writrun/ exists at %s; writrun update refreshes an adopted kit\n", root)
			return 1, true
		}
	}
	ctx.Root, ctx.Adopted = root, adopted
	return 0, false
}

func lookup(cmds []Command, name string) (Command, bool) {
	for _, c := range cmds {
		if c.Name == name {
			return c, true
		}
	}
	return Command{}, false
}

// Product is what the binary calls itself, in one place, because a
// second copy is a second name. The command a person types is `writrun`
// and is unchanged; this is the name the product carries, which the
// repository, the module and the formula already use
// (decisions/runtime/0014-the-product-is-writrun-cli.md).
const Product = "writrun-cli"

// helpGroups is --help's grouping: what a person is doing, in the order
// docs/product/screens/help.excalidraw draws it. Thirteen rows in table
// order name every command and teach none of them.
//
// It is not the entry screen's grouping, which is the queue's own
// sections (cmd/writrun/main.go). Two surfaces, two questions: the
// screen lists the work, and this list is read by someone meeting the
// commands for the first time.
var helpGroups = []struct {
	name  string
	names []string
}{
	{"GETTING STARTED", []string{"init", "doctor", "config"}},
	{"DOING THE WORK", []string{"list", "take", "work", "status", "finish"}},
	{"WRITING THE RULES", []string{"author", "amend", "report"}},
	{"KEEPING IT CURRENT", []string{"update", "uninstall"}},
}

// help is the grouped table plus the docs' address. Every row's text is
// the command table's own summary — the string the entry screen shows,
// so the two surfaces cannot drift (spec-0039).
func help(f Frame) {
	fmt.Fprintln(f.Stdout, Product+" — run a project by WritRun, from the command line.")
	fmt.Fprintln(f.Stdout, "What is written, runs.")

	width := 0
	for _, c := range f.Commands {
		if len(c.Name) > width {
			width = len(c.Name)
		}
	}
	grouped := map[string]bool{}
	for _, g := range helpGroups {
		var rows []Command
		for _, n := range g.names {
			if c, ok := lookup(f.Commands, n); ok {
				rows = append(rows, c)
				grouped[n] = true
			}
		}
		if len(rows) == 0 {
			continue
		}
		fmt.Fprintln(f.Stdout)
		fmt.Fprintln(f.Stdout, g.name)
		for _, c := range rows {
			helpRow(f, width, c)
		}
	}
	// A command no group names is listed anyway, under no heading: the
	// help answers for the table it was handed, and a command missing
	// from the grouping is a gap a test names rather than one a reader
	// has to notice by its absence.
	var ungrouped []Command
	for _, c := range f.Commands {
		if !grouped[c.Name] {
			ungrouped = append(ungrouped, c)
		}
	}
	if len(ungrouped) > 0 {
		fmt.Fprintln(f.Stdout)
		for _, c := range ungrouped {
			helpRow(f, width, c)
		}
	}

	fmt.Fprintln(f.Stdout)
	fmt.Fprintln(f.Stdout)
	fmt.Fprintln(f.Stdout, "Run any command with no arguments and it explains itself first.")
	fmt.Fprintln(f.Stdout, "Docs: "+docsAddress)
}

// helpRow is one command's row: its name and the table's own summary.
func helpRow(f Frame, width int, c Command) {
	fmt.Fprintf(f.Stdout, "  %-*s   %s\n", width, c.Name, c.Summary)
}

// commandHelp answers `writrun <command> --help`: the long description
// and nothing else. It answers anywhere — the need is the command's to
// enforce when it runs, and what a command is for is readable outside
// an adopted repository.
func commandHelp(f Frame, name string) int {
	cmd, ok := lookup(f.Commands, name)
	if !ok {
		fmt.Fprintf(f.Stderr, "writrun: unknown command %q\n", name)
		usage(f.Stderr)
		return 2
	}
	writeAbout(f.Stdout, cmd.Name, cmd.About)
	return 0
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "usage: writrun [--version] [--help] [--yes] [--no-color] <command> [--] [args]")
}

// openScreen answers `writrun` with no command. The screen needs a
// terminal at both ends; without one there is no screen to open, and
// the help is what the rule prescribes rather than a fallback this
// invented. Where the terminal is there and the kit is not, the screen
// is the first run's (screens/README.md, spec-0038).
//
// It returns the process's exit code. A command that refuses inside the
// session does not end it and does not decide it: the reader saw the
// refusal, read it, and came back — leaving the screen is what ends the
// run, and leaving is not a failure.
// spawner is the frame's port with the session's flags already on it,
// which is the form a command wants: `config` names a key and never a
// flag it did not parse. nil stays nil — a frame without the port hands
// a command nothing to fall back from.
func spawner(f Frame, noColor, yes bool) func([]string) error {
	if f.Again == nil {
		return nil
	}
	return func(args []string) error { return f.Again(withFlags(args, noColor, yes)) }
}

// sessionArgs is the command line the session would have typed: the
// frame's flags as this run received them, then the command and the one
// argument a row may carry.
//
// The flags lead because the parser takes the first bare word as the
// command name, and a session running under `--yes` must not be asked a
// question the parent had already answered.
func sessionArgs(name string, rest []string, noColor, yes bool) []string {
	return withFlags(append([]string{name}, rest...), noColor, yes)
}

// withFlags prepends the frame's flags to a command line the session is
// about to hand to a process of its own. It is the one place they are
// spelled, so a caller naming a command never has to know them.
func withFlags(args []string, noColor, yes bool) []string {
	var flags []string
	if noColor {
		flags = append(flags, "--no-color")
	}
	if yes {
		flags = append(flags, "--yes")
	}
	return append(flags, args...)
}

// openFirstRun answers `writrun` with no command where `.writrun/` is
// absent: the screen that says what this is, what the environment
// answers, and what to do next. A binary built without that screen
// prints the help, as it did before there was one (spec-0038).
//
// The command a key chose runs after the screen has closed, not inside
// it: `init` asks questions, and a terminal program's input reader can
// outlive the program it reads for (decision 0015, report-0040). There
// is no session here — the screen's whole offer is one command — so
// closing it is enough, and the spawn is what makes *alone* a fact
// where the port is wired.
func openFirstRun(f Frame, noColor, yes bool, root string) int {
	if f.FirstRun == nil {
		help(f)
		return 0
	}
	ctx := &Ctx{
		Stdout:   f.Stdout,
		Stderr:   f.Stderr,
		Stdin:    f.Stdin,
		Terminal: f.Terminal,
		Yes:      yes,
		Color:    colorEnabled(f.Terminal.InteractiveOut(), noColor, f.Getenv),
		Version:  f.Version,
		Root:     root,
		Again:    spawner(f, noColor, yes),
	}
	name, err := f.FirstRun(ctx)
	if err != nil {
		fmt.Fprintf(f.Stderr, "writrun: %v\n", err)
		return 1
	}
	if name == "" {
		return 0
	}
	if f.Again != nil {
		if err := f.Again(sessionArgs(name, nil, noColor, yes)); err == nil {
			return 0
		}
	}
	return dispatch(f, noColor, yes, name, nil)
}

func openScreen(f Frame, noColor, yes bool) int {
	if !f.Terminal.InteractiveIn() || !f.Terminal.InteractiveOut() {
		help(f)
		return 0
	}
	// Adoption is read rather than enforced: outside one the screen is
	// the first run's, not the refusal NeedAdopted would print. This is
	// the one caller that wants the fact without the verdict.
	wd, err := f.Getwd()
	if err != nil {
		fmt.Fprintf(f.Stderr, "writrun: %v\n", err)
		return 1
	}
	root, adopted, err := f.FindRepo(wd)
	if err != nil || !adopted {
		return openFirstRun(f, noColor, yes, root)
	}
	if f.Screen == nil {
		help(f)
		return 0
	}
	ctx := &Ctx{
		Stdout:   f.Stdout,
		Stderr:   f.Stderr,
		Stdin:    f.Stdin,
		Terminal: f.Terminal,
		Yes:      yes,
		Color:    colorEnabled(f.Terminal.InteractiveOut(), noColor, f.Getenv),
		Version:  f.Version,
		Root:     root,
		Adopted:  adopted,
		Again:    spawner(f, noColor, yes),
	}
	// The screen runs each chosen command through this, and the command
	// reports itself on the terminal the screen released — so nothing is
	// handed back to say, and an exit code is not one either. The one a
	// command answers belongs to `writrun <command>`, where it is a
	// script's to read; in a session there is no script to read it.
	err = f.Screen(ctx, func(name, arg string, out io.Writer) {
		var rest []string
		if arg != "" {
			rest = []string{arg}
		}
		// A command that asks runs as its own process, so that nothing
		// of it is left reading this one's terminal when the screen
		// comes back (spec-0044). A captured command asks nothing, opens
		// no terminal program, and has no reader to outlive it — it
		// stays here, where its output can be paged.
		//
		// Everything the reader sees is the screen's still: the pause,
		// the released terminal, the line naming what is running, and
		// the wait for the return. Only where the command runs changes.
		if out == nil && f.Again != nil {
			if err := f.Again(sessionArgs(name, rest, noColor, yes)); err == nil {
				return
			}
		}
		g := f
		// A captured command writes where the screen can page it, and
		// both streams go to the one place: a reader reads one account,
		// in the order it was written, not a report with its warnings
		// filed somewhere else.
		if out != nil {
			g.Stdout, g.Stderr = out, out
		}
		dispatch(g, noColor, yes, name, rest)
	})
	if err != nil {
		fmt.Fprintf(f.Stderr, "writrun: %v\n", err)
		return 1
	}
	return 0
}
