#!/usr/bin/env bash
. "$(dirname "$0")/../../cli_lib.sh"

# An unknown flag is usage on stderr and a non-zero exit
# (spec-0001, edge cases).
check "an unknown flag aborts naming itself" 2 "unknown flag --bogus" \
  -- "$WRITRUN" --bogus

finish
