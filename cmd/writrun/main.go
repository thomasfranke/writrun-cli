// Command writrun is the porcelain for WritRun: it packages the
// methodology's own scripts and files into human-shaped commands.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"time"

	"github.com/thomasfranke/writrun-cli/internal/agentx"
	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/command/amendcmd"
	"github.com/thomasfranke/writrun-cli/internal/command/authorcmd"
	"github.com/thomasfranke/writrun-cli/internal/command/doctorcmd"
	"github.com/thomasfranke/writrun-cli/internal/command/finishcmd"
	"github.com/thomasfranke/writrun-cli/internal/command/initcmd"
	"github.com/thomasfranke/writrun-cli/internal/command/listcmd"
	"github.com/thomasfranke/writrun-cli/internal/command/reportcmd"
	"github.com/thomasfranke/writrun-cli/internal/command/statuscmd"
	"github.com/thomasfranke/writrun-cli/internal/command/takecmd"
	"github.com/thomasfranke/writrun-cli/internal/command/uninstallcmd"
	"github.com/thomasfranke/writrun-cli/internal/command/updatecmd"
	"github.com/thomasfranke/writrun-cli/internal/command/workcmd"
	"github.com/thomasfranke/writrun-cli/internal/forge"
	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/kit"
	"github.com/thomasfranke/writrun-cli/internal/kitfetch"
	"github.com/thomasfranke/writrun-cli/internal/screen"
	"github.com/thomasfranke/writrun-cli/internal/term"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
	"github.com/thomasfranke/writrun-cli/internal/wrepo"
)

// version is stamped from the tag at release time; without that stamp
// buildVersion falls back to the module version the toolchain recorded,
// so `go install ...@latest` still names its release and a source build
// names its commit.
var version = "dev"

func buildVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return version
}

// writrunTag is the WritRun tag this release pins.
const writrunTag = "v0.0.07"

func main() {
	disk := vfs.OS{}
	os.Exit(command.Run(command.Frame{
		Version:    buildVersion(),
		WritRunTag: writrunTag,
		Commands:   commands(),
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
		Terminal:   terminal(),
		FindRepo:   func(dir string) (string, bool, error) { return wrepo.Find(disk, dir) },
		Getenv:     os.Getenv,
		Getwd:      os.Getwd,
		Screen:     openScreen,
	}, os.Args[1:]))
}

// commands is the production command table. WRITRUN_SOURCE is the
// suite's seam: it points the kit fetch at a local WritRun clone;
// empty means the canonical repository.
func commands() []command.Command {
	gh := forge.Client{}
	disk := vfs.OS{}
	source := os.Getenv("WRITRUN_SOURCE")
	// Named for what it is, not for its package: `kit` is the script
	// runner's package name, and a local of that name shadows it.
	fetcher := kitfetch.Clone{Files: disk, Git: gitx.Run}
	return []command.Command{
		initcmd.New(initcmd.Deps{
			Tag:      writrunTag,
			Source:   source,
			Git:      gitx.Run,
			Gh:       gh.Run,
			LookPath: exec.LookPath,
			Files:    disk,
			Kit:      fetcher,
		}),
		updatecmd.New(updatecmd.Deps{
			Tag:    writrunTag,
			Source: source,
			Git:    gitx.Run,
			Files:  disk,
			Kit:    fetcher,
		}),
		doctorcmd.New(doctorcmd.Deps{
			Scripts:  kit.Run,
			Gh:       gh.Run,
			Files:    disk,
			LookPath: exec.LookPath,
		}),
		uninstallcmd.New(uninstallcmd.Deps{Git: gitx.Run, Files: disk}),
		listcmd.New(listcmd.Deps{Script: kit.Run}),
		workcmd.New(workcmd.Deps{
			Git:     gitx.Run,
			Scripts: kit.Run,
			Launch:  agentx.Run,
		}),
		statuscmd.New(statuscmd.Deps{
			Tag:     writrunTag,
			Git:     gitx.Run,
			Files:   disk,
			Scripts: kit.Run,
		}),
		takecmd.New(takecmd.Deps{Scripts: kit.Run}),
		authorcmd.New(authorcmd.Deps{
			Scripts: kit.Run,
			Files:   disk,
			Git:     gitx.Run,
			Gh:      gh.Run,
		}),
		finishcmd.New(finishcmd.Deps{
			Scripts: kit.Run,
			Files:   disk,
			Git:     gitx.Run,
			Gh:      gh.Run,
			Now:     time.Now,
			Die:     finishcmd.Die,
		}),
		amendcmd.New(amendcmd.Deps{
			Scripts: kit.Run,
			Files:   disk,
			Git:     gitx.Run,
			Gh:      gh.Run,
			Getenv:  os.Getenv,
		}),
		reportcmd.New(reportcmd.Deps{Scripts: kit.Run, Files: disk}),
	}
}

