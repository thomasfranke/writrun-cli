#!/usr/bin/env bash
. "$(dirname "$0")/../../finish_lib.sh"

# A project that keeps a provenance ledger has a third writer in the
# sequence: `record_provenance.sh` appends an entry to the task file at
# step 3, between the completion writes and the gates. A refusal undoes
# that append too — the entry records an act that did not happen — and
# it does so on both starting states, including the one where the worker
# had already dated the task by hand and step 2 therefore wrote nothing
# to that file at all (spec-0017; product/pull-requests/shape.md).
# WRITRUN_TTY_IN stands in for the terminal.
make_repo
ledger_on
TASK="$TARGET/work/tasks/task-0001-a-thing.md"
SPEC="$TARGET/work/specs/spec-0001-a-thing.md"
hand_dated "$TASK" 2026-01-02T00:00:00Z
on_branch task/0001-a-thing
cd "$TARGET" || exit 1

tree_is_clean() {
  local out
  out=$(git_q -C "$TARGET" status --porcelain)
  [ -z "$out" ] && return 0
  printf 'the refused finish left:\n%s\n' "$out"
  return 1
}

# --- the date was already there, so step 2 wrote nothing to the task ---

printf 'n' > "$WORK/keys"
export WRITRUN_TTY_IN="$WORK/keys"
check "the ledger appended before the question" 1 "appended to work/tasks/task-0001-a-thing.md" \
  -- finish_cmd --by human --login octocat
unset WRITRUN_TTY_IN

check "the decline is reported as one" 0 "" \
  -- grep -q "declined — nothing changed" "$FINISH_OUT"
check "git status --porcelain is empty after the no" 0 "" \
  -- tree_is_clean
check "the ledger entry is gone with the rest" 1 "" \
  -- grep -q "octocat" "$TASK"
check "the date the worker wrote by hand is untouched" 0 "2026-01-02T00:00:00Z" \
  -- field completed "$TASK"
check "the spec is the approved one it was" 0 "approved" \
  -- field status "$SPEC"

# --- and where this run wrote the date itself -------------------------

sed 's/^completed: 2026-01-02T00:00:00Z$/completed: null/' "$TASK" > "$TASK.tmp"
mv "$TASK.tmp" "$TASK"
git_q -C "$TARGET" commit -q -am "the date is the command's to write again"

printf 'n' > "$WORK/keys"
export WRITRUN_TTY_IN="$WORK/keys"
check "the run wrote the date and the ledger entry" 1 "wrote completed:" \
  -- finish_cmd --by agent --model claude-opus-5 --login octocat
unset WRITRUN_TTY_IN

check "the ledger appended on this path too" 0 "" \
  -- grep -q "appended to work/tasks/task-0001-a-thing.md" "$FINISH_OUT"
check "git status --porcelain is empty after the no" 0 "" \
  -- tree_is_clean
check "the ledger entry is gone with the date" 1 "" \
  -- grep -q "octocat" "$TASK"
check "the task carries no completion date" 0 "null" \
  -- field completed "$TASK"

# --- a yes keeps every one of the three ------------------------------

printf 'y' > "$WORK/keys"
export WRITRUN_TTY_IN="$WORK/keys"
check "the confirmed finish completes" 0 "wrote completed:" \
  -- finish_cmd --by agent --model claude-opus-5 --login octocat
unset WRITRUN_TTY_IN

check "the ledger entry stands after a yes" 0 "" \
  -- grep -q "octocat" "$TASK"
check "the spec is implemented after a yes" 0 "implemented" \
  -- field status "$SPEC"
check "nothing was put back on the way to the forge" 1 "" \
  -- grep -q "restored " "$FINISH_OUT"

finish
