#!/usr/bin/env bash
# check_unique_ids.sh — a queue id this change mints must be one nobody
# else already claims.
#
# Usage: check_unique_ids.sh <diff-range> <owner/repo> <pr-number>
#   `gh` must be on PATH and authenticated (GH_TOKEN in CI; a stub in the
#   test suite).
#
# An id is unique across the queue *and* across every open pull request
# (docs/technical/README.md#task-schema). The generator mints from one
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

TAB=$(printf '\t')

# gh defaults to 30 open pull requests, and a silently truncated list
# reports a claimed id as free — the exact lie this check exists to
# prevent. The limit is raised, and *hitting* it is reported rather than
# passed off as a complete answer.
PR_FETCH_LIMIT=200

# The left end of the range, which is the branch this change is measured
# against — `A...B`, `A..B`, and a bare ref all name it differently.
case "$RANGE" in
  *...*)
    left="${RANGE%%...*}"
    right="${RANGE##*...}"
    if ! BASE=$(git merge-base "${left:-HEAD}" "${right:-HEAD}" 2>&1); then
      echo "git merge-base ${left:-HEAD} ${right:-HEAD} failed:" >&2
      printf '%s\n' "$BASE" | head -n 2 >&2
      exit 3
    fi
    ;;
  *..*) BASE="${RANGE%%..*}" ;;
  *)    BASE="$RANGE" ;;
esac

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

mine=""
git_read "git diff --name-only --diff-filter=A ${RANGE} -- work/tasks work/specs work/reports" \
  diff --name-only --diff-filter=A "$RANGE" -- 'work/tasks/*.md' 'work/specs/*.md' 'work/reports/*.md'
while IFS= read -r f; do
  [ -n "$f" ] || continue
  k=$(queue_id "$f")
  [ -n "$k" ] || continue
  mine="${mine}${k}${TAB}${f}"$'\n'
done <<EOF
$GIT_OUT
EOF

if [ -z "$mine" ]; then
  echo "This change adds no queue file — nothing claims an id."
  exit 0
fi

# --- what the base branch already holds -----------------------------------

held=""
git_read "git ls-tree -r --name-only ${BASE} -- work/tasks work/specs work/reports" \
  ls-tree -r --name-only "$BASE" -- work/tasks work/specs work/reports
while IFS= read -r f; do
  [ -n "$f" ] || continue
  k=$(queue_id "$f")
  [ -n "$k" ] || continue
  held="${held}${k}${TAB}${f}"$'\n'
done <<EOF
$GIT_OUT
EOF

# --- what other open pull requests claim ----------------------------------
#
# Per pull request, because only the API's file list carries `status`, and
# without it a modification would read as a claim. The pull request being
# checked is skipped: its own additions are the ones under examination and
# must not collide with themselves.

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
        --jq '.[] | select(.status == "added") | .filename' 2>/dev/null || true)
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

collisions=0
while IFS="$TAB" read -r kind num file; do
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
  echo "(docs/technical/README.md#task-schema). A number a branch has not" >&2
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
