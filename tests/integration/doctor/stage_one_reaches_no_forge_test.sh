#!/usr/bin/env bash
. "$(dirname "$0")/../../doctor_lib.sh"

# A project is never judged against machinery it did not enable. Stage 2
# is previewed and stage 3 is not: the preview reaches one rung, and the
# rung above it says so rather than being left to look like a clean bill
# (spec-0004, edge cases; spec-0036).
make_repo 1
forge_healthy
cd "$TARGET" || exit 1

"$WRITRUN" doctor > "$WORK/out" 2>&1

check "stage 1 exits 0" 0 "" -- "$WRITRUN" doctor
check "the rung above the preview says it is not previewed" 0 \
  "Stage 3 — Issues: not previewed — one rung at a time." -- cat "$WORK/out"
check "no Issues read was made" 1 "" -- grep -q "has_issues" "$GH_DIR/calls"
check "the exit status answers for the declaration alone" 0 \
  "Every assumption up to stage 1 holds." -- cat "$WORK/out"

finish
