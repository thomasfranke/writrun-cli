#!/usr/bin/env bash
. "$(dirname "$0")/../../finish_lib.sh"

# report-0018's reproduction: a SIGTERM into a slowed `preflight.sh`,
# between the completion writes and the confirmation. The undo was
# ordinary control flow and a signal returned through none of it, so the
# run died at 143 with both queue files changed. It now runs the same
# undo every other non-success end runs, and still dies of the signal
# (spec-0021, acceptance criteria).
make_repo
on_branch task/0001-a-thing
slow_script .writrun/scripts/stage-1-tasks-and-specs/preflight.sh 30
cd "$TARGET" || exit 1

TASK="$TARGET/work/tasks/task-0001-a-thing.md"
SPEC="$TARGET/work/specs/spec-0001-a-thing.md"

check "the run dies of the signal, at the status the signal gives" 143 "" \
  -- finish_into_a_signal TERM --yes

check "git status --porcelain is empty after the terminate" 0 "" \
  -- tree_is_clean
check "the spec is the approved one it was" 0 "approved" \
  -- field status "$SPEC"
check "the task carries no completion date" 0 "null" \
  -- field completed "$TASK"
check "the task's status line is untouched" 0 "in-progress" \
  -- field status "$TASK"
check "the spec is reported put back" 0 "" \
  -- grep -q "restored work/specs/spec-0001-a-thing.md" "$FINISH_OUT"
check "the task is reported put back" 0 "" \
  -- grep -q "restored work/tasks/task-0001-a-thing.md" "$FINISH_OUT"
check "the pull request was never marked ready" 1 "" \
  -- grep -q "pr ready" "$GH_LOG"

finish
