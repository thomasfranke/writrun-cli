#!/usr/bin/env bash
. "$(dirname "$0")/../../init_lib.sh"

# A fenced section a kit before v0.0.04 grafted is named and left
# exactly as it is: from v0.0.04 the whole of AGENTS.md is the
# project's, so a refresh does not rewrite it (spec-0025, step 6).
make_target "$TARGET"
cd "$TARGET" || exit 1
"$WRITRUN" init --stage 1 --yes > /dev/null 2>&1 || { echo "FAIL  the fixture could not adopt"; exit 1; }
age_kit "$TARGET"
cat > AGENTS.md <<'EOF'
# AGENTS.md

The project's own paragraph.

## WritRun — working the queue

<!-- writrun:begin
     This section is WritRun's flow. -->

### Picking work

The flow's text.

<!-- writrun:end -->
EOF
cp AGENTS.md "$WORK/agents.before"
git_q add -A && git_q commit -q -m "chore: the kit, one release back"

check "the plan names the stale section" 0 "still carries a writrun:begin/writrun:end section" \
  -- "$WRITRUN" update --yes
check "the refresh proceeded" 0 "$TAG" -- cat .writrun/VERSION
check "AGENTS.md is byte-identical" 0 "" -- cmp "$WORK/agents.before" AGENTS.md

finish
