#!/usr/bin/env bash
. "$(dirname "$0")/../../author_lib.sh"

# The composition is shown, the question is asked, and a no leaves
# nothing behind — no branch, no push, no pull request
# (product/pull-requests/shape.md). WRITRUN_TTY_IN stands in for the
# terminal.
make_repo
authoring_change
printf 'n' > "$WORK/keys"
export WRITRUN_TTY_IN="$WORK/keys"
cd "$TARGET" || exit 1

check "the composition is shown before the question" 1 "branch: docs/derived-work" \
  -- "$WRITRUN" author --title "$TITLE"

unset WRITRUN_TTY_IN
check "no branch reached the origin" 1 "" \
  -- git -C "$ORIGIN" rev-parse --verify --quiet refs/heads/docs/derived-work
check "no pull request was opened" 1 "" \
  -- grep -q "pr create" "$GH_LOG"
check "the queue is untouched" 0 "" \
  -- git diff --quiet -- work

finish
