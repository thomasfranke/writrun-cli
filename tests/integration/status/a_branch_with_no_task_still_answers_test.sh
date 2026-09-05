#!/usr/bin/env bash
. "$(dirname "$0")/../../status_lib.sh"

# A branch carrying no task is said plainly, and the reports and the kit
# are still answered — only the completion checks are skipped, because
# there is no completion to be short of (spec-0013, acceptance criteria).
make_repo
report report-0001 open
on_branch main
cd "$TARGET" || exit 1

"$WRITRUN" status > "$WORK/main.out" 2>&1

check "the branch is named" 0 "Branch   main" -- cat "$WORK/main.out"
check "carrying no task is said" 0 "Task     none — this branch carries no task" -- cat "$WORK/main.out"
check "no check is named" 1 "" -- grep -q "^Checks" "$WORK/main.out"
check "the reports are still counted" 0 "Reports  1 open" -- cat "$WORK/main.out"
check "the kit is still named" 0 "Kit      WritRun v0.0.03" -- cat "$WORK/main.out"
check "answering is exit 0" 0 "" -- "$WRITRUN" status

# A detached HEAD is the same answer, said as what it is.
on_branch --detach
"$WRITRUN" status > "$WORK/detached.out" 2>&1
check "a detached HEAD is said plainly" 0 "Branch   detached HEAD — no branch" -- cat "$WORK/detached.out"
check "and it carries no task either" 0 "Task     none" -- cat "$WORK/detached.out"

finish
