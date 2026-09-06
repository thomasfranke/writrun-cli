#!/usr/bin/env bash
# flip_approved_specs.sh — records an approval into the spec files that
# earned it: flips `status: draft` to `status: approved`.
#
# Usage: flip_approved_specs.sh <diff-range>
#
# Two kinds of spec qualify, and only these:
#
#   added — a spec the change introduces. An implementation PR also touches
#     specs, and a spec deliberately parked in draft on main must not be
#     approved by a change that merely edited it — modified files do not
#     qualify by themselves.
#
#   re-drafted — an amendment returned an approved spec to draft in this
#     same change (special flow: a spec changes after its approval). The
#     review approving the PR is assent to the amended content, so it flips
#     back too; net status on merge is unchanged and the diff that lands is
#     just the amended body.
#
# Only the front-matter block is touched — a spec in this methodology's own
# repo can legitimately quote `status: draft` in its body.
#
# Mutates the working tree only and prints one "approved <file>" line per
# flip; committing is the caller's job. Always exits 0.

set -euo pipefail
RANGE="${1:?usage: flip_approved_specs.sh <diff-range>}"

. "$(dirname "$0")/queue_lib.sh"

# The approved→draft move is read from the front matter at the range's
# two ends, never grepped out of the diff text — a body edit swapping
# quoted `status:` lines is not an amendment.
ql_range_ends "$RANGE"
BASE="$QL_BASE"

# This one mutates and otherwise always exits 0, so a swallowed failure
# does not merely misreport — it records no approval at all, silently, on
# a merge that granted one.
ql_git_read "git diff --name-only --diff-filter=A ${RANGE} -- 'work/specs/*.md'" \
  diff --name-only --diff-filter=A "$RANGE" -- 'work/specs/*.md'
added="$QL_GIT_OUT"

redrafted=""
ql_git_read "git diff --name-only --diff-filter=M ${RANGE} -- 'work/specs/*.md'" \
  diff --name-only --diff-filter=M "$RANGE" -- 'work/specs/*.md'
for s in $QL_GIT_OUT; do
  [ -f "$s" ] || continue
  [ "$(ql_fm_field_in status < "$s")" = "draft" ] || continue
  [ "$(git show "${BASE}:$s" 2>/dev/null | ql_fm_field_in status)" = "approved" ] || continue
  redrafted="$redrafted $s"
done

for s in $added $redrafted; do
  [ -f "$s" ] || continue
  fm_end=$(awk 'NR > 1 && /^---$/ { print NR; exit }' "$s")
  [ -n "$fm_end" ] || continue
  awk -v end="$fm_end" \
    'NR <= end && $0 == "status: draft" { found = 1 } END { exit !found }' \
    "$s" || continue
  # -i.bak, then rm: portable across GNU sed and macOS sed.
  sed -i.bak "1,${fm_end} s/^status: draft$/status: approved/" "$s"
  rm -f "${s}.bak"
  echo "approved $s"
done
