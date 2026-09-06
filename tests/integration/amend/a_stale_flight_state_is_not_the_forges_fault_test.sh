#!/usr/bin/env bash
. "$(dirname "$0")/../../amend_lib.sh"

# task-0012 reads as in flight but its pull request is gone — merged,
# closed, or never opened. The forge answers perfectly well; it just
# names nothing working the task. The terminal said so correctly while
# the body said "the forge did not answer" and claimed a suspension that
# does not exist, which the kit's own gate passes over: it reads the
# same state as a stale flight and asks for no reference. Nothing
# downstream would have caught the false sentence.
make_repo
export WRITRUN_PR_LIST="$(printf '77\tdocs/unrelated\tsomeone\t[Docs] Something else')"
cd "$TARGET" || exit 1

check "the amendment is opened" 0 "Amended spec-0011" \
  -- amend_cmd spec-0011 --title "$TITLE" --yes

check "the composition said no pull request works it" 0 "" \
  -- grep -q "no open pull request works it" "$AMEND_OUT"
check "the body does not blame a forge that answered" 1 "" \
  -- grep -q "the forge did not answer" "$GH_LOG"
check "the body claims no suspension" 1 "" \
  -- grep -q "Suspends" "$GH_LOG"
check "the body names the stale flight state instead" 0 "" \
  -- grep -qF "task-0012 reads as in flight, but no open pull request works it" "$GH_LOG"
check "no hand-checking was announced" 1 "" \
  -- grep -q "check by hand" "$AMEND_OUT"

# The gate agrees: nothing is owed, so the body owes no reference.
CHECK=".writrun/scripts/stage-2-pull-requests/check_amendment_reference.sh"
export GH_PR_LIST="$(printf '77\tdocs/unrelated\t[Docs] Something else\n')"
unset WRITRUN_PR_LIST
gate() { PR_BODY="$(cat "$GH_BODY")" bash "$CHECK" main...HEAD owner/repo 99; }
check "the kit's own gate asks for no reference" 0 "flight state is stale" \
  -- gate

finish
