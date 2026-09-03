#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# A dirty tree aborts before anything happens: no call, no tag.
release_setup
printf 'uncommitted\n' >> README.md
out=$(bash "$RELEASE_SH" 2>&1); code=$?
if [ "$code" -ne 0 ] &&
   printf '%s' "$out" | grep -q 'not clean' &&
   [ -z "$(git tag --list)" ] &&
   [ ! -f "$WORK/calls.log" ]; then
  echo "ok    a dirty tree aborts with nothing mutated"; pass=$((pass + 1))
else
  echo "FAIL  a dirty tree aborts with nothing mutated"
  printf '%s\n' "$out" | sed 's/^/      | /'
  fail=$((fail + 1))
fi

finish
