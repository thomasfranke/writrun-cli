#!/usr/bin/env bash
. "$(dirname "$0")/../../status_lib.sh"

# On a task branch the answer names the task the branch carries, the
# spec and its status, the verdict of the completion checks, the reports
# awaiting triage and the kit's tag (spec-0013, steps).
make_repo
report report-0001 open
cd "$TARGET" || exit 1

"$WRITRUN" status > "$WORK/answer.out" 2>&1

check "the branch is named" 0 "Branch   task/0014-status-command" -- cat "$WORK/answer.out"
check "the task is named with its title" 0 "task-0014  in-progress  Answer where the work stands" \
  -- cat "$WORK/answer.out"
check "the spec is named with its status" 0 "Spec     spec-0013  approved" -- cat "$WORK/answer.out"
check "the checks are run and their verdict given" 0 "Checks   all pass" -- cat "$WORK/answer.out"
check "the open reports are counted" 0 "Reports  1 open" -- cat "$WORK/answer.out"
check "the kit's tag is named" 0 "Kit      WritRun $PINNED" -- cat "$WORK/answer.out"
check "answering is exit 0" 0 "" -- "$WRITRUN" status

finish
