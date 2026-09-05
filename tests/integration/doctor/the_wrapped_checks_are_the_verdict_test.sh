#!/usr/bin/env bash
. "$(dirname "$0")/../../doctor_lib.sh"

# The queue's front matter and the settings file are judged by the
# repository's own checks, and their reporting reaches the reader in
# their own words — no second opinion is formed on either
# (docs/about.md, product/rules.md).
make_repo 1
cd "$TARGET" || exit 1

cat > "$TARGET/work/tasks/task-0001-broken.md" <<'EOF'
---
id: task-0001
status: ready
---

# A task whose front matter is not canonical
EOF

"$WRITRUN" doctor > "$WORK/front.out" 2>&1
check "a refusing front-matter check breaks the run" 1 "" -- "$WRITRUN" doctor
check "the refusing check is named" 0 "check_front_matter.sh — it refuses" -- cat "$WORK/front.out"
check "the check's own words are carried" 0 "MALFORMED: work/tasks/task-0001-broken.md" \
  -- cat "$WORK/front.out"
rm -f "$TARGET/work/tasks/task-0001-broken.md"

printf '{"stage": 1, "deep": {"deeper": {"x": 1}}}\n' > "$TARGET/.writrun/settings.json"
"$WRITRUN" doctor > "$WORK/settings.out" 2>&1
check "a refusing settings check breaks the run" 1 "" -- "$WRITRUN" doctor
check "a settings file the readers cannot see is named" 0 \
  "check_settings.sh — it refuses" -- cat "$WORK/settings.out"
check "the settings check's own words are carried" 0 "REJECTED" -- cat "$WORK/settings.out"

settings 1
check "both checks passing leaves stage 1 clear" 0 "Stage 1 — files: all clear." -- "$WRITRUN" doctor

finish
