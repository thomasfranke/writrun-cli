#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# patch — the default — moves the third digit only.
release_setup
git tag -a v0.0.10 -m v0.0.10
out=$(bash "$RELEASE_SH" patch 2>&1); code=$?
if [ "$code" -eq 0 ] &&
   git tag --list | grep -qx 'v0.0.11'; then
  echo "ok    patch bumps the third digit (v0.0.10 -> v0.0.11)"; pass=$((pass + 1))
else
  echo "FAIL  patch bumps the third digit (v0.0.10 -> v0.0.11)"
  printf '%s\n' "$out" | sed 's/^/      | /'
  fail=$((fail + 1))
fi

finish
