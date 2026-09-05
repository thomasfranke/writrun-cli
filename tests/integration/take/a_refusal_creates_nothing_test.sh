#!/usr/bin/env bash
. "$(dirname "$0")/../../take_lib.sh"

# Exit 1 — the script refuses an ineligible task, and the wrapper passes
# the reason and the code through having created nothing (spec-0008,
# acceptance criteria).
make_repo true false
cd "$TARGET" || exit 1

check "the script's refusal is the wrapper's refusal" 1 "REFUSED" \
  -- "$WRITRUN" take task-0003 --title "$TITLE" --yes

check "no branch was cut" 1 "" \
  -- git rev-parse --verify --quiet refs/heads/task/0003-a-third-thing
check "nothing reached the forge" 1 "" \
  -- grep -q "pr create" "$GH_LOG"
check "the task file is unchanged in the working tree" 0 "" \
  -- git diff --quiet -- work/tasks

# A title the declared style refuses is the script's word too.
check "a badly styled title is refused verbatim" 1 "does not read as the declared" \
  -- "$WRITRUN" take task-0001 --title "just some words" --yes

finish
