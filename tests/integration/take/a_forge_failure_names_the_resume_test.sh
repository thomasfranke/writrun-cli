#!/usr/bin/env bash
. "$(dirname "$0")/../../take_lib.sh"

# Exit 3 — the branch is pushed and `gh pr create` fails, which is the
# one state the act must not leave behind. The wrapper passes the exit
# code, the reason and the exact --resume invocation through (spec-0008,
# acceptance criteria).
make_repo true false
export GH_PR_CREATE_EXIT=1
cd "$TARGET" || exit 1

check "the forge failure is the wrapper's exit 3" 3 "gh pr create failed" \
  -- take task-0001 --title "$TITLE" --slug a-thing --yes

check "the exact resume invocation is shown" 0 "" \
  -- grep -q -- "--slug a-thing --resume --confirm" "$TAKE_OUT"
check "the branch it left behind is named" 0 "" \
  -- grep -q "task/0001-a-thing is pushed but has no pull request" "$TAKE_OUT"
check "the task file is unchanged in the working tree" 0 "" \
  -- git diff --quiet main -- work/tasks

finish
