#!/usr/bin/env bash
. "$(dirname "$0")/../../finish_lib.sh"

# spec-0002 promises a doc this branch never touched, so check_deltas.sh
# says MISSING and the sequence stops there: no status written, no
# completed date, nothing on the forge (spec-0010, acceptance criteria).
make_repo
on_branch task/0002-another-thing
cd "$TARGET" || exit 1

check "the delta check refuses in its own words" 1 "MISSING: spec-0002's promised change" \
  -- finish_cmd --yes

check "the spec was not marked implemented" 0 "approved" \
  -- field status work/specs/spec-0002-another-thing.md
check "no completed date was written" 0 "null" \
  -- field completed work/tasks/task-0002-another-thing.md
check "the task's status line is untouched" 0 "in-progress" \
  -- field status work/tasks/task-0002-another-thing.md
check "preflight never ran" 1 "" \
  -- grep -q "PREFLIGHT" "$FINISH_OUT"
check "the forge was never reached" 1 "" \
  -- grep -q "pr " "$GH_LOG"

finish
