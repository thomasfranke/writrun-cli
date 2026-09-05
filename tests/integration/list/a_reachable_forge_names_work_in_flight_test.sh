#!/usr/bin/env bash
. "$(dirname "$0")/../../list_lib.sh"

# The forge answers: a task whose pull request is open is named as in
# flight, not offered — and a queue with nothing left to take still
# exits 0 (spec-0006, steps).
make_repo
task task-0001 ready high spec-0001 "A task somebody already took"
spec spec-0001 approved
forge_online 7 "task/0001-a-task" someone "[TASK-0001] A task somebody already took"
cd "$TARGET" || exit 1

"$WRITRUN" list > "$WORK/inflight.out" 2>&1

check "the open pull request is named" 0 "#7 by @someone" \
  -- grep "task-0001" "$WORK/inflight.out"
check "the taken task is not offered" 0 "Nothing is available." \
  -- cat "$WORK/inflight.out"
check "nothing available is still exit 0" 0 "" -- "$WRITRUN" list
check "a reachable forge prints no unreachable warning" 1 "" \
  -- grep -q "could not reach GitHub" "$WORK/inflight.out"

finish
