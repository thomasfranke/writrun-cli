#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# No CHANGELOG.md exists: the cut creates it with a title and the first
# section under it, and the release commit carries it.
release_setup
out=$(bash "$RELEASE_SH" 2>&1); code=$?
if [ "$code" -eq 0 ] &&
   [ "$(sed -n 1p CHANGELOG.md)" = "# Changelog" ] &&
   grep -q "^## v0.1.0 — $(date -u +%Y-%m-%d)\$" CHANGELOG.md &&
   git show --stat --format= HEAD | grep -q 'CHANGELOG.md'; then
  echo "ok    the cut creates CHANGELOG.md and commits it"; pass=$((pass + 1))
else
  echo "FAIL  the cut creates CHANGELOG.md and commits it"
  printf '%s\n' "$out" | sed 's/^/      | /'
  fail=$((fail + 1))
fi

finish
