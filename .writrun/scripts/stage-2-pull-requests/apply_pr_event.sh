#!/usr/bin/env bash
# apply_pr_event.sh — turns one pull-request forge event into the status
# write it implies, via flip_task_status.sh. The workflow wires events to
# this; the edge table lives in the flip script and in
# product/stage-2-pull-requests/statuses.md.
#
# Usage: apply_pr_event.sh <event>
#   <event>: opened | reopened | ready_for_review | converted_to_draft |
#            review_requested | changes_requested | closed
#
# Context arrives in env, never argv interpolation — a head branch name
# and a login are a fork's to choose, and both are data here:
#   PR_HEAD_REF   the head branch (task/NNNN-* names the task; anything
#                 else exits without writing)
#   PR_AUTHOR     the pull request author's login (forge-authenticated)
#   PR_DRAFT      true|false
#   PR_MERGED     true|false (closed only)
#   GH_REPO       owner/repo, for the survivor query
#   GH_TOKEN      lets `gh` ask the forge about surviving pull requests
#
# On close-without-merge, the forge is asked whether another open pull
# request still works the same task: with a survivor, the newest one's
# author and draftness are re-recorded instead of landing the task —
# never a silent skip, or taken_by strands on the closed PR's author.
# When the forge cannot answer, the task lands: a queue that briefly
# forgets a survivor heals at that survivor's next event, while a task
# stranded in-flight with no PR heals never.
#
# Exits 0 in every no-op case (not a task branch, no legal edge, merged
# close); 3 only for usage errors. Mutates the working tree; the caller
# commits.
#
# Portable bash 3.2, POSIX awk/sed. See the standing rule in
# docs/technical/decisions/.

set -euo pipefail

EVENT="${1:?usage: apply_pr_event.sh <event>}"

FLIP="$(dirname "$0")/flip_task_status.sh"

# The task id is the head branch's — validated as data before any use.
case "${PR_HEAD_REF:-}" in
  task/[0-9]*) ;;
  *) echo "head '${PR_HEAD_REF:-}' names no task branch — nothing to record"; exit 0 ;;
esac
TASK=$(printf '%s' "$PR_HEAD_REF" | sed -E 's|^task/([0-9]+).*|task-\1|')

draftness() {   # $1: true|false -> draft|ready
  [ "$1" = "true" ] && printf 'draft' || printf 'ready'
}

case "$EVENT" in
  opened|reopened)
    bash "$FLIP" take "$TASK" "${PR_AUTHOR:?}" "$(draftness "${PR_DRAFT:-true}")"
    ;;
  ready_for_review)
    bash "$FLIP" review "$TASK"
    ;;
  converted_to_draft|changes_requested)
    bash "$FLIP" rework "$TASK"
    ;;
  review_requested)
    # GitHub fires this on drafts too (CODEOWNERS auto-requests); a
    # draft's task is being worked, not reviewed.
    if [ "${PR_DRAFT:-true}" = "true" ]; then
      echo "review requested on a draft — not an in-review signal"
      exit 0
    fi
    bash "$FLIP" review "$TASK"
    ;;
  closed)
    if [ "${PR_MERGED:-false}" = "true" ]; then
      echo "closed by merging — the merge recording owns this move"
      exit 0
    fi
    # A surviving open PR on the same task supersedes the landing. The
    # match is by number, zero-padding stripped — every id reader in
    # this machine normalizes, and a survivor spelling `task/019-` must
    # not be invisible to a close on `task/0019-`.
    num=$(printf '%s' "$TASK" | sed 's/^task-0*//')
    survivor=""
    if [ -n "${GH_TOKEN:-}" ] && command -v gh >/dev/null 2>&1; then
      survivor=$(gh pr list --repo "${GH_REPO:?}" --state open \
        --json number,headRefName,author,isDraft \
        --jq "[.[] | select(.headRefName | test(\"^task/0*${num}-\"))] | sort_by(.number) | last | if . == null then \"\" else \"\(.author.login) \(.isDraft)\" end" \
        2>/dev/null || printf '')
    fi
    if [ -n "$survivor" ]; then
      s_login=${survivor%% *}
      s_draft=${survivor##* }
      echo "a surviving pull request still works ${TASK} — recording it instead"
      bash "$FLIP" take "$TASK" "$s_login" "$(draftness "$s_draft")"
    else
      bash "$FLIP" land "$TASK"
    fi
    ;;
  *)
    echo "usage: apply_pr_event.sh <opened|reopened|ready_for_review|converted_to_draft|review_requested|changes_requested|closed>" >&2
    exit 3
    ;;
esac

exit 0
