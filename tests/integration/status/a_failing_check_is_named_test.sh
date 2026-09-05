#!/usr/bin/env bash
. "$(dirname "$0")/../../status_lib.sh"

# The first check that fails is named in the script's own words — the
# stage `finish` would stop at — and naming it is an answer, not a
# failure of status (spec-0013, acceptance criteria).
make_repo
break_the_first_check
cd "$TARGET" || exit 1

"$WRITRUN" status > "$WORK/failing.out" 2>&1
code=$?

check "the failing stage is named" 0 "Checks   PREFLIGHT STOPPED at 1/3 front matter" -- cat "$WORK/failing.out"
check "its failure is named" 0 "exit 1" -- grep "^Checks" "$WORK/failing.out"
check "a failing check is still an answer" 0 "" -- test "$code" -eq 0
check "the rest of the answer survives it" 0 "Reports  none open" -- cat "$WORK/failing.out"

finish
