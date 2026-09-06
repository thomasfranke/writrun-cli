#!/usr/bin/env bash
. "$(dirname "$0")/../../finish_lib.sh"

# The composition is shown, the question is asked, and a no leaves
# nothing behind — not the forge, and not the two completion edits the
# order puts before the question. The whole sentence, asserted over the
# working tree itself (spec-0010, steps 2 and 5; spec-0017;
# product/pull-requests/shape.md). WRITRUN_TTY_IN stands in for the
# terminal.
make_repo
on_branch task/0001-a-thing
printf 'n' > "$WORK/keys"
export WRITRUN_TTY_IN="$WORK/keys"
cd "$TARGET" || exit 1

TASK="$TARGET/work/tasks/task-0001-a-thing.md"
SPEC="$TARGET/work/specs/spec-0001-a-thing.md"

# The whole guarantee in one predicate: a tree with nothing to commit.
# It prints what it found so a failure names the file it was left in.
tree_is_clean() {
  local out
  out=$(git_q -C "$TARGET" status --porcelain)
  [ -z "$out" ] && return 0
  printf 'the refused finish left:\n%s\n' "$out"
  return 1
}

check "the composition is shown before the question" 1 "pull request   #7" \
  -- finish_cmd

unset WRITRUN_TTY_IN
check "the decline is reported as one" 0 "" \
  -- grep -q "declined — nothing changed" "$FINISH_OUT"
check "the pull request was not marked ready" 1 "" \
  -- grep -q "pr ready" "$GH_LOG"
check "the pull request was read, and only read" 0 "" \
  -- grep -q "pr view" "$GH_LOG"

check "git status --porcelain is empty after the no" 0 "" \
  -- tree_is_clean
check "the spec is the approved one it was" 0 "approved" \
  -- field status "$SPEC"
check "the task carries no completion date" 0 "null" \
  -- field completed "$TASK"
check "the spec is reported put back" 0 "" \
  -- grep -q "restored work/specs/spec-0001-a-thing.md" "$FINISH_OUT"
check "the task is reported put back" 0 "" \
  -- grep -q "restored work/tasks/task-0001-a-thing.md" "$FINISH_OUT"

# A rerun after a no is a first run: the decline left nothing for it to
# report as already done (spec-0017, acceptance criteria).
printf 'y' > "$WORK/keys"
export WRITRUN_TTY_IN="$WORK/keys"
check "the rerun after a no completes" 0 "wrote completed:" \
  -- finish_cmd
unset WRITRUN_TTY_IN

check "the rerun wrote rather than reported unchanged" 1 "" \
  -- grep -q "unchanged:" "$FINISH_OUT"
check "the rerun marked the pull request ready" 0 "" \
  -- grep -q "pr ready" "$GH_LOG"
check "the spec is implemented after the rerun" 0 "implemented" \
  -- field status "$SPEC"

finish
