#!/usr/bin/env bash
. "$(dirname "$0")/../../finish_lib.sh"

# The green path over the repository's own scripts: the deltas check
# passes, the spec reaches `implemented`, the task's `completed` date is
# written, preflight runs, and the draft is marked ready on the yes
# (spec-0010, tests required).
make_repo
on_branch task/0001-a-thing
cd "$TARGET" || exit 1

check "the finish reports the checks it ran" 0 "PREFLIGHT OK" \
  -- finish_cmd --yes

check "the delta check ran first" 0 "" \
  -- grep -q "OK: diff matches the promised deltas of spec-0001" "$FINISH_OUT"
check "the spec is implemented" 0 "implemented" \
  -- field status work/specs/spec-0001-a-thing.md
check "the task carries a completed date" 0 "" \
  -- grep -qE '^completed: 2[0-9]{3}-' work/tasks/task-0001-a-thing.md
check "the task's status line is untouched" 0 "in-progress" \
  -- field status work/tasks/task-0001-a-thing.md
check "the pull request was marked ready" 0 "" \
  -- grep -qx "pr ready 7" "$GH_LOG"
check "the composition was shown before the act" 0 "" \
  -- grep -q "pull request   #7" "$FINISH_OUT"

finish
