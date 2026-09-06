#!/usr/bin/env bash
. "$(dirname "$0")/../../amend_lib.sh"

# The ordinary pre-implementation amendment: nothing references the spec
# from flight, so nothing is suspended and no reference is owed — the
# whole point of that flow is that it costs nothing (spec-0011, edge
# cases).
make_repo
task_file 0012 amend-command ready spec-0011 "Amend a spec"
git_q -C "$TARGET" add -A
git_q -C "$TARGET" commit -q -m "the task rests"
git_q -C "$TARGET" push -q origin main
in_flight_pr
cd "$TARGET" || exit 1

check "the amendment is opened" 0 "Amended spec-0011" \
  -- amend_cmd spec-0011 --title "$TITLE" --yes

check "the composition said nothing is suspended" 0 "" \
  -- grep -q "suspends   nothing" "$AMEND_OUT"
check "the body claims no suspension" 1 "" \
  -- grep -q "Suspends #" "$GH_LOG"
check "the body says so in as many words" 0 "" \
  -- grep -qF "suspends nothing" "$GH_LOG"

finish
