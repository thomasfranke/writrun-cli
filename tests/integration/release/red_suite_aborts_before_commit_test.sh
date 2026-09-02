#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# A red suite aborts the release before the commit: only the stamp is
# left dirty in the tree — no commit, no tag, no push, no forge release.
release_setup
printf '#!/usr/bin/env bash\n[ "${1:-}" = tests ] && exit 1\nexit 0\n' > "$WORK/stub-bin/make"
out=$(bash "$RELEASE_SH" 2>&1); code=$?
if [ "$code" -ne 0 ] &&
   [ "$(git rev-list --count HEAD)" = "1" ] &&
   [ -z "$(git tag --list)" ] &&
   ! grep -q 'gh release create' "$WORK/calls.log" 2>/dev/null &&
   [ -n "$(git status --porcelain)" ]; then
  echo "ok    a red suite aborts before the commit, leaving only the stamp dirty"; pass=$((pass + 1))
else
  echo "FAIL  a red suite aborts before the commit, leaving only the stamp dirty"
  printf '%s\n' "$out" | sed 's/^/      | /'
  fail=$((fail + 1))
fi

finish
