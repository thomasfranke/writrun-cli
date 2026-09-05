#!/usr/bin/env bash
. "$(dirname "$0")/../../harness.sh"

# The configuration, validated by the tool that consumes it — a
# deprecated property or an unknown field fails here, not on the tag.
#
# CI installs goreleaser before the suite (tests.yml,
# release-readiness.yml). A session without it gets a named skip rather
# than a green that proved nothing; nothing here releases.
if ! command -v goreleaser >/dev/null 2>&1; then
  echo "ok    the release configuration is valid (skipped: goreleaser is not installed)"
  finish
fi

cd "$REPO_ROOT" || exit 1
check "the release configuration is valid" 0 "configuration file(s) validated" \
  -- goreleaser check

finish
