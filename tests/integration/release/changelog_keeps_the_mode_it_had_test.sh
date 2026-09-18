#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# The cut writes back through the existing file. An `mv` from mktemp
# would carry 0600 onto it, and git records only the exec bit — so a
# changelog nobody but its author can read would reach main looking
# untouched.
release_setup
printf '# Changelog\n\n## v0.1.0 — 2026-01-01\n\n### chore\n\n- chore: the first cut\n\n' > CHANGELOG.md
git add -A >/dev/null && git commit -qm "chore: seed the changelog"
git tag -a v0.1.0 -m v0.1.0
git commit -q --allow-empty -m "feat(queue): read the queue"
release_publish
# 0640, not the mode a fresh file lands on, so the assertion reads the
# mode this file had and not one the cut got for free.
chmod 0640 CHANGELOG.md
out=$(bash "$RELEASE_SH" 2>&1); code=$?
if [ "$code" -eq 0 ] &&
   grep -q '^## v0.1.1' CHANGELOG.md &&
   [ "$(ls -l CHANGELOG.md | cut -c1-10)" = "-rw-r-----" ]; then
  echo "ok    the cut leaves the changelog's mode alone"; pass=$((pass + 1))
else
  echo "FAIL  the cut leaves the changelog's mode alone"
  printf '%s\n' "$out" | sed 's/^/      | /'
  ls -l CHANGELOG.md | sed 's/^/      > /'
  fail=$((fail + 1))
fi

finish
