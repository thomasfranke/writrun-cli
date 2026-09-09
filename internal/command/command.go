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
// prints, its need, and the work itself.
type Command struct {
	Name    string
	Summary string
	Need    Need
	Run     func(ctx *Ctx, args []string) error
	// AsksNothing says this command reads the terminal for nothing: it
	// writes its answer and returns. Only such a command can have its
	// output captured and paged, because a captured question would wait
	// on a reader who cannot see it.
	//
	// The zero value is the safe one. A command added without a thought
	// here keeps the terminal to itself, which is only ever slower to
	// read — never a question asked into the dark.
	AsksNothing bool
}
