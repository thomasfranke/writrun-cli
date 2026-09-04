#!/usr/bin/env bash
. "$(dirname "$0")/../../cli_lib.sh"

# An unknown subcommand is usage on stderr and a non-zero exit
# (spec-0001, edge cases).
check "an unknown subcommand aborts naming itself" 2 'unknown command "bogus"' \
  -- "$WRITRUN" bogus
check "usage is shown with the refusal" 2 "usage: writrun" \
  -- "$WRITRUN" bogus

finish
