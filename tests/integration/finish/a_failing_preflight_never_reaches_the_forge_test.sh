#!/usr/bin/env bash
. "$(dirname "$0")/../../finish_lib.sh"

# A queue file this branch never touched is missing a schema field, so
# preflight's first stage refuses. The completion edits stand — they are
# step 2 and they already happened — but nothing reaches the forge
# (spec-0010, acceptance criteria).
make_repo
on_branch task/0001-a-thing
grep -v '^priority: ' "$TARGET/work/tasks/task-0002-another-thing.md" > "$TARGET/task-0002.tmp"
mv "$TARGET/task-0002.tmp" "$TARGET/work/tasks/task-0002-another-thing.md"
git_q -C "$TARGET" commit -q -am "a queue file the schema refuses"
cd "$TARGET" || exit 1

check "preflight stops the sequence" 1 "PREFLIGHT STOPPED at 1/3" \
  -- finish_cmd --yes

check "the forge was never reached" 1 "" \
  -- grep -q "pr " "$GH_LOG"
check "the completion edits stand" 0 "implemented" \
  -- field status work/specs/spec-0001-a-thing.md
check "the task's status line is untouched" 0 "in-progress" \
  -- field status work/tasks/task-0001-a-thing.md

finish
