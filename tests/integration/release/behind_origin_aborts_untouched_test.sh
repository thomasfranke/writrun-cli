#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# A main behind origin tags a history the forge never saw, and the push
# at the end is where it would find out — after the tag is public.
release_setup
git commit -q --allow-empty -m "landed on the forge"
git push -q origin main
git reset -q --hard HEAD~1

out=$(bash "$RELEASE_SH" 2>&1); code=$?
if [ "$code" -ne 0 ] &&
   printf '%s' "$out" | grep -q 'behind origin/main' &&
   [ ! -f CHANGELOG.md ] &&
   [ -z "$(git tag --list)" ] &&
   [ ! -f "$WORK/calls.log" ]; then
  echo "ok    a main behind origin aborts with nothing mutated"; pass=$((pass + 1))
else
  echo "FAIL  a main behind origin aborts with nothing mutated"
  printf '%s\n' "$out" | sed 's/^/      | /'
  fail=$((fail + 1))
fi

finish
