#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# The first cut counts from nothing, and the bump decides it like any
# other. A constant would make `make release patch` answer a number it
# was not asked for and say nothing about overriding — which is what it
# did, and how this was found (technical/versioning/release.md).
for pair in "patch v0.0.1" "minor v0.1.0" "major v1.0.0"; do
  set -- $pair
  bump=$1 want=$2

  release_setup
  out=$(bash "$RELEASE_SH" "$bump" 2>&1); code=$?
  if [ "$code" -eq 0 ] &&
     git tag --list | grep -qx "$want" &&
     printf '%s\n' "$out" | grep -q "release: none -> $want ($bump)"; then
    echo "ok    the first $bump cuts $want, and says so"; pass=$((pass + 1))
  else
    echo "FAIL  the first $bump cuts $want, and says so"
    printf '%s\n' "$out" | sed 's/^/      | /'
    fail=$((fail + 1))
  fi
done

finish
