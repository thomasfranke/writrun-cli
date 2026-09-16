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
			help(f)
			return 0
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
		Again:    spawner(f, noColor, yes),
	}

	if code, failed := resolveNeed(f, cmd.Need, ctx); failed {
		return code
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

// help is one line per command plus the docs' address — it restates
// nothing (product/rules.md).
func help(f Frame) {
	fmt.Fprintln(f.Stdout, Product+" — the porcelain for WritRun.")
	if len(f.Commands) > 0 {
		fmt.Fprintln(f.Stdout)
		width := 0
		for _, c := range f.Commands {
			if len(c.Name) > width {
				width = len(c.Name)
			}
		}
		for _, c := range f.Commands {
			fmt.Fprintf(f.Stdout, "  %-*s  %s\n", width, c.Name, c.Summary)
		}
	}
	fmt.Fprintln(f.Stdout)
	fmt.Fprintln(f.Stdout, "Docs: "+docsAddress)
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "usage: writrun [--version] [--help] [--yes] [--no-color] <command> [--] [args]")
}

// openScreen answers `writrun` with no command. The screen needs a
// terminal at both ends and an adopted repository; without either there
// is no screen to open, and the help is what the rule prescribes rather
// than a fallback this invented (screens/README.md).
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

func openScreen(f Frame, noColor, yes bool) int {
	if f.Screen == nil || !f.Terminal.InteractiveIn() || !f.Terminal.InteractiveOut() {
		help(f)
		return 0
	}
	// Adoption is read rather than enforced: outside one the rule asks
	// for the help, not for the refusal NeedAdopted would print. This is
	// the one caller that wants the fact without the verdict.
	wd, err := f.Getwd()
	if err != nil {
		fmt.Fprintf(f.Stderr, "writrun: %v\n", err)
		return 1
	}
	root, adopted, err := f.FindRepo(wd)
	if err != nil || !adopted {
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
