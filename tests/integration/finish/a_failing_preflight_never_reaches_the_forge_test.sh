#!/usr/bin/env bash
. "$(dirname "$0")/../../finish_lib.sh"

# A queue file this branch never touched is missing a schema field, so
# preflight's first stage refuses. The gates saw the completion edits —
# that is why they run after them — and the refusal then puts them back,
# so a run the gates stopped leaves the tree as it found it (spec-0010,
# acceptance criteria; spec-0017).
make_repo
on_branch task/0001-a-thing
grep -v '^priority: ' "$TARGET/work/tasks/task-0002-another-thing.md" > "$TARGET/task-0002.tmp"
mv "$TARGET/task-0002.tmp" "$TARGET/work/tasks/task-0002-another-thing.md"
git_q -C "$TARGET" commit -q -am "a queue file the schema refuses"
cd "$TARGET" || exit 1

tree_is_clean() {
  local out
  out=$(git_q -C "$TARGET" status --porcelain)
  [ -z "$out" ] && return 0
  printf 'the stopped finish left:\n%s\n' "$out"
  return 1
}

check "preflight stops the sequence" 1 "PREFLIGHT STOPPED at 1/3" \
  -- finish_cmd --yes

check "the forge was never reached" 1 "" \
  -- grep -q "pr " "$GH_LOG"
check "the completion edits are put back" 0 "approved" \
  -- field status work/specs/spec-0001-a-thing.md
check "the completion date is put back" 0 "null" \
  -- field completed work/tasks/task-0001-a-thing.md
check "git status --porcelain is empty after the stop" 0 "" \
  -- tree_is_clean
check "the task's status line is untouched" 0 "in-progress" \
  -- field status work/tasks/task-0001-a-thing.md

finish
