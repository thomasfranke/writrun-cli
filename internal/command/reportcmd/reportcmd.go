// Package reportcmd is `writrun report`: an observation recorded into
// `work/reports/` before the conversation it was noticed in ends. The
// file is the methodology's own generator's — the id minted in
// sequence, the status born `open`, nothing triaged — and this command
// adds the two things the generator has no flag for: the reporter's
// paragraph, and the question that precedes a write
// (docs/product/reports/report.md, docs/product/rules.md).
//
// It records and stops. Triage is a judgement, so no flag here sets a
// route; and recording is exempt from the one-kind-per-change rule, so
// nothing is branched, committed or opened on the forge.
package reportcmd

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/kit"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// generator is the authority this command wraps: the creation skill's
// own scaffolder, which owns the id, the front matter and the file's
// name. Nothing here mints, numbers or validates in its place.
const generator = ".writrun/skills/writrun-create-task-and-spec/new.sh"

// reports is where a recorded report lives — the one directory the
// generator may have named back.
const reports = "work/reports/"

// Deps is the wiring report needs beyond the frame's Ctx.
type Deps struct {
	// Scripts runs the adopted repository's own scripts.
	Scripts kit.Runner
	// Files is the filesystem, for the one edit the generator leaves to
	// its caller: its placeholder body replaced by the reporter's.
	Files vfs.FS
}

// New returns the report command wired with its dependencies.
func New(d Deps) command.Command {
	return command.Command{
		Name:    "report",
		Summary: "record an observation into work/reports/: no triage, no branch, no pull request",
		Need:    command.NeedAdopted,
		Run: func(ctx *command.Ctx, args []string) error {
			return run(ctx, d, args)
		},
	}
}

func run(ctx *command.Ctx, d Deps, args []string) error {
	given, flags, err := split(args)
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("report", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	body := fs.String("body", "", "the observation, in the reporter's own words")
	slug := fs.String("slug", "", "the filename's subject words")
	docRef := fs.String("doc-ref", "", "the doc this observation is against, as path#anchor")
	if err := fs.Parse(flags); err != nil {
		return err
	}

	// Neither answer is inspected here. A title the generator refuses
	// and a slug it refuses are its refusals, and a second opinion on
	// them would be a second authority (docs/about.md).
	title, err := ctx.AskInput("What did you notice?", given, "the title as an argument")
	if err != nil {
		return err
	}
	observation, err := ctx.AskInput("What was observed, and the evidence at hand:", *body, "--body")
	if err != nil {
		return err
	}

	// One file, and it cannot be named before it exists — the id is the
	// generator's to mint. So the question shows what is known: the
	// title, and where it is about to land (docs/product/rules.md).
	fmt.Fprintf(ctx.Stdout, "Recording into %s\n\n  %s\n\n", reports, title)
	if err := ctx.AskConfirm("Record this observation?"); err != nil {
		return err
	}

	mint := []string{"report", title}
	if *slug != "" {
		mint = append(mint, "--slug", *slug)
	}
	if *docRef != "" {
		mint = append(mint, "--doc-ref", *docRef)
	}

	// The generator's reporting is the user's, as it produced it; the
	// copy is read only for the path it named.
	var said bytes.Buffer
	if err := d.Scripts(ctx.Root, io.MultiWriter(ctx.Stdout, &said), ctx.Stderr, generator, mint...); err != nil {
		return passthrough(err)
	}
	file, ok := created(said.String())
	if !ok {
		return fmt.Errorf("%s named no file it created — nothing was recorded", generator)
	}
	return fill(d.Files, filepath.Join(ctx.Root, file), file, observation)
}

// split separates the title from the flags. Go's flag package stops at
// the first operand, so `report "A title" --body x` would leave the
// body unparsed; the title is lifted out first and the rest parsed
// whole (the shape takecmd uses for its task id).
func split(args []string) (string, []string, error) {
	title := ""
	var flags []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case valued(a):
			flags = append(flags, a)
			if i+1 < len(args) {
				flags = append(flags, args[i+1])
				i++
			}
		case strings.HasPrefix(a, "-"):
			flags = append(flags, a)
		case title == "":
			title = a
		default:
			return "", nil, fmt.Errorf("two titles given (%q and %q) — report takes one", title, a)
		}
	}
	return title, flags, nil
}

// valued reports whether the flag takes the argument after it.
func valued(a string) bool {
	switch a {
	case "--body", "-body", "--slug", "-slug", "--doc-ref", "-doc-ref":
		return true
	}
	return false
}

// created reads back the path the generator named — its line is
// "Created work/reports/report-0009-a-thing.md (report-0009)". The
// path is taken from that line rather than composed here: the id is
// minted inside the script and the slug is its own slugify's when none
// was given, so this command cannot know the name before it is told.
func created(said string) (string, bool) {
	const prefix = "Created "
	for _, line := range strings.Split(said, "\n") {
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		path := strings.TrimSpace(line[len(prefix):])
		if i := strings.Index(path, " ("); i >= 0 {
			path = path[:i]
		}
		if strings.HasPrefix(path, reports) && !strings.Contains(path, "..") {
			return path, true
		}
	}
	return "", false
}

// passthrough hands the generator's verdict up unedited: the frame
// turns an error carrying an exit code into that exit code, having
// reported nothing over what the script already said.
func passthrough(err error) error {
	var verdict interface{ ExitCode() int }
	if errors.As(err, &verdict) && verdict.ExitCode() > 0 {
		return err
	}
	return fmt.Errorf("running %s: %w", generator, err)
}
