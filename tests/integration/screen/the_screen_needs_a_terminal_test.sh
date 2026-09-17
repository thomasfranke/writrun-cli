#!/usr/bin/env bash
. "$(dirname "$0")/../../list_lib.sh"

# `writrun` with no command opens a screen, and prints what --help
# prints in the one case where a screen cannot exist: no terminal
# (docs/product/screens/README.md, spec-0020, spec-0038). Which screen
# it opens is the kit's answer — the entry screen inside an adoption,
# the first run outside one — and both need a terminal at either end.
#
# The suite has no terminal, so what it reaches here is the fallback,
# and the fallback is the rule's own answer rather than a degraded one.
# The screens themselves are driven through a pty in tests/e2e/screen/.
make_repo
task task-0001 ready medium "[]" "A thing to do"
cd "$TARGET" || exit 1

check "no terminal prints what --help prints" 0 "run a project by WritRun" \
  -- "$WRITRUN"

check "and it names the commands, as the help does" 0 "take " \
  -- "$WRITRUN"

# Nothing is written: the screen reads, and every change goes through
# the command it dispatches. The fixture's own queue file is untracked
# before the run, so what is asserted is that the run changed nothing —
# not that the tree was clean to begin with.
BEFORE=$(git_q status --porcelain)
"$WRITRUN" > /dev/null 2>&1
AFTER=$(git_q status --porcelain)
check "the run changed nothing in the working tree" 0 "" \
  -- test "$BEFORE" = "$AFTER"

cd "$WORK" || exit 1
mkdir -p bare && cd bare || exit 1
check "outside an adoption, with no terminal, it prints the help too" 0 "run a project by WritRun" \
  -- "$WRITRUN"

# And it opens nothing there: the first-run screen needs a terminal at
# both ends like every other, so a run with neither says what --help
# says and introduces nothing.
# grep exits 1 when it finds nothing, which is the answer wanted here.
check "and it draws none of the first run's screen" 1 "" \
  -- sh -c '"$0" | grep -q "NO KIT HERE"' "$WRITRUN"

finish
