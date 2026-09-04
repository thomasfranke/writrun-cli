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

# The approved→draft move is read from the front matter at the range's
# two ends, never grepped out of the diff text — a body edit swapping
# quoted `status:` lines is not an amendment.
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

# This one mutates and otherwise always exits 0, so a swallowed failure
# does not merely misreport — it records no approval at all, silently, on
# a merge that granted one.
git_read "git diff --name-only --diff-filter=A ${RANGE} -- work/specs" \
  diff --name-only --diff-filter=A "$RANGE" -- 'work/specs/*.md'
added="$GIT_OUT"

redrafted=""
git_read "git diff --name-only --diff-filter=M ${RANGE} -- work/specs" \
  diff --name-only --diff-filter=M "$RANGE" -- 'work/specs/*.md'
for s in $GIT_OUT; do
  [ -f "$s" ] || continue
  [ "$(fm_field status < "$s")" = "draft" ] || continue
  [ "$(git show "${BASE}:$s" 2>/dev/null | fm_field status)" = "approved" ] || continue
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
