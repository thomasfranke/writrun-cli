#!/usr/bin/env bash
. "$(dirname "$0")/../../finish_lib.sh"

# The same window on the interrupt: `--yes` means no prompt is reading
# the terminal, so the Ctrl-C at a run whose silence reads as a hang is
# a real SIGINT and the guard is what answers it. The tree ends as it
# started and the status is the signal's 130 (spec-0021, acceptance
# criteria).
make_repo
on_branch task/0001-a-thing
slow_script .writrun/scripts/stage-1-tasks-and-specs/preflight.sh 30
cd "$TARGET" || exit 1

TASK="$TARGET/work/tasks/task-0001-a-thing.md"
SPEC="$TARGET/work/specs/spec-0001-a-thing.md"

check "the run dies of the signal, at the status the signal gives" 130 "" \
  -- finish_into_a_signal INT --yes

check "git status --porcelain is empty after the interrupt" 0 "" \
  -- tree_is_clean
check "the spec is the approved one it was" 0 "approved" \
  -- field status "$SPEC"
check "the task carries no completion date" 0 "null" \
  -- field completed "$TASK"
check "the spec is reported put back" 0 "" \
  -- grep -q "restored work/specs/spec-0001-a-thing.md" "$FINISH_OUT"
check "the task is reported put back" 0 "" \
  -- grep -q "restored work/tasks/task-0001-a-thing.md" "$FINISH_OUT"

finish
