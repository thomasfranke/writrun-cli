#!/usr/bin/env bash
# record_task_status.sh — the merge's half of the transition machine:
# moves every task the merge affected to where the queue now says it
# belongs, on the working tree, in the same recording commit as the spec
# flips and date stamps (product/stage-2-pull-requests/statuses.md).
#
# Usage: record_task_status.sh <diff-range> [carried-task-id...]
#   The carried ids are the tasks whose work this merge took. Passed as
#   arguments, or — when none are given and the caller exported
#   PR_HEAD_REF / PR_TITLE — derived from the head branch's own id
#   (task/NNNN-*) and every [TASK-NNNN] tag leading the title, so a
#   merge whose diff never touched a task file still lands it, and the
#   workflow stays wiring with no parsing of its own. A claim over
#   QL_CARRIED_MAX comes back as the helper's over-ceiling sentinel:
#   the carried ids are dropped with the refusal printed, the
#   range-derived scope still records, and the exit stays 0 — a merged
#   close fires no second event, and the Commit step behind this is
#   success-gated, so a red exit would lose the range's writes.
#
# Per task in scope — added or modified by the range, referenced by a
# spec the range touched, or carried — one of three moves, in this
# order:
#
#   done    its `completed` date is written: the worker declared
#           finishing and the merge took the work. `taken_by` stays —
#           the record of who completed it.
#   land    carried, in flight, no declaration: one spec of several is
#           work taken, not finished — to ready, or backlog if any of
#           its specs is draft, derived now, never assumed. taken_by
#           cleared. **Only a carried task lands**: a merge that merely
#           touches an in-flight task's spec — an amendment landing
#           while the work rides another, still-open pull request —
#           says nothing about that work, and a task it pulled back to
#           ready would read as free while somebody's PR is open. The
#           in-flight state belongs to the task's own pull request's
#           events, and to no other merge.
#   settle  it rests in backlog or ready: re-derived from its specs as
#           they now stand — the approval this merge recorded, or the
#           amendment it landed, may have moved the answer. An empty
#           spec_ref settles to ready: no approval event exists for it,
#           and backlog must not be a trap.
#
# blocked and dropped are a person's and are never touched. A task
# already where it belongs writes nothing.
#
# Output contract: one `moved <file>: <old> -> <new>` line per write on
# stdout, and — when $GITHUB_OUTPUT is set — `changed=0|1`,
# `tasks=<id ...>` and `scope=<id ...>` appended there, so a workflow
# reads results instead of scraping prose. Always exits 0, except 3 when
# git cannot read the range — an unreadable range is not an empty one.
#
# `tasks` and `scope` answer different questions and a caller wants both.
# `tasks` is what this run *moved*; `scope` is every task the merge put
# in front of it, moved or not. A task the merge created and settled
# where it already belonged writes no `moved` line and still owes its
# mirror a label, so the projection reads `scope` — deriving it a second
# time from the range would be a second chance to disagree with this
# one (docs/product/stage-3-github-issues/labels.md).
#
# Portable bash 3.2, POSIX awk/sed — no gawk extensions. See the
# standing rule in docs/technical/decisions/.

set -euo pipefail

. "$(dirname "$0")/queue_lib.sh"

RANGE="${1:?usage: record_task_status.sh <diff-range> [carried-task-id...]}"
shift
CARRIED_IDS="$*"
if [ -z "$CARRIED_IDS" ]; then
  CARRIED_IDS=$(ql_carried_from_env)
fi
case "$CARRIED_IDS" in
  over-ceiling:*)
    # The claim is refused, the event is not: the diff range is the
    # repository's own evidence and the title is only the author's
    # claim, so the range-derived scope below still records — and the
    # exit stays 0 once it has, because the approve workflow's Commit
    # step is success-gated and a merged close fires no second event. A
    # red exit here would lose the range's writes with the claim's.
    echo "the head branch and title claim ${CARRIED_IDS#over-ceiling:} distinct tasks — the ceiling is ${QL_CARRIED_MAX}."
    echo "The carried set is refused; only what the merge's own diff range proves is recorded."
    CARRIED_IDS=""
    ;;
