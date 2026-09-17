#!/usr/bin/env bash
. "$(dirname "$0")/../../doctor_lib.sh"

# Stage 1 reads two files: `writrun/gates.md`, whose every row a project
# must answer, and AGENTS.md, whose absence is what breaks a flow
# (spec-0025, step 7).
make_repo 1
cd "$TARGET" || exit 1

gates "<!-- TODO — default: human writes or reviews before merge -->"
check "a gate left as a placeholder is named" 1 \
  "unanswered: Writing or changing anything under docs/" \
  -- "$WRITRUN" doctor
check "a placeholder breaks a flow" 1 "breaking a flow" -- "$WRITRUN" doctor

# A project that never wrote its own gates defers to the kit's default,
# which answers every gate the cautious way — so an absent file is a
# complete answer rather than a missing one (WritRun v0.0.07's two
# homes). It was a breaking finding while the gates were a TODO in the
# adopter's file that no update could correct.
gates
rm -f "$TARGET/writrun/gates.md"
check "an absent gates file is answered by the kit's default" 0 \
  "Stage 1 — files: 9 of 9 met." -- "$WRITRUN" doctor

gates
legacy_agents
check "a stale fenced section is named" 0 \
  "!  AGENTS.md — a writrun:begin/writrun:end section is stale" -- "$WRITRUN" doctor
check "a stale section breaks nothing" 0 "none breaking a flow" -- "$WRITRUN" doctor

agents
rm -f "$TARGET/AGENTS.md"
check "a missing entry point is named" 1 "AGENTS.md — the agents' entry point is missing" \
  -- "$WRITRUN" doctor

agents
check "an answered table holds" 0 "Stage 1 — files: 9 of 9 met." -- "$WRITRUN" doctor

finish
