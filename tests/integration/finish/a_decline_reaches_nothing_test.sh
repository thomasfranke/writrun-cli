#!/usr/bin/env bash
. "$(dirname "$0")/../../finish_lib.sh"

# The composition is shown, the question is asked, and a no leaves the
# forge untouched — the pull request stays a draft (spec-0010, step 5;
# product/pull-requests/shape.md). WRITRUN_TTY_IN stands in for the
# terminal.
make_repo
on_branch task/0001-a-thing
printf 'n' > "$WORK/keys"
export WRITRUN_TTY_IN="$WORK/keys"
cd "$TARGET" || exit 1

check "the composition is shown before the question" 1 "pull request   #7" \
  -- finish_cmd

unset WRITRUN_TTY_IN
check "the decline is reported as one" 0 "" \
  -- grep -q "declined — nothing changed" "$FINISH_OUT"
check "the pull request was not marked ready" 1 "" \
  -- grep -q "pr ready" "$GH_LOG"
check "the pull request was read, and only read" 0 "" \
  -- grep -q "pr view" "$GH_LOG"

finish
