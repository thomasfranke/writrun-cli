#!/usr/bin/env bash
. "$(dirname "$0")/../../init_lib.sh"

# A dirty tree is refused before anything happens — the adoption must
# be the only change (spec-0002, steps).
make_target "$TARGET"
printf 'dirt\n' > "$TARGET/uncommitted.txt"
cd "$TARGET" || exit 1

check "a dirty tree is refused by name" 1 "dirty" \
  -- "$WRITRUN" init --stage 1 --yes
check "no .writrun/ was written" 1 "" \
  -- test -d .writrun

finish
