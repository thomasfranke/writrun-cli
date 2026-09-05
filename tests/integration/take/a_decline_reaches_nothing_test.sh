#!/usr/bin/env bash
. "$(dirname "$0")/../../take_lib.sh"

# Exit 2 — composed and waiting: the composition is shown, the question
# is asked, and a no leaves the forge untouched (spec-0008, acceptance
# criteria). WRITRUN_TTY_IN stands in for the terminal.
make_repo true false
printf 'n' > "$WORK/keys"
export WRITRUN_TTY_IN="$WORK/keys"
cd "$TARGET" || exit 1

check "the composition is shown before the question" 1 "branch: task/0001-a-thing" \
  -- "$WRITRUN" take task-0001 --title "$TITLE"

unset WRITRUN_TTY_IN
check "no branch was cut" 1 "" \
  -- git rev-parse --verify --quiet refs/heads/task/0001-a-thing
check "no branch reached the origin" 1 "" \
  -- git -C "$ORIGIN" rev-parse --verify --quiet refs/heads/task/0001-a-thing
check "no pull request was opened" 1 "" \
  -- grep -q "pr create" "$GH_LOG"
check "the task file is unchanged in the working tree" 0 "" \
  -- git diff --quiet -- work/tasks

finish
