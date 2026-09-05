#!/usr/bin/env bash
. "$(dirname "$0")/../../init_lib.sh"

# A kit recording a tag newer than the one this binary pins is not
# refreshed backwards: a downgrade is a deliberate act update does not
# offer (spec-0003, edge cases).
make_target "$TARGET"
cd "$TARGET" || exit 1
"$WRITRUN" init --stage 1 --yes > /dev/null 2>&1 || { echo "FAIL  the fixture could not adopt"; exit 1; }
printf 'v9.9.9\n' > .writrun/VERSION
git_q add -A && git_q commit -q -m "chore: a kit from the future"

check "update refuses to move backwards" 1 "downgrade" \
  -- "$WRITRUN" update --yes
check "the recorded tag is untouched" 0 "v9.9.9" -- cat .writrun/VERSION
check "nothing was written" 0 "" -- git_q diff --quiet

finish
