#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# When the stamp already holds the next number (a release re-cut after a
# failed push), there is nothing to commit — the tag still lands on HEAD.
release_setup
printf 'v0.0.01\n' > VERSION
git add -A >/dev/null
git commit -qm "pre-stamped"
before=$(git rev-list --count HEAD)
out=$(bash "$RELEASE_SH" 2>&1); code=$?
if [ "$code" -eq 0 ] &&
   [ "$(git rev-list --count HEAD)" = "$before" ] &&
   [ "$(git describe --tags)" = "v0.0.01" ]; then
  echo "ok    an unchanged stamp skips the commit but still tags HEAD"; pass=$((pass + 1))
else
  echo "FAIL  an unchanged stamp skips the commit but still tags HEAD"
  printf '%s\n' "$out" | sed 's/^/      | /'
  fail=$((fail + 1))
fi

finish
