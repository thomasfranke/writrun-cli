#!/usr/bin/env bash
. "$(dirname "$0")/../../list_lib.sh"

# The queue files are the authority: a forge that answers nothing
# changes no group, and its warning reaches the reader unedited
# (spec-0006, acceptance criteria).
make_repo
task task-0001 ready high spec-0001 "A task that may be taken"
spec spec-0001 approved
forge_offline
cd "$TARGET" || exit 1

"$WRITRUN" list > "$WORK/offline.out" 2>&1

check "the task is still available" 0 "task-0001" \
  -- grep "Available — any of these may be taken:" -A2 "$WORK/offline.out"
check "the script's warning is passed through" 0 "could not reach GitHub" \
  -- cat "$WORK/offline.out"
check "an unreachable forge is not a failure" 0 "" -- "$WRITRUN" list
check "the warning survives a filter" 0 "could not reach GitHub" \
  -- "$WRITRUN" list --held

finish
