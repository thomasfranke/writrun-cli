#!/usr/bin/env bash
# flip_task_status.sh — the transition machine: applies one forge event
# to one task's status line, on the working tree, exactly as the edge
# table in product/stage-2-pull-requests/statuses.md draws it.
#
# Usage:
#   flip_task_status.sh take   <task-id> <login> <draft|ready>
#   flip_task_status.sh review <task-id>
#   flip_task_status.sh rework <task-id>
#   flip_task_status.sh land   <task-id>
#
#   take    a pull request opened, reopened, or superseded another —
#           in-progress when it is a draft, in-review when it is not;
#           taken_by set to its author. Legal from backlog, ready, or an
#           in-flight state (the newest pull request wins).
#   review  the pull request was marked ready, or review was re-requested
#           on one already marked ready: in-progress -> in-review.
#   rework  converted back to draft, or a review requested changes:
#           in-review -> in-progress.
#   land    the task leaves flight with no surviving pull request: to
#           ready — or backlog if any of its specs is draft, derived at
#           exit time, never assumed — and taken_by cleared.
#
# **An event with no legal edge writes nothing and exits 0.** A stale
# replay, a reopen against a done task, a close echoed twice — an echo is
# not an error, and it never marches a task backwards. The one loud exit
# (3) is a usage error: an unknown mode, or arguments missing.
#
# Mutates the working tree only and prints one `moved` line per write;
# committing is the caller's job, one commit per forge event.
#
# Portable bash 3.2, POSIX awk/sed — no gawk extensions. See the standing
# rule in docs/technical/decisions/.

set -euo pipefail

. "$(dirname "$0")/queue_lib.sh"

MODE="${1:-}"
TASK="${2:-}"
[ -n "$MODE" ] && [ -n "$TASK" ] \
  || { echo "usage: flip_task_status.sh <take|review|rework|land> <task-id> [login] [draft|ready]" >&2; exit 3; }

FILE=$(ql_task_file "$TASK")
if [ -z "$FILE" ]; then
  num=$(ql_task_num "$TASK")
  [ -n "$num" ] || { echo "'${TASK}' names no task id" >&2; exit 0; }
  echo "task-${num} resolves to no file — nothing to move" >&2
  exit 0
fi

CUR=$(ql_fm_field status "$FILE")

move() {   # move <new-status> [taken_by-value]
  ql_set_field "$FILE" status "$1"
  [ -n "${2:-}" ] && ql_set_field "$FILE" taken_by "$2"
  echo "moved ${FILE}: ${CUR} -> $1"
}

echo_only() { echo "no edge: ${MODE} against '${CUR}' on ${FILE} — an echo, nothing written"; }

case "$MODE" in
  take)
    LOGIN="${3:-}"
    DRAFTNESS="${4:-draft}"
    [ -n "$LOGIN" ] || { echo "usage: flip_task_status.sh take <task-id> <login> <draft|ready>" >&2; exit 3; }
    printf '%s' "$LOGIN" | grep -qE '^[A-Za-z0-9-]+(\[bot\])?$' \
      || { echo "'${LOGIN}' is not a bare forge login — nothing written" >&2; exit 0; }
    dest=in-progress
    [ "$DRAFTNESS" = "ready" ] && dest=in-review
    case "$CUR" in
      backlog|ready|in-progress|in-review)
        if [ "$CUR" = "$dest" ] && [ "$(ql_fm_field taken_by "$FILE")" = "$LOGIN" ]; then
          echo "already there: ${FILE} is ${dest}, taken by ${LOGIN}"
        else
          move "$dest" "$LOGIN"
        fi ;;
      *) echo_only ;;
    esac
    ;;
  review)
    case "$CUR" in
      in-progress) move in-review ;;
      *) echo_only ;;
    esac
    ;;
  rework)
    case "$CUR" in
      in-review) move in-progress ;;
      *) echo_only ;;
    esac
    ;;
  land)
    case "$CUR" in
      in-progress|in-review) move "$(ql_resting "$FILE")" null ;;
      *) echo_only ;;
    esac
    ;;
  *)
    echo "usage: flip_task_status.sh <take|review|rework|land> <task-id> [login] [draft|ready]" >&2
    exit 3
    ;;
esac

exit 0
