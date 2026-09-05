#!/usr/bin/env bash
. "$(dirname "$0")/../../status_lib.sh"

# Reads only: nothing on the repository changes, on a task branch or off
# one, whether the checks pass or stop (spec-0013, acceptance criteria).
make_repo
report report-0001 open
cd "$TARGET" || exit 1
git_q add -A
git_q commit -q -m "the queue"
before=$(git_q rev-parse HEAD)

"$WRITRUN" status > /dev/null 2>&1
on_branch main
"$WRITRUN" status > /dev/null 2>&1
on_branch task/0014-status-command
break_the_first_check
"$WRITRUN" status > /dev/null 2>&1
git_q checkout -q -- work/tasks

dirty=$(git_q status --porcelain)
check "the working tree is untouched" 0 "" -- test -z "$dirty"
check "no commit was made" 0 "" -- test "$(git_q rev-parse HEAD)" = "$before"

finish
