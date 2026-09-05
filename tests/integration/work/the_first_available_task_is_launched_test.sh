#!/usr/bin/env bash
. "$(dirname "$0")/../../work_lib.sh"

# With no argument the task launched is the first of the lister's
# available group — the algorithm's order, run by the algorithm's own
# script (spec-0007, steps).
make_repo
configure_agent
cd "$TARGET" || exit 1
RESOLVED=$(pwd -P)

check "the run says which task it launched on" 0 "task-0001 — launching" -- work
check "the agent was started with one argument" 0 "argc=1" -- launched_argument
check "the first available task is the one worked" 0 "Work task-0001 in this repository" \
  -- launched_argument
check "the agent runs from the repository root" 0 "cwd=$RESOLVED" -- launched_argument

# The agent's own exit is the command's: nothing is restated, and a
# failed agent is not reported as a successful launch.
export AGENT_EXIT=3
check "the agent's exit travels up" 3 "" -- work
unset AGENT_EXIT

finish
