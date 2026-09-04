#!/usr/bin/env bash
. "$(dirname "$0")/../../cli_lib.sh"

# --help answers anywhere, prints the docs' address, and a bare
# `writrun` is the same answer (product/rules.md).
mkdir -p "$WORK/nowhere"
cd "$WORK/nowhere" || exit 1

check "writrun --help answers outside any repository" 0 "Docs: https://" \
  -- "$WRITRUN" --help
check "a bare writrun answers with the same help" 0 "porcelain for WritRun" \
  -- "$WRITRUN"

finish
