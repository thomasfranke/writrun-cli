#!/usr/bin/env bash
. "$(dirname "$0")/../../report_lib.sh"

# Recording is exempt from the one-kind-per-change rule: the file is
# written wherever the reporter is, so no branch is cut, nothing is
# committed, and nothing reaches the forge
# (docs/product/reports/report.md).
make_repo
cd "$TARGET" || exit 1

check "the observation is recorded" 0 "Created $RECORDED" \
  -- report "$TITLE" --body "$BODY" --slug a-finding --yes

check "the branch is the one the reporter was on" 0 "main" \
  -- git rev-parse --abbrev-ref HEAD
check "no second branch was cut" 0 "1" \
  -- sh -c 'git branch --list | wc -l | tr -d " "'
check "the origin holds main and nothing else" 0 "1" \
  -- sh -c 'git -C "$ORIGIN" for-each-ref --format="%(refname)" refs/heads | wc -l | tr -d " "'
check "no pull request was opened" 1 "" -- grep -q "pr create" "$GH_LOG"
check "nothing was committed" 0 "the kit and the queue" \
  -- git log -1 --pretty=%s
check "the file is there to be committed by the change already open" 0 "?? $RECORDED" \
  -- git status --porcelain

finish
