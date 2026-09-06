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

	Terminal Terminal
	// FindRepo walks up from a directory to the git toplevel; adopted
	// says whether `.writrun/` is there.
	FindRepo func(dir string) (root string, adopted bool, err error)
	Getenv   func(string) string
	Getwd    func() (string, error)
	// Screen opens the no-command queue screen and returns the command
	// it dispatched to, empty when the user left without choosing. It
	// is a field rather than a call so the frame keeps no dependency on
	// the screen's engine, and so a suite can drive the routing without
	// one (screen.md, spec-0020).
	//
	// nil is a binary built without a screen: the no-command path then
	// prints the help, which is what it printed before there was one.
	Screen func(ctx *Ctx) (name string, arg string, err error)
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
			fmt.Fprintf(f.Stdout, "writrun %s (pins WritRun %s)\n", f.Version, f.WritRunTag)
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

	if name == "" {
		code, dispatched, cmdName, cmdArg := openScreen(f, noColor, yes)
		if !dispatched {
			return code
		}
		name, rest = cmdName, nil
		if cmdArg != "" {
			rest = []string{cmdArg}
		}
	}

	cmd, ok := lookup(f.Commands, name)
	if !ok {
		fmt.Fprintf(f.Stderr, "writrun: unknown command %q\n", name)
		usage(f.Stderr)
		return 2
	}

	ctx := &Ctx{
		Stdout:   f.Stdout,
		Stderr:   f.Stderr,
		Terminal: f.Terminal,
		Yes:      yes,
		Color:    colorEnabled(f.Terminal.InteractiveOut(), noColor, f.Getenv),
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

// help is one line per command plus the docs' address — it restates
// nothing (product/rules.md).
func help(f Frame) {
	fmt.Fprintln(f.Stdout, "writrun — the porcelain for WritRun.")
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
// than a fallback this invented (screen.md).
//
// It returns the exit code to use when nothing was dispatched, and the
// command to run when something was.
func openScreen(f Frame, noColor, yes bool) (code int, dispatched bool, name, arg string) {
	if f.Screen == nil || !f.Terminal.InteractiveIn() || !f.Terminal.InteractiveOut() {
		help(f)
		return 0, false, "", ""
	}
	// Adoption is read rather than enforced: outside one the rule asks
	// for the help, not for the refusal NeedAdopted would print. This is
	// the one caller that wants the fact without the verdict.
	wd, err := f.Getwd()
	if err != nil {
		fmt.Fprintf(f.Stderr, "writrun: %v\n", err)
		return 1, false, "", ""
	}
	root, adopted, err := f.FindRepo(wd)
	if err != nil || !adopted {
		help(f)
		return 0, false, "", ""
	}
	ctx := &Ctx{
		Stdout:   f.Stdout,
		Stderr:   f.Stderr,
		Terminal: f.Terminal,
		Yes:      yes,
		Color:    colorEnabled(f.Terminal.InteractiveOut(), noColor, f.Getenv),
		Root:     root,
		Adopted:  adopted,
	}
	name, arg, err = f.Screen(ctx)
	if err != nil {
		fmt.Fprintf(f.Stderr, "writrun: %v\n", err)
		return 1, false, "", ""
	}
	if name == "" {
		return 0, false, "", ""
	}
	return 0, true, name, arg
}
