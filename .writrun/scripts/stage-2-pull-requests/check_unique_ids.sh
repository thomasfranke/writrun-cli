#!/usr/bin/env bash
# check_unique_ids.sh — a queue id this change mints must be one nobody
# else already claims.
#
# Usage: check_unique_ids.sh <diff-range> <owner/repo> <pr-number>
#   `gh` must be on PATH and authenticated (GH_TOKEN in CI; a stub in the
#   test suite).
#
# An id is unique across the queue *and* across every open pull request
# (docs/technical/schemas/task.md#task-schema). The generator mints from one
# branch's view, so two branches cut from the same authority branch both
# see the same highest id and both take the next one — and until now
# nothing rejected it: the collision surfaced at the second merge, after
# both changes were written.
#
# Two claimants, and the message must name which, because the fix is
# renumbering and that is only obvious once you know what you collided
# with:
#
#   the base branch — the id is already an id. Whoever holds it holds it;
#     the change renumbers.
#
#   another open pull request — neither number is an id yet (identity
#     begins at the merge), so either side could renumber. This one is
#     the one being checked, so this one is the one told.
#
# Only files the range **adds** claim anything. Modifying a queue file is
# not a claim: the id it carries is already on the base branch, where it
# belongs to whoever put it there.
#
# Ids are read from the filename prefix, never from front matter — the
# subject slug that follows is not part of identity, and a file whose two
# spellings disagree is check_front_matter.sh's business, not this one's.
#
# Best-effort on the forge half, deliberately: no `gh`, no network, or no
# auth still checks the base branch — the deterministic half — and says
# its view was narrow rather than reporting a clean pass it cannot back.
# A silently narrower scan is how the collision this check exists for
# happened in the first place.
#
# Exit codes: 0 no collision; 1 a collision, named; 3 usage error.
#
# Portable bash 3.2, POSIX awk/sed — no gawk extensions, no associative
# arrays. See the standing rule in docs/technical/decisions/.

set -euo pipefail

RANGE="${1:?usage: check_unique_ids.sh <diff-range> <owner/repo> <pr-number>}"
REPO="${2:?usage: check_unique_ids.sh <diff-range> <owner/repo> <pr-number>}"
PR="${3:?usage: check_unique_ids.sh <diff-range> <owner/repo> <pr-number>}"

# `ql_row_fields`, the reader every tab-delimited row in this repository
# goes through — the rows below are assembled here and never leave, which
# is exactly why a private parse was tempting and why it is refused: the
# collapse is a property of `read`, not of where the row came from. Also
# `ql_range_ends` and `ql_git_read`, for the reason the lib's header
# gives: private copies of these drifted before.
. "$(dirname "$0")/queue_lib.sh"

TAB=$(printf '\t')

# gh defaults to 30 open pull requests, and a silently truncated list
# reports a claimed id as free — the exact lie this check exists to
# prevent. The limit is raised, and *hitting* it is reported rather than
# passed off as a complete answer.
PR_FETCH_LIMIT=200

# The left end of the range, which is the branch this change is measured
# against — `A...B`, `A..B`, and a bare ref all name it differently.
ql_range_ends "$RANGE"
BASE="$QL_BASE"

# queue_id <path> — "<kind><TAB><number>" for a queue file, where kind is
# task, spec or report and number is the id's digits with leading zeros
# dropped. Prints nothing for anything else: a README, a path outside
# work/, or a filename whose prefix is malformed.
#
# The three kinds number independently — report-0001 and task-0001 are
# not a collision — so the kind travels with the number and every
# comparison below is on the pair.
queue_id() {
  printf '%s' "$1" | tr '[:upper:]' '[:lower:]' | awk -F/ -v tab="$TAB" '
    NF != 3 || $1 != "work" { next }
    $2 == "tasks"   { kind = "task" }
    $2 == "specs"   { kind = "spec" }
    $2 == "reports" { kind = "report" }
    kind == ""      { next }
    {
      name = $3
      sub(/\.md$/, "", name)
      if (name !~ "^" kind "-[0-9]+(-.+)?$") next
      n = name
      sub("^" kind "-0*", "", n)
      sub(/-.*$/, "", n)
      if (n == "") n = "0"
      print kind tab (n + 0)
    }'
}

# --- what this change claims ---------------------------------------------

# **A rename is a claim, and a release.** A queue filename is an id plus
# a subject slug, so renumbering a file changes its path and git pairs it
# as a rename rather than a modification — invisible to `--diff-filter=A`,
# which is how a change that frees an id and then claims it was refused
# for colliding with itself. So the claim side reads additions and
# renames both: a rename's destination is a claim like any other, and its
# source is a release, subtracted from the base below.
mine=""
released=""
ql_git_read "git diff --name-status --diff-filter=AR ${RANGE} -- 'work/tasks/*.md' 'work/specs/*.md' 'work/reports/*.md'" \
  diff --name-status --diff-filter=AR "$RANGE" -- 'work/tasks/*.md' 'work/specs/*.md' 'work/reports/*.md'
while IFS= read -r row; do
  [ -n "$row" ] || continue
  # A status row is status<TAB>path, or status<TAB>src<TAB>dst for a
  # rename. A row with no tab is not one of ours; skipping it is the same
  # answer the filename-shape guard gives a path that is not a queue file.
  case "$row" in *"$TAB"*) ;; *) continue ;; esac
  status=${row%%"$TAB"*}
  rest=${row#*"$TAB"}
  case "$status" in
    # R comes scored — R100, R087 — and the score is the heuristic's
    # confidence, not part of the verdict.
    R*)
      ql_row_fields 2 "$rest" || continue
      src="$QL_F1"; dst="$QL_F2"
      k=$(queue_id "$src")
      [ -n "$k" ] && released="${released}${k}"$'\n'
      ;;
    A)
      src=""; dst="$rest"
      ;;
    *) continue ;;
  esac
  k=$(queue_id "$dst")
  [ -n "$k" ] || continue
  mine="${mine}${k}${TAB}${dst}"$'\n'
