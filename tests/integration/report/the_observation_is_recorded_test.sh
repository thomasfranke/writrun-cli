#!/usr/bin/env bash
. "$(dirname "$0")/../../report_lib.sh"

# The generator mints the file — the id in sequence, the status born
# open, nothing triaged — and the reporter's paragraph stands where its
# placeholder stood (docs/product/reports/report.md).
make_repo
cd "$TARGET" || exit 1

check "the generator's own report is the command's" 0 "Created $RECORDED" \
  -- report "$TITLE" --body "$BODY" --slug a-finding --doc-ref technical/testing/suites.md --yes

check "the file the generator named is there" 0 "" -- test -f "$RECORDED"
check "the id continues the sequence" 0 "" -- grep -qx "id: report-0002" "$RECORDED"
check "the status is born open" 0 "" -- grep -qx "status: open" "$RECORDED"
check "nothing is triaged" 0 "" -- grep -qx "triaged: null" "$RECORDED"
check "no task claims it" 0 "" -- grep -qx "task_ref: \[\]" "$RECORDED"
check "the doc-ref reached the front matter" 0 "" \
  -- grep -qx "doc_ref: technical/testing/suites.md" "$RECORDED"
check "the title is the heading" 0 "" -- grep -qxF "# $TITLE" "$RECORDED"
check "the observation is the body" 0 "" -- grep -qF "$BODY" "$RECORDED"
check "the placeholder did not survive" 1 "" -- grep -q "TODO" "$RECORDED"

# The slug is the filename's subject and never part of the id; omitted,
# the generator derives one from the title.
check "an omitted slug is the generator's to derive" 0 "Created work/reports/report-0003" \
  -- report "Another finding entirely" --body "$BODY" --yes

finish
