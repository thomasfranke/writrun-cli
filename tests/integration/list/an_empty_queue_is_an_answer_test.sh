#!/usr/bin/env bash
. "$(dirname "$0")/../../list_lib.sh"

# Nothing available is an answer to the question asked, not a failure:
# the lister's own message, exit 0 (spec-0006, edge cases).
make_repo
cd "$TARGET" || exit 1

check "an empty queue exits 0 with the script's own message" 0 "Nothing is available." \
  -- "$WRITRUN" list

finish
