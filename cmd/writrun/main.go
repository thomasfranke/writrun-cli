// Command writrun is the porcelain for WritRun: it packages the
// methodology's own scripts and files into human-shaped commands.
package main

import (
	"bytes"
	"os"
	"os/exec"
	"runtime/debug"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/command/initcmd"
	"github.com/thomasfranke/writrun-cli/internal/command/listcmd"
	"github.com/thomasfranke/writrun-cli/internal/command/takecmd"
	"github.com/thomasfranke/writrun-cli/internal/command/uninstallcmd"
	"github.com/thomasfranke/writrun-cli/internal/command/updatecmd"
	"github.com/thomasfranke/writrun-cli/internal/forge"
	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/kit"
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
const writrunTag = "v0.0.03"

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
	}, os.Args[1:]))
}

// commands is the production command table. WRITRUN_SOURCE is the
// suite's seam: it points the kit fetch at a local WritRun clone;
// empty means the canonical repository.
func commands() []command.Command {
	gh := forge.Client{}
	disk := vfs.OS{}
	source := os.Getenv("WRITRUN_SOURCE")
	return []command.Command{
		initcmd.New(initcmd.Deps{
			Tag:      writrunTag,
			Source:   source,
			Git:      gitx.Run,
			Gh:       gh.Run,
			LookPath: exec.LookPath,
			Files:    disk,
		}),
		updatecmd.New(updatecmd.Deps{
			Tag:    writrunTag,
			Source: source,
			Git:    gitx.Run,
			Files:  disk,
		}),
		uninstallcmd.New(uninstallcmd.Deps{Git: gitx.Run, Files: disk}),
		listcmd.New(listcmd.Deps{Script: kit.Run}),
		takecmd.New(takecmd.Deps{Scripts: kit.Run}),
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
