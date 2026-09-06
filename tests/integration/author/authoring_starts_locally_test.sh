#!/usr/bin/env bash
. "$(dirname "$0")/../../author_lib.sh"

# A diff already on a pushed branch is refused: authoring starts
# locally, and the one carve-out is --resume, which finishes an act
# whose pull request never opened (spec-0009, edge cases).
make_repo
authoring_change
git_q -C "$TARGET" push -q -u origin docs/derived-work
cd "$TARGET" || exit 1

check "a pushed branch is refused" 1 "authoring starts locally" \
  -- author --title "$TITLE" --yes
check "no check ran" 1 "" \
  -- grep -q "all canonical" "$AUTHOR_OUT"
check "the forge was never reached" 1 "" \
  -- grep -q . "$GH_LOG"

check "--resume finishes the act instead" 0 "pull request open and ready for review" \
  -- author --title "$TITLE" --resume --yes
check "the pull request was opened" 0 "" \
  -- grep -q -- "pr create" "$GH_LOG"
check "it was not opened as a draft" 1 "" \
  -- grep -q -- "--draft" "$GH_LOG"

finish
