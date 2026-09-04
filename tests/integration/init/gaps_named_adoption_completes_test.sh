#!/usr/bin/env bash
. "$(dirname "$0")/../../init_lib.sh"

# The chosen stage's checks name each gap and the adoption still
# completes (spec-0002, acceptance criteria): this target has no About
# file and no real chapters, and exits zero anyway.
make_target "$TARGET"
cd "$TARGET" || exit 1

check "init completes despite the gaps" 0 "named, not fixed" \
  -- "$WRITRUN" init --stage 1 --yes
check "the kit landed regardless" 0 "" \
  -- test -d .writrun

finish
