#!/usr/bin/env bash
. "$(dirname "$0")/../../list_lib.sh"

# Reads only: nothing about the queue changes (spec-0006, acceptance
# criteria).
make_repo
task task-0001 ready   high   spec-0001 "A task that may be taken"
task task-0002 backlog medium spec-0002 "A task whose spec is not approved"
spec spec-0001 approved
spec spec-0002 draft
report report-0001 open "Something that was noticed"
forge_online
cd "$TARGET" || exit 1
git_q add -A
git_q commit -q -m "the queue"

"$WRITRUN" list > /dev/null 2>&1
"$WRITRUN" list --available > /dev/null 2>&1
dirty=$(git_q status --porcelain)

check "the working tree is untouched" 0 "" -- test -z "$dirty"

finish
