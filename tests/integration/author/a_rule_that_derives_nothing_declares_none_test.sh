#!/usr/bin/env bash
. "$(dirname "$0")/../../author_lib.sh"

# A rule that derives no work says so under the contract marker: an
# empty declaration and a forgotten one look identical, which is what
# the kit's own door refuses (spec-0009, acceptance criteria).
make_repo
on_branch docs/derives-nothing
rule_written
committed "docs(product): the declaration is the section"
cd "$TARGET" || exit 1

check "the run opened the pull request" 0 "" \
  -- author --title "$TITLE" --yes

check "the body declares none" 0 "" \
  -- grep -q "none" "$GH_BODY"
check "the table's header is not there to be filled" 1 "" \
  -- grep -q "What it implements" "$GH_BODY"
check "the opened body passes check_derived_work.sh" 0 "declared as none" \
  -- env PR_BODY="$(cat "$GH_BODY")" \
     bash .writrun/scripts/stage-2-pull-requests/check_derived_work.sh origin/main...HEAD

# The same door, asked about a body this command did not write: the
# declaration is what it reads, so its absence is a refusal — which is
# what makes the line above an assertion and not a coincidence.
check "a body without the declaration is refused by that same door" 1 "" \
  -- env PR_BODY="## What" \
     bash .writrun/scripts/stage-2-pull-requests/check_derived_work.sh origin/main...HEAD

finish
