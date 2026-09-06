#!/usr/bin/env bash
. "$(dirname "$0")/../../amend_lib.sh"

# Step 3: task-0012 is in flight on #42, so the body carries the line
# `check_amendment_reference.sh` accepts — that pull request's number,
# and that task's id (spec-0011, acceptance criteria).
make_repo
in_flight_pr
cd "$TARGET" || exit 1

check "the amendment is opened" 0 "Amended spec-0011" \
  -- amend_cmd spec-0011 --title "$TITLE" --yes

check "the composition named the suspension before the act" 0 "" \
  -- grep -q "suspends   task-0012 on #42" "$AMEND_OUT"
check "the body carries the line the check accepts" 0 "" \
  -- grep -qF "Suspends #42 — task-0012 waits on this amendment." "$GH_LOG"
check "no other task is named" 1 "" \
  -- grep -qF "task-0014 waits on this amendment" "$GH_LOG"
check "no task's status moved" 0 "in-progress" \
  -- field status work/tasks/task-0012-amend-command.md
check "no task's taken_by moved" 0 "null" \
  -- field taken_by work/tasks/task-0012-amend-command.md

finish
