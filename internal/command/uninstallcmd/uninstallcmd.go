// Package uninstallcmd is `writrun uninstall`: what init installed,
// removed. The queue and the project's docs are not the kit's and
// survive it — what the methodology helped write belongs to the
// repository (docs/product/adoption/uninstall.md, spec-0005).
package uninstallcmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/hook"
	"github.com/thomasfranke/writrun-cli/internal/kitpaths"
	"github.com/thomasfranke/writrun-cli/internal/pointer"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// Deps is the wiring uninstall needs beyond the frame's Ctx.
type Deps struct {
	// Git runs one git invocation — the hooks directory is git's to
	// name, not a path this command may assume.
	Git gitx.Runner
	// Files is the filesystem this command reads and writes through.
	Files vfs.FS
}

// New returns the uninstall command wired with its dependencies.
func New(d Deps) command.Command {
	return command.Command{
		Name:    "uninstall",
		Summary: "remove WritRun, keep everything it helped you write",
		About: command.About{
			Sentence: []string{
				"remove WritRun, and keep everything it helped you",
				"write",
			},
			Parts: []command.AboutPart{
				{Label: "why you would", Lines: []string{
					"The methodology is not for this project after",
					"all, or you are moving it somewhere else.",
				}},
				{Label: "what it does", Lines: []string{
					"Removes the kit's own files \u2014 it knows them by",
					"the `writrun-` prefix they carry \u2014 and shows",
					"both what goes and what stays before it asks.",
				}},
				{Label: "what it never", Lines: []string{
					"Touches your tasks, your specs, your reports or",
					"your docs. Those are your record, not the",
					"tool's.",
				}},
				{Label: "what comes next", Lines: []string{
					"The repository keeps working; it simply has no",
					"kit in it, and `init` can put one back.",
				}},
			},
		},
		Need: command.NeedAdopted,
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
	r, err := plan(d.Files, ctx.Root, hookAt)
	if err != nil {
		return err
	}
	if err := ctx.AskPlan(r.plan()); err != nil {
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
	disk vfs.FS
	root string

	dirs  []string // kit-owned directories to delete
	files []string // kit-owned files to delete
	gone  []string // named in the kit's inventory, already not there

	hookAt    string
	hookState hook.State

	// agents is what AGENTS.md becomes: nil with agentsWhole set means
	// the file was nothing but the kit's skeleton.
	agents      []byte
	agentsWhole bool
	agentsKept  bool // no WritRun section found; the file is the project's alone
}

func plan(files vfs.FS, root, hookAt string) (*removal, error) {
	r := &removal{disk: files, root: root, hookAt: hookAt}

	for _, dir := range kitpaths.RemoveDirs {
		if _, err := files.Stat(filepath.Join(root, dir)); err == nil {
			r.dirs = append(r.dirs, dir)
		} else {
			r.gone = append(r.gone, dir)
		}
	}
	for _, rel := range kitpaths.RemoveFiles() {
		if _, err := files.Stat(filepath.Join(root, filepath.FromSlash(rel))); err == nil {
			r.files = append(r.files, rel)
		} else {
			r.gone = append(r.gone, rel)
		}
	}
	// The kit prefixes its own files in the two `.github` folders, and
	// uninstall has no template to read: the namespace is what tells
	// them from the project's, so a tag that added a workflow is
	// removed without a Go change
	// (docs/technical/engineering/coupling.md).
	namespaced, err := kitFilesIn(files, root)
	if err != nil {
		return nil, err
	}
	r.files = append(r.files, namespaced...)

	state, err := hook.Inspect(files, hookAt)
	if err != nil {
		return nil, err
	}
	r.hookState = state

	agents, err := files.ReadFile(filepath.Join(root, "AGENTS.md"))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		r.agentsKept = true
		r.gone = append(r.gone, "AGENTS.md")
	case err != nil:
		return nil, fmt.Errorf("reading AGENTS.md: %w", err)
	default:
		out, only, sectionErr := pointer.Remove(agents)
		switch {
		case sectionErr != nil:
			// No section to cut: whatever this file is, it is the
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

// plan is both sets as the rows of one screen, because the
// confirmation is about both: what goes, and what a person is being
// promised will stay.
//
// The screen's row names the act and the path; what the path is moves
// into the pane under it, where a reader can read it whole. The printed
// form keeps every clause on the row, because a transcript has no pane
// (docs/product/screens/adoption/uninstall.excalidraw, spec-0040).
func (r *removal) plan() command.Plan {
	p := command.Plan{
		Verb:     "remove",
		Question: "Remove the WritRun kit from this repository?",
	}
	var printed []string
	line := func(text string) {
		p.Rows = append(p.Rows, command.PlanRow{Text: text})
		printed = append(printed, text)
	}
	// shown is the row on the screen, written is the row in a
	// transcript, and detail is what the pane says about it.
	row := func(shown, written, detail string) {
		p.Rows = append(p.Rows, command.PlanRow{Text: shown, Detail: detail, Selects: true})
		printed = append(printed, written)
	}

	line("writrun uninstall — the plan; nothing is removed before the confirmation:")
	line("")
	for _, dir := range r.dirs {
		text := fmt.Sprintf("  remove       %s/ — the kit, whole", dir)
		row(text, text,
			dir+"/ — the kit's own tree, whole. Nothing this project wrote lives "+
				"under it: every file there arrived with a tag and is replaced by one.")
	}
	for _, rel := range r.files {
		text := fmt.Sprintf("  remove       %s", rel)
		row(text, text, fileDetail(rel))
	}
	switch {
	case r.agentsWhole:
		text := "  remove       AGENTS.md — nothing in it but the kit's own section"
		row(text, text,
			"AGENTS.md — the file holds the kit's section and nothing else, so "+
				"removing the section removes the file. A line of your own in it "+
				"would have made this an edit instead.")
	case r.agents != nil:
		row("  edit         AGENTS.md — WritRun's section only",
			"  edit         AGENTS.md — WritRun's section only; every byte outside it stays",
			"AGENTS.md — WritRun's section only; every byte outside it stays, "+
				"because the file is the project's and the section is the kit's.")
	case r.agentsKept:
		row("  kept         AGENTS.md — no WritRun section found",
			"  kept         AGENTS.md — no WritRun section found; left as the project wrote it",
			"AGENTS.md — no WritRun section is in it, so there is nothing of the "+
				"kit's to cut. The file is left as the project wrote it.")
	}
	switch r.hookState {
	case hook.Ours:
		row(fmt.Sprintf("  remove       %s", r.hookDisplay()),
			fmt.Sprintf("  remove       %s — the commit-msg hook the adoption installed", r.hookDisplay()),
			r.hookDisplay()+" — the commit-msg hook `init` installed, byte for "+
				"byte. A hook you wrote is never removed: it is recognised, and kept.")
	case hook.Foreign:
		row(fmt.Sprintf("  kept         %s", r.hookDisplay()),
			fmt.Sprintf("  kept         %s — the installed hook is not the one init writes; it is another project's to remove", r.hookDisplay()),
			r.hookDisplay()+" — the installed hook is not the one `init` writes, "+
				"so it is another project's to remove and this leaves it alone.")
	case hook.Absent:
		row(fmt.Sprintf("  kept         %s", r.hookDisplay()),
			fmt.Sprintf("  kept         %s — no commit-msg hook is installed", r.hookDisplay()),
			r.hookDisplay()+" — no commit-msg hook is installed, so there is "+
				"nothing here to remove.")
	}
	for _, rel := range r.gone {
		text := fmt.Sprintf("  already gone %s", rel)
		row(text, text,
			rel+" — the kit's inventory names it and this repository does not "+
				"have it. Nothing is done about a file that is already gone.")
	}
	line("")
	for _, rel := range kitpaths.Keep {
		text := fmt.Sprintf("  stays        %s/ — the project's, not the kit's", rel)
		row(text, text,
			rel+"/ — the project's, not the kit's. Nothing under it is read, "+
				"moved or removed: what you wrote through the methodology "+
				"outlives the tooling.")
	}
	line("")
	p.Printed = printed
	return p
}

// fileDetail is what one removed file is. A file in the folders the kit
// shares with the project is known by the prefix it carries, and that
// is the fact worth saying: it is why a workflow of your own is never
// in this list (docs/product/screens/adoption/uninstall.excalidraw).
func fileDetail(rel string) string {
	for _, dir := range kitpaths.NamespacedDirs() {
		if strings.HasPrefix(rel, dir+"/") {
			return rel + " — the kit's own, known by the `writrun-` prefix it " +
				"carries. A workflow this project wrote carries no such prefix and " +
				"is never in this set; one a later tag added always is."
		}
	}
	return rel + " — the kit's own, named in its inventory. Removing it leaves " +
		"nothing of the kit behind at this address."
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
		if err := r.disk.RemoveAll(filepath.Join(r.root, dir)); err != nil {
			return fmt.Errorf("removing %s: %w", dir, err)
		}
	}
	for _, rel := range r.files {
		if err := r.disk.Remove(filepath.Join(r.root, filepath.FromSlash(rel))); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("removing %s: %w", rel, err)
		}
	}
	agentsPath := filepath.Join(r.root, "AGENTS.md")
	switch {
	case r.agentsWhole:
		if err := r.disk.Remove(agentsPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("removing AGENTS.md: %w", err)
		}
	case r.agents != nil:
		if err := r.disk.WriteFile(agentsPath, r.agents, 0o644); err != nil {
			return fmt.Errorf("editing AGENTS.md: %w", err)
		}
	}
	if r.hookState == hook.Ours {
		if err := r.disk.Remove(r.hookAt); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("removing the commit-msg hook: %w", err)
		}
	}
	return nil
}

// kitFilesIn lists the kit's namespaced files present in the folders it
// shares with the project, sorted, as slash-separated paths.
func kitFilesIn(files vfs.FS, root string) ([]string, error) {
	var out []string
	for _, dir := range kitpaths.NamespacedDirs() {
		err := files.WalkDir(filepath.Join(root, filepath.FromSlash(dir)), func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					return nil
				}
				return err
			}
			if entry.IsDir() {
				return nil
			}
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			if slash := filepath.ToSlash(rel); kitpaths.Namespaced(slash) {
				out = append(out, slash)
			}
			return nil
		})
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}
	sort.Strings(out)
	return out, nil
}
