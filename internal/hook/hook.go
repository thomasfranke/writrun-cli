// Package hook owns the commit-message hook the adoption installs:
// its text, where git keeps it, and whether the one on disk is still
// the one this kit wrote. init writes it, uninstall removes it, and
// only a hook recognised as the kit's own is ever removed.
package hook

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// Script is the commit-msg hook init installs. The vocabulary is
// read from the kit's check_observance.sh at commit time, not baked in
// here, so editing those two lines is the whole customization — the
// hook and the door can never disagree (conventions/commits.md).
const Script = `#!/usr/bin/env bash
# commit-msg — installed by writrun init. Validates the Conventional
# subject against the TYPES/SCOPES lines of the kit's
# check_observance.sh. It validates; it never writes a message
# (docs/product/adoption/init.md).

subject=$(awk '!/^#/ && NF { print; exit }' "$1")

# Git's own generated shapes pass untouched — the convention governs
# what a person or an agent writes, not what a merge writes.
case "$subject" in
  "Merge "*|"Revert "*|"fixup! "*|"squash! "*) exit 0 ;;
esac

top=$(git rev-parse --show-toplevel 2>/dev/null) || exit 0
observance="$top/.writrun/scripts/stage-2-pull-requests/check_observance.sh"
# A hook outliving the kit blocks nothing — validating against a
# vocabulary that is gone would refuse every commit for a fault that is
# the hook's own.
[ -f "$observance" ] || exit 0

TYPES=$(sed -n 's/^TYPES="\(.*\)"$/\1/p' "$observance" | head -n1)
SCOPES=$(sed -n 's/^SCOPES="\(.*\)"$/\1/p' "$observance" | head -n1)
[ -n "$TYPES" ] || exit 0

type=$(printf '%s' "$subject" | sed -nE 's/^([a-z]+)(\([a-z0-9-]+\))?!?: .+$/\1/p')
if [ -z "$type" ]; then
  echo "commit-msg: '$subject' is not a Conventional subject — type(scope): imperative summary" >&2
  exit 1
fi
case " $TYPES " in *" $type "*) ;; *)
  echo "commit-msg: the type '$type' is outside the vocabulary ($TYPES)" >&2
  exit 1 ;;
esac
scope=$(printf '%s' "$subject" | sed -nE 's/^[a-z]+\(([a-z0-9-]+)\)!?: .+$/\1/p')
if [ -n "$scope" ] && [ -n "$SCOPES" ]; then
  case " $SCOPES " in *" $scope "*) ;; *)
    echo "commit-msg: the scope '$scope' is outside the vocabulary ($SCOPES)" >&2
    exit 1 ;;
  esac
fi
exit 0
`

// Path resolves where the commit-msg hook lives, through git so a
// worktree's redirected hooks directory is the one answered.
func Path(root string, git gitx.Runner) (string, error) {
	out, err := git(root, "rev-parse", "--git-path", "hooks/commit-msg")
	if err != nil {
		return "", fmt.Errorf("resolving the hooks directory: %w", err)
	}
	p := strings.TrimSpace(out)
	if !filepath.IsAbs(p) {
		p = filepath.Join(root, p)
	}
	return p, nil
}

// RefuseForeign refuses where a commit-msg hook already exists:
// whatever installed it owns it, and overwriting would trade one
// project's convention for another's silently (spec-0002, edge cases).
func RefuseForeign(files vfs.FS, path string) error {
	if _, err := files.Stat(path); err == nil {
		return fmt.Errorf("a commit-msg hook is already installed at %s — writrun init refuses to overwrite it; move it aside and rerun", path)
	}
	return nil
}

// Install writes the hook executable. The directory may not exist yet
// in a repository that never had a hook installed.
func Install(files vfs.FS, path string) error {
	if err := files.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("installing the commit-msg hook: %w", err)
	}
	if err := files.WriteFile(path, []byte(Script), 0o755); err != nil {
		return fmt.Errorf("installing the commit-msg hook: %w", err)
	}
	return nil
}

// State is what uninstall found where the hook should be.
type State int

const (
	// Absent — nothing is installed there.
	Absent State = iota
	// Ours — byte-for-byte the hook this kit writes; removable.
	Ours
	// Foreign — something else's, whatever wrote it; left standing and
	// said so (spec-0005).
	Foreign
)

// Inspect reports whether the hook on disk is the one Install writes.
// Anything that is not byte-identical is Foreign — a hook a project
// edited is a hook a project owns.
func Inspect(files vfs.FS, path string) (State, error) {
	content, err := files.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Absent, nil
	}
	if err != nil {
		return Absent, fmt.Errorf("reading the commit-msg hook: %w", err)
	}
	if bytes.Equal(content, []byte(Script)) {
		return Ours, nil
	}
	return Foreign, nil
}
