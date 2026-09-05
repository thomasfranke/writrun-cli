#!/usr/bin/env bash
. "$(dirname "$0")/../../doctor_lib.sh"

# Every stage-1 finding names the file and what is expected of it, and
# every one of them breaks a flow (spec-0004, step 3).
make_repo 1
cd "$TARGET" || exit 1

rm -f "$TARGET/docs/about.md"
check "the About file is named" 1 "docs/about.md — an About file is required" -- "$WRITRUN" doctor
printf '# About\n' > "$TARGET/docs/about.md"

rm -f "$TARGET/docs/product/rules.md"
check "a product folder holding only a README is named" 1 \
  "docs/product/ — at least one real product doc" -- "$WRITRUN" doctor
printf '# Rules\n' > "$TARGET/docs/product/rules.md"

rm -f "$TARGET/docs/technical/boundaries.md"
check "a technical folder holding only a README is named" 1 \
  "docs/technical/ — at least one real technical doc" -- "$WRITRUN" doctor
printf '# Boundaries\n' > "$TARGET/docs/technical/boundaries.md"

rm -rf "$TARGET/work/specs"
check "the split is named where a queue folder is gone" 1 \
  "work/specs/ — the docs/ and work/ split" -- "$WRITRUN" doctor
mkdir -p "$TARGET/work/specs"

rm -f "$TARGET/.writrun/VERSION"
check "an unrecorded tag is named" 1 ".writrun/VERSION — the kit's tag is not recorded" \
  -- "$WRITRUN" doctor
printf 'main\n' > "$TARGET/.writrun/VERSION"
check "an unreadable tag is named" 1 "is not a readable tag" -- "$WRITRUN" doctor
printf 'v0.0.03\n' > "$TARGET/.writrun/VERSION"

check "a repaired repository holds again" 0 "Stage 1 — files: all clear." -- "$WRITRUN" doctor

finish
