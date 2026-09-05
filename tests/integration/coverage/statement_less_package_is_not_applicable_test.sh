#!/usr/bin/env bash
. "$(dirname "$0")/../../coverage_lib.sh"

# A package holding no statements divides by zero in a naive check. It
# is reported as not applicable and does not fail the run (spec-0014).
coverage_setup
profile_block a 100 1
profile_block consts 0 0

check "a statement-less package is reported not applicable" 0 \
  "internal/consts                                  n/a  no statements" -- gate
check "a statement-less package does not fail the run" 0 \
  "coverage over internal/: 100.0% (floor 90%)" -- gate

finish
