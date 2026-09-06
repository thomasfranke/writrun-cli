#!/usr/bin/env bash
. "$(dirname "$0")/../../amend_lib.sh"

# An amendment records that an approval is in question; it does not work
# the task riding on it. So the branch carries no `task-NNNN` id and the
# title carries no `[TASK-NNNN]` tag — a name that read as flight would
# have the machinery report the amendment as the work
# (conventions/branches.md, conventions/prs.md).
make_repo
in_flight_pr
cd "$TARGET" || exit 1

check "the amendment is opened" 0 "Amended spec-0011" \
  -- amend_cmd spec-0011 --title "$TITLE" --yes

check "the branch is the id-less one that was composed" 0 "" \
  -- git -C "$ORIGIN" rev-parse --verify --quiet refs/heads/docs/amend-command
check "no task-shaped branch reached the origin" 1 "" \
  -- git -C "$ORIGIN" rev-parse --verify --quiet refs/heads/task/0012-amend-command
git -C "$TARGET" rev-parse --abbrev-ref HEAD > "$WORK/branch"
check "the branch name carries no digits of the task" 1 "" \
  -- grep -qE "task|0012" "$WORK/branch"
check "the title carries no task tag" 1 "" \
  -- grep -q -- "\[TASK-" "$GH_ARGS"

finish
