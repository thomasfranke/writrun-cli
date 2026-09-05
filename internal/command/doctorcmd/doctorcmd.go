// Package doctorcmd is `writrun doctor`: every assumption the
// methodology makes of an adopted repository, reported by the stage
// that makes it and judged only up to the stage the repository
// declares. It reports; it never repairs — there is no `--fix`, and no
// run of it writes anything (docs/product/adoption/doctor.md,
// spec-0004).
package doctorcmd

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/kit"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// The adopted repository's own scripts, which this command runs and
// never reimplements: the settings reader that says which stage is
// declared, and the two checks whose exit code is the whole verdict on
// the queue's front matter and on the settings file's shape.
const (
	settingsReader    = ".writrun/scripts/stage-2-pull-requests/read_setting.sh"
	frontMatterScript = ".writrun/skills/writrun-check-front-matter/check_front_matter.sh"
	settingsScript    = ".writrun/scripts/stage-2-pull-requests/check_settings.sh"
)

// Deps is the wiring doctor needs beyond the frame's Ctx: the four
// ports it reads the world through, and nothing else.
type Deps struct {
	// Scripts runs one of the adopted repository's own scripts.
	Scripts kit.Runner
	// Gh asks the forge, and is reached only from stage 2.
	Gh func(args ...string) (string, error)
	// Files is the filesystem; doctor only ever reads through it.
	Files vfs.FS
	// LookPath probes the PATH for the wrapped scripts' requirements.
	LookPath func(name string) (string, error)
}

// New returns the doctor command wired with its dependencies.
func New(d Deps) command.Command {
	return command.Command{
		Name:    "doctor",
		Summary: "report whether the repository still satisfies what the methodology assumes",
		Need:    command.NeedAdopted,
		Run: func(ctx *command.Ctx, args []string) error {
			return run(ctx, d, args)
		},
	}
}

// verdict is doctor's own exit. Every finding is on stdout already, so
// what travels up carries a code and nothing to restate: the frame
// turns an error carrying an exit code into that code and prints
// nothing over it (internal/command/run.go).
type verdict int

func (v verdict) Error() string { return fmt.Sprintf("exit status %d", int(v)) }
func (v verdict) ExitCode() int { return int(v) }

func run(ctx *command.Ctx, d Deps, args []string) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q — doctor examines the whole repository and takes none", fs.Arg(0))
	}

	stage, unreadable := declaredStage(ctx.Root, d)
	found := examine(ctx.Root, stage, d, unreadable)
	render(ctx.Stdout, stage, found)
	if breaking(found) > 0 {
		return verdict(1)
	}
	return nil
}

// declaredStage asks the repository's own settings reader which stage
// is declared. The reader documents its own defaults and never fails on
// an absent file, so an error here means the kit itself cannot answer:
// the stage falls back to 1 — the least machinery, so nothing is judged
// against what the project may not have enabled — and the fault is
// reported as a stage-1 finding rather than swallowed.
func declaredStage(root string, d Deps) (stage int, unreadable []finding) {
	var said bytes.Buffer
	if err := d.Scripts(root, &said, &said, settingsReader, "stage"); err != nil {
		return 1, []finding{{
			stage:  1,
			level:  breaks,
			text:   settingsReader + " — the declared stage could not be read; stage 1 was assumed, so nothing above it was examined",
			detail: said.String(),
		}}
	}
	n, err := strconv.Atoi(strings.TrimSpace(said.String()))
	if err != nil || n < 1 || n > 3 {
		return 1, []finding{{
			stage: 1,
			level: breaks,
			text:  fmt.Sprintf(".writrun/settings.json — the declared stage reads as %q; 1, 2 or 3 is expected, and stage 1 was assumed", strings.TrimSpace(said.String())),
		}}
	}
	return n, nil
}

// examine runs the groups from stage 0 up to the declared one — a
// project is never judged against machinery it did not enable
// (product/adoption/doctor.md).
func examine(root string, stage int, d Deps, found []finding) []finding {
	found = append(found, stage0(d)...)
	if stage >= 1 {
		found = append(found, stage1(root, d)...)
	}
	if stage >= 2 {
		forge, reachable := stage2(d)
		found = append(found, forge...)
		if stage >= 3 {
			found = append(found, stage3(d, reachable)...)
		}
	}
	return found
}
