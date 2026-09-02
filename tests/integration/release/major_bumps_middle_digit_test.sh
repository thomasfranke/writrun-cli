#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# major moves the middle digit and zeroes the third.
release_setup
git tag -a v0.1.9 -m v0.1.9
out=$(bash "$RELEASE_SH" major 2>&1); code=$?
if [ "$code" -eq 0 ] &&
   [ "$(cat VERSION)" = "v0.2.0" ] &&
   git tag --list | grep -qx 'v0.2.0'; then
  echo "ok    major bumps the middle digit (v0.1.9 -> v0.2.0)"; pass=$((pass + 1))
else
  echo "FAIL  major bumps the middle digit (v0.1.9 -> v0.2.0)"
  printf '%s\n' "$out" | sed 's/^/      | /'
  fail=$((fail + 1))
fi

finish
