#!/usr/bin/env bash
. "$(dirname "$0")/../../init_lib.sh"

# A decline at the confirmation leaves the repository untouched
# (spec-0002, acceptance criteria). WRITRUN_TTY_IN stands in for the
# terminal: the one byte answers the confirm with "no".
make_target "$TARGET"
printf 'n' > "$WORK/keys"
export WRITRUN_TTY_IN="$WORK/keys"
cd "$TARGET" || exit 1

check "a declined confirmation exits non-zero saying so" 1 "declined" \
  -- "$WRITRUN" init --stage 1

unset WRITRUN_TTY_IN
check "no .writrun/ was written" 1 "" \
  -- test -d .writrun
check "no hook was installed" 1 "" \
  -- test -f .git/hooks/commit-msg
check "no AGENTS.md was written" 1 "" \
  -- test -f AGENTS.md

finish
