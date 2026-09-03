#!/usr/bin/env bash
# stamp_task_dates.sh — writes the machinery's two dates onto the task
# files a merge just moved.
#
# Usage: stamp_task_dates.sh <diff-range> <timestamp>
#   The timestamp is the merge commit's own, never today's — the caller
#   reads it with
#   `TZ=UTC0 git show -s --format=%cd --date=format:%Y-%m-%dT%H:%M:%SZ`.
#
# A task carries four dates and **who writes each is part of the
# contract** (docs/product/pipeline.md#flows-and-statuses). `created` and
# `completed` are a person's, written on the branch. `queued` and
# `merged` are these: a hand-written date cannot honestly record a merge,
# because it would have to be typed before the event it describes and
# would be wrong by however long review takes.
#
#   queued  — the range **adds** the task file. The merge is what put it
#             in the queue.
#   merged  — the range writes the task's **`completed` date** — the
#             worker's declaration of finishing. The merge is what took
#             its work; the status flip it implies is
#             record_task_status.sh's, in the same recording commit.
#
# The transition is read from the front matter at the range's two ends,
# never from the diff text — the rule check_state.sh and
# flip_approved_specs.sh already follow, because a task body may quote a
# `status:` line at column 0 and this repository's own docs do.
#
# **Only a `null` field is written.** A date already recorded is history,
# and a second merge touching the same file must not restate it.
#
# Mutates the working tree only and prints one "stamped" line per field
# written; committing is the caller's job. Always exits 0, except 3 for a
# usage error — a merge that moved no task is the ordinary case, not a
# failure.
#
# Portable bash 3.2, POSIX awk/sed — no gawk extensions, no associative
# arrays. See the standing rule in docs/technical/decisions/.

set -euo pipefail

RANGE="${1:?usage: stamp_task_dates.sh <diff-range> <timestamp>}"
STAMP="${2:?usage: stamp_task_dates.sh <diff-range> <timestamp>}"

printf '%s' "$STAMP" \
  | grep -qE '^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$' \
  || { echo "'$STAMP' is not an RFC 3339 UTC timestamp spelled with Z" >&2; exit 3; }

# The left end of the range, which is what the file is measured against —
# `A...B`, `A..B`, and a bare ref all name it differently.
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

# fm_field <field> — stdin is a task file; prints the field's value from
# the front-matter block alone, so a body line spelling `status:` at
# column 0 is not read as one.
fm_field() {
  awk -v f="$1" '
    NR == 1 { if ($0 != "---") exit; next }
    /^---$/ { exit }
    sub("^" f ": *", "") { sub(/[[:space:]]*$/, ""); print; exit }
  '
}

# stamp <file> <field> — writes STAMP into <field> when it holds null.
stamp() {
  local f="$1" field="$2" cur
  cur=$(fm_field "$field" < "$f")
  if [ -z "$cur" ]; then
    echo "SKIPPED: ${f} has no '${field}' field — the canonical check will name it" >&2
    return 0
  fi
  [ "$cur" = "null" ] || return 0
  awk -v field="$field" -v stamp="$STAMP" '
    NR == 1 && $0 == "---" { infm = 1; print; next }
    infm && /^---$/        { infm = 0; print; next }
    infm && index($0, field ":") == 1 { print field ": " stamp; next }
    { print }
  ' "$f" > "${f}.tmp" && mv "${f}.tmp" "$f"
  echo "stamped ${field} on ${f}"
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

# is_dated <file> — the task carries a written completed date.
is_dated() {
  local v
  v=$(fm_field completed < "$1")
  [ -n "$v" ] && [ "$v" != "null" ]
}

# A task added by the range entered the queue at this merge — and if it
# arrived with its completed date already written (tracked work shipped
# with its own change), the same merge took its work too.
git_read "git diff --name-only --diff-filter=A ${RANGE} -- work/tasks" \
  diff --name-only --diff-filter=A "$RANGE" -- 'work/tasks/*.md'
while IFS= read -r f; do
  [ -n "$f" ] || continue
  [ -f "$f" ] || continue
  stamp "$f" queued
  if is_dated "$f"; then stamp "$f" merged; fi
done <<EOF
$GIT_OUT
EOF

# A task the range modified is stamped only for the declaration it
# actually carried: a date already written at the base is history, not a
# finishing here.
git_read "git diff --name-only --diff-filter=M ${RANGE} -- work/tasks" \
  diff --name-only --diff-filter=M "$RANGE" -- 'work/tasks/*.md'
while IFS= read -r f; do
  [ -n "$f" ] || continue
  [ -f "$f" ] || continue
  is_dated "$f" || continue
  was=$(git show "${BASE}:${f}" 2>/dev/null | fm_field completed || true)
  if [ -n "$was" ] && [ "$was" != "null" ]; then continue; fi
  stamp "$f" merged
done <<EOF
$GIT_OUT
EOF

exit 0
