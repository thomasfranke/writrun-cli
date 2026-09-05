#!/usr/bin/env bash
. "$(dirname "$0")/../../init_lib.sh"

# A damaged fence stops the refresh before anything is written
# (spec-0003, acceptance criteria).
make_target "$TARGET"
cd "$TARGET" || exit 1
"$WRITRUN" init --stage 1 --yes > /dev/null 2>&1 || { echo "FAIL  the fixture could not adopt"; exit 1; }
age_kit "$TARGET"
# The closing marker goes; the opening one stays.
grep -v "writrun:end" AGENTS.md > AGENTS.md.tmp && mv AGENTS.md.tmp AGENTS.md
git_q add -A && git_q commit -q -m "chore: the kit, one release back"

check "update refuses a damaged fence" 1 "fenced section" \
  -- "$WRITRUN" update --yes
check "the tag is still the old one" 0 "$OLD_TAG" -- cat .writrun/VERSION
check "nothing was written" 0 "" -- git_q diff --quiet

finish
