#!/usr/bin/env bash
. "$(dirname "$0")/../../list_lib.sh"

# There is no queue where the kit was never installed: the refusal
# names what is missing (spec-0001, the frame's needs).
mkdir -p "$WORK/plain"
cd "$WORK/plain" || exit 1
git_q init -q

check "list runs only where .writrun/ is" 1 "not an adopted repository" \
  -- "$WRITRUN" list

finish
