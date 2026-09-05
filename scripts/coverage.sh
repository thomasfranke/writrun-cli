#!/usr/bin/env bash
# coverage.sh — the unit tier with its two coverage gates: run it, print
# the percentages, and fail below either floor.
#
#   coverage.sh [total-floor]      default: 90
#
#   COVERAGE_FLOOR           total floor; the argument overrides it
#   COVERAGE_PACKAGE_FLOOR   per-package floor, default 80
#   COVERAGE_PROFILE         profile path, default cover.out
#   COVERAGE_GATE_ONLY=1     gate the profile already there, run no test
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
#
# The percentages are summed from the profile rather than read from
# `go tool cover -func`: -coverpkg writes one copy of each block per test
# binary, so a block is counted once and is covered when any copy covered
# it. Two floors need the per-package split anyway, and -func reports one
# percentage per function.

set -euo pipefail

FLOOR="${1:-${COVERAGE_FLOOR:-90}}"
PACKAGE_FLOOR="${COVERAGE_PACKAGE_FLOOR:-80}"
PROFILE="${COVERAGE_PROFILE:-cover.out}"

if [ "${COVERAGE_GATE_ONLY:-0}" != "1" ]; then
  go test -race -shuffle=on -timeout=10m \
    -coverpkg=./internal/... -coverprofile="$PROFILE" ./...
fi

awk -v floor="$FLOOR" -v pkg_floor="$PACKAGE_FLOOR" '
  NR == 1 && $1 == "mode:" { next }
  NF == 3 {
    stmt[$1] = $2
    if ($3 + 0 > 0) hit[$1] = 1
    next
  }
  { printf "unreadable profile line %d: %s\n", NR, $0; bad = 1 }

  END {
    if (bad) exit 1

    for (block in stmt) {
      pkg = block
      sub(/:[0-9]+\.[0-9]+,[0-9]+\.[0-9]+$/, "", pkg)
      sub(/\/[^\/]*$/, "", pkg)
      if (match(pkg, /\/internal\//)) pkg = substr(pkg, RSTART + 1)

      if (!(pkg in total)) { total[pkg] = 0; covered[pkg] = 0; names[++n] = pkg }
      total[pkg] += stmt[block]
      grand_total += stmt[block]
      if (block in hit) { covered[pkg] += stmt[block]; grand_covered += stmt[block] }
    }

    if (n == 0) { print "the profile names no package"; exit 1 }

    for (i = 2; i <= n; i++) {
      name = names[i]
      for (j = i - 1; j >= 1 && names[j] > name; j--) names[j + 1] = names[j]
      names[j + 1] = name
    }

    if (grand_total == 0) { print "the profile counts no statements"; exit 1 }
    pct = 100 * grand_covered / grand_total
    printf "coverage over internal/: %.1f%% (floor %s%%)\n", pct, floor

    for (i = 1; i <= n; i++) {
      pkg = names[i]
      if (total[pkg] == 0) {
        printf "  %-44s     n/a  no statements\n", pkg
        continue
      }
      each = 100 * covered[pkg] / total[pkg]
      printf "  %-44s %7.1f%% (floor %s%%)\n", pkg, each, pkg_floor
      if (each + 0 < pkg_floor + 0) below[++nb] = sprintf("%s at %.1f%%", pkg, each)
    }

    for (i = 1; i <= nb; i++)
      printf "below the per-package floor: %s (floor %s%%)\n", below[i], pkg_floor
    if (pct + 0 < floor + 0) print "below the gate"

    if (nb > 0 || pct + 0 < floor + 0) exit 1
  }' "$PROFILE"
