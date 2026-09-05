#!/usr/bin/env bash
. "$(dirname "$0")/../../coverage_lib.sh"

# Both packages clear the per-package floor and the total does not, so
# only the global gate can be what fails the run (spec-0014).
coverage_setup
profile_block a 85 1
profile_block a 15 0
profile_block b 85 1
profile_block b 15 0

check "the total below its floor names the percentage and the floor" 1 \
  "coverage over internal/: 85.0% (floor 90%)" -- gate
check "the total below its floor fails the run" 1 "below the gate" -- gate

finish