done <<EOF
$QL_GIT_OUT
EOF

if [ -z "$mine" ]; then
  echo "This change adds no queue file — nothing claims an id."
  exit 0
fi

# --- what the base branch already holds -----------------------------------

held=""
ql_git_read "git ls-tree -r --name-only ${BASE} -- work/tasks work/specs work/reports" \
  ls-tree -r --name-only "$BASE" -- work/tasks work/specs work/reports
while IFS= read -r f; do
  [ -n "$f" ] || continue
  k=$(queue_id "$f")
  [ -n "$k" ] || continue
  held="${held}${k}${TAB}${f}"$'\n'
done <<EOF
$QL_GIT_OUT
EOF

# An id whose only holder on the base is a file this change renamed away
# is not held any more — the other half of the same blindness, and the
# half that makes a renumber and the claim it frees one change instead of
# two. Subtracted by id, never by path: the point of a renumber is that
# the path changed, and a rename that moves only the slug releases and
# claims the same id, which cancels exactly as it should.
if [ -n "$released" ]; then
  kept=""
  while IFS= read -r row; do
    ql_row_fields 3 "$row" || continue
    if printf '%s' "$released" | awk -F"$TAB" -v k="$QL_F1" -v n="$QL_F2" \
         '$1 == k && $2 == n { found = 1 } END { exit !found }'; then
      continue
    fi
    kept="${kept}${row}"$'\n'
  done <<EOF
$held
EOF
  held="$kept"
fi

# --- what other open pull requests claim ----------------------------------
#
# Per pull request, because only the API's file list carries `status`, and
# without it a modification would read as a claim — and a rename would
# read as nothing at all. The destination of a rename is asked for and
# the source is not: what this reader wants to know is which ids somebody
# else has taken, and a pull request that renames a file away has not
# released that id to anyone. It still holds it until it merges.
#
# The pull request being checked is skipped: its own additions are the
# ones under examination and must not collide with themselves.

claimed=""
forge_view="none"
truncated=""

if command -v gh >/dev/null 2>&1; then
  if numbers=$(gh pr list --repo "$REPO" --state open \
        --limit "$PR_FETCH_LIMIT" --json number \
        --jq '.[].number' 2>/dev/null); then
    forge_view="gh"
    count=0
    for n in $numbers; do
      count=$((count + 1))
      if [ "$n" = "$PR" ]; then continue; fi
      files=$(gh api "repos/${REPO}/pulls/${n}/files" --paginate \
        --jq '.[] | select(.status == "added" or .status == "renamed") | .filename' 2>/dev/null || true)
      while IFS= read -r f; do
        [ -n "$f" ] || continue
        k=$(queue_id "$f")
        [ -n "$k" ] || continue
        claimed="${claimed}${k}${TAB}#${n}"$'\n'
      done <<EOF
$files
EOF
    done
    if [ "$count" -ge "$PR_FETCH_LIMIT" ]; then truncated=yes; fi
  fi
fi

# --- the verdict ----------------------------------------------------------

# The rows go through `ql_row_fields`, never `IFS="$TAB" read`. These
# three fields are assembled here rather than read off a forge, but the
# parse is the same one and the collapse is the same collapse — a tab is
# IFS whitespace, so an empty field would take the next field's place and
# the verdict would name the wrong file. One reader for every row this
# repository splits on tabs is what keeps the class closed; the hazard is
# written out in the helper's header.
collisions=0
while IFS= read -r row; do
  ql_row_fields 3 "$row" || continue
  kind="$QL_F1"; num="$QL_F2"; file="$QL_F3"
  [ -n "$kind" ] || continue

  other=$(printf '%s' "$held" | awk -F"$TAB" -v k="$kind" -v n="$num" \
    '$1 == k && $2 == n { print $3; exit }')
  if [ -n "$other" ]; then
    echo "COLLISION: ${file} claims an id the base branch already holds" >&2
    echo "  held there by ${other}" >&2
    collisions=$((collisions + 1))
    continue
  fi

  other=$(printf '%s' "$claimed" | awk -F"$TAB" -v k="$kind" -v n="$num" \
    '$1 == k && $2 == n { print $3; exit }')
  if [ -n "$other" ]; then
    echo "COLLISION: ${file} claims an id open pull request ${other} also adds" >&2
    collisions=$((collisions + 1))
  fi
done <<EOF
$mine
EOF

if [ "$collisions" -gt 0 ]; then
  echo "" >&2
  echo "An id is unique across the queue and every open pull request" >&2
  echo "(docs/technical/schemas/task.md#task-schema). A number a branch has not" >&2
  echo "merged is not yet an id, so renumbering costs nothing — renumber" >&2
  echo "this change's files, and the reference to them, above the ids" >&2
  echo "named here." >&2
  exit 1
fi

if [ "$forge_view" = "none" ]; then
  echo "No id collides with the base branch."
  echo "The forge did not answer, so open pull requests were not consulted" >&2
  echo "— an id another one already claims would not have been seen." >&2
  exit 0
fi

if [ -n "$truncated" ]; then
  echo "No id collides with the base branch or the open pull requests read."
  echo "The open-pull-request list hit its ${PR_FETCH_LIMIT} limit, so it may" >&2
  echo "be incomplete." >&2
  exit 0
fi

echo "No id collides with the base branch or any open pull request."
