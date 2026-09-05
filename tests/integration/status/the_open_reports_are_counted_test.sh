#!/usr/bin/env bash
. "$(dirname "$0")/../../status_lib.sh"

# Step 4 counts the reports awaiting triage and nothing else: a report
# triage already ended is not waiting, and the README beside them is not
# a report (spec-0013, steps).
make_repo
cd "$TARGET" || exit 1

check "an empty queue of reports is said" 0 "Reports  none open" -- "$WRITRUN" status

report report-0001 open
report report-0002 open
report report-0003 declined
report report-0004 tracked
check "only the open ones are counted" 0 "Reports  2 open, waiting to be triaged" \
  -- "$WRITRUN" status

finish
