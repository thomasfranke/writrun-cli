#!/usr/bin/env bash
. "$(dirname "$0")/../../finish_lib.sh"

# The forge has no pull request for this branch: its own words reach the
# user, and nothing is marked ready. The checks before it ran, so the
# reporting says exactly how far the sequence got.
make_repo
on_branch task/0001-a-thing
export GH_PR_VIEW_FAILS=1
cd "$TARGET" || exit 1

check "the forge's own words reach the user" 1 "no pull requests found for branch" \
  -- finish_cmd --yes

check "preflight had already passed" 0 "" \
  -- grep -q "PREFLIGHT OK" "$FINISH_OUT"
check "nothing was marked ready" 1 "" \
  -- grep -q "pr ready" "$GH_LOG"

finish
