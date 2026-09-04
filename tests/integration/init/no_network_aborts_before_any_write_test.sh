#!/usr/bin/env bash
. "$(dirname "$0")/../../init_lib.sh"

# An unreachable source aborts before any write, naming the fetch
# failure (spec-0002, edge cases).
make_target "$TARGET"
export WRITRUN_SOURCE="$WORK/nowhere"
cd "$TARGET" || exit 1

check "the fetch failure aborts naming itself" 1 "nothing was written" \
  -- "$WRITRUN" init --stage 1 --yes
check "no .writrun/ was written" 1 "" \
  -- test -d .writrun

finish
