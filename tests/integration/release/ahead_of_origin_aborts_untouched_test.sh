#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# A main ahead of origin carries commits no pull request reviewed, and
# the tag would close over them.
release_setup
git commit -q --allow-empty -m "written here, never pushed"

out=$(bash "$RELEASE_SH" 2>&1); code=$?
if [ "$code" -ne 0 ] &&
   printf '%s' "$out" | grep -q 'ahead of origin/main' &&
   [ ! -f CHANGELOG.md ] &&
   [ -z "$(git tag --list)" ] &&
   [ ! -f "$WORK/calls.log" ]; then
  echo "ok    a main ahead of origin aborts with nothing mutated"; pass=$((pass + 1))
else
  echo "FAIL  a main ahead of origin aborts with nothing mutated"
  printf '%s\n' "$out" | sed 's/^/      | /'
  fail=$((fail + 1))
fi

finish
