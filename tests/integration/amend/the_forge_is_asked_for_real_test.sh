#!/usr/bin/env bash
. "$(dirname "$0")/../../amend_lib.sh"

# Every other case supplies the open pull requests through the kit's
# WRITRUN_PR_LIST seam, which short-circuits before `gh` is reached — so
# the argument list the command actually sends the forge was exercised
# nowhere, and a wrong `--json` or `--jq` would have passed the whole
# suite. Here the seam is unset and the stub answers `pr list` for real.
make_repo
unset WRITRUN_PR_LIST
# The four tab-separated fields the command asks `gh` to print.
export GH_PR_LIST="$(printf '42\ttask/0012-amend-command\tsomeone\t[TASK-0012] Amend a spec\n')"
cd "$TARGET" || exit 1

check "the amendment is opened without the seam" 0 "Amended spec-0011" \
  -- amend_cmd spec-0011 --title "$TITLE" --yes

check "the forge really was listed" 0 "" \
  -- grep -q "^pr list" "$GH_LOG"
check "the open state is asked for" 0 "" \
  -- grep -qx -- "--state" "$GH_ARGS"
check "the fields are the ones the composition reads" 0 "" \
  -- grep -qx -- "number,headRefName,author,title" "$GH_ARGS"
check "the jq prints them tab-separated" 0 "" \
  -- grep -qF -- '\(.number)\t\(.headRefName)\t\(.author.login)\t\(.title)' "$GH_ARGS"
check "the limit is the one the kit's own check settled on" 0 "" \
  -- grep -qx -- "200" "$GH_ARGS"

# And the answer was read: #42 is the pull request the gate asks for.
check "the number came out of the forge's own answer" 0 "" \
  -- grep -qF "Suspends #42 — task-0012 waits on this amendment." "$GH_LOG"
check "no hand-checking was announced" 1 "" \
  -- grep -q "check by hand" "$AMEND_OUT"

finish
