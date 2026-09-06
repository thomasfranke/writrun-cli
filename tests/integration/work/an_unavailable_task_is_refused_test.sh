#!/usr/bin/env bash
. "$(dirname "$0")/../../work_lib.sh"

# A task the queue holds back is refused in the lister's own words —
# the section it sits under and the line it sits on, quoted rather than
# summarised — and nothing is launched (spec-0007, acceptance criteria).
make_repo
configure_agent
cd "$TARGET" || exit 1

check "the refusal names the lister's section" 1 "Held back:" -- work task-0003
check "the refusal carries the lister's own reason" 1 "spec-0003 is draft" \
  -- work task-0003
check "nothing was launched" 0 "" -- test ! -s "$AGENT_LOG"

# A task no section of the answer holds is not available either.
check "a task in no section is refused" 1 "task-0042 is in no section" -- work task-0042
check "still nothing was launched" 0 "" -- test ! -s "$AGENT_LOG"

# A named available task is the one worked, wherever it sits in the order.
check "the named available task is the one launched" 0 "task-0002 — launching" \
  -- work task-0002
check "the agent got the task it was named" 0 "Work task-0002 in this repository" \
  -- launched_argument

finish