esac

err=$(mktemp "${TMPDIR:-/tmp}/writrun-git.XXXXXX")
if ! CHANGED=$(git diff --name-only "$RANGE" 2>"$err"); then
  echo "git diff --name-only ${RANGE} failed:" >&2
  head -n 2 "$err" >&2
  rm -f "$err"
  exit 3
fi
rm -f "$err"

# The scope: task files the range touched, tasks of specs it touched,
# and the carried ones.
SCOPE=""
add_scope() {
  case " $SCOPE " in *" $1 "*) ;; *) SCOPE="$SCOPE $1" ;; esac
}
while IFS= read -r f; do
  case "$f" in
    work/tasks/task-*.md) [ -f "$f" ] && add_scope "$f" ;;
    work/specs/spec-*.md)
      [ -f "$f" ] || continue
      tref=$(ql_fm_field task_ref "$f")
      [ -n "$tref" ] || continue
      tf=$(ql_task_file "$tref")
      [ -n "$tf" ] && add_scope "$tf"
      ;;
  esac
done <<EOF
$CHANGED
EOF

CARRIED_FILES=""
for cid in $CARRIED_IDS; do
  tf=$(ql_task_file "$cid")
  [ -n "$tf" ] || continue
  add_scope "$tf"
  CARRIED_FILES="$CARRIED_FILES $tf"
done

is_carried() {
  case " $CARRIED_FILES " in *" $1 "*) return 0 ;; esac
  return 1
}

MOVED=""
note_move() {   # note_move <file> <old> <new>
  echo "moved $1: $2 -> $3"
  id=$(basename "$1" .md | sed -E 's/^(task-[0-9]+).*/\1/')
  case " $MOVED " in *" $id "*) ;; *) MOVED="${MOVED:+$MOVED }$id" ;; esac
}

for f in $SCOPE; do
  st=$(ql_fm_field status "$f")
  case "$st" in
    blocked|dropped) continue ;;   # a person's, never the machinery's
  esac

  cdate=$(ql_fm_field completed "$f")
  if [ -n "$cdate" ] && [ "$cdate" != "null" ]; then
    if [ "$st" != "done" ]; then
      ql_set_field "$f" status done
      note_move "$f" "$st" done
    fi
    continue
  fi

  case "$st" in
    in-progress|in-review)
      # Only the merge that carried this task's work lands it; any
      # other merge leaves the in-flight state to the task's own pull
      # request's events.
      if is_carried "$f"; then
        dest=$(ql_resting "$f")
        ql_set_field "$f" status "$dest"
        ql_set_field "$f" taken_by null
        note_move "$f" "$st" "$dest"
      fi
      ;;
    backlog|ready)
      dest=$(ql_resting "$f")
      if [ "$dest" != "$st" ]; then
        ql_set_field "$f" status "$dest"
        note_move "$f" "$st" "$dest"
      fi
      ;;
  esac
done

if [ -n "${GITHUB_OUTPUT:-}" ]; then
  if [ -n "$MOVED" ]; then
    printf 'changed=1\n' >> "$GITHUB_OUTPUT"
  else
    printf 'changed=0\n' >> "$GITHUB_OUTPUT"
  fi
  printf 'tasks=%s\n' "$MOVED" >> "$GITHUB_OUTPUT"
  SCOPE_IDS=""
  for f in $SCOPE; do
    id=$(basename "$f" .md | sed -E 's/^(task-[0-9]+).*/\1/')
    case " $SCOPE_IDS " in
      *" $id "*) ;;
      *) SCOPE_IDS="${SCOPE_IDS:+$SCOPE_IDS }$id" ;;
    esac
  done
  printf 'scope=%s\n' "$SCOPE_IDS" >> "$GITHUB_OUTPUT"
fi

exit 0
