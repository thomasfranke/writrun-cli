#!/usr/bin/env bash
. "$(dirname "$0")/../../take_lib.sh"

# Exit 2 then exit 0 — the composition is printed even under --yes, and
# the confirmed rerun performs exactly the act it showed (spec-0008,
# edge cases). The forge is the stub throughout.
make_repo true false
cd "$TARGET" || exit 1

check "the take reports the script's own words" 0 "Took task-0001" \
  -- take task-0001 --title "$TITLE" --slug a-thing --yes

check "the composition was shown before the confirmed rerun" 0 "" \
  -- grep -q "branch: task/0001-a-thing" "$TAKE_OUT"
check "the branch reached the origin" 0 "" \
  -- git -C "$ORIGIN" rev-parse --verify --quiet refs/heads/task/0001-a-thing
check "the draft pull request was opened" 0 "" \
  -- grep -q -- "pr create --draft" "$GH_LOG"
check "the pull request title carries the task tag" 0 "" \
  -- grep -qF -- "[TASK-0001] $TITLE" "$GH_LOG"
check "the task file is unchanged in the working tree" 0 "" \
  -- git diff --quiet main -- work/tasks

finish
