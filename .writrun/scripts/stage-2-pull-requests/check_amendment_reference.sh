#!/usr/bin/env bash
# check_amendment_reference.sh — an amendment cut under an open pull
# request names the pull request it suspends.
#
# Usage: check_amendment_reference.sh <diff-range> <owner/repo> <pr-number>
#   The PR body arrives via $PR_BODY — through the environment, never
#   inline interpolation: it is attacker-controlled text on a fork PR.
#   `gh` must be on PATH and authenticated (GH_TOKEN in CI; a stub in the
#   test suite).
#
# Returning an approved spec to `draft` suspends the task riding it: a
# spec whose approval is in question authorizes nothing, so the work
# waits until the amendment merges
# (docs/product/stage-2-pull-requests/statuses.md#an-amendment-under-an-open-pull-request).
# No status moves and none should — flight belongs to the task's own pull
# request's events — so the only record of the pause is the relation
# between the two pull requests, and the forge is where relations between
# pull requests live.
#
# The requirement is symmetric in the docs and asymmetric here on
# purpose: the suspended pull request predates the amendment and cannot
# have named it at open, so its side stays convention. This gate holds
# the side that can be held.
#
# Only a task **in flight** is suspended. Amending the spec of a `ready`,
# `backlog` or `blocked` task is the ordinary pre-implementation
# amendment flow, which this check must leave exactly as it was — the
# whole point of that flow is that it costs nothing.
#
# Best-effort on the forge half, deliberately, the same contract
# check_unique_ids.sh states: without an answer the number to name cannot
# be known, so the check says its view was narrow rather than failing a
# change over a question it could not ask.
#
# Exit codes: 0 nothing owed, or the reference is present; 1 an
# amendment suspends an in-flight task and does not name its pull
# request; 3 usage error, or a range that could not be read.
#
# Portable bash 3.2, POSIX awk/sed — no gawk extensions, no associative
# arrays. See the standing rule in docs/technical/decisions/.

set -euo pipefail

RANGE="${1:?usage: check_amendment_reference.sh <diff-range> <owner/repo> <pr-number>}"
REPO="${2:?usage: check_amendment_reference.sh <diff-range> <owner/repo> <pr-number>}"
PR="${3:?usage: check_amendment_reference.sh <diff-range> <owner/repo> <pr-number>}"

. "$(dirname "$0")/queue_lib.sh"

TAB=$(printf '\t')

# gh defaults to 30 open pull requests, and a silently truncated list
# would report a suspended task's pull request as nonexistent — the check
# passing precisely where it should fire.
PR_FETCH_LIMIT=200

# The left end of the range — the branch this change is measured against.
ql_range_ends "$RANGE"
BASE="$QL_BASE"

# --- what this change returns to draft ------------------------------------
#
# Read from the front matter at the range's two ends, never grepped out of
# the diff text: a spec body quoting `status: draft` at column 0 is prose,
# not an amendment.

ql_git_read "git diff --name-only ${RANGE} -- 'work/specs/*.md'" \
  diff --name-only "$RANGE" -- 'work/specs/*.md'

# Read line by line and never with `for s in $QL_GIT_OUT`: word splitting
# turns one path containing a space into two paths that exist nowhere,
# each skipped by the `-f` test below — an amendment dropped in silence,
# which for a gate is the same failure as reading nothing at all.
touched="$QL_GIT_OUT"

amended=""
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
  [ "$(ql_fm_field status "$s")" = "draft" ] || continue

  # What the spec was at the base, through ql_git_read for the reason its
  # own comment gives: `$(git … ) || was=""` cannot tell "this spec is
  # new on the branch" from "git could not be read", and the second one
  # silently becomes "nothing is suspended" — the gate passing exactly
  # where it must fire. `ls-tree` separates the two: absent from a tree
  # it could read is an answer, a tree it could not read is not.
  ql_git_read "git ls-tree ${BASE} -- ${s}" ls-tree "$BASE" -- "$s"
  if [ -z "$QL_GIT_OUT" ]; then
    continue    # not in the base tree: a new spec, never an amendment
  fi
  ql_git_read "git show ${BASE}:${s}" show "${BASE}:$s"
  was=$(printf '%s\n' "$QL_GIT_OUT" | ql_fm_field_in status)

  case "$was" in approved|implemented) ;; *) continue ;; esac
  amended="${amended}$(ql_fm_field id "$s")"$'\n'
done <<EOF
$touched
EOF

if [ -z "$amended" ]; then
  echo "This change returns no approved spec to draft — nothing is suspended."
  exit 0
fi

# --- which tasks that suspends --------------------------------------------
#
# Only tasks in flight. The task files are read from the checkout: an
# amendment touches work/specs and nothing else, so the two ends agree.

