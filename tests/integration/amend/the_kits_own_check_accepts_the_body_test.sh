#!/usr/bin/env bash
. "$(dirname "$0")/../../amend_lib.sh"

# spec-0011's Definition of Done, read literally: the pull request this
# command opens passes `check_amendment_reference.sh`. The composed body
# is handed to the kit's own gate, over the change the command made, with
# the forge answering the way it answers in CI — three tab-separated
# fields, which is what that script asks `gh` for.
make_repo
in_flight_pr
cd "$TARGET" || exit 1

check "the amendment is opened" 0 "Amended spec-0011" \
  -- amend_cmd spec-0011 --title "$TITLE" --yes

CHECK=".writrun/scripts/stage-2-pull-requests/check_amendment_reference.sh"
export GH_PR_LIST="$(printf '42\ttask/0012-amend-command\t[TASK-0012] Amend a spec\n')"
unset WRITRUN_PR_LIST

gate() { PR_BODY="$(cat "$GH_BODY")" bash "$CHECK" main...HEAD owner/repo 99; }
check "the kit's own gate accepts the composed body" 0 "#42 is named" \
  -- gate

# And the same gate over the same change with an empty body fails, so
# the case above is the body's doing and not the gate's silence.
gate_empty() { PR_BODY="" bash "$CHECK" main...HEAD owner/repo 99; }
check "a body naming nothing is rejected by the same gate" 1 "Suspends #42" \
  -- gate_empty

finish
