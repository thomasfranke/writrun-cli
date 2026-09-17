#!/usr/bin/env bash
. "$(dirname "$0")/../../finish_lib.sh"

# Step 1 judges the tree, so a promise kept and not yet committed is
# read rather than refused.
#
# This is the twin of `a_missing_delta_stops_before_any_write`: same
# branch, same spec, same promised document — the difference is that
# here the author has written it and not committed it, which is the
# state a person is in when they run `finish`. On the two-ended range
# the edit is invisible and `check_deltas.sh` answers MISSING, refusing
# a promise that was in fact kept. The range is bare, so its diff
# reaches the working tree (spec-0045, report-0014).
make_repo
on_branch task/0002-another-thing
cd "$TARGET" || exit 1

# spec-0002 promises product/promised.md. Write it and stage it, which
# is what `git add` of a new chapter leaves behind — `git diff <ref>`
# reads the index and the tree for a tracked path and is blind to an
# untracked one, so staging is what makes a new file part of the change
# rather than part of the debris.
printf '# Promised\n\nThe chapter the completion owed.\n' > docs/product/promised.md
git_q -C "$TARGET" add docs/product/promised.md

check "the promised document is uncommitted" 1 "" -- tree_is_clean
check "the finish reads it and runs through" 0 "PREFLIGHT OK" -- finish_cmd --yes
check "the delta check judged the promise" 0 "" \
  -- grep -q "OK: diff matches the promised deltas of spec-0002" "$FINISH_OUT"
check "nothing was reported missing" 1 "" -- grep -q "MISSING" "$FINISH_OUT"
check "the spec reached implemented" 0 "implemented" \
  -- field status work/specs/spec-0002-another-thing.md

finish
