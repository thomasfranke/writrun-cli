#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# major moves the first digit, zeroing the rest.
release_setup
git tag -a v0.9.15 -m v0.9.15
out=$(bash "$RELEASE_SH" major 2>&1); code=$?
if [ "$code" -eq 0 ] &&
   git tag --list | grep -qx 'v1.0.0'; then
  echo "ok    major bumps the first digit (v0.9.15 -> v1.0.0)"; pass=$((pass + 1))
else
  echo "FAIL  major bumps the first digit (v0.9.15 -> v1.0.0)"
  printf '%s\n' "$out" | sed 's/^/      | /'
  fail=$((fail + 1))
fi

finish
