#!/usr/bin/env bash
. "$(dirname "$0")/../../init_lib.sh"

# In an adopted repository init refuses and points at update
# (spec-0002, acceptance criteria).
make_target "$TARGET"
mkdir -p "$TARGET/.writrun"
cd "$TARGET" || exit 1

check "init refuses an adopted repository, pointing at update" 1 "writrun update" \
  -- "$WRITRUN" init --stage 1 --yes

finish
