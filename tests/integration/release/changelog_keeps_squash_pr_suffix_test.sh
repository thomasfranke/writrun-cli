#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# The (#NN) the forge appends to a squash subject is the only hop an
# entry has back to its pull request: it is kept as written.
release_setup
git tag -a v0.1.0 -m v0.1.0
git commit -q --allow-empty -m "feat(queue): name the reports (#42)"
out=$(bash "$RELEASE_SH" 2>&1); code=$?
if [ "$code" -eq 0 ] &&
   grep -q '^### feat$' CHANGELOG.md &&
   grep -qx -- '- feat(queue): name the reports (#42)' CHANGELOG.md; then
  echo "ok    a squash subject keeps its (#NN) hop back to the pull request"; pass=$((pass + 1))
else
  echo "FAIL  a squash subject keeps its (#NN) hop back to the pull request"
  printf '%s\n' "$out" | sed 's/^/      | /'
  fail=$((fail + 1))
fi

finish
