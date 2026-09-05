#!/usr/bin/env bash
. "$(dirname "$0")/../../finish_lib.sh"

# task-0003 carries no spec: there are no deltas to check and no Outcome
# to record, and the completion date is written and preflight still
# gates all the same (spec-0010, edge cases).
make_repo
on_branch task/0003-a-third-thing
cd "$TARGET" || exit 1

check "the finish says why no deltas were checked" 0 "task-0003 carries no spec" \
  -- finish_cmd --yes

check "no delta check ran" 1 "" \
  -- grep -q "promised deltas of spec" "$FINISH_OUT"
check "preflight still ran" 0 "" \
  -- grep -q "PREFLIGHT OK" "$FINISH_OUT"
check "the task carries a completed date" 0 "" \
  -- grep -qE '^completed: 2[0-9]{3}-' work/tasks/task-0003-a-third-thing.md
check "the task's status line is untouched" 0 "in-progress" \
  -- field status work/tasks/task-0003-a-third-thing.md
check "the pull request was marked ready" 0 "" \
  -- grep -qx "pr ready 7" "$GH_LOG"

finish
