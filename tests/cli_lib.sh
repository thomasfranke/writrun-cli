#!/usr/bin/env bash
# cli_lib.sh — the fixture behind the CLI integration cases
# (tests/integration/cli/): the binary compiled into a throwaway
# workspace, driven against fixture directories. Each case sources this
# and runs standalone, or under tests/run.sh.
#
# A runner that exports WRITRUN_BIN_DIR (tests/run.sh, the make
# targets) gets one build shared by every case; standalone, the case
# builds into its own workspace.

. "$(dirname "${BASH_SOURCE[0]}")/harness.sh"

WORK=$(mktemp -d)
WRITRUN="${WRITRUN_BIN_DIR:-$WORK}/writrun"

if [ ! -x "$WRITRUN" ]; then
  if ! BUILD_ERR=$(cd "$REPO_ROOT" && go build -o "$WRITRUN" ./cmd/writrun 2>&1); then
    echo "FAIL  go build ./cmd/writrun"
    printf '%s\n' "$BUILD_ERR" | sed 's/^/      | /'
    fail=$((fail + 1))
    finish
  fi
fi

# PINNED is the WritRun tag this binary pins, read from the binary
# itself — a fixture that spelled it out would have to be edited on
# every bump, and would be the only place the two could disagree.
PINNED=$("$WRITRUN" --version | sed -n 's/.*pins WritRun \(v[0-9.]*\).*/\1/p')
