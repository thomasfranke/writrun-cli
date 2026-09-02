#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# `patch` is SemVer's word, not this project's — the vocabulary is
# minor|major|epoch and anything else aborts.
release_setup
out=$(bash "$RELEASE_SH" patch 2>&1); code=$?
if [ "$code" -ne 0 ] &&
   printf '%s' "$out" | grep -q 'pick one of minor|major|epoch' &&
   [ -z "$(git tag --list)" ]; then
  echo "ok    an unknown bump word aborts (patch is not in the vocabulary)"; pass=$((pass + 1))
else
  echo "FAIL  an unknown bump word aborts (patch is not in the vocabulary)"
  printf '%s\n' "$out" | sed 's/^/      | /'
  fail=$((fail + 1))
fi

finish
