#!/usr/bin/env bash
. "$(dirname "$0")/../../coverage_lib.sh"

# -coverpkg writes one copy of every block per test binary. A block is
# counted once and is covered when any copy covered it — summing the
# copies would report a package that one binary covered as half covered
# (spec-0014).
coverage_setup
profile_block a 100 0 7
profile_block a 100 1 7

check "a block repeated per test binary is counted once" 0 \
  "internal/a                                     100.0% (floor 80%)" -- gate

finish
