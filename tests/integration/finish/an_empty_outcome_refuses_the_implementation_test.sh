#!/usr/bin/env bash
. "$(dirname "$0")/../../finish_lib.sh"

# The spec still carries the generator's placeholder, so the outcome was
# never recorded: the spec is not marked implemented, the completion
# date is not written, and nothing after step 2 runs (spec-0010, step 2).
make_repo
on_branch task/0001-a-thing
spec_file 0001 a-thing task-0001 approved none none
git_q -C "$TARGET" commit -q -am "the spec, unfinished"
cd "$TARGET" || exit 1

check "the refusal names the spec's Outcome" 1 "spec-0001's ## Outcome is empty" \
  -- finish_cmd --yes

check "the delta check still ran and passed" 0 "" \
  -- grep -q "OK: diff matches the promised deltas" "$FINISH_OUT"
check "the spec was not marked implemented" 0 "approved" \
  -- field status work/specs/spec-0001-a-thing.md
check "no completed date was written" 0 "null" \
  -- field completed work/tasks/task-0001-a-thing.md
check "preflight never ran" 1 "" \
  -- grep -q "PREFLIGHT" "$FINISH_OUT"
check "the forge was never reached" 1 "" \
  -- grep -q "pr " "$GH_LOG"

finish
