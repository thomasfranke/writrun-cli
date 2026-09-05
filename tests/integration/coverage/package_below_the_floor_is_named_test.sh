#!/usr/bin/env bash
. "$(dirname "$0")/../../coverage_lib.sh"

# The total stays over its floor and one package does not: the average
# is what hid two 0% packages, so the per-package floor is checked from
# the same profile and names what failed (spec-0014).
coverage_setup
profile_block high 950 1
profile_block low 50 1
profile_block low 50 0

check "the total over its floor does not fail the run" 1 \
  "coverage over internal/: 95.2% (floor 90%)" -- gate
check "a package under its floor is named with its percentage" 1 \
  "below the per-package floor: internal/low at 50.0% (floor 80%)" -- gate

finish
