#!/usr/bin/env bash
. "$(dirname "$0")/../../work_lib.sh"

# work selects and launches. It writes nothing to the queue, cuts no
# branch and opens nothing on the forge — the launched agent takes the
# task, exactly as a human would (spec-0007, scope).
make_repo
configure_agent
cd "$TARGET" || exit 1
before=$(git_q rev-parse HEAD)

work > /dev/null 2>&1
work task-0002 > /dev/null 2>&1
work task-0003 > /dev/null 2>&1
work task-0042 > /dev/null 2>&1

dirty=$(git_q status --porcelain)
check "the working tree is untouched" 0 "" -- test -z "$dirty"
check "no commit was made" 0 "" -- test "$(git_q rev-parse HEAD)" = "$before"
check "no branch was cut" 0 "" -- test "$(git_q for-each-ref --format='%(refname)' refs/heads)" = "refs/heads/main"
check "no pull request was opened" 1 "" -- grep -q "pr create" "$GH_LOG"

finish
