#!/usr/bin/env bash
. "$(dirname "$0")/../../report_lib.sh"

# Free text is typed, not navigated (docs/product/rules.md). One form
# reads the keys per run, the way every case driving the pseudo-terminal
# does: --yes answers the confirmation, and whichever of the two texts
# was given as a flag is not asked.
make_repo
cd "$TARGET" || exit 1

printf 'Typed at the prompt.\r' > "$WORK/keys"
export WRITRUN_TTY_IN="$WORK/keys"
check "the typed observation is recorded" 0 "Created work/reports/report-0002" \
  -- report "A typed finding" --slug a-finding --yes
check "the typed observation is the body" 0 "" \
  -- grep -qF "Typed at the prompt." "$RECORDED"
check "the placeholder did not survive" 1 "" -- grep -q "TODO" "$RECORDED"

printf 'A second typed finding\r' > "$WORK/keys"
check "the typed title is recorded" 0 "Created work/reports/report-0003" \
  -- report --body "$BODY" --slug another-finding --yes
check "the typed title is the heading" 0 "" \
  -- grep -qxF "# A second typed finding" work/reports/report-0003-another-finding.md

unset WRITRUN_TTY_IN
finish
