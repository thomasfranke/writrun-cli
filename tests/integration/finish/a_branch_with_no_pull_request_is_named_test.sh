#!/usr/bin/env bash
. "$(dirname "$0")/../../finish_lib.sh"

# The forge has no pull request for this branch: its own words reach the
# user, nothing is marked ready, and the completion edits are put back —
# a forge that will not answer is one more end that is not a success
# (spec-0017). The checks before it ran, so the reporting says exactly
# how far the sequence got.
make_repo
on_branch task/0001-a-thing
export GH_PR_VIEW_FAILS=1
cd "$TARGET" || exit 1

check "the forge's own words reach the user" 1 "no pull requests found for branch" \
  -- finish_cmd --yes

check "preflight had already passed" 0 "" \
  -- grep -q "PREFLIGHT OK" "$FINISH_OUT"
check "nothing was marked ready" 1 "" \
  -- grep -q "pr ready" "$GH_LOG"
check "the completion edits are put back" 0 "approved" \
  -- field status work/specs/spec-0001-a-thing.md
check "the completion date is put back" 0 "null" \
  -- field completed work/tasks/task-0001-a-thing.md

finish
