#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# v0.0.10 is newer than v0.0.9 by version sort though older by the
# lexicographic one — the bump must start from v0.0.10.
release_setup
git tag -a v0.0.9 -m v0.0.9
git tag -a v0.0.10 -m v0.0.10
out=$(bash "$RELEASE_SH" patch 2>&1); code=$?
if [ "$code" -eq 0 ] &&
   git tag --list | grep -qx 'v0.0.11'; then
  echo "ok    the latest tag is found by version sort, not lexicographic"; pass=$((pass + 1))
else
  echo "FAIL  the latest tag is found by version sort, not lexicographic"
  printf '%s\n' "$out" | sed 's/^/      | /'
  fail=$((fail + 1))
fi

finish
