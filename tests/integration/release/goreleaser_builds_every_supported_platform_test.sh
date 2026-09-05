#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# The matrix GoReleaser declares is the one technical/runtime/platforms.md
# names — macOS, Linux, Windows — built without cgo, so every artifact on
# the release is static.

check "the release builds for macOS, Linux and Windows" 0 "" \
  -- equals "$(yaml_list goos)" "darwin,linux,windows"
check "each platform is built for both architectures" 0 "" \
  -- equals "$(yaml_list goarch)" "amd64,arm64"
check "the binaries are static: cgo is off" 0 "" \
  -- grep -qE '^ +- CGO_ENABLED=0$' "$GORELEASER_YAML"

finish
