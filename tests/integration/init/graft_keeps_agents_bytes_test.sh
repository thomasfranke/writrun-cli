#!/usr/bin/env bash
. "$(dirname "$0")/../../init_lib.sh"

# An existing AGENTS.md is grafted, every byte outside the fenced
# section unchanged (spec-0002, acceptance criteria).
make_target "$TARGET"
printf '# My own AGENTS.md\n\nRules I already had, byte for byte.\n' > "$TARGET/AGENTS.md"
(cd "$TARGET" && git_q add . && git_q commit -q -m "agents")
cp "$TARGET/AGENTS.md" "$WORK/agents.before"
cd "$TARGET" || exit 1

check "init adopts with an existing AGENTS.md" 0 "graft" \
  -- "$WRITRUN" init --stage 1 --yes

before_bytes=$(wc -c < "$WORK/agents.before")
head -c "$before_bytes" AGENTS.md > "$WORK/agents.head"
check "every byte before the graft survives verbatim" 0 "" \
  -- cmp "$WORK/agents.before" "$WORK/agents.head"
check "the fenced section is now in place" 0 "writrun:end" \
  -- grep "writrun:end" AGENTS.md

finish
