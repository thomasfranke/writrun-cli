#!/usr/bin/env bash
. "$(dirname "$0")/../../author_lib.sh"

# The green path over the repository's own scripts: the three checks run
# in their fixed order, the composition is shown, and the pull request
# opens **ready** — an authoring pull request has no work to announce
# (spec-0009, acceptance criteria).
make_repo
authoring_change
cd "$TARGET" || exit 1

check "the run reports the pull request it opened" 0 "pull request open and ready for review" \
  -- author --title "$TITLE" --yes

check "the front-matter check ran" 0 "" \
  -- grep -q "all canonical" "$AUTHOR_OUT"
check "the doc-shapes check ran" 0 "" \
  -- grep -q "shown shape(s)" "$AUTHOR_OUT"
check "the state check ran over the range" 0 "" \
  -- grep -q "no forbidden lifecycle transition in origin/main...HEAD" "$AUTHOR_OUT"
check "the composition was shown before the act" 0 "" \
  -- grep -q "branch: docs/derived-work" "$AUTHOR_OUT"
check "the branch reached the origin" 0 "" \
  -- git -C "$ORIGIN" rev-parse --verify --quiet refs/heads/docs/derived-work
check "the pull request was opened" 0 "" \
  -- grep -q -- "pr create" "$GH_LOG"
check "it was not opened as a draft" 1 "" \
  -- grep -q -- "--draft" "$GH_LOG"
check "the title carries no task tag" 1 "" \
  -- grep -q -- "TASK-" "$GH_LOG"
check "the title is the one that was given" 0 "" \
  -- grep -qF -- "--title $TITLE" "$GH_LOG"

finish
