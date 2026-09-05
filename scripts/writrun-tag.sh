#!/usr/bin/env bash
# writrun-tag.sh — print the WritRun tag this build pins.
#
#   writrun-tag.sh
#
# The tag has one writer: the `writrunTag` constant in
# cmd/writrun/main.go, which the binary prints under `--version`. The
# release notes name the same tag (.goreleaser.yaml), and reading it
# from the source is what keeps the notes and the binary from
# disagreeing.
#
# Exit 1 when the constant is absent: a release that cannot name its
# pinned tag is not published (technical/architecture.md).
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
src="$root/cmd/writrun/main.go"

[ -f "$src" ] || { echo "writrun-tag: $src not found" >&2; exit 1; }

tag="$(sed -n 's/^const writrunTag = "\([^"]*\)".*/\1/p' "$src" | head -n 1)"
[ -n "$tag" ] || { echo "writrun-tag: no writrunTag constant in $src" >&2; exit 1; }

printf '%s\n' "$tag"
