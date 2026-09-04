package command

import (
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
		help(f)
		return 0
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
			fmt.Fprintf(f.Stderr, "writrun: already adopted — .writrun/ exists at %s\n", root)
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
	fmt.Fprintln(w, "usage: writrun [--version] [--help] [--yes] [--no-color] <command> [args]")
}
