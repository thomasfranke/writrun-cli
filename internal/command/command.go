// Package command is the frame every writrun command plugs into:
// dispatch, adopted-repository detection, the interaction helpers, and
// the reporting discipline of docs/product/rules.md — implemented once.
package command

// Need declares a command's relationship to the adopted repository.
type Need int

const (
	// NeedAny runs anywhere; the repository, when there is one, is
	// resolved best-effort.
	NeedAny Need = iota
	// NeedAdopted requires `.writrun/` at the git toplevel.
	NeedAdopted
	// NeedAbsent requires a repository not yet adopted (init).
	NeedAbsent
)

// Command is one subcommand: its name, the one-line summary --help
// prints, the long description it explains itself with, its need, and
// the work itself.
type Command struct {
	Name    string
	Summary string
	// About is the long description: what the command is for, in plain
	// words, printed by `writrun <command> --help` and by a bare run of
	// a command that would otherwise ask (spec-0039). A command without
	// one is a command a newcomer cannot learn from the binary, which
	// is what the check over the table refuses.
	About About
	Need  Need
	Run   func(ctx *Ctx, args []string) error
	// AsksNothing says this command reads the terminal for nothing: it
	// writes its answer and returns. Only such a command can have its
	// output captured and paged, because a captured question would wait
	// on a reader who cannot see it.
	//
	// The zero value is the safe one. A command added without a thought
	// here keeps the terminal to itself, which is only ever slower to
	// read — never a question asked into the dark.
	AsksNothing bool
	// Daily says this command does its work bare and is run often, so a
	// bare run of it explains nothing (spec-0039). An explanation a
	// person meets every day is read once and skipped after, and
	// `writrun <command> --help` is where it stays reachable.
	//
	// It is not AsksNothing, though the two named the same four commands
	// until `doctor` became a screen. AsksNothing answers whether the
	// screen may capture this command's output — `doctor` reads the
	// terminal now, so it may not. Daily answers whether a person still
	// needs telling what it is for, and `doctor` is run as often as it
	// ever was. One field could not answer both without lying about one
	// of them (report-0044).
	Daily bool
}
