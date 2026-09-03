#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# No tag exists yet: the first release is v0.0.01 by decree, whatever the
# bump would be — and the whole path runs in order: the forge guard
# before anything, the suite next, then the commit, the tag, the push,
# the forge release.
release_setup
out=$(bash "$RELEASE_SH" 2>&1); code=$?
if [ "$code" -eq 0 ] &&
   [ "$(sed -n 1p "$WORK/calls.log")" = "gh auth status" ] &&
   [ "$(sed -n 2p "$WORK/calls.log")" = "make tests" ] &&
   git log -1 --format=%s | grep -q 'chore(release): v0.0.01' &&
   git tag --list | grep -qx 'v0.0.01' &&
   git ls-remote --tags origin | grep -q 'refs/tags/v0.0.01' &&
   grep -q 'gh release create v0.0.01' "$WORK/calls.log"; then
  echo "ok    the first release is v0.0.01: tested, tagged, pushed, published"; pass=$((pass + 1))
else
  echo "FAIL  the first release is v0.0.01: tested, tagged, pushed, published"
  printf '%s\n' "$out" | sed 's/^/      | /'
  fail=$((fail + 1))
fi

finish
