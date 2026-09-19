#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# Neither side contains the other, so no pull and no push settles it —
# the guard says so rather than naming a move that would not work.
release_setup
git commit -q --allow-empty -m "landed on the forge"
git push -q origin main
git reset -q --hard HEAD~1
git commit -q --allow-empty -m "written here instead"

out=$(bash "$RELEASE_SH" 2>&1); code=$?
if [ "$code" -ne 0 ] &&
   printf '%s' "$out" | grep -q 'diverged from origin/main' &&
   [ ! -f CHANGELOG.md ] &&
   [ -z "$(git tag --list)" ] &&
   [ ! -f "$WORK/calls.log" ]; then
  echo "ok    a main diverged from origin aborts with nothing mutated"; pass=$((pass + 1))
else
  echo "FAIL  a main diverged from origin aborts with nothing mutated"
  printf '%s\n' "$out" | sed 's/^/      | /'
  fail=$((fail + 1))
fi

finish
