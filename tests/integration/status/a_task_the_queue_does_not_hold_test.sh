#!/usr/bin/env bash
. "$(dirname "$0")/../../status_lib.sh"

# A branch named like a task the queue does not hold is named as
# unknown, never invented — and with nothing resolved there is no
# completion to check (spec-0013, edge cases).
make_repo
cd "$TARGET" || exit 1
git_q checkout -q -b task/0099-a-task-nobody-wrote

"$WRITRUN" status > "$WORK/unknown.out" 2>&1

check "the id is named as unknown" 0 "Task     task-0099 — the queue holds no such task" \
  -- cat "$WORK/unknown.out"
check "no spec is invented for it" 1 "" -- grep -q "^Spec" "$WORK/unknown.out"
check "no check is claimed to have run" 1 "" -- grep -q "^Checks" "$WORK/unknown.out"
check "the rest is still answered" 0 "Kit      WritRun $PINNED" -- cat "$WORK/unknown.out"

finish
