#!/usr/bin/env bash
# project_pr_tasks.sh — relabels the mirror of every task a pull request
# carries, by projecting the queue as it stands on the checked-out
# authority branch — the one label derivation, shared with the merge
# path (rederive_labels.sh; docs/product/stage-3-github-issues/labels.md).
#
# Usage: project_pr_tasks.sh <owner/repo> [event]
#   Run from a checkout of the authority branch *after* the Stage-2
#   recording pushed the event's status writes — the whole point is to
#   read what the machinery just wrote, never to derive a second answer
#   from the event's own fields. A predecessor (reflect_progress.sh)
#   did exactly that, and its private mapping could contradict the file.
#
# The carried ids come from env, as data (PR_HEAD_REF, PR_TITLE — a
# fork's to write): the head branch's task and every [TASK-NNNN] tag
# leading the title. No id means nothing to project, which is not an
# error. A claim over QL_CARRIED_MAX comes back as the helper's
# over-ceiling sentinel: nothing is projected and the exit is non-zero
# — a relabelling pass over dozens of mirrors would be the same refused
# claim wearing Stage 3's clothes.
#
# **Except on a close, which is why the event is an argument.** The
# ceiling bounds claiming, and a close claims nothing: the Stage-2
# recorder has already released the tasks it named, and this pass only
# restates what the queue now says about them. Refusing here turned the
# `reflect` job red and left those mirrors reading `in-progress` over a
# queue that says `ready` — a mirror ahead of the file, which nothing
# comes back to heal. That is the same exemption the merge recorder and
# the close arm of the in-flight recorder already carry, for the same
# reason, stated as one rule in
# `docs/technical/decisions/pull-requests/0069-a-close-releases-what-it-cannot-claim.md`
# — the entry that *extends* 0068, whose constant and whole-set refusal
# both stand.
#
# The event arrives as an argument rather than a sixth `PR_*` name: the
# `PR_*` enumeration is the contract `technical/distribution/checks.md`
# holds the caller to, and what a script is asked to do is the caller's
# own word, not a field of the pull request.
#
# This path mints nothing and says so — `--minted` with nothing behind
# it. Almost every miss it sees is a finding about the repository. The
# exception is a mirror another workflow minted between this run's read
# and its lookup. Either is answered from one read of the list, and
# healed by the next pull-request event if it was not.
#
# Exit codes: 0 done (including nothing to do, and an over-ceiling
# close); 1 the claim is over the ceiling on an event that claims,
# nothing projected; 3 usage error.
#
# Portable bash 3.2, POSIX awk/sed. See the standing rule in
# docs/technical/decisions/.

set -euo pipefail

REPO="${1:?usage: project_pr_tasks.sh <owner/repo> [event]}"
EVENT="${2:-}"

HERE="$(cd "$(dirname "$0")" && pwd)"
. "$HERE/../stage-2-pull-requests/queue_lib.sh"

carried=$(ql_carried_from_env)
if [ -z "$carried" ]; then
  echo "the pull request names no task — nothing to project"
  exit 0
fi
case "$carried" in
  over-ceiling:*)
    over="${carried#over-ceiling:}"
    if [ "$EVENT" = "closed" ]; then
      # Said, not hidden — but on stdout and green: the projection is
      # correct and the claim it rides was already met red at the event
      # that made it. The ids are re-read through the same helper with
      # the ceiling lifted to the count it just reported, never a second
      # parser, and the assignment lives in the substitution's own
      # subshell.
      echo "the head branch and title claim ${over} distinct tasks — the ceiling is ${QL_CARRIED_MAX}."
      echo "A close releases rather than claims, so the mirrors it released are still projected."
      carried=$(QL_CARRIED_MAX="$over" ql_carried_from_env)
    else
      echo "the head branch and title claim ${over} distinct tasks — the ceiling is ${QL_CARRIED_MAX}." >&2
      echo "Nothing was projected. Retitle the pull request to what the work carries:" >&2
      echo "the edit re-fires the recording, and the projection follows it." >&2
      exit 1
    fi
    ;;
esac

# One projector: rederive_labels reads the queue files and restates
# them, one to one, closing terminal states.
# shellcheck disable=SC2086
bash "$HERE/rederive_labels.sh" "$REPO" $carried --minted
