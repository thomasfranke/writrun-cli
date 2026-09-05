#!/usr/bin/env bash
. "$(dirname "$0")/../../list_lib.sh"

# A filter chooses sections and nothing else: the tasks it shows keep
# the group and the order the unfiltered run gave them (spec-0006).
make_repo
task task-0001 ready   high   spec-0001 "A task that may be taken"
task task-0002 backlog medium spec-0002 "A task whose spec is not approved"
spec spec-0001 approved
spec spec-0002 draft
report report-0001 open "Something that was noticed"
forge_online
cd "$TARGET" || exit 1

"$WRITRUN" list --available > "$WORK/available.out" 2>&1
"$WRITRUN" list --held      > "$WORK/held.out" 2>&1
"$WRITRUN" list --reports   > "$WORK/reports.out" 2>&1

check "--available keeps the available section" 0 "task-0001" \
  -- grep "Available — any of these may be taken:" -A2 "$WORK/available.out"
check "--available drops the held-back section" 1 "" \
  -- grep -q "Held back:" "$WORK/available.out"
check "--available drops the reports section" 1 "" \
  -- grep -q "report-0001" "$WORK/available.out"

check "--held keeps the held-back section" 0 "task-0002" \
  -- grep "Held back:" -A2 "$WORK/held.out"
check "--held drops the available section" 1 "" \
  -- grep -q "Available —" "$WORK/held.out"

check "--reports keeps the open reports" 0 "report-0001" \
  -- grep "Open reports" -A2 "$WORK/reports.out"
check "--reports drops the available section" 1 "" \
  -- grep -q "task-0001" "$WORK/reports.out"

check "a filter does not change the exit code" 0 "" -- "$WRITRUN" list --held

finish
