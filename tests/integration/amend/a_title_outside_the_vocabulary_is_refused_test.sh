#!/usr/bin/env bash
. "$(dirname "$0")/../../amend_lib.sh"

# `--type wibble` composes the title `[Wibble][Specs] …`, which the
# project's own check_observance.sh refuses. The refusal has to arrive
# here — before the branch is cut, and long before the forge sees it
# (spec-0023; report-0022).
make_repo
in_flight_pr
cd "$TARGET" || exit 1

check "the door refuses the composed title in its own words" 1 "outside the vocabulary" \
  -- amend_cmd spec-0011 --type wibble --title "$TITLE" --yes

check "the working tree carries nothing" 0 "" \
  -- test -z "$(git status --porcelain)"
check "the spec still reads approved" 0 "approved" \
  -- field status work/specs/spec-0011-amend-command.md
check "no branch was cut" 1 "" \
  -- git rev-parse --verify --quiet refs/heads/wibble/amend-command
check "nothing reached the origin" 1 "" \
  -- git -C "$ORIGIN" rev-parse --verify --quiet refs/heads/wibble/amend-command
check "no pull request was opened" 1 "" \
  -- grep -q "pr create" "$GH_LOG"

# And a type the vocabulary declares passes the same door, so the
# refusal above is the vocabulary's doing and not the check's silence.
check "a declared type is opened" 0 "Amended spec-0011" \
  -- amend_cmd spec-0011 --title "$TITLE" --yes

finish
