// Command writrun is the porcelain for WritRun: it packages the
// methodology's own scripts and files into human-shaped commands.
package main

import (
	"os"
	"runtime/debug"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/term"
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
	os.Exit(command.Run(command.Frame{
		Version:    buildVersion(),
		WritRunTag: writrunTag,
		Commands:   nil,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
		Terminal:   term.New(),
		FindRepo:   wrepo.Find,
		Getenv:     os.Getenv,
		Getwd:      os.Getwd,
	}, os.Args[1:]))
}
