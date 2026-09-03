#!/usr/bin/env bash
# stage_gate.sh — does this project's adoption stage reach far enough for
# the job about to run?
#
# Usage: stage_gate.sh <required-stage>
#   Run from the repository root. Writes `run=true|false` to $GITHUB_OUTPUT
#   when that variable is set, so the steps after it can be guarded with
#   `if: steps.gate.outputs.run == 'true'`.
#
# The stages are ordered and cumulative, which is why one value gates four
# workflows: Stage 3 without Stage 2 would ask for a projection that
# pull-request events drive, with no pull requests to drive it
# (product/adoption.md#three-stages).
#
#   1  tasks and specs   no workflow runs
#   2  pull requests     writrun check, writrun approve
#   3  GitHub issues     adds writrun issues, writrun progress
#
# **The setting is what stops the machinery, not deleting the files.** They
# stay installed and inert — one way to say a thing rather than two free to
# disagree, which is the reversal of 0041 this implements.
#
# It always exits 0. A gate that failed the job would report "did not run"
# as a red check, and a project that chose a lower stage did nothing wrong.
# It says why instead, every time, so a silent absence is never mistaken
# for a silent success.
#
# Exit codes: 0 always, except 3 for a usage error.
#
# Portable bash 3.2, POSIX awk/sed. See the standing rule in
# docs/technical/decisions/.

set -euo pipefail

NEED="${1:-}"
case "$NEED" in
  1|2|3) ;;
  *) echo "usage: stage_gate.sh <1|2|3>" >&2; exit 3 ;;
esac

HERE=$(bash "$(dirname "$0")/read_setting.sh" stage)

# The address the value actually came from, so the off-switch message below
# names a file the adopter has. An unmoved file is still honoured through
# the bridge (decision 0053), and pointing them at the address they have
# not adopted yet would send them to a path that does not exist.
SETTINGS=".writrun/settings.json"
if [ ! -f "$SETTINGS" ] && [ -f ".writrun/conventions/settings.json" ]; then
  SETTINGS=".writrun/conventions/settings.json"
fi

case "$HERE" in
  1|2|3) ;;
  *)
    echo "stage '${HERE}' is outside the vocabulary; reading it as the" >&2
    echo "documented default. check_settings.sh is what names that fault." >&2
    HERE=3
    ;;
esac

report() { [ -n "${GITHUB_OUTPUT:-}" ] && printf 'run=%s\n' "$1" >> "$GITHUB_OUTPUT"; return 0; }

if [ "$HERE" -ge "$NEED" ]; then
  echo "stage is ${HERE}, which reaches ${NEED} — running."
  report true
  exit 0
fi

echo "stage is ${HERE}, which stops below ${NEED} — not running."
echo "This job is off because ${SETTINGS} says so,"
echo "not because anything failed. Raise 'stage' to turn it on"
echo "(docs/product/adoption.md#three-stages)."
report false
