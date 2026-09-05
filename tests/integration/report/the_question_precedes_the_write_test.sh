#!/usr/bin/env bash
. "$(dirname "$0")/../../report_lib.sh"

# The write is a change to the repository, so it is shown and confirmed
# first, and every question has a flag that answers it
# (docs/product/rules.md). WRITRUN_TTY_IN stands in for the terminal.
make_repo
cd "$TARGET" || exit 1

printf 'n' > "$WORK/keys"
export WRITRUN_TTY_IN="$WORK/keys"

check "a declined confirmation exits non-zero saying so" 1 "declined" \
  -- report "$TITLE" --body "$BODY" --slug a-finding
check "the decline recorded nothing" 1 "" -- test -f "$RECORDED"
check "the question showed what would be recorded" 0 "" \
  -- grep -qF "$TITLE" "$REPORT_OUT"

unset WRITRUN_TTY_IN
check "without a terminal the confirmation names its flag" 1 "\-\-yes" \
  -- report "$TITLE" --body "$BODY" --slug a-finding
check "without a terminal the observation names its flag" 1 "\-\-body" \
  -- report "$TITLE" --yes
check "without a terminal the title is required" 1 "the title as an argument" \
  -- report --body "$BODY" --yes
check "no question left a file behind" 1 "" -- test -f "$RECORDED"

finish
