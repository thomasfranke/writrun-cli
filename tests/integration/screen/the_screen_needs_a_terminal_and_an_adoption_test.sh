#!/usr/bin/env bash
. "$(dirname "$0")/../../list_lib.sh"

# `writrun` with no command opens the queue as a screen, and prints what
# --help prints in the two cases where a screen cannot exist: no
# terminal, and no adopted repository (docs/product/screen.md,
# spec-0020). The suite has no terminal, so both cases it can reach are
# the fallbacks — and the fallback is the rule's own answer, not a
# degraded one.
make_repo
task task-0001 ready medium "[]" "A thing to do"
cd "$TARGET" || exit 1

check "no terminal prints what --help prints" 0 "the porcelain for WritRun" \
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
check "outside an adoption it prints the help too" 0 "the porcelain for WritRun" \
  -- "$WRITRUN"

finish
