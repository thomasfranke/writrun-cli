#!/usr/bin/env bash
. "$(dirname "$0")/../../status_lib.sh"

# Step 5 compares the kit's recorded tag with the tag this client pins:
# a difference names both values and nothing bridges it — no refresh, no
# rewrite of the file (spec-0013, acceptance criteria).
make_repo
kit_tag v0.0.01
cd "$TARGET" || exit 1

"$WRITRUN" status > "$WORK/mismatch.out" 2>&1

check "the recorded tag is named" 0 "v0.0.01" -- grep "^Kit" "$WORK/mismatch.out"
check "the pinned tag is named beside it" 0 "v0.0.03" -- grep "^Kit" "$WORK/mismatch.out"
check "a mismatch is not a failure" 0 "" -- "$WRITRUN" status
check "nothing bridged it" 0 "v0.0.01" -- cat "$TARGET/.writrun/VERSION"

# The same release spelled two ways is one release, not a mismatch.
kit_tag v0.0.3
check "a spelling is not a difference" 0 "Kit      WritRun v0.0.3 — the tag this client pins" \
  -- "$WRITRUN" status

finish
