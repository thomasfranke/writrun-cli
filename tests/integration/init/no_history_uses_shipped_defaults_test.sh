#!/usr/bin/env bash
. "$(dirname "$0")/../../init_lib.sh"

# A repository with no Conventional history and no contributing guide
# gets the shipped defaults, and the plan says so (spec-0002, edge
# cases).
make_target "$TARGET" "just some words" "more plain words"
cd "$TARGET" || exit 1

check "the plan says the defaults stand" 0 "shipped defaults" \
  -- "$WRITRUN" init --stage 1 --yes
check "the shipped vocabulary survives in the door" 0 'TYPES="docs feat fix refactor chore"' \
  -- grep TYPES= .writrun/scripts/stage-2-pull-requests/check_observance.sh

finish
