#!/usr/bin/env bash
. "$(dirname "$0")/../../cli_lib.sh"

# `--help` groups its rows by what a person is doing, and every row's
# text is the command table's own summary — the same string the entry
# screen shows (spec-0039, steps 5 and 6).
mkdir -p "$WORK/nowhere"
cd "$WORK/nowhere" || exit 1

"$WRITRUN" --help > "$WORK/help.out" 2>&1

check "getting started comes first" 0 "^GETTING STARTED" -- cat "$WORK/help.out"
check "doing the work is a group" 0 "^DOING THE WORK" -- cat "$WORK/help.out"
check "writing the rules is a group" 0 "^WRITING THE RULES" -- cat "$WORK/help.out"
check "keeping it current is a group" 0 "^KEEPING IT CURRENT" -- cat "$WORK/help.out"
check "a row carries the command table's own summary" 0 \
  "  status      where does the work on this branch stand?" -- cat "$WORK/help.out"
check "the explanation every command offers is named" 0 \
  "^Run any command with no arguments and it explains itself first." -- cat "$WORK/help.out"

finish
