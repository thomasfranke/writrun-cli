#!/usr/bin/env bash
. "$(dirname "$0")/../../take_lib.sh"

# With no task id and a terminal, the available tasks are arrow-selected
# (spec-0008, acceptance criteria): a down arrow and enter land the
# second of them. --title answers the one free-text question, so the
# selection is the only form reading the keys.
make_repo true false
printf '\033[B\r' > "$WORK/keys"
export WRITRUN_TTY_IN="$WORK/keys"
cd "$TARGET" || exit 1

check "the selected task is the one composed" 0 "branch: task/0002-another-thing" \
  -- take --title "$TITLE" --yes

check "the selection offered only the available group" 0 "" \
  -- grep -q "Took task-0002" "$TAKE_OUT"

unset WRITRUN_TTY_IN
check "without a terminal the task id is required" 1 "the task id as an argument" \
  -- take --title "$TITLE" --yes
check "without a terminal the title is required" 1 "\-\-title" \
  -- take task-0001

finish
