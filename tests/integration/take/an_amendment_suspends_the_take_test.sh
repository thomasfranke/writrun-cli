#!/usr/bin/env bash
. "$(dirname "$0")/../../take_lib.sh"

# The kit's WRITRUN_PR_FILES seam supplies what an id-less pull request
# touches: one amending the task's spec suspends the take, and the
# wrapper passes that refusal through (spec-0008, tests required).
make_repo true true
export WRITRUN_PR_LIST="$(printf '42\tspec/0001-a-thing\tsomeone\tdocs(specs): reword the contract')"
export WRITRUN_PR_FILES="$(printf '42\twork/specs/spec-0001-a-thing.md')"
cd "$TARGET" || exit 1

check "the amendment suspends the take" 1 "suspended: pull request #42 amends spec-0001" \
  -- take task-0001 --title "$TITLE"

check "no branch was cut" 1 "" \
  -- git rev-parse --verify --quiet refs/heads/task/0001-a-thing
check "no pull request was opened" 1 "" \
  -- grep -q "pr create" "$GH_LOG"

finish
