#!/usr/bin/env bash
. "$(dirname "$0")/../../doctor_lib.sh"

# A forge that will not answer has told doctor nothing about the
# repository: the checks it could not make are reported as unread and
# none of them fails the run (spec-0004, acceptance criteria).
make_repo 3
forge_offline
cd "$TARGET" || exit 1

"$WRITRUN" doctor > "$WORK/out" 2>&1

check "an unreachable forge is not a failure" 0 "" -- "$WRITRUN" doctor
check "the cause is named" 0 "unread   gh is not authenticated" -- cat "$WORK/out"
check "stage 3 says what it could not check" 0 \
  "unread   whether Issues are enabled was not read" -- cat "$WORK/out"
check "nothing was made to break" 0 "none breaking a flow" -- cat "$WORK/out"
check "no read piled onto the one fault" 1 "" \
  -- grep -q "allow_squash_merge" "$GH_DIR/calls"

finish
