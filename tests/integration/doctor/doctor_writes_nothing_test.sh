#!/usr/bin/env bash
. "$(dirname "$0")/../../doctor_lib.sh"

# Reports only: no run writes anything, and there is no flag that
# repairs. A run that found something is the one to check — a command
# that fixed on its way would fix precisely there (spec-0004,
# acceptance criteria).
make_repo 3
forge_healthy
cd "$TARGET" || exit 1

"$WRITRUN" doctor > /dev/null 2>&1
clean=$(git_q status --porcelain)
check "a clean run touches nothing" 0 "" -- test -z "$clean"

rm -f "$TARGET/docs/about.md"
git_q add -A
git_q commit -q -m "the About file goes"
"$WRITRUN" doctor > /dev/null 2>&1
dirty=$(git_q status --porcelain)
check "a run with findings touches nothing" 0 "" -- test -z "$dirty"

check "there is no repair flag" 1 "not defined: -fix" -- "$WRITRUN" doctor --fix
check "there is no stage argument" 1 "unexpected argument" -- "$WRITRUN" doctor 2

finish
