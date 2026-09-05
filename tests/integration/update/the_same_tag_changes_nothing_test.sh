#!/usr/bin/env bash
. "$(dirname "$0")/../../init_lib.sh"

# The recorded tag equals the target: say so and change nothing
# (spec-0003, acceptance criteria).
make_target "$TARGET"
cd "$TARGET" || exit 1
"$WRITRUN" init --stage 1 --yes > /dev/null 2>&1 || { echo "FAIL  the fixture could not adopt"; exit 1; }
git_q add -A && git_q commit -q -m "chore: adopt"
before=$(git_q rev-parse HEAD)

check "update stands down on the same tag" 0 "Already at WritRun $TAG" \
  -- "$WRITRUN" update --yes
check "nothing in the tree changed" 0 "" -- git_q diff --quiet
check "the commit is still the same" 0 "$before" -- git_q rev-parse HEAD

finish
