#!/usr/bin/env bash
. "$(dirname "$0")/../../doctor_lib.sh"

# Stage 1 reads two files: `.writrun/gates.md`, whose every row a project
# must answer, and AGENTS.md, whose absence is what breaks a flow
# (spec-0025, step 7).
make_repo 1
cd "$TARGET" || exit 1

gates "<!-- TODO — default: human writes or reviews before merge -->"
check "a gate left as a placeholder is named" 1 \
  "the gate for Writing or changing anything under docs/ is unanswered" \
  -- "$WRITRUN" doctor
check "a placeholder breaks a flow" 1 "breaking a flow" -- "$WRITRUN" doctor

gates
rm -f "$TARGET/.writrun/gates.md"
check "a missing gates file is named" 1 \
  ".writrun/gates.md — the project's gate answers are missing" -- "$WRITRUN" doctor

gates
legacy_agents
check "a stale fenced section is named" 0 \
  "a writrun:begin/writrun:end section is still there" -- "$WRITRUN" doctor
check "a stale section breaks nothing" 0 "none breaking a flow" -- "$WRITRUN" doctor

agents
rm -f "$TARGET/AGENTS.md"
check "a missing entry point is named" 1 "AGENTS.md — the agents' entry point is missing" \
  -- "$WRITRUN" doctor

agents
check "an answered table holds" 0 "Stage 1 — files: all clear." -- "$WRITRUN" doctor

finish