// terminal is the production terminal. WRITRUN_TTY_IN is the suite's
// pseudo-terminal: a file of key bytes driving the forms, so the
// guarded flows — a decline, an arrow-selected stage — are exercisable
// through the compiled binary, where no real terminal exists. The
// bytes are handed over as a plain reader, never the *os.File: a
// regular file cannot join Linux's epoll interest list, and the
// non-file reader is what selects the form library's fallback input
// path — the same one the headless unit tests exercise.
func terminal() term.Terminal {
	t := term.New()
	if p := os.Getenv("WRITRUN_TTY_IN"); p != "" {
		if keys, err := os.ReadFile(p); err == nil {
			t.In = bytes.NewReader(keys)
		}
	}
	return t
}

// openScreen is the frame's screen port in production: the entry screen
// built from the command table the binary already carries, with the
// queue one keystroke in — read by the selection skill's own lister,
// the same authority `writrun list` wraps, so the two cannot become two
// answers about one queue.
func openScreen(ctx *command.Ctx) (string, string, error) {
	action, err := screen.Open(entryScreen(ctx), func() (string, error) {
		return listing(ctx)
	}, os.Stdin, os.Stdout)
	if err != nil {
		return "", "", err
	}
	return action.Command, action.Arg, nil
}

// listing is the lister's output, captured rather than streamed: it is
// the screen's content, not a message to print behind it.
//
// Exit 1 is not a failure. The lister's last statement is
// `[ -n "$available" ]`, so its status answers "is anything available",
// and 1 is the answer "no" — the reading listcmd and takecmd both make
// of it. Treating it as a failure closed the screen on the one state
// its own footer was written for (internal/screen/model.go).
func listing(ctx *command.Ctx) (string, error) {
	var out, errb bytes.Buffer
	if err := kit.Run(ctx.Root, &out, &errb, nil, listerScript); err != nil && exitCode(err) != 1 {
		if msg := strings.TrimSpace(errb.String()); msg != "" {
			return "", fmt.Errorf("%s", msg)
		}
		return "", err
	}
	return out.String(), nil
}

// entryGroups is what the screen lists, in the order it lists them: the
// grouping docs/product/README.md gives the same commands, because one
// set grouped two ways is two answers about one tool. A command absent
// from here is a command the screen does not offer, and `init` is the
// only one — it refuses inside an adoption, which is the only place
// this screen opens.
var entryGroups = []struct {
	name  string
	names []string
}{
	{"tasks", []string{"list", "take", "work", "status", "finish"}},
	{"authoring", []string{"author", "amend"}},
	{"reports", []string{"report"}},
	{"adoption", []string{"doctor", "update", "uninstall"}},
}

// entryScreen fills the screen from the command table and the facts the
// context already holds. Every summary is the table's own — nothing
// here writes a second description of a command.
func entryScreen(ctx *command.Ctx) screen.Entry {
	summaries := map[string]string{}
	for _, c := range commands() {
		summaries[c.Name] = c.Summary
	}
	e := screen.Entry{Header: header(ctx)}
	for _, g := range entryGroups {
		group := screen.Group{Name: g.name}
		for _, n := range g.names {
			if s, there := summaries[n]; there {
				group.Rows = append(group.Rows, screen.Command{Name: n, Summary: s})
			}
		}
		if len(group.Rows) > 0 {
			e.Groups = append(e.Groups, group)
		}
	}
	e.Stage = stageLines(ctx)
	return e
}

// exitCode reads the script's own verdict off the error the runner
// returned; -1 says the runner failed before the script spoke, which is
// not a verdict to map.
func exitCode(err error) int {
	var verdict interface{ ExitCode() int }
	if errors.As(err, &verdict) && verdict.ExitCode() > 0 {
		return verdict.ExitCode()
	}
	return -1
}

// listerScript is the selection skill's lister, named here as listcmd
// names it — one path, two callers, and neither reimplements what it
// decides.
const listerScript = kit.ListTasks

// header is the screen's first line. Every fact in it is a cheap read:
// no check runs to open this screen, so a version, a pinned tag and the
// current branch are all it may say.
func header(ctx *command.Ctx) string {
	line := command.Product + " " + buildVersion() + " · pins WritRun " + writrunTag
	if b, err := gitx.Run(ctx.Root, "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		if name := strings.TrimSpace(b); name != "" {
			line += " · branch " + name
		}
	}
	return line
}

// stageLines are the declared stage and the conduct flags that decide
// what the commands below will do, read through the kit's own reader.
// Where a value cannot be read the line is left out: a screen that
// cannot say the stage says nothing about it rather than guessing.
func stageLines(ctx *command.Ctx) []string {
	read := func(key string) string {
		var out bytes.Buffer
		if err := kit.Run(ctx.Root, &out, io.Discard, nil, kit.ReadSetting, key); err != nil {
			return ""
		}
		return strings.TrimSpace(out.String())
	}
	stage := read("stage")
	if stage == "" {
		return nil
	}
	lines := []string{"STAGE " + stage}
	// `ask` and `auto` are what the flag means to the reader: a false
	// conduct flag is a command that composes and waits for a yes.
	word := func(key string) string {
		switch read(key) {
		case "true":
			return "auto"
		case "false":
			return "ask"
		}
		return "?"
	}
	lines = append(lines, "  commit "+word("stage_2.auto_commit")+
		" · push "+word("stage_2.auto_push")+
		" · pull request "+word("stage_2.auto_pr"))
	return lines
}
