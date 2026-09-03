#!/usr/bin/env bash
# check_recorded_approvals.sh — an approval recorded in the diff must be
# backed by a real review on the pull request.
#
# Usage: check_recorded_approvals.sh <diff-range> <owner/repo> <pr-number>
#   `gh` must be on PATH and authenticated (GH_TOKEN in CI; a stub in the
#   test suite).
#
# Two shapes of diff carry an approval the change itself recorded, and both
# are indistinguishable from an abuse without the forge:
#
#   born approved — a spec the change *adds* with `status: approved`. No
#     `-status: draft` line exists for check_state.sh's rule A to read; the
#     legitimate flip (recorded after a maintainer approved, by `writrun
#     approve` or a fork contributor's own hand per CONTRIBUTING.md) and
#     self-approval look identical.
#
#   edited under approval — a spec *modified* with no status move, final
#     status `approved`. Either the net result of a legal amendment
#     (approved -> draft -> re-approved within the same PR) or a silent
#     edit of content a human assented to.
#
# The difference in both cases is the review, and only the forge holds it:
# exit 0 only if the PR carries an approving review from someone with
# authority — the same associations `writrun approve` requires.

set -euo pipefail
RANGE="${1:?usage: check_recorded_approvals.sh <diff-range> <owner/repo> <pr-number>}"
REPO="${2:?usage: check_recorded_approvals.sh <diff-range> <owner/repo> <pr-number>}"
PR="${3:?usage: check_recorded_approvals.sh <diff-range> <owner/repo> <pr-number>}"

# Statuses are read from the front-matter block at the two ends of the
# range, never grepped out of the diff text — a spec body legitimately
# quotes `status:` lines at column 0 (this repository's own docs do), and
# a quoted line must neither trigger this check nor exempt a spec from it.
case "$RANGE" in
  *...*)
    left="${RANGE%%...*}"
    right="${RANGE##*...}"
    # The same rule as the diff below: a merge-base that could not be
    # computed is not a base of "nothing", it is an unanswered question.
    if ! BASE=$(git merge-base "${left:-HEAD}" "${right:-HEAD}" 2>&1); then
      echo "git merge-base ${left:-HEAD} ${right:-HEAD} failed:" >&2
      printf '%s\n' "$BASE" | head -n 2 >&2
      exit 3
    fi
    ;;
  *..*) BASE="${RANGE%%..*}" ;;
  *)    BASE="$RANGE" ;;
esac
fm_field() {
  awk -v f="$1" '
    NR == 1 { if ($0 != "---") exit; next }
    /^---$/ { exit }
    sub("^" f ": *", "") { sub(/[[:space:]]*$/, ""); print; exit }
  '
}

# git_read <label> <git-args...> — runs git and leaves its stdout in
# GIT_OUT. On failure it prints what git said and exits 3, because a
# check that could not read its input must never report the empty result
# as a clean one: `$(git … || true)` yields exactly the same empty string
# whether nothing matched or nothing ran, and two of these checks are
# gates (spec-0013).
#
# **Never call this inside a command substitution.** The `exit` would end
# only the subshell, and the caller would go on reading the empty value
# this exists to prevent — the very shape of the bug being removed.
GIT_OUT=""
git_read() {
  local label="$1" err
  shift
  err=$(mktemp "${TMPDIR:-/tmp}/writrun-git.XXXXXX")
  if ! GIT_OUT=$(git "$@" 2>"$err"); then
    echo "${label} failed:" >&2
    head -n 2 "$err" >&2
    rm -f "$err"
    exit 3
  fi
  rm -f "$err"
}

born=""
git_read "git diff --name-only --diff-filter=A ${RANGE} -- work/specs" \
  diff --name-only --diff-filter=A "$RANGE" -- 'work/specs/*.md'
for s in $GIT_OUT; do
  [ -f "$s" ] || continue
  st=$(fm_field status < "$s")
  [ "$st" = "approved" ] && born="$born $s"
done

edited=""
git_read "git diff --name-only --diff-filter=M ${RANGE} -- work/specs" \
  diff --name-only --diff-filter=M "$RANGE" -- 'work/specs/*.md'
for s in $GIT_OUT; do
  [ -f "$s" ] || continue
  st=$(fm_field status < "$s")
  [ "$st" = "approved" ] || continue
  # "No status move" means the base end already said approved — a real
  # move is rule A's or the flip's business, not this check's.
  old=$(git show "${BASE}:$s" 2>/dev/null | fm_field status || true)
  [ "$old" = "approved" ] || continue
  edited="$edited $s"
done

if [ -z "$born" ] && [ -z "$edited" ]; then
  echo "No approval recorded by this change needs verifying."
  exit 0
fi

n=$(gh api "repos/${REPO}/pulls/${PR}/reviews" \
  --paginate \
  --jq '[.[] | select(.state == "APPROVED")
             | select(.author_association == "OWNER"
                   or .author_association == "MEMBER"
                   or .author_association == "COLLABORATOR")] | length' \
  | awk '{ sum += $1 } END { print sum + 0 }')

if [ "$n" -gt 0 ]; then
  echo "An authorized approving review exists; accepted for:${born}${edited}"
  exit 0
fi

if [ -n "$born" ]; then
  echo "FORBIDDEN: these specs enter the change already 'approved':" >&2
  echo " ${born}" >&2
fi
if [ -n "$edited" ]; then
  echo "FORBIDDEN: these approved specs were edited with no status move:" >&2
  echo " ${edited}" >&2
  echo "Content under an approval never changes silently — amend" >&2
  echo "through draft (docs/product/stage-1-tasks-and-specs/conflicts.md)." >&2
fi
echo "No approving review from an owner, member, or collaborator" >&2
echo "exists on this pull request. draft -> approved is a human gate" >&2
echo "(docs/product/stage-1-tasks-and-specs/gates.md) — the transition, or the" >&2
echo "amended content, is accepted once the review exists." >&2
exit 1
