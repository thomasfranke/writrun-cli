#!/usr/bin/env bash
# coverage.sh — the unit tier with its coverage gate: run it, print the
# percentage, and fail below the floor.
#
#   coverage.sh [floor]        default: 85
#
# One home for the gate on purpose. It lived inline in tests.yml, so
# `make tests` passed locally while CI failed on the same tree — the
# percentage is only reported where the gate is read, and both now read
# this file (technical/testing/tiers.md).
#
# The profile counts statements in ./internal/ whoever exercised them,
# so a package with no _test.go of its own still shows the coverage the
# other packages give it. cover.out is left behind for
# `go tool cover -html=cover.out`.

set -euo pipefail

FLOOR="${1:-${COVERAGE_FLOOR:-85}}"
PROFILE="${COVERAGE_PROFILE:-cover.out}"

go test -coverpkg=./internal/... -coverprofile="$PROFILE" ./...

go tool cover -func="$PROFILE" | awk -v floor="$FLOOR" '
  $1 == "total:" {
    seen = 1
    sub(/%/, "", $3)
    printf "coverage over internal/: %s%% (floor %s%%)\n", $3, floor
    if ($3 + 0 < floor + 0) { print "below the gate"; exit 1 }
  }
  END { if (!seen) { print "no total: line in the cover output"; exit 1 } }'
