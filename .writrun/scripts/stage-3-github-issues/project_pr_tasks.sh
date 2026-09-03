#!/usr/bin/env bash
# project_pr_tasks.sh — relabels the mirror of every task a pull request
# carries, by projecting the queue as it stands on the checked-out
# authority branch — the one label derivation, shared with the merge
# path (rederive_labels.sh; docs/product/stage-3-github-issues/labels.md).
#
# Usage: project_pr_tasks.sh <owner/repo>
#   Run from a checkout of the authority branch *after* the Stage-2
#   recording pushed the event's status writes — the whole point is to
#   read what the machinery just wrote, never to derive a second answer
#   from the event's own fields. A predecessor (reflect_progress.sh)
#   did exactly that, and its private mapping could contradict the file.
#
# The carried ids come from env, as data (PR_HEAD_REF, PR_TITLE — a
# fork's to write): the head branch's task and every [TASK-NNNN] tag
# leading the title. No id means nothing to project, which is not an
# error.
#
# Exit codes: 0 done (including nothing to do); 3 usage error.
#
# Portable bash 3.2, POSIX awk/sed. See the standing rule in
# docs/technical/decisions/.

set -euo pipefail

REPO="${1:?usage: project_pr_tasks.sh <owner/repo>}"

HERE="$(cd "$(dirname "$0")" && pwd)"
. "$HERE/../stage-2-pull-requests/queue_lib.sh"

carried=$(ql_carried_from_env)
if [ -z "$carried" ]; then
  echo "the pull request names no task — nothing to project"
  exit 0
fi

# One projector: rederive_labels reads the queue files and restates
# them, one to one, closing terminal states.
# shellcheck disable=SC2086
bash "$HERE/rederive_labels.sh" "$REPO" $carried
