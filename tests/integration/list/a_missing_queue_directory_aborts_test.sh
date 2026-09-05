#!/usr/bin/env bash
. "$(dirname "$0")/../../list_lib.sh"

# No queue to read is a refusal, and the cause is the script's own words
# (spec-0006, acceptance criteria).
make_repo
rm -rf "$TARGET/work/tasks"
cd "$TARGET" || exit 1

check "a missing queue directory aborts naming the cause" 3 "No such directory" \
  -- "$WRITRUN" list

finish
