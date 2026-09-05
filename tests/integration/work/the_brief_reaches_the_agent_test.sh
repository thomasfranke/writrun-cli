#!/usr/bin/env bash
. "$(dirname "$0")/../../work_lib.sh"

# The agent receives the brief the methodology's own script assembled,
# unedited, and is pointed at the project's instructions with it
# (spec-0007, acceptance criteria).
make_repo
configure_agent
cd "$TARGET" || exit 1

work task-0001 > /dev/null 2>&1

# The brief is the script's, not this suite's: it is run here and the
# launched argument is compared against what it printed.
bash .writrun/skills/writrun-select-next-task/brief.sh task-0001 > "$WORK/brief.out" 2>/dev/null

check "the launch carried the brief's header" 0 "task-0001  ready  medium  specs: spec-0001 approved" \
  -- launched_argument
check "the launch carried the task's body" 0 "One paragraph of brief" -- launched_argument
check "the launch carried the spec" 0 "spec-0001 — the contract" -- launched_argument
check "the launch carried every line the script printed" 0 "" \
  -- bash -c 'while IFS= read -r line; do
       [ -n "$line" ] || continue
       grep -qF -- "$line" "$AGENT_LOG" || { echo "missing: $line"; exit 1; }
     done < "'"$WORK"'/brief.out"'
check "the agent is pointed at the project's instructions" 0 "AGENTS.md at the repository root" \
  -- launched_argument

finish
