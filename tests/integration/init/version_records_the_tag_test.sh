#!/usr/bin/env bash
. "$(dirname "$0")/../../init_lib.sh"

# When init completes, .writrun/VERSION names the tag the kit came
# from (spec-0002, acceptance criteria).
make_target "$TARGET"
cd "$TARGET" || exit 1

check "init adopts" 0 "" -- "$WRITRUN" init --stage 1 --yes
check ".writrun/VERSION names the pinned tag" 0 "$TAG" \
  -- cat .writrun/VERSION

finish
