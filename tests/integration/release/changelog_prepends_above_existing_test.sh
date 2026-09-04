#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# Newest first: the new section lands above every existing one and below
# the preamble the file opens with — the older section is kept, not
# rewritten.
release_setup
printf '# Changelog\n\nPreamble stays put.\n\n## v0.1.0 — 2026-01-01\n\n### chore\n\n- chore: the first cut\n\n' > CHANGELOG.md
git add -A >/dev/null && git commit -qm "chore: seed the changelog"
git tag -a v0.1.0 -m v0.1.0
git commit -q --allow-empty -m "feat(queue): read the queue"
out=$(bash "$RELEASE_SH" 2>&1); code=$?
new=$(grep -n '^## v0.1.1' CHANGELOG.md | cut -d: -f1)
old=$(grep -n '^## v0.1.0' CHANGELOG.md | cut -d: -f1)
pre=$(grep -n '^Preamble stays put.$' CHANGELOG.md | cut -d: -f1)
if [ "$code" -eq 0 ] && [ -n "$new" ] && [ -n "$old" ] && [ -n "$pre" ] &&
   [ "$pre" -lt "$new" ] && [ "$new" -lt "$old" ] &&
   grep -q '^- feat(queue): read the queue$' CHANGELOG.md &&
   grep -q '^- chore: the first cut$' CHANGELOG.md; then
  echo "ok    a new section prepends above the existing one, below the preamble"; pass=$((pass + 1))
else
  echo "FAIL  a new section prepends above the existing one, below the preamble"
  printf '%s\n' "$out" | sed 's/^/      | /'
  fail=$((fail + 1))
fi

finish
