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
}
