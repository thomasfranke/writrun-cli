#!/usr/bin/env bash
. "$(dirname "$0")/../../init_lib.sh"

# Content on both sides of the fence survives byte-identical; only the
# fenced section goes (spec-0005, acceptance criteria).
make_target "$TARGET"
printf '# Ours\n\nBefore the fence.\n' > "$TARGET/AGENTS.md"
(cd "$TARGET" && git_q add . && git_q commit -q -m "agents")
cd "$TARGET" || exit 1
"$WRITRUN" init --stage 1 --yes > /dev/null 2>&1 || { echo "FAIL  the fixture could not adopt"; exit 1; }
printf '\n## After the fence\n\nTrailing prose.\n' >> AGENTS.md

check "uninstall edits AGENTS.md rather than removing it" 0 "edit         AGENTS.md" \
  -- "$WRITRUN" uninstall --yes

check "what came before the fence survives" 0 "Before the fence." -- cat AGENTS.md
check "what came after the fence survives" 0 "Trailing prose." -- cat AGENTS.md
check "the fence is gone" 1 "" -- grep -q "writrun:begin" AGENTS.md

finish
