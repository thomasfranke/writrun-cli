// Package configcmd is `writrun config`: the adopter's own settings,
// shown and changed through the kit's own reader and checker. It knows
// no key, no allowed value and no schema — the kit declares all three,
// and this command asks (docs/product/config.md, spec-0031).
package configcmd

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/kit"
	"github.com/thomasfranke/writrun-cli/internal/palette"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// Deps is the wiring config needs beyond the frame's Ctx.
type Deps struct {
	// Scripts runs the adopted repository's own scripts: the reader
	// that answers a key, and the checker that judges a write.
	Scripts kit.Runner
	// Files is the settings file itself — read to list the keys, and
	// written when one changes.
	Files vfs.FS
}

// New returns the config command wired with its dependencies.
func New(d Deps) command.Command {
	return command.Command{
		Name:    "config",
		Summary: "the adopter's settings, and one changed",
		Need:    command.NeedAdopted,
		Run: func(ctx *command.Ctx, args []string) error {
			return run(ctx, d, args)
		},
	}
}

func run(ctx *command.Ctx, d Deps, args []string) error {
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 2 {
		return fmt.Errorf("config takes a key and a value, and was given %d arguments", fs.NArg())
	}

	path := filepath.Join(ctx.Root, filepath.FromSlash(kit.Settings))
	before, err := d.Files.ReadFile(path)
	if err != nil {
		return fmt.Errorf("%s is absent — the kit's reader documents its defaults and keeps working without it, and writing one is adoption's act, not this command's", kit.Settings)
	}

	keys := keysOf(string(before))
	if fs.NArg() == 0 {
		return show(ctx, d, keys)
	}

	key := fs.Arg(0)
	if !named(keys, key) {
		return fmt.Errorf("%q is not a key %s declares — `writrun config` lists them", key, kit.Settings)
	}
	value := fs.Arg(1)
	if value == "" {
		if value, err = ctx.AskInput("The value for "+key+":", "", "the value as a second argument"); err != nil {
			return err
		}
	}
	return set(ctx, d, path, before, key, strings.TrimSpace(value))
}

// show prints every key under the section that owns it, with the value
// the kit's reader answers. The sections and their order are the file's
// own: a list in Go would be a second answer about which keys exist.
func show(ctx *command.Ctx, d Deps, keys []key) error {
	p := palette.New(ctx.Color)
	section := ""
	for _, k := range keys {
		if k.section != section {
			section = k.section
			label := "THE STAGE"
			if section != "" {
				label = strings.ToUpper(strings.ReplaceAll(section, "_", " "))
			}
			fmt.Fprintf(ctx.Stdout, "\n%s\n", p.Heading(label))
		}
		value, err := read(d, ctx.Root, k)
		if err != nil {
			return err
		}
		fmt.Fprintf(ctx.Stdout, "  %-20s %s\n", k.name, p.Declared(value))
	}
	fmt.Fprintf(ctx.Stdout, "\n%s\n", p.Dim("writrun config <key> <value> changes one; "+
		kit.CheckSettings+" judges it"))
	return nil
}

// set writes one value, then hands the file to the kit's checker. The
// write is kept only where the checker accepts it, and a refusal puts
// the previous bytes back — the adopter's file is never left in a shape
// its own kit calls invalid.
func set(ctx *command.Ctx, d Deps, path string, before []byte, key, value string) error {
	k, _ := find(keysOf(string(before)), key)
	was, err := read(d, ctx.Root, k)
	if err != nil {
		return err
	}
	if was == value {
		fmt.Fprintf(ctx.Stdout, "%s is already %q — nothing to write.\n", key, value)
		return nil
	}

	fmt.Fprintf(ctx.Stdout, "%s\n  %s → %s\nin %s\n\n", key, was, value, kit.Settings)
	if err := ctx.AskConfirm("Write it?"); err != nil {
		return err
	}

	next, err := rewrite(string(before), k, value)
	if err != nil {
		return err
	}
	info, err := d.Files.Stat(path)
	if err != nil {
		return err
	}
	if err := d.Files.WriteFile(path, []byte(next), info.Mode().Perm()); err != nil {
		return err
	}

	var said bytes.Buffer
	if err := d.Scripts(ctx.Root, &said, &said, nil, kit.CheckSettings); err != nil {
		// The checker's own words carry the vocabulary this command does
		// not hold, so they reach the user unedited.
		if restoreErr := d.Files.WriteFile(path, before, info.Mode().Perm()); restoreErr != nil {
			return fmt.Errorf("%s refused the change and restoring %s failed: %w", kit.CheckSettings, kit.Settings, restoreErr)
		}
		fmt.Fprint(ctx.Stderr, said.String())
		return fmt.Errorf("%s refused the change — %s is exactly as it was", kit.CheckSettings, kit.Settings)
	}
	fmt.Fprintf(ctx.Stdout, "%s is %s.\n", key, value)
	return nil
}

// read answers one key through the kit's own reader, which is the only
// thing that knows how the file is shaped.
func read(d Deps, root string, k key) (string, error) {
	var out, errb bytes.Buffer
	if err := d.Scripts(root, &out, &errb, nil, kit.ReadSetting, k.dotted()); err != nil {
		if msg := strings.TrimSpace(errb.String()); msg != "" {
			return "", fmt.Errorf("reading %s: %s", k.dotted(), msg)
		}
		return "", fmt.Errorf("reading %s: %w", k.dotted(), err)
	}
	return strings.TrimSpace(out.String()), nil
}
