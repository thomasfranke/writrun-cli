#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# The tag is the only place the number is written, so the build stamps
# it into `main.version` — the seam cmd/writrun/main.go reads before
# falling back to the module version `go install` records. The `v` stays
# on: `writrun --version` prints the string as given.
#
# The flags are read out of the configuration and linked for real, so a
# renamed variable or a wrong template fails here rather than on the
# release.

check "the version is stamped from the tag" 0 "" \
  -- grep -qF -- '-X main.version={{ .Tag }}' "$GORELEASER_YAML"

ldflags="$(sed -n 's/^ *- \(-s -w -X main\.version=.*\)$/\1/p' "$GORELEASER_YAML")"
stamped="${ldflags//\{\{ .Tag \}\}/v9.9.9}"

WORK=$(mktemp -d)
if (cd "$REPO_ROOT" && go build -ldflags "$stamped" -o "$WORK/writrun" ./cmd/writrun); then
  check "the linked binary names the stamped tag" 0 "writrun v9.9.9" \
    -- "$WORK/writrun" --version
else
  echo "FAIL  the linked binary names the stamped tag: go build failed"
  fail=$((fail + 1))
fi

finish
