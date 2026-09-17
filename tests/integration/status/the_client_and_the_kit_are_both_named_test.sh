#!/usr/bin/env bash
. "$(dirname "$0")/../../status_lib.sh"

# Two facts, not one: the binary that answered and the tag its scripts
# came from (spec-0043, steps 1 and 3). The client's version is read
# where `--version` reads it, so the two surfaces cannot disagree about
# what is running — which is what this case compares.
make_repo
cd "$TARGET" || exit 1

VERSION=$("$WRITRUN" --version | awk '{print $2}')
"$WRITRUN" status > "$WORK/answer.out" 2>&1

check "the client that answered is named" 0 "Client   writrun-cli $VERSION" \
  -- cat "$WORK/answer.out"
check "the kit's tag is named beside it" 0 "Kit      WritRun $PINNED" \
  -- cat "$WORK/answer.out"
check "answering is still exit 0" 0 "" -- "$WRITRUN" status

finish
