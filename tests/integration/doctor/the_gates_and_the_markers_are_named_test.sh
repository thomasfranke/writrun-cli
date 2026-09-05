#!/usr/bin/env bash
. "$(dirname "$0")/../../doctor_lib.sh"

# AGENTS.md carries two assumptions: the fence a refresh writes through,
# and the four gates the methodology requires a project to answer
# (spec-0004, step 3).
make_repo 1
cd "$TARGET" || exit 1

agents "<!-- TODO — default: human writes or reviews before merge -->"
check "a gate left as a placeholder is named" 1 \
  "the gate for who writes or reviews a change under docs/ is still a placeholder" \
  -- "$WRITRUN" doctor
check "a placeholder breaks a flow" 1 "breaking a flow" -- "$WRITRUN" doctor

agents
sed -i.bak '/writrun:end/d' "$TARGET/AGENTS.md" && rm -f "$TARGET/AGENTS.md.bak"
check "damaged markers are named" 1 \
  "the fenced writrun:begin/writrun:end markers are damaged" -- "$WRITRUN" doctor

rm -f "$TARGET/AGENTS.md"
check "a missing entry point is named" 1 "AGENTS.md — the agents' entry point is missing" \
  -- "$WRITRUN" doctor

agents
check "an answered table holds" 0 "Stage 1 — files: all clear." -- "$WRITRUN" doctor

finish
