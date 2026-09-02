#!/usr/bin/env bash
# run.sh — discovers and runs every case in the suite.
#
#   bash tests/run.sh
#
# Layout: one tier per directory — unit/ for script internals,
# integration/ for the automation scripts, e2e/ for whole-path runs
# against a copy of this repository — inside it one directory per script
# under test, one file per behaviour, suffixed `_test.sh`. Every case
# sources the fixture for its domain (tests/release_lib.sh, layered on
# tests/harness.sh) and also runs standalone:
#
#   bash tests/integration/release/minor_bumps_third_digit_test.sh

set -uo pipefail

TESTS_DIR="$(cd "$(dirname "$0")" && pwd)"

pass=0
fail=0

for tier in "$TESTS_DIR"/*/; do
  [ -d "$tier" ] || continue
  tname=$(basename "$tier")
  for dir in "$tier"*/; do
    [ -d "$dir" ] || continue
    printf '%s/%s\n' "$tname" "$(basename "$dir")"
    for case_file in "$dir"*_test.sh; do
      [ -e "$case_file" ] || continue
      if bash "$case_file"; then
        pass=$((pass + 1))
      else
        fail=$((fail + 1))
      fi
    done
    echo
  done
done

printf '%s case files passed, %s failed\n' "$pass" "$fail"
[ "$fail" -eq 0 ]
