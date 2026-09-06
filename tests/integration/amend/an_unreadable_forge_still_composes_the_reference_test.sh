#!/usr/bin/env bash
. "$(dirname "$0")/../../amend_lib.sh"

# Best-effort on the forge half, the same contract
# check_amendment_reference.sh states: without an answer the pull
# request's number cannot be known, so the command says the reference
# must be checked by hand and still composes it from the queue
# (spec-0011, acceptance criteria).
make_repo
export GH_PR_LIST_EXIT=1
cd "$TARGET" || exit 1

check "the amendment is still opened" 0 "Amended spec-0011" \
  -- amend_cmd spec-0011 --title "$TITLE" --yes

check "the narrow view is named" 0 "" \
  -- grep -q "check by hand" "$AMEND_OUT"
check "the reference is composed from the queue" 0 "" \
  -- grep -qF "Suspends the open pull request working task-0012" "$GH_LOG"
check "no number was invented" 1 "" \
  -- grep -q "Suspends #" "$GH_LOG"

finish
