#!/usr/bin/env bash
. "$(dirname "$0")/../../doctor_lib.sh"

# A project is never judged against machinery it did not enable: at
# stage 1 no forge read is made at all, and the stand-down is said
# rather than left to look like a clean bill (spec-0004, edge cases).
make_repo 1
forge_healthy
cd "$TARGET" || exit 1

"$WRITRUN" doctor > "$WORK/out" 2>&1

check "stage 1 exits 0" 0 "" -- "$WRITRUN" doctor
check "the forge stand-down is said" 0 \
  "Stage 2 — the forge: not examined — the repository declares stage 1." -- cat "$WORK/out"
check "the Issues stand-down is said" 0 \
  "Stage 3 — Issues: not examined — the repository declares stage 1." -- cat "$WORK/out"
check "no forge read was made" 1 "" -- test -s "$GH_DIR/calls"

finish
