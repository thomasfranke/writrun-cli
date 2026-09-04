#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# `epoch` was WritRun's word, not SemVer's — the vocabulary is
# patch|minor|major and anything else aborts.
release_setup
out=$(bash "$RELEASE_SH" epoch 2>&1); code=$?
if [ "$code" -ne 0 ] &&
   printf '%s' "$out" | grep -q 'pick one of patch|minor|major' &&
   [ -z "$(git tag --list)" ]; then
  echo "ok    an unknown bump word aborts (epoch is not in the vocabulary)"; pass=$((pass + 1))
else
  echo "FAIL  an unknown bump word aborts (epoch is not in the vocabulary)"
  printf '%s\n' "$out" | sed 's/^/      | /'
  fail=$((fail + 1))
fi

finish
