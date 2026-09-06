// Package kittag is the tag `.writrun/VERSION` records: where the file
// is, what it says, and the two questions asked of two tags — whether
// they name one release, and which release comes first. Every command
// that names that file names it through Path, and `update`, `status`
// and `doctor` read its contents through Read. `init` reads the file at
// Path too, for the stage-1 gap it words itself (initcmd.checkFiles).
package kittag

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// Rel is where an adopted repository records the kit's tag, relative to
// the repository root and slash-separated — the form a refresh compares
// its plan in.
const Rel = ".writrun/VERSION"

// Path is where an adopted repository records the kit's tag. Every
// reader and every writer of that file asks here, so the location has
// one definition.
func Path(root string) string {
	return filepath.Join(root, filepath.FromSlash(Rel))
}

// Read returns the recorded tag, trimmed. The error is the
// filesystem's own, unwrapped: each command words its own refusal.
//
// An unrecorded tag is not an error here — it reads as the empty
// string, because the three callers answer it differently: `update`
// refuses a refresh with no starting point, `status` names the file,
// and `doctor` reports it as an unreadable tag.
func Read(files vfs.FS, root string) (string, error) {
	raw, err := files.ReadFile(Path(root))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(raw)), nil
}

// Compare orders two tags: -1 when a precedes b, 0 when they are the
// same release, 1 when a follows b. The components are read as numbers,
// so `v0.0.03` and `v0.0.3` are one release and `v0.0.10` follows
// `v0.0.9` — which string order gets wrong.
//
// A tag neither side can parse compares equal only to itself: an
// unreadable version is not a reason to refuse a refresh, but it is no
// reason to call it a downgrade either.
func Compare(a, b string) int {
	if a == b {
		return 0
	}
	x, okA := Components(a)
	y, okB := Components(b)
	if !okA || !okB {
		return 1 // unreadable: treat as a move forward, never a downgrade
	}
	for i := 0; i < len(x) || i < len(y); i++ {
		var xi, yi int
		if i < len(x) {
			xi = x[i]
		}
		if i < len(y) {
			yi = y[i]
		}
		switch {
		case xi < yi:
			return -1
		case xi > yi:
			return 1
		}
	}
	return 0
}

// SameRelease reports whether two tags name one release. The components
// are read as numbers, so `v0.0.03` and `v0.0.3` are the same release
// and a mismatch is never announced over a spelling. Two tags neither
// side can read are the same only when they are the same text.
//
// It is `Compare(a, b) == 0` and nothing else: Compare answers 0 only
// on two tags it read as one release, and 1 — never 0 — on a tag it
// could not read, so no ordering reaches a caller asking only whether
// two tags match. The name exists so that caller never spells the
// comparison as an ordering.
func SameRelease(a, b string) bool {
	return Compare(a, b) == 0
}

// Components reads a tag's numbers; ok is false for anything that is
// not a dot-separated run of digits behind an optional `v`. It is
// deliberately lax where Readable is strict: a comparison that refused
// a spelling would call a refresh a downgrade over one.
func Components(tag string) ([]int, bool) {
	t := strings.TrimPrefix(strings.TrimSpace(tag), "v")
	if t == "" {
		return nil, false
	}
	parts := strings.Split(t, ".")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, false
		}
		out = append(out, n)
	}
	return out, true
}

// Readable reports whether a recorded tag can be read as a WritRun
// release: a leading `v` and two or more all-digit components.
// `v0.0.03` is one, and so is a two-component tag a later release may
// carry. This is what a health report asks of the file, where Compare
// asks only what it can order.
//
// The tag is taken already trimmed — Read returns one — so surrounding
// space is not trimmed here, where Components trims its own input.
func Readable(tag string) bool {
	if !strings.HasPrefix(tag, "v") {
		return false
	}
	parts := strings.Split(strings.TrimPrefix(tag, "v"), ".")
	if len(parts) < 2 {
		return false
	}
	for _, part := range parts {
		if part == "" {
			return false
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return false
			}
		}
	}
	return true
}
