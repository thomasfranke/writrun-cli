// Package takecmd is `writrun take`: the methodology's own
// `take_task.sh`, run with the human's confirmation around its conduct
// gate. It reimplements no check — the script decides eligibility, the
// title's style, the branch, the body and the act, and this command
// collects the two answers the script needs and maps its four exit
// codes (docs/product/pull-requests/take.md, spec-0008).
package takecmd

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/kit"
)

// takeScript is the authority this command wraps; listScript is the
// authority on which tasks may be offered for selection. Both are the
// adopted repository's own.
const (
	takeScript = ".writrun/scripts/stage-2-pull-requests/take_task.sh"
	listScript = ".writrun/skills/writrun-select-next-task/list_tasks.sh"
)

// availableHeader opens the lister's Available section — the group the
// selection offers, and the only one it offers.
const availableHeader = "Available"

// Deps is the wiring take needs beyond the frame's Ctx.
type Deps struct {
	// Scripts runs the adopted repository's own scripts.
	Scripts kit.Runner
}

// New returns the take command wired with its dependencies.
func New(d Deps) command.Command {
	return command.Command{
		Name:    "take",
		Summary: "take a task: the branch pushed and the draft pull request opened",
		Need:    command.NeedAdopted,
		Run: func(ctx *command.Ctx, args []string) error {
			return run(ctx, d, args)
		},
	}
}

func run(ctx *command.Ctx, d Deps, args []string) error {
	id, flags, err := split(args)
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("take", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	title := fs.String("title", "", "the summary the pull-request title carries")
	slug := fs.String("slug", "", "the branch's subject words")
	if err := fs.Parse(flags); err != nil {
		return err
	}

	// Nothing here is validated: an id that resolves to no task and a
	// title the declared style refuses are both the script's refusals,
	// and a second opinion on them is a second authority (spec-0008).
	if id == "" {
		if id, err = selectTask(ctx, d); err != nil {
			return err
		}
	}
	summary, err := ctx.AskInput("The summary after the task tag:", *title, "--title")
	if err != nil {
		return err
	}

	take := []string{id, "--title", summary}
	if *slug != "" {
		take = append(take, "--slug", *slug)
	}

	err = d.Scripts(ctx.Root, ctx.Stdout, ctx.Stderr, takeScript, take...)
	if exitCode(err) != 2 {
		// 0 the act is done, 1 a refusal, 3 a git or forge failure with
		// the resume it named — every one of them already reported in
		// the script's own words, on the stream it chose.
		return passthrough(err)
	}

	// 2: composed and waiting. The composition is on stdout already, so
	// the question is all this adds, and `--yes` answers it.
	if err := ctx.AskConfirm("Push the branch and open the draft pull request?"); err != nil {
		return err
	}
	err = d.Scripts(ctx.Root, ctx.Stdout, ctx.Stderr, takeScript, append(take, "--confirm")...)
	if exitCode(err) == 2 {
		return fmt.Errorf("%s composed again under --confirm — nothing reached the forge", takeScript)
	}
	return passthrough(err)
}

// split separates the task id from the flags. Go's flag package stops
// at the first operand, so `take task-0009 --title x` would leave the
// title unparsed; the id is lifted out first and the rest parsed whole.
func split(args []string) (string, []string, error) {
	id := ""
	var flags []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--title" || a == "-title" || a == "--slug" || a == "-slug":
			flags = append(flags, a)
			if i+1 < len(args) {
				flags = append(flags, args[i+1])
				i++
			}
		case strings.HasPrefix(a, "-"):
			flags = append(flags, a)
		case id == "":
			id = a
		default:
			return "", nil, fmt.Errorf("two task ids given (%q and %q) — take takes one", id, a)
		}
	}
	return id, flags, nil
}

// selectTask offers the lister's Available group for arrow selection.
// The group is the lister's answer, not this command's: what may be
// taken is steps 2–4 of the selection algorithm, and they run there.
func selectTask(ctx *command.Ctx, d Deps) (string, error) {
	var listing bytes.Buffer
	err := d.Scripts(ctx.Root, &listing, ctx.Stderr, listScript)
	// The lister's own listing says what is available and what is held
	// back, so a stop shows it rather than summarising it. Exit 1 is
	// its "nothing is available"; anything else is a lister that could
	// not answer at all.
	if err != nil && exitCode(err) != 1 {
		fmt.Fprint(ctx.Stdout, listing.String())
		return "", fmt.Errorf("running %s: %w", listScript, err)
	}
	ids, labels := parseAvailable(listing.String())
	if len(ids) == 0 {
		fmt.Fprint(ctx.Stdout, listing.String())
		return "", errors.New("no task is available to take")
	}
	i, err := ctx.AskSelect("Take which task?", labels, "", "the task id as an argument")
	if err != nil {
		return "", err
	}
	return ids[i], nil
}

// parseAvailable reads the ids and the lines of the lister's Available
// section. The section ends at the first line that is not one of its
// entries, so the groups below it are never offered.
func parseAvailable(listing string) (ids []string, labels []string) {
	inside := false
	for _, line := range strings.Split(listing, "\n") {
		if strings.HasPrefix(line, availableHeader) {
			inside = true
			continue
		}
		if !inside {
			continue
		}
		if !strings.HasPrefix(line, "  ") || strings.TrimSpace(line) == "" {
			break
		}
		entry := strings.TrimSpace(line)
		ids = append(ids, strings.Fields(entry)[0])
		labels = append(labels, entry)
	}
	return ids, labels
}

// exitCode reads the script's own verdict off the error the runner
// returned; -1 says the runner failed before the script spoke, which
// is not a verdict to map.
func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var verdict interface{ ExitCode() int }
	if errors.As(err, &verdict) && verdict.ExitCode() > 0 {
		return verdict.ExitCode()
	}
	return -1
}

// passthrough hands the script's verdict up unedited: the frame turns
// an error carrying an exit code into that exit code, having reported
// nothing over what the script already said.
func passthrough(err error) error {
	if err == nil {
		return nil
	}
	if exitCode(err) < 0 {
		return fmt.Errorf("running %s: %w", takeScript, err)
	}
	return err
}
