#!/usr/bin/env bash
. "$(dirname "$0")/../../amend_lib.sh"

# The composition is shown, the question is asked, and a no leaves the
# repository exactly as it was: no half-written status, no orphan
# branch, nothing on the forge
# (docs/product/pull-requests/shape.md; report-0015).
#
# This is where amend differs from finish deliberately: its one queue
# edit lands with the push rather than before the question, which it can
# because no later step reads it. WRITRUN_TTY_IN stands in for the
# terminal.
make_repo
in_flight_pr
printf 'n' > "$WORK/keys"
export WRITRUN_TTY_IN="$WORK/keys"
cd "$TARGET" || exit 1

check "the composition is shown before the question" 1 "branch     docs/amend-command" \
  -- amend_cmd spec-0011 --title "$TITLE"

unset WRITRUN_TTY_IN
check "the decline is reported as one" 0 "" \
  -- grep -q "declined — nothing changed" "$AMEND_OUT"
check "the working tree carries nothing" 0 "" \
  -- test -z "$(git status --porcelain)"
check "the spec still reads approved" 0 "approved" \
  -- field status work/specs/spec-0011-amend-command.md
check "no branch was cut" 1 "" \
  -- git rev-parse --verify --quiet refs/heads/docs/amend-command
check "nothing reached the origin" 1 "" \
  -- git -C "$ORIGIN" rev-parse --verify --quiet refs/heads/docs/amend-command
check "no pull request was opened" 1 "" \
  -- grep -q "pr create" "$GH_LOG"

finish
