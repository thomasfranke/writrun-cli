#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# Every release names the WritRun tag it pins. The notes footer reads it
# from $WRITRUN_TAG, and scripts/writrun-tag.sh fills that variable from
# cmd/writrun/main.go — the one place the tag is written, and the one
# the binary prints under --version.

check "the notes footer names the pinned tag" 0 "" \
  -- grep -qF -- 'Pins WritRun `{{ .Env.WRITRUN_TAG }}`' "$GORELEASER_YAML"

tag="$(bash "$WRITRUN_TAG_SH")"
check "the tag comes from the constant the binary prints" 0 "" \
  -- grep -qF "const writrunTag = \"$tag\"" "$MAIN_GO"

# A source without the constant is a release that could not name its
# tag: the reader fails instead of printing nothing.
WORK=$(mktemp -d)
mkdir -p "$WORK/scripts" "$WORK/cmd/writrun"
cp "$WRITRUN_TAG_SH" "$WORK/scripts/writrun-tag.sh"
printf 'package main\n' > "$WORK/cmd/writrun/main.go"
check "a source without the constant fails loudly" 1 "no writrunTag constant" \
  -- bash "$WORK/scripts/writrun-tag.sh"

finish
