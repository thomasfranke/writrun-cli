#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# A subject the vocabulary does not know is listed verbatim under Other,
# never dropped — judging it is the door's job at the pull request, not
# the tag's.
release_setup
git tag -a v0.0.01 -m v0.0.01
git commit -q --allow-empty -m "wip: poke at it"
git commit -q --allow-empty -m "fix: a real one"
out=$(bash "$RELEASE_SH" 2>&1); code=$?
if [ "$code" -eq 0 ] &&
   grep -q '^### fix$' CHANGELOG.md &&
   grep -q '^- fix: a real one$' CHANGELOG.md &&
   grep -q '^### Other$' CHANGELOG.md &&
   grep -q '^- wip: poke at it$' CHANGELOG.md; then
  echo "ok    a non-conventional subject lands verbatim under Other"; pass=$((pass + 1))
else
  echo "FAIL  a non-conventional subject lands verbatim under Other"
  printf '%s\n' "$out" | sed 's/^/      | /'
  fail=$((fail + 1))
fi

finish
