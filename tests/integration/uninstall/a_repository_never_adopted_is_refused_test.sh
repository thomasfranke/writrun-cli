#!/usr/bin/env bash
. "$(dirname "$0")/../../init_lib.sh"

# Uninstall runs only where `.writrun/` is present (spec-0005,
# acceptance criteria).
make_target "$TARGET"
cd "$TARGET" || exit 1

check "uninstall refuses a repository never adopted" 1 "not an adopted repository" \
  -- "$WRITRUN" uninstall --yes
check "nothing was written" 0 "" -- git_q diff --quiet

finish
