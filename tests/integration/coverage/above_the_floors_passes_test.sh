#!/usr/bin/env bash
. "$(dirname "$0")/../../coverage_lib.sh"

# Every package over both floors: the gate prints the percentages and
# exits 0 (spec-0014).
coverage_setup
profile_block a 40 1
profile_block b 60 1

check "a profile over both floors passes" 0 "coverage over internal/: 100.0% (floor 90%)" \
  -- gate

finish
