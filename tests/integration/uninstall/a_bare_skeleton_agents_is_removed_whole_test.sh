#!/usr/bin/env bash
. "$(dirname "$0")/../../init_lib.sh"

# An AGENTS.md that is nothing but the kit's own skeleton goes with the
# kit, named in the shown set (spec-0005, edge cases).
make_target "$TARGET"
cd "$TARGET" || exit 1
"$WRITRUN" init --stage 1 --yes > /dev/null 2>&1 || { echo "FAIL  the fixture could not adopt"; exit 1; }
# The skeleton init wrote, reduced to the section alone.
awk '/^## WritRun/,0' AGENTS.md > AGENTS.md.tmp && mv AGENTS.md.tmp AGENTS.md

check "uninstall names it for removal" 0 "nothing in it but the kit" \
  -- "$WRITRUN" uninstall --yes
check "AGENTS.md is gone" 1 "" -- test -e AGENTS.md

finish
