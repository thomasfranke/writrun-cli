// Package doctorcmd is `writrun doctor`: every requirement the
// methodology makes of an adopted repository, named and marked by the
// stage that makes it, judged only up to the stage the repository
// declares, and the rung above that declaration previewed. It reports;
// it never repairs — there is no `--fix`, and no run of it writes
// anything (docs/product/adoption/doctor.md, spec-0004, spec-0036).
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
	"github.com/thomasfranke/writrun-cli/internal/palette"
	"github.com/thomasfranke/writrun-cli/internal/screen"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// The adopted repository's own scripts, which this command runs and
// never reimplements: the settings reader that says which stage is
// declared, and the two checks whose exit code is the whole verdict on
// the queue's front matter and on the settings file's shape.
const (
	settingsReader    = kit.ReadSetting
	frontMatterScript = kit.CheckFrontMatter
	settingsScript    = kit.CheckSettings
)

// Deps is the wiring doctor needs beyond the frame's Ctx: the four
// ports it reads the world through, and the one line it cannot compose
// itself.
type Deps struct {
	// Scripts runs one of the adopted repository's own scripts.
	Scripts kit.Runner
	// Gh asks the forge, and is reached from stage 2, or from the preview
	// of it a stage-1 repository gets.
	Gh func(args ...string) (string, error)
	// Files is the filesystem; doctor only ever reads through it.
	Files vfs.FS
	// LookPath probes the PATH for the wrapped scripts' requirements.
	LookPath func(name string) (string, error)
	// Header is the screen's first line: the product, its version, the
	// tag it pins, the branch. The caller composes it, because every fact
	// in it is one the caller already holds and none is doctor's
	// (internal/screen, Entry.Header).
	Header func(root string) string
}

// New returns the doctor command wired with its dependencies.
func New(d Deps) command.Command {
	return command.Command{
		Name:    "doctor",
		Summary: "check this repository still satisfies WritRun",
		About: command.About{
			Sentence: []string{
				"check this repository still satisfies what WritRun",
				"assumes",
			},
			Parts: []command.AboutPart{
				{Label: "why you would", Lines: []string{
					"Something stopped working, or you want to know",
					"whether you are ready to move up a stage.",
				}},
				{Label: "what it does", Lines: []string{
					"Reads every requirement the stages make \u2014",
					"programs on your PATH, documents in place,",
					"GitHub settings \u2014 and marks each one met or",
					"not.",
				}},
				{Label: "what it never", Lines: []string{
					"Repairs. It names the file or the setting and",
					"what is expected of it; the change is yours to",
					"make.",
				}},
				{Label: "what comes next", Lines: []string{
					"A stage above yours is previewed, so you can",
					"see what declaring it would ask of this",
					"repository.",
				}},
			},
		},
		Need:  command.NeedAdopted,
		Daily: true,
		Run: func(ctx *command.Ctx, args []string) error {
			return run(ctx, d, args)
		},
	}
}

// verdict is doctor's own exit. Every fault is on stdout already, so
// what travels up carries a code and nothing to restate: the frame
// turns an error carrying an exit code into that code and prints
// nothing over it (internal/command/run.go).
type verdict int

func (v verdict) Error() string { return fmt.Sprintf("exit status %d", int(v)) }
func (v verdict) ExitCode() int { return int(v) }

func run(ctx *command.Ctx, d Deps, args []string) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	atStage := fs.Int("at", 0, "examine up to this stage instead of the declared one")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q — doctor examines the whole repository and takes none", fs.Arg(0))
	}
	if *atStage != 0 && (*atStage < 1 || *atStage > 3) {
		return fmt.Errorf("--at %d is not a stage — 1, 2 or 3", *atStage)
	}

	// A terminal gets the screen the drawing gives; anything else gets
	// the report, because a script reading `writrun doctor` must keep
	// reading it (spec-0037).
	if ctx.Terminal.InteractiveIn() && ctx.Terminal.InteractiveOut() {
		return browse(ctx, d, *atStage)
	}

	declared, examined, preview, found := examineAll(ctx.Root, d, *atStage)
	render(ctx.Stdout, palette.New(ctx.Color), declared, examined, preview, found)
	if breaking(at(found, upTo(declared)...)) > 0 {
		return verdict(1)
	}
	return nil
}

// examineAll is one whole run: which stage is declared, how far the
// examination reached, which rung it previewed, and every requirement
// all of that produced.
//
// **The declaration decides the verdict; `--at` decides only how far the
// examination reaches.** A stage nobody declared cannot fail a build, or
// asking what a stage would require would be a way to break one
// (product/adoption/doctor.md).
func examineAll(root string, d Deps, atStage int) (declared, examined, preview int, found []requirement) {
	declared, unreadable := declaredStage(root, d)
	examined = declared
	if atStage != 0 {
		examined = atStage
	}
	// The rung above what was examined, previewed through the same code
	// path `--at` uses: no check is written twice for it (spec-0036).
	if examined < 3 {
		preview = examined + 1
	}
	found = append(unreadable, examine(root, examined, d)...)
	if preview > 0 {
		found = append(found, at(examine(root, preview, d), preview)...)
	}
	return declared, examined, preview, found
}

