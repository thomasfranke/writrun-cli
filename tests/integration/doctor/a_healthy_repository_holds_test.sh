#!/usr/bin/env bash
. "$(dirname "$0")/../../doctor_lib.sh"

# Every assumption met: exit 0, and each stage says so by name — the
# report is read stage by stage (spec-0004, acceptance criteria).
make_repo 3
forge_healthy
cd "$TARGET" || exit 1

"$WRITRUN" doctor > "$WORK/out" 2>&1

check "every assumption holding exits 0" 0 "" -- "$WRITRUN" doctor
check "the declared stage opens the report" 0 "Stage 3 is declared" -- cat "$WORK/out"
check "the environment is clear" 0 "Stage 0 — environment: all clear." -- cat "$WORK/out"
check "the files are clear" 0 "Stage 1 — files: all clear." -- cat "$WORK/out"
check "the forge is clear" 0 "Stage 2 — the forge: all clear." -- cat "$WORK/out"
check "Issues are clear" 0 "Stage 3 — Issues: all clear." -- cat "$WORK/out"
check "the summary agrees" 0 "Every assumption up to stage 3 holds." -- cat "$WORK/out"
check "the report says it repairs nothing" 0 "doctor reports; it repairs nothing." -- cat "$WORK/out"

finish
