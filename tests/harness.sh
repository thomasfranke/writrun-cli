#!/usr/bin/env bash
# harness.sh — the assertion core every fixture layers on: paths,
# counters, the check/finish pair, and the temp-dir cleanup. Case files
# never source this directly unless they need no repository at all —
# they source the fixture for their domain (release_lib.sh), which
# sources this.
#
# No framework and no dependency beyond git and the same POSIX awk/sed
# the scripts themselves are held to.

set -uo pipefail

TESTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$TESTS_DIR/.." && pwd)"

pass=0
fail=0
WORK=""

cleanup() { [ -n "$WORK" ] && rm -rf "$WORK"; }
trap cleanup EXIT

# check <name> <expected-exit> <expected-substring-or-empty> -- <cmd...>
check() {
  local name="$1" want_code="$2" want_text="$3"; shift 4
  local out code
  out=$("$@" 2>&1); code=$?

  if [ "$code" -ne "$want_code" ]; then
    printf 'FAIL  %s\n      expected exit %s, got %s\n' "$name" "$want_code" "$code"
    printf '%s\n' "$out" | sed 's/^/      | /'
    fail=$((fail + 1)); return
  fi
  if [ -n "$want_text" ] && ! printf '%s' "$out" | grep -q "$want_text"; then
    printf 'FAIL  %s\n      expected output to contain: %s\n' "$name" "$want_text"
    printf '%s\n' "$out" | sed 's/^/      | /'
    fail=$((fail + 1)); return
  fi
  printf 'ok    %s\n' "$name"
  pass=$((pass + 1))
}

# finish — every case file's last line: exit with the case's verdict.
finish() { [ "$fail" -eq 0 ]; exit $?; }
