package updatecmd

import (
	"strconv"
	"strings"
)

// compareTags orders two WritRun tags: -1 when a precedes b, 0 when
// they are the same release, 1 when a follows b. The components are
// read as numbers, so `v0.0.03` and `v0.0.3` are one release and
// `v0.0.10` follows `v0.0.9` — which string order gets wrong.
//
// A tag neither side can parse compares equal only to itself: an
// unreadable version is not a reason to refuse a refresh, but it is no
// reason to call it a downgrade either.
func compareTags(a, b string) int {
	if a == b {
		return 0
	}
	x, okA := parseTag(a)
	y, okB := parseTag(b)
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

func parseTag(tag string) ([]int, bool) {
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
