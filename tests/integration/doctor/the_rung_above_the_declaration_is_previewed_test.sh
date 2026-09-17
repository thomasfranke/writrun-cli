#!/usr/bin/env bash
. "$(dirname "$0")/../../doctor_lib.sh"

# The question a person actually asks is "can I move up yet", and the
# report answers it: the rung above the declaration is previewed, its
# requirements are counted, and none of them reaches the exit status
# (spec-0036).
make_repo 1
forge_healthy
cd "$TARGET" || exit 1

"$WRITRUN" doctor > "$WORK/reach.out" 2>&1

check "a previewed stage exits 0" 0 "" -- "$WRITRUN" doctor
check "the header names the rung previewed" 0 \
  "Stage 1 is declared — stages 0–1 examined, stage 2 previewed." -- cat "$WORK/reach.out"
check "the previewed group counts its rows" 0 \
  "Stage 2 — the forge, previewed: 6 of 6 met." -- cat "$WORK/reach.out"
check "the rung above the preview says it is not previewed" 0 \
  "Stage 3 — Issues: not previewed — one rung at a time." -- cat "$WORK/reach.out"
check "the report answers whether the stage is within reach" 0 \
  "Stage 2 is within reach: its 6 requirements are met." -- cat "$WORK/reach.out"
check "the preview is the first forge read the run makes" 0 "" \
  -- grep -q "allow_squash_merge" "$GH_DIR/calls"

# A previewed requirement that is unmet is reported and still exits 0: a
# stage nobody declared cannot fail a build.
forge_healthy
gh_reply "api repos/{owner}/{repo} --jq .allow_squash_merge" "false"
"$WRITRUN" doctor > "$WORK/unmet.out" 2>&1
check "an unmet previewed requirement does not fail the run" 0 "" -- "$WRITRUN" doctor
check "the previewed count says what is unmet" 0 \
  "Stage 2 — the forge, previewed: 5 of 6 met." -- cat "$WORK/unmet.out"
check "the report says the stage is out of reach" 0 \
  "Stage 2 is not within reach: 1 requirement unmet." -- cat "$WORK/unmet.out"

finish
