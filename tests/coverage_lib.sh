#!/usr/bin/env bash
# coverage_lib.sh — the fixture for scripts/coverage.sh: a profile
# written by hand, gated without running a test.
#
# COVERAGE_GATE_ONLY is what makes the gate testable — the percentages
# come from a fixture profile, so a case states the numbers it is about
# instead of arranging Go packages that happen to produce them.

. "$(dirname "${BASH_SOURCE[0]}")/harness.sh"

COVERAGE_SH="$REPO_ROOT/scripts/coverage.sh"
PROFILE=""
blockline=0

# coverage_setup — an empty profile in a temp directory.
coverage_setup() {
  WORK=$(mktemp -d)
  PROFILE="$WORK/cover.out"
  blockline=0
  printf 'mode: set\n' > "$PROFILE"
}

# profile_block <package> <statements> <covered 0|1> [line]
# One block under internal/<package>. Repeat a line number to write the
# copies -coverpkg produces, one per test binary.
profile_block() {
  local pkg="$1" stmts="$2" covered="$3" line="${4:-}"
  if [ -z "$line" ]; then blockline=$((blockline + 1)); line=$blockline; fi
  printf 'example.com/m/internal/%s/f.go:%d.1,%d.2 %s %s\n' \
    "$pkg" "$line" "$line" "$stmts" "$covered" >> "$PROFILE"
}

# gate [total-floor] — the gate over the fixture profile, no test run.
gate() {
  COVERAGE_GATE_ONLY=1 COVERAGE_PROFILE="$PROFILE" bash "$COVERAGE_SH" "$@"
}
