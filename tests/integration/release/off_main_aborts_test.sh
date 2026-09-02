#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# Tags are cut from main — any other branch aborts before anything happens.
release_setup
git checkout -qb feature
out=$(bash "$RELEASE_SH" 2>&1); code=$?
if [ "$code" -ne 0 ] &&
   printf '%s' "$out" | grep -q 'cut from main' &&
   [ -z "$(git tag --list)" ]; then
  echo "ok    a branch other than main aborts with nothing mutated"; pass=$((pass + 1))
else
  echo "FAIL  a branch other than main aborts with nothing mutated"
  printf '%s\n' "$out" | sed 's/^/      | /'
  fail=$((fail + 1))
fi

finish
