#!/usr/bin/env bash
. "$(dirname "$0")/../../cli_lib.sh"

# --version answers anywhere — here, a directory that is no repository
# at all — and names both the client's version and the pinned WritRun
# tag (product/rules.md).
mkdir -p "$WORK/nowhere"
cd "$WORK/nowhere" || exit 1

check "writrun --version answers outside any repository" 0 "pins WritRun v" \
  -- "$WRITRUN" --version
check "the client's own version is named" 0 "writrun dev" \
  -- "$WRITRUN" --version

finish
