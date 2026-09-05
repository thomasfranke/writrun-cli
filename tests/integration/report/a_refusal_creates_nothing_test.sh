#!/usr/bin/env bash
. "$(dirname "$0")/../../report_lib.sh"

# The generator refuses in its own words and with its own exit code,
# and the wrapper adds nothing to either. Triage is not this command's
# to set, so the flags that would set one are not flags it has
# (docs/product/reports/report.md).
make_repo
cd "$TARGET" || exit 1

check "a slug outside the filename contract is refused verbatim" 3 "outside the filename contract" \
  -- report "$TITLE" --body "$BODY" --slug "Not A Slug" --yes
check "the refusal created nothing" 0 "1" \
  -- sh -c 'ls work/reports | wc -l | tr -d " "'

check "a route is not a flag this command has" 1 "flag provided but not defined" \
  -- report "$TITLE" --body "$BODY" --tracked --yes
check "a priority is not a flag this command has" 1 "flag provided but not defined" \
  -- report "$TITLE" --body "$BODY" --priority=high --yes
check "nothing a refusal touched reached the queue" 0 "1" \
  -- sh -c 'ls work/reports | wc -l | tr -d " "'

finish
