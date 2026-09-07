#!/usr/bin/env bash
# check_promised_deltas.sh — runs the delta contract for every spec this
# change implements.
#
# Usage: check_promised_deltas.sh <diff-range>
#
# An authoring change has no spec to check against — it ships no behaviour
# and promised no deltas; identified by the absence of any spec the change
# moved to `implemented`. Otherwise, one check_deltas.sh call with every
# implemented spec: MISSING is judged per spec, UNDECLARED against the
# union of their promises — checking each spec alone against the whole diff
# would report every sibling spec's promise as undeclared and fail a
# legitimate multi-spec completion.

set -euo pipefail
RANGE="${1:?usage: check_promised_deltas.sh <diff-range>}"

# Resolved from this script's own location, so it works from inside the
# throwaway repositories the test suite builds.
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
CHECK_DELTAS="$REPO_ROOT/.writrun/skills/writrun-check-spec-deltas/check_deltas.sh"

. "$(dirname "$0")/queue_lib.sh"

# "The specs this change implements" is read from the front matter at the
# range's two ends, never grepped out of the diff text — a spec body
# quoting `status: implemented` at column 0 is not an implementation and
# must not have its promises checked against this diff.
ql_range_ends "$RANGE"
BASE="$QL_BASE"

# Read line by line and never with `for s in $QL_GIT_OUT`: word splitting
# turns one spec path containing a space into two paths that exist
# nowhere, each skipped by the `-f` test below — an implemented spec
# dropped in silence, which for a gate is the same failure as reading
# nothing at all. The accumulator is newline-separated for the same
# reason, and a here-document rather than a pipe so the loop can write
# into it at all.
implemented=""
ql_git_read "git diff --name-only ${RANGE} -- 'work/specs/*.md'" \
  diff --name-only "$RANGE" -- 'work/specs/*.md'
while IFS= read -r s; do
  [ -n "$s" ] || continue

  # git quotes a path holding control characters or non-ASCII bytes. A
  # spec path never needs it, so a quoted one is a path this check cannot
  # parse, and an unparseable input is refused rather than skipped.
  case "$s" in
    '"'*)
      echo "cannot read the changed path ${s} — refusing rather than skipping it" >&2
      exit 3
      ;;
  esac

  [ -f "$s" ] || continue
  [ "$(ql_fm_field_in status < "$s")" = "implemented" ] || continue
  [ "$(git show "${BASE}:$s" 2>/dev/null | ql_fm_field_in status)" = "implemented" ] && continue
  implemented="${implemented}${s}"$'\n'
done <<TOUCHED
${QL_GIT_OUT}
TOUCHED

if [ -z "$implemented" ]; then
  echo "No spec reached 'implemented' — authoring change, deltas not applicable."
  exit 0
fi

ids=""
while IFS= read -r s; do
  [ -n "$s" ] || continue
  id=$(sed -n 's/^id: *//p' "$s" | head -n1)
  ids="${ids:+${ids},}${id}"
done <<IMPLEMENTED
${implemented}
IMPLEMENTED
echo "Checking ${ids}"
bash "$CHECK_DELTAS" "$ids" "$RANGE"
