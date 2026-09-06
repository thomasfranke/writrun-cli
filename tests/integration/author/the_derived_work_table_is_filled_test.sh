#!/usr/bin/env bash
. "$(dirname "$0")/../../author_lib.sh"

# The body is the template's authoring half: the Derived-work table
# filled from the tasks and specs the diff adds, the implementing half
# gone, and the whole thing readable by the door it will meet
# (spec-0009, acceptance criteria and Definition of Done).
make_repo
authoring_change
cd "$TARGET" || exit 1

check "the run opened the pull request" 0 "" \
  -- author --title "$TITLE" --yes

check "the table lists the derived task and its spec" 0 "" \
  -- grep -qF "| task-0001 | spec-0001 | Declare the derived work |" "$GH_BODY"
check "the template's placeholder row is gone" 1 "" \
  -- grep -q "task-NNNN" "$GH_BODY"
check "the contract marker heading is kept" 0 "" \
  -- grep -qx "## Derived work" "$GH_BODY"
check "the implementing half is gone" 1 "" \
  -- grep -qx "## Spec" "$GH_BODY"
check "the writrun markers are kept" 0 "" \
  -- grep -q -- "<!-- writrun:end -->" "$GH_BODY"

# The kit's own door, run over exactly what was opened.
check "the opened body passes check_derived_work.sh" 0 "Derived work present" \
  -- env PR_BODY="$(cat "$GH_BODY")" \
     bash .writrun/scripts/stage-2-pull-requests/check_derived_work.sh origin/main...HEAD

finish
