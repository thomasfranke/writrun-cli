// Package uninstallcmd is `writrun uninstall`: what init installed,
// removed. The queue and the project's docs are not the kit's and
// survive it — what the methodology helped write belongs to the
// repository (docs/product/adoption/uninstall.md, spec-0005).
package uninstallcmd

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/fence"
	"github.com/thomasfranke/writrun-cli/internal/hook"
	"github.com/thomasfranke/writrun-cli/internal/kitpaths"
)

// Deps is the wiring uninstall needs beyond the frame's Ctx.
type Deps struct {
	// Git runs one git invocation — the hooks directory is git's to
	// name, not a path this command may assume.
	Git hook.GitRunner
}

// New returns the uninstall command wired with its dependencies.
func New(d Deps) command.Command {
	return command.Command{
		Name:    "uninstall",
		Summary: "remove the WritRun kit, keeping the project's record",
		Need:    command.NeedAdopted,
		Run: func(ctx *command.Ctx, args []string) error {
			return run(ctx, d, args)
		},
	}
}

func run(ctx *command.Ctx, d Deps, args []string) error {
	fs := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}

	hookAt, err := hook.Path(ctx.Root, d.Git)
	if err != nil {
		return err
	}
	r, err := plan(ctx.Root, hookAt)
	if err != nil {
		return err
	}
	r.render(ctx.Stdout)
	if err := ctx.AskConfirm("Remove the WritRun kit from this repository?"); err != nil {
		return err
	}
	if err := r.apply(); err != nil {
		return fmt.Errorf("%w — the removal is partial; rerun writrun uninstall to finish it", err)
	}
	fmt.Fprintln(ctx.Stdout, "Removed the WritRun kit. The queue and the project's docs are untouched.")
	return nil
}

// removal is the whole plan: what goes, what was already gone, and
// what stays — computed before anything is deleted and shown before
// the confirmation.
type removal struct {
	root string

	dirs  []string // kit-owned directories to delete
	files []string // kit-owned files to delete
	gone  []string // named in the kit's inventory, already not there

	hookAt    string
	hookState hook.State

	// agents is what AGENTS.md becomes: nil with agentsDelete set
	// means the file was nothing but the kit's skeleton.
	agents      []byte
	agentsWhole bool
	agentsKept  bool // no fence found; the file is the project's alone
}

func plan(root, hookAt string) (*removal, error) {
	r := &removal{root: root, hookAt: hookAt}

	for _, dir := range kitpaths.RemoveDirs {
		if _, err := os.Stat(filepath.Join(root, dir)); err == nil {
			r.dirs = append(r.dirs, dir)
		} else {
			r.gone = append(r.gone, dir)
		}
	}
	for _, rel := range kitpaths.RemoveFiles() {
		if _, err := os.Stat(filepath.Join(root, rel)); err == nil {
			r.files = append(r.files, rel)
		} else {
			r.gone = append(r.gone, rel)
		}
	}

	state, err := hook.Inspect(hookAt)
	if err != nil {
		return nil, err
	}
	r.hookState = state

	agents, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	switch {
	case os.IsNotExist(err):
		r.agentsKept = true
		r.gone = append(r.gone, "AGENTS.md")
	case err != nil:
		return nil, fmt.Errorf("reading AGENTS.md: %w", err)
	default:
		out, only, fenceErr := fence.Remove(agents)
		switch {
		case fenceErr != nil:
			// No fence to cut: whatever this file is, it is the
			// project's, and uninstall does not guess at its shape.
			r.agentsKept = true
		case only:
			r.agentsWhole = true
		default:
			r.agents = out
		}
	}
	return r, nil
}

// render shows both sets, because the confirmation is about both: what
// goes, and what a person is being promised will stay.
func (r *removal) render(w io.Writer) {
	fmt.Fprintln(w, "writrun uninstall — the plan; nothing is removed before the confirmation:")
	fmt.Fprintln(w)
	for _, dir := range r.dirs {
		fmt.Fprintf(w, "  remove       %s/ — the kit, whole\n", dir)
	}
	for _, rel := range r.files {
		fmt.Fprintf(w, "  remove       %s\n", rel)
	}
	switch {
	case r.agentsWhole:
		fmt.Fprintln(w, "  remove       AGENTS.md — nothing in it but the kit's own section")
	case r.agents != nil:
		fmt.Fprintln(w, "  edit         AGENTS.md — the fenced section only; every byte outside it stays")
	case r.agentsKept:
		fmt.Fprintln(w, "  kept         AGENTS.md — no fenced WritRun section found; left as the project wrote it")
	}
	switch r.hookState {
	case hook.Ours:
		fmt.Fprintf(w, "  remove       %s — the commit-msg hook the adoption installed\n", r.hookDisplay())
	case hook.Foreign:
		fmt.Fprintf(w, "  kept         %s — the installed hook is not the one init writes; it is another project's to remove\n", r.hookDisplay())
	case hook.Absent:
		fmt.Fprintf(w, "  kept         %s — no commit-msg hook is installed\n", r.hookDisplay())
	}
	for _, rel := range r.gone {
		fmt.Fprintf(w, "  already gone %s\n", rel)
	}
	fmt.Fprintln(w)
	for _, rel := range kitpaths.Keep {
		fmt.Fprintf(w, "  stays        %s/ — the project's, not the kit's\n", rel)
	}
	fmt.Fprintln(w)
}

func (r *removal) hookDisplay() string {
	rel, err := filepath.Rel(r.root, r.hookAt)
	if err != nil || len(rel) > 1 && rel[0] == '.' && rel[1] == '.' {
		return r.hookAt
	}
	return filepath.ToSlash(rel)
}

// apply performs exactly the rendered plan.
func (r *removal) apply() error {
	for _, dir := range r.dirs {
		if err := os.RemoveAll(filepath.Join(r.root, dir)); err != nil {
			return fmt.Errorf("removing %s: %w", dir, err)
		}
	}
	for _, rel := range r.files {
		if err := os.Remove(filepath.Join(r.root, rel)); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("removing %s: %w", rel, err)
		}
	}
	agentsPath := filepath.Join(r.root, "AGENTS.md")
	switch {
	case r.agentsWhole:
		if err := os.Remove(agentsPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("removing AGENTS.md: %w", err)
		}
	case r.agents != nil:
		if err := os.WriteFile(agentsPath, r.agents, 0o644); err != nil {
			return fmt.Errorf("editing AGENTS.md: %w", err)
		}
	}
	if r.hookState == hook.Ours {
		if err := os.Remove(r.hookAt); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("removing the commit-msg hook: %w", err)
		}
	}
	return nil
}
