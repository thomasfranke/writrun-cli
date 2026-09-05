#!/usr/bin/env bash
# run.sh — discovers and runs every case in the suite.
#
#   bash tests/run.sh
#
# Layout: one tier per directory — integration/ for the compiled binary
# against fixture repositories, e2e/ for whole-path runs against a copy
# of this repository — inside it one directory per subject under test,
# one file per behaviour, suffixed `_test.sh`. The unit tier is Go,
# beside the code (technical/testing/tiers.md). Every case sources the
# fixture for its domain, layered on tests/harness.sh, and also runs
# standalone:
#
#   bash tests/integration/release/minor_bumps_middle_digit_test.sh

set -uo pipefail

TESTS_DIR="$(cd "$(dirname "$0")" && pwd)"

# One compiled binary for the whole run: the first CLI case builds it
# here and the rest reuse it (tests/cli_lib.sh).
WRITRUN_BIN_DIR="$(mktemp -d)"
export WRITRUN_BIN_DIR
trap 'rm -rf "$WRITRUN_BIN_DIR"' EXIT

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
