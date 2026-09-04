#!/usr/bin/env bash
. "$(dirname "$0")/../../init_lib.sh"

# A commit-msg hook another tool installed is refused, named, never
# overwritten (spec-0002, edge cases).
make_target "$TARGET"
mkdir -p "$TARGET/.git/hooks"
printf '#!/bin/sh\nexit 0\n' > "$TARGET/.git/hooks/commit-msg"
chmod +x "$TARGET/.git/hooks/commit-msg"
cd "$TARGET" || exit 1

check "the foreign hook is refused by name" 1 "commit-msg hook is already installed" \
  -- "$WRITRUN" init --stage 1 --yes
check "the foreign hook survives untouched" 0 "" \
  -- grep -q "exit 0" .git/hooks/commit-msg
check "no .writrun/ was written" 1 "" \
  -- test -d .writrun

finish
