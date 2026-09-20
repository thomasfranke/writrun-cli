#!/usr/bin/env bash
# The repository this case needs is the one the doctor cases build: the
# kit's own reader and checker, a project half that satisfies stage 1,
# and the forge stubbed. The subject is `config`, which is why the case
# lives here (spec-0041).
. "$(dirname "$0")/../../doctor_lib.sh"

# Raising the stage shows what the new stage would require of this
# repository — doctor's own answers, in doctor's own marks — and then
# writes. It never refuses: a check that could not be read is unread,
# and the stage stays a declaration.
make_repo 1
forge_healthy
cd "$TARGET" || exit 1

"$WRITRUN" --yes config stage 2 > "$WORK/raise.out" 2>&1

check "the raise is previewed and written" 0 "" -- cat "$WORK/raise.out"
check "the preview opens by naming the stage" 0 \
  "stage 2 — pull requests. doctor previews what the stage would" -- cat "$WORK/raise.out"
check "the requirements are shown in doctor's marks" 0 \
  "✓  gh on the PATH" -- cat "$WORK/raise.out"
check "the counts are the preview's own" 0 \
  "7 of 7 met, 0 unmet, 0 unread." -- cat "$WORK/raise.out"
check "the stage was written" 0 "stage is 2." -- cat "$WORK/raise.out"
check "the settings file holds the new stage" 0 '"stage": 2' \
  -- cat "$TARGET/writrun/settings.json"
check "the kit's own checker was asked" 0 "" -- bash "$TARGET/$SETTINGS_CHECK"

# An unmet requirement is shown and refuses nothing.
settings 1
gh_reply "api repos/{owner}/{repo} --jq .allow_squash_merge" "false"
"$WRITRUN" --yes config stage 2 > "$WORK/unmet.out" 2>&1
check "an unmet requirement is marked" 0 "✗  squash merging is on" -- cat "$WORK/unmet.out"
check "an unmet requirement does not refuse the declaration" 0 "stage is 2." \
  -- cat "$WORK/unmet.out"

# Nothing new is required of the repository when the stage is lowered,
# so nothing is previewed.
settings 3
: > "$GH_DIR/calls"
"$WRITRUN" --yes config stage 1 > "$WORK/lower.out" 2>&1
check "a lowered stage is written" 0 "stage is 1." -- cat "$WORK/lower.out"
check "a lowered stage previews nothing" 1 "" -- grep -q "doctor previews" "$WORK/lower.out"
check "a lowered stage reaches no forge" 1 "" -- test -s "$GH_DIR/calls"

finish
