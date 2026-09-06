#!/usr/bin/env bash
. "$(dirname "$0")/../../queue_lib.sh"

# One reader, so one answer: every command that reads the queue's front
# matter is run over every state a queue file can be in and every id
# form a person types, and each cell is held by its exit code, what it
# left under work/, whether it reached the fake forge or the bare
# origin, and what it said (spec-0022, tests required).
#
# The golden is the enumeration in machine form. A cell that moves is a
# behaviour change, and this diff is where it has to be argued for.

OUT="$WORK/matrix.txt"
GOLDEN="$(dirname "$0")/matrix.golden"

matrix "$OUT"

check "every cell answers as the enumeration says it does" 0 "" \
  -- diff -u "$GOLDEN" "$OUT"

finish
