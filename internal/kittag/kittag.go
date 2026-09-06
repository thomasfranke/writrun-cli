// Package kittag is the tag `.writrun/VERSION` records: where the file
// is, what it says, and the two questions asked of two tags — whether
// they name one release, and which release comes first. `update`,
// `status` and `doctor` all read that file, so they read it from here.
package kittag

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// Path is where an adopted repository records the kit's tag.
func Path(root string) string {
	return filepath.Join(root, ".writrun", "VERSION")
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
// This is not `Compare(a, b) == 0`: ordering carries the rule that an
// unreadable tag is a move forward, and a caller asking only whether
// two tags match must not inherit it.
func SameRelease(a, b string) bool {
	if a == b {
		return true
	}
	x, okA := Components(a)
	y, okB := Components(b)
	if !okA || !okB {
		return false
	}
	for i := 0; i < len(x) || i < len(y); i++ {
		var xi, yi int
		if i < len(x) {
			xi = x[i]
		}
		if i < len(y) {
			yi = y[i]
		}
		if xi != yi {
			return false
		}
	}
	return true
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
