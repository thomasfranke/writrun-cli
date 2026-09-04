#!/usr/bin/env bash
. "$(dirname "$0")/../../cli_lib.sh"

# --version answers anywhere — here, a directory that is no repository
# at all — and names both the client's version and the pinned WritRun
# tag (product/rules.md). The suite builds from the git tree, so the
# toolchain stamps a module version and the client names it.
mkdir -p "$WORK/nowhere"
cd "$WORK/nowhere" || exit 1

check "writrun --version answers outside any repository" 0 "pins WritRun v" \
  -- "$WRITRUN" --version
check "the client's own version is named" 0 "writrun v" \
  -- "$WRITRUN" --version

finish
