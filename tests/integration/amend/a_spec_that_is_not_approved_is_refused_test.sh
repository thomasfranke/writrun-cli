#!/usr/bin/env bash
. "$(dirname "$0")/../../amend_lib.sh"

# Step 1 of spec-0011: amend returns an *approved* spec to draft, so a
# draft one is refused — and the refusal is the whole command. Nothing
# is asked, nothing reaches the forge, nothing is written.
make_repo
in_flight_pr
cd "$TARGET" || exit 1

check "the draft spec is refused, naming what it found" 1 "spec-0013 is 'draft'" \
  -- amend_cmd spec-0013 --title "$TITLE" --yes

check "the spec still reads draft, untouched" 0 "draft" \
  -- field status work/specs/spec-0013-another-thing.md
check "nothing reached the forge" 1 "" \
  -- grep -q "pr create" "$GH_LOG"
check "no branch was cut" 1 "" \
  -- git rev-parse --verify --quiet refs/heads/docs/another-thing
check "the working tree is untouched" 0 "" \
  -- test -z "$(git status --porcelain)"

finish
