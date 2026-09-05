#!/usr/bin/env bash
. "$(dirname "$0")/../../take_lib.sh"

# The kit's WRITRUN_PR_LIST seam supplies the open pull requests, so the
# forge is never asked: a task already in flight is the script's refusal
# and the wrapper's exit 1 (spec-0008, tests required). Both conduct
# flags are true here, because the forge reads sit after the gate.
make_repo true true
export WRITRUN_PR_LIST="$(printf '41\ttask/0001-a-thing\tsomeone\t[TASK-0001] A thing to do')"
cd "$TARGET" || exit 1

check "a task in flight is refused" 1 "already in flight on pull request #41" \
  -- take task-0001 --title "$TITLE"

check "no branch was cut" 1 "" \
  -- git rev-parse --verify --quiet refs/heads/task/0001-a-thing
check "no pull request was opened" 1 "" \
  -- grep -q "pr create" "$GH_LOG"

finish
