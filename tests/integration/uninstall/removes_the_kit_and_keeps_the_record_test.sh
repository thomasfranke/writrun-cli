#!/usr/bin/env bash
. "$(dirname "$0")/../../init_lib.sh"

# Adopt, then uninstall: the removal set is gone, the keep set is
# byte-identical (spec-0005, acceptance criteria).
make_target "$TARGET"
printf '# My own AGENTS.md\n\nRules I already had.\n' > "$TARGET/AGENTS.md"
(cd "$TARGET" && git_q add . && git_q commit -q -m "agents")
cp "$TARGET/AGENTS.md" "$WORK/agents.before"
cd "$TARGET" || exit 1
"$WRITRUN" init --stage 1 --yes > /dev/null 2>&1 || { echo "FAIL  the fixture could not adopt"; exit 1; }

# The project's record, written after the adoption.
printf 'id: task-0001\n' > work/tasks/task-0001-a-task.md
printf '# Our own chapter\n' > docs/product/a-chapter.md
cp work/tasks/task-0001-a-task.md "$WORK/task.before"
cp docs/product/a-chapter.md "$WORK/chapter.before"

hook_at=$(git_q rev-parse --git-path hooks/commit-msg)
check "the hook is installed before the removal" 0 "" -- test -f "$hook_at"

check "uninstall shows both sets and removes" 0 "stays        work/" \
  -- "$WRITRUN" uninstall --yes

check ".writrun/ is gone" 1 "" -- test -e .writrun
check "WRITRUN.md is gone" 1 "" -- test -e WRITRUN.md
check "the instructions doc is gone" 1 "" -- test -e docs/writrun-instructions.md
for wf in approve check issues progress; do
  check "the $wf workflow is gone" 1 "" -- test -e ".github/workflows/writrun-$wf.yml"
done
check "the commit-msg hook is gone" 1 "" -- test -f "$hook_at"

check "the queue survives byte-identical" 0 "" -- cmp "$WORK/task.before" work/tasks/task-0001-a-task.md
check "the project's chapter survives byte-identical" 0 "" -- cmp "$WORK/chapter.before" docs/product/a-chapter.md
check "AGENTS.md is back to what the project wrote" 0 "" -- cmp "$WORK/agents.before" AGENTS.md

finish
