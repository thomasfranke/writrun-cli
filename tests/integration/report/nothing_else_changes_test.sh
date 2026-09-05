#!/usr/bin/env bash
. "$(dirname "$0")/../../report_lib.sh"

# A recorded report changes nothing else: the queue, the statuses and
# every file already tracked are exactly as they were
# (docs/product/reports/report.md).
make_repo
cd "$TARGET" || exit 1

check "the observation is recorded" 0 "" \
  -- report "$TITLE" --body "$BODY" --slug a-finding --yes

check "no tracked file was touched" 0 "" -- git diff --quiet
check "the task's status is untouched" 0 "" \
  -- grep -qx "status: ready" work/tasks/task-0001-a-thing.md
check "the earlier report was not triaged" 0 "" \
  -- grep -qx "status: open" work/reports/report-0001-an-earlier-finding.md
check "the new file is the only thing that appeared" 0 "1" \
  -- sh -c 'git status --porcelain | wc -l | tr -d " "'

finish
