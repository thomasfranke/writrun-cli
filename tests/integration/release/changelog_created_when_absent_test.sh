#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# No CHANGELOG.md exists: the cut creates it with a title and the first
# section under it, and the release commit carries it beside the stamp.
release_setup
out=$(bash "$RELEASE_SH" 2>&1); code=$?
if [ "$code" -eq 0 ] &&
   [ "$(sed -n 1p CHANGELOG.md)" = "# Changelog" ] &&
   grep -q "^## v0.0.01 — $(date -u +%Y-%m-%d)\$" CHANGELOG.md &&
   git show --stat --format= HEAD | grep -q 'CHANGELOG.md' &&
   git show --stat --format= HEAD | grep -q 'VERSION'; then
  echo "ok    the cut creates CHANGELOG.md and commits it with the stamp"; pass=$((pass + 1))
else
  echo "FAIL  the cut creates CHANGELOG.md and commits it with the stamp"
  printf '%s\n' "$out" | sed 's/^/      | /'
  fail=$((fail + 1))
fi

finish
