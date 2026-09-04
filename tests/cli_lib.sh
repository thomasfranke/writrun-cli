#!/usr/bin/env bash
# cli_lib.sh — the fixture behind the CLI integration cases
# (tests/integration/cli/): the binary compiled into a throwaway
# workspace, driven against fixture directories. Each case sources this
# and runs standalone, or under tests/run.sh.

. "$(dirname "${BASH_SOURCE[0]}")/harness.sh"

WORK=$(mktemp -d)
WRITRUN="$WORK/writrun"

if ! BUILD_ERR=$(cd "$REPO_ROOT" && go build -o "$WRITRUN" ./cmd/writrun 2>&1); then
  echo "FAIL  go build ./cmd/writrun"
  printf '%s\n' "$BUILD_ERR" | sed 's/^/      | /'
  fail=$((fail + 1))
  finish
fi
