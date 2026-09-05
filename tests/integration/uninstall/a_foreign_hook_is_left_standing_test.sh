#!/usr/bin/env bash
. "$(dirname "$0")/../../init_lib.sh"

# A hook that is not the one init writes is another project's to
# remove: left standing, and said so (spec-0005).
make_target "$TARGET"
cd "$TARGET" || exit 1
"$WRITRUN" init --stage 1 --yes > /dev/null 2>&1 || { echo "FAIL  the fixture could not adopt"; exit 1; }

hook_at=$(git_q rev-parse --git-path hooks/commit-msg)
printf '#!/bin/sh\n# somebody else wrote this\nexit 0\n' > "$hook_at"

check "uninstall names the hook it will not touch" 0 "not the one init writes" \
  -- "$WRITRUN" uninstall --yes
check "the foreign hook is still there" 0 "somebody else wrote this" -- cat "$hook_at"
check ".writrun/ is gone all the same" 1 "" -- test -e .writrun

finish
