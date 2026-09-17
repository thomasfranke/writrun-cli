#!/usr/bin/env bash
. "$(dirname "$0")/../../list_lib.sh"

# `list` is run daily and a daily explanation is noise: a bare run does
# its work and prints no description, terminal or not. `--help` is where
# the description is asked for (spec-0039, step 4).
make_repo
printf '\033' > "$WORK/keys"
export WRITRUN_TTY_IN="$WORK/keys"
cd "$TARGET" || exit 1

"$WRITRUN" list > "$WORK/list.out" 2>&1
check "a bare run explains nothing" 1 "" -- grep -q "why you would" "$WORK/list.out"
check "it did its work" 0 "Nothing is available." -- cat "$WORK/list.out"
check "--help is where the description is" 0 \
  "^list — see what work is waiting, and what is blocked" -- "$WRITRUN" list --help

finish
