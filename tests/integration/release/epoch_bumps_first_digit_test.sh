#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# epoch — historic milestones only — moves the first digit, zeroing the rest.
release_setup
git tag -a v0.9.15 -m v0.9.15
out=$(bash "$RELEASE_SH" epoch 2>&1); code=$?
if [ "$code" -eq 0 ] &&
   [ "$(cat VERSION)" = "v1.0.00" ] &&
   git tag --list | grep -qx 'v1.0.00'; then
  echo "ok    epoch bumps the first digit (v0.9.15 -> v1.0.00)"; pass=$((pass + 1))
else
  echo "FAIL  epoch bumps the first digit (v0.9.15 -> v1.0.00)"
  printf '%s\n' "$out" | sed 's/^/      | /'
  fail=$((fail + 1))
fi

finish