suspended=""    # task-id<TAB>spec-id
for spec in $amended; do
  [ -n "$spec" ] || continue
  for t in work/tasks/*.md; do
    [ -f "$t" ] || continue
    case "$(basename "$t")" in README.md|readme.md) continue ;; esac
    st=$(ql_fm_field status "$t")
    case "$st" in in-progress|in-review) ;; *) continue ;; esac
    refs=$(ql_fm_field spec_ref "$t" | tr -d '[]' | tr ',' ' ')
    for r in $refs; do
      [ "$r" = "$spec" ] || continue
      suspended="${suspended}$(ql_fm_field id "$t")${TAB}${spec}"$'\n'
    done
  done
done

if [ -z "$suspended" ]; then
  echo "The amended specs belong to no task in flight — the ordinary"
  echo "pre-implementation amendment flow, which owes no reference."
  exit 0
fi

# --- the pull request each suspended task is riding -----------------------

pr_lines=""
forge_view="none"
truncated=""
if command -v gh >/dev/null 2>&1; then
  if pr_lines=$(gh pr list --repo "$REPO" --state open \
        --limit "$PR_FETCH_LIMIT" --json number,headRefName,title \
        --jq '.[] | "\(.number)\t\(.headRefName)\t\(.title)"' 2>/dev/null); then
    forge_view="gh"
    # An `if`, not a trailing `&&`: as the last statement of this block it
    # would hand its own false to the enclosing `if`, and `set -e` would
    # end the run on the ordinary case of a list that fits.
    if [ "$(printf '%s\n' "$pr_lines" | grep -c .)" -ge "$PR_FETCH_LIMIT" ]; then
      truncated=yes
    fi
  fi
fi

if [ "$forge_view" = "none" ]; then
  echo "An amendment here suspends a task in flight, but the forge did not" >&2
  echo "answer, so the pull request it must name could not be identified." >&2
  echo "Check by hand that the body references it." >&2
  exit 0
fi

# --- the rows this check can read -----------------------------------------
#
# The carried set of every open pull request, resolved once. Once,
# because the lookup below runs per suspended task and the answer is the
# same every time — and because a row skipped for claiming too much owes
# one notice per pull request, not one per task that asked about it.
#
# Another pull request's over-ceiling claim is its author's fault, and
# failing this act over it would let one hostile title stop every
# amendment. The row is skipped, with the notice on stderr, and the fact
# that one was skipped is remembered: what the lookup can no longer
# answer, the verdict must not report as an answer.
#
# The rows go through `ql_row_fields`, never `IFS="$TAB" read`: a tab is
# IFS whitespace, so an empty field would vanish and shift every field
# after it — a head branch this check cannot read as a task, and a
# suspended pull request reported as one that does not exist. The helper's
# header carries the whole hazard.
readable=""
skipped=""
while IFS= read -r row; do
  ql_row_fields 3 "$row" || continue
  num="$QL_F1"; branch="$QL_F2"; ptitle="$QL_F3"
  [ -n "$num" ] || continue
  carried=$(ql_carried_of "$branch" "${ptitle:-}")
  case "$carried" in
    over-ceiling:*)
      echo "pull request #${num} claims ${carried#over-ceiling:} distinct tasks — over the ceiling of ${QL_CARRIED_MAX}; its row is skipped" >&2
      skipped="${skipped}#${num} "
      continue
      ;;
  esac
  readable="${readable}${num}${TAB}${carried}"$'\n'
done <<EOF
$pr_lines
EOF

# task_pr <task-id> — the number of the open pull request working that
# task, or nothing. This pull request is skipped: the amendment is not
# the work, and a queue/ branch carries no id anyway.
task_pr() {
  local want="$1" num carried c row
  while IFS= read -r row; do
    ql_row_fields 2 "$row" || continue
    num="$QL_F1"; carried="$QL_F2"
    [ -n "$num" ] || continue
    [ "$num" = "$PR" ] && continue
    for c in $carried; do
      if [ "$(ql_task_num "$c")" = "$(ql_task_num "$want")" ]; then
        printf '%s' "$num"; return 0
      fi
    done
  done <<EOF
$readable
EOF
  return 0
}

# --- the verdict ----------------------------------------------------------
#
# `#N` is what a person writes and what the forge renders as a link; the
# full URL is the other spelling, and both count. Nothing else does — the
# task id alone would be satisfied by the sentence that names the spec.

missing=0
seen=""
while IFS= read -r row; do
  ql_row_fields 2 "$row" || continue
  task="$QL_F1"; spec="$QL_F2"
  [ -n "$task" ] || continue
  case "$seen" in *" $task "*) continue ;; esac
  seen="${seen} ${task} "

  num=$(task_pr "$task")
  if [ -z "$num" ] && [ -n "$skipped" ]; then
    # A row was skipped, so "no open pull request works it" is not
    # something this check knows — the skipped one may be exactly it.
    # Said rather than passed off as a clean answer, and still not
    # failed: the best-effort contract above holds for a question that
    # cannot be asked, however it came to be unaskable.
    echo "${task} reads as in flight and no readable pull request works it." >&2
    echo "Skipped for claiming over the ceiling: ${skipped% }. The one to name" >&2
    echo "may be among them. Check by hand that this body references it." >&2
    continue
  fi
  if [ -z "$num" ]; then
    echo "${task} reads as in flight but no open pull request works it —"
    echo "nothing to name. Its flight state is stale, not this change's business."
    continue
  fi

  if printf '%s' "${PR_BODY:-}" \
    | grep -qE "(#${num}([^0-9]|\$))|(/pull/${num}([^0-9]|\$))"; then
    echo "${task} is suspended by this amendment, and #${num} is named."
    continue
  fi

  echo "This change returns ${spec} to draft while ${task} rides #${num}." >&2
  echo "That suspends the task: until this merges, the work cannot advance," >&2
  echo "and nothing but these two pull requests records the wait." >&2
  echo "Name it — add a line to this pull request's body:" >&2
  echo "  Suspends #${num} — ${task} waits on this amendment." >&2
  missing=$((missing + 1))
done <<EOF
$suspended
EOF

if [ "$missing" -gt 0 ]; then
  echo "" >&2
  echo "The two pull requests name each other" >&2
  echo "(docs/product/stage-2-pull-requests/statuses.md#an-amendment-under-an-open-pull-request)." >&2
  exit 1
fi

if [ -n "$truncated" ]; then
  echo "The open-pull-request list hit its ${PR_FETCH_LIMIT} limit, so it may" >&2
  echo "be incomplete." >&2
fi
