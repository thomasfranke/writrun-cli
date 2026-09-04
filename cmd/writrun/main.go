// Command writrun is the porcelain for WritRun: it packages the
// methodology's own scripts and files into human-shaped commands.
package main

import (
	"os"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/term"
	"github.com/thomasfranke/writrun-cli/internal/wrepo"
)

// version is stamped from the tag at release time; a source build says so.
var version = "dev"

// writrunTag is the WritRun tag this release pins.
const writrunTag = "v0.0.03"

func main() {
	os.Exit(command.Run(command.Frame{
		Version:    version,
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
