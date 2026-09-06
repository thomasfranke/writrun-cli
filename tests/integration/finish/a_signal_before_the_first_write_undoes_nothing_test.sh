#!/usr/bin/env bash
. "$(dirname "$0")/../../finish_lib.sh"

# The window opens at the first completion write, not at the command.
# A SIGTERM into the deltas check — step 1, before anything is written —
# has nothing to put back, so it puts nothing back and says nothing
# (spec-0021, acceptance criteria).
make_repo
on_branch task/0001-a-thing
slow_script .writrun/skills/writrun-check-spec-deltas/check_deltas.sh 30
cd "$TARGET" || exit 1

TASK="$TARGET/work/tasks/task-0001-a-thing.md"
SPEC="$TARGET/work/specs/spec-0001-a-thing.md"

check "the run dies of the signal, at the status the signal gives" 143 "" \
  -- finish_into_a_signal TERM --yes

check "git status --porcelain is empty before the first write" 0 "" \
  -- tree_is_clean
check "the spec is the approved one it was" 0 "approved" \
  -- field status "$SPEC"
check "the task carries no completion date" 0 "null" \
  -- field completed "$TASK"
check "nothing was reported put back" 1 "" \
  -- grep -q "restored " "$FINISH_OUT"
check "no failure to put anything back was reported" 1 "" \
  -- grep -q "The working tree is left changed" "$FINISH_OUT"

finish
