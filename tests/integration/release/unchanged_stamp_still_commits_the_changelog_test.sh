#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# When the stamp already holds the next number (a release re-cut after a
# failed push) the stamp write is a no-op — but the cut always composes
# a section, so there is always something to commit, and the tag lands
# on the commit that carries it.
release_setup
printf 'v0.0.01\n' > VERSION
git add -A >/dev/null
git commit -qm "pre-stamped"
before=$(git rev-list --count HEAD)
out=$(bash "$RELEASE_SH" 2>&1); code=$?
if [ "$code" -eq 0 ] &&
   [ "$(cat VERSION)" = "v0.0.01" ] &&
   [ "$(git rev-list --count HEAD)" = "$((before + 1))" ] &&
   git show --stat --format= HEAD | grep -q 'CHANGELOG.md' &&
   [ "$(git describe --tags)" = "v0.0.01" ]; then
  echo "ok    an unchanged stamp still earns a commit — the changelog carries it"; pass=$((pass + 1))
else
  echo "FAIL  an unchanged stamp still earns a commit — the changelog carries it"
  printf '%s\n' "$out" | sed 's/^/      | /'
  fail=$((fail + 1))
fi

finish
