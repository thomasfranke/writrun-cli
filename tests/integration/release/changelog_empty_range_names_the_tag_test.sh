#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# The tag is cut on a HEAD the last tag already points at: the section
# says so rather than standing as an empty heading.
release_setup
git tag -a v0.1.0 -m v0.1.0
out=$(bash "$RELEASE_SH" 2>&1); code=$?
if [ "$code" -eq 0 ] &&
   grep -q '^## v0.1.1' CHANGELOG.md &&
   grep -q '^No commit landed between v0.1.0 and this tag.$' CHANGELOG.md &&
   ! grep -q '^### ' CHANGELOG.md; then
  echo "ok    an empty range writes a section naming the tag, never an empty heading"; pass=$((pass + 1))
else
  echo "FAIL  an empty range writes a section naming the tag, never an empty heading"
  printf '%s\n' "$out" | sed 's/^/      | /'
  fail=$((fail + 1))
fi

finish
