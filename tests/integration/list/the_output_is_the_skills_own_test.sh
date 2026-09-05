#!/usr/bin/env bash
. "$(dirname "$0")/../../list_lib.sh"

# A queue of mixed statuses, rendered: what the binary prints is what
# the lister prints, byte for byte (spec-0006, definition of done).
make_repo
task task-0001 ready   high   spec-0001 "A task that may be taken"
task task-0002 backlog medium spec-0002 "A task whose spec is not approved"
task task-0003 blocked low    spec-0003 "A task somebody stopped" "waiting on the forge"
spec spec-0001 approved
spec spec-0002 draft
spec spec-0003 approved
report report-0001 open "Something that was noticed"
forge_online
cd "$TARGET" || exit 1

"$WRITRUN" list > "$WORK/binary.out" 2> "$WORK/binary.err"
bash "$LISTER" > "$WORK/script.out" 2>&1

check "the binary prints the skill's listing unedited" 0 "" \
  -- diff -u "$WORK/script.out" "$WORK/binary.out"
check "something available exits 0" 0 "Available — any of these may be taken:" \
  -- "$WRITRUN" list
check "the available task is named" 0 "task-0001" -- grep "task-0001" "$WORK/binary.out"
check "the held-back reason is shown verbatim" 0 "blocked: waiting on the forge" \
  -- grep "task-0003" "$WORK/binary.out"
check "the unapproved spec holds its task back" 0 "spec-0002 is draft" \
  -- grep "task-0002" "$WORK/binary.out"
check "the open report waits for triage" 0 "report-0001" \
  -- grep "report-0001" "$WORK/binary.out"
check "nothing was written to stderr" 0 "" -- test ! -s "$WORK/binary.err"

# Running from a subdirectory is the same answer: the lister runs from
# the repository root, whatever directory the person is in.
mkdir -p "$TARGET/docs/deep"
cd "$TARGET/docs/deep" || exit 1
"$WRITRUN" list > "$WORK/deep.out" 2>&1
check "a subdirectory gets the same listing" 0 "" \
  -- diff -u "$WORK/binary.out" "$WORK/deep.out"

finish