// Preview is what `config` shows before it raises the stage: one stage's
// requirements, examined and marked as the report marks them, and the
// three counts the question quotes. The checks are doctor's own, so
// `config` keeps no copy of one (spec-0041).
func Preview(d Deps, colour bool) func(root string, stage int, w io.Writer) (met, unmet, unread int, err error) {
	return func(root string, stage int, w io.Writer) (int, int, int, error) {
		if stage < 0 || stage > 3 {
			return 0, 0, 0, fmt.Errorf("%d is not a stage — 1, 2 or 3", stage)
		}
		group := at(examine(root, stage, d), stage)
		p := palette.New(colour)
		for _, r := range group {
			fmt.Fprintf(w, "         %s  %s\n", paint(p, r.mark), r.text())
		}
		u := counted(group, unread)
		m := counted(group, met)
		return m, len(group) - m - u, u, nil
	}
}

// declaredStage asks the repository's own settings reader which stage
// is declared. The reader documents its own defaults and never fails on
// an absent file, so an error here means the kit itself cannot answer:
// the stage falls back to 1 — the least machinery, so nothing is judged
// against what the project may not have enabled — and the fault is
// reported as a stage-1 requirement rather than swallowed.
func declaredStage(root string, d Deps) (stage int, unreadable []requirement) {
	var said bytes.Buffer
	if err := d.Scripts(root, &said, &said, nil, settingsReader, "stage"); err != nil {
		return 1, []requirement{{
			stage: 1, name: kit.Settings, mark: breaks,
			note:   "the declared stage could not be read; stage 1 was assumed, so nothing above it was examined",
			detail: said.String(),
		}}
	}
	n, err := strconv.Atoi(strings.TrimSpace(said.String()))
	if err != nil || n < 1 || n > 3 {
		return 1, []requirement{{
			stage: 1, name: kit.Settings, mark: breaks,
			note: fmt.Sprintf("the declared stage reads as %q; 1, 2 or 3 is expected, and stage 1 was assumed", strings.TrimSpace(said.String())),
		}}
	}
	return n, nil
}

// examine runs the groups from stage 0 up to the one named — a project
// is never judged against machinery it did not enable
// (product/adoption/doctor.md).
func examine(root string, stage int, d Deps) []requirement {
	found := stage0(d)
	if stage >= 1 {
		found = append(found, stage1(root, d)...)
	}
	if stage >= 2 {
		forge, reachable := stage2(root, d)
		found = append(found, forge...)
		if stage >= 3 {
			found = append(found, stage3(d, reachable)...)
		}
	}
	return found
}

// browse opens the report as a screen: the same rows, a cursor that
// stops on every one of them, and the document's own sentence about the
// requirement under it (spec-0037).
//
// The requirements are read again on every `r` rather than remembered,
// because whether one holds is the repository's answer and not this
// screen's.
func browse(ctx *command.Ctx, d Deps, atStage int) error {
	load := func() (screen.Requirements, error) {
		return requirementsScreen(ctx.Root, d, atStage), nil
	}
	return screen.OpenRequirements(load, ctx.Stdin, ctx.Stdout)
}

// requirementsScreen is one whole run as the screen shows it.
func requirementsScreen(root string, d Deps, atStage int) screen.Requirements {
	declared, examined, preview, found := examineAll(root, d, atStage)
	s := screen.Requirements{Line: screenLine(declared, examined, preview)}
	if d.Header != nil {
		s.Header = d.Header(root)
	}
	for st := 0; st <= 3; st++ {
		group := at(found, st)
		s.Groups = append(s.Groups, screen.RequirementGroup{
			Name: screenHeading(st, examined, preview, group),
			Rows: rowsOf(group),
		})
	}
	return s
}

// rowsOf is one group as the screen shows it: the glyph, the row's own
// text, and the sentence the document carries about it.
func rowsOf(group []requirement) []screen.Requirement {
	rows := make([]screen.Requirement, 0, len(group))
	for _, r := range group {
		rows = append(rows, screen.Requirement{
			Mark:    glyphs[r.mark],
			Text:    r.text(),
			Name:    r.name,
			Explain: explain(r),
		})
	}
	return rows
}

// screenLine is the screen's second line: the declaration, the range
// examined and the rung previewed, in the shape the drawing gives
// (docs/product/screens/adoption/doctor.excalidraw).
func screenLine(declared, examined, preview int) string {
	line := fmt.Sprintf("STAGE %d · %s — stages 0–%d examined", declared, stageNames[declared], examined)
	if preview == 0 {
		return line + ", no rung above it"
	}
	return fmt.Sprintf("%s, stage %d previewed", line, preview)
}

// screenHeading names one group on the screen: the stage, its subject,
// whether it was previewed, and the count — in the shape the drawing
// gives (docs/product/screens/adoption/doctor.excalidraw).
func screenHeading(s, examined, preview int, group []requirement) string {
	name := strings.ToUpper(fmt.Sprintf("Stage %d · %s", s, stageNames[s]))
	if preview > 0 && s == preview {
		return name + ", PREVIEWED — " + fmt.Sprintf("%d of %d met", counted(group, met), len(group))
	}
	if s > examined {
		return name + " — not previewed, one rung at a time"
	}
	return name + " — " + fmt.Sprintf("%d of %d met", counted(group, met), len(group))
}
