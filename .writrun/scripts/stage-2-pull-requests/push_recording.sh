#!/usr/bin/env bash
# push_recording.sh — lands a composed recording commit on the branch.
#
# Usage: push_recording.sh <branch>
#   Run from the repository root, on a checkout already carrying the
#   recording commit. The caller composes and commits; this script only
#   makes the commit reach the branch.
#
# The loss this closes: two recordings racing to one branch. The push
# is refused because the branch moved under a rebase that predates it,
# so the answer is to read the branch again and push again. Never
# --force and never --force-with-lease, on any attempt: the recording
# is an addition to the branch's history, not a replacement of it, and
# the lease flag is the one that fails on exactly the race being
# survived.
#
# The budget is five attempts, sized by the largest burst practice has
# produced — the five drafts a five-task batch opens seconds apart. The
# worst-placed of five runs can lose four races before its turn. One
# fetch and one push per attempt: the budget belongs to the run, never
# to each obstacle.
#
# A failure here is one of three things, and separating them is this
# script's whole judgement. The remote answered and refused — a ruleset
# on the branch, a token without write, a required check. The remote
# answered and the branch had moved — the race. Or the remote was never
# reached — a network blip, a forge 500, a proxy timeout. Only the first
# is permanent. Calling the third by the first's name loses the
# recording to the one class a second attempt clears for free, and sends
# the operator hunting a conflict or a ruleset that never existed.
#
# A retry is earned, never assumed. Where the remote answered, the fetch
# that opens the next attempt shows whether the tip left the commit the
# refused push was rebased onto; an unmoved tip means the refusal was
# never a race, and the run fails at once naming the branch as unmoved.
# Movement is the one version-proof fact: git and the forge word their
# refusals differently across versions, so no stderr is read.
#
# Where the remote was never reached, movement proves nothing — the tip
# is where it was because the push never arrived — so that test is not
# armed and the attempt is simply spent. Which of the two happened is
# read from git's exit status, never its wording: a push exits 1 when
# the remote answered about the refs, and dies with something else when
# it did not. A pull says it by whether a rebase is in progress
# afterwards. Five instant attempts do not outlast an outage and no
# sleep is added to make them: the run ends red naming the remote as
# unreached, which is the report a silently lost recording never gave.
#
# And no attempt sleeps. Where the obstacle is a sibling, every retry
# spent is that sibling's recording landing, so the loop only loses
# while the queue advances. Where it is an outage, a sleep would only
# make the run take longer to say so.
#
# A conflicting rebase aborts back to the recording commit and fails at
# once: the same commit meets the same conflict, and the tree must not
# be left carrying markers in the queue files the projection reads from
# disk. The abort is checked rather than assumed. An abort that failed
# over a tree still carrying markers must not be reported as a restored
# one — that report is the exact outcome this abort exists to prevent,
# and the mirror steps that follow parse those files.
#
# Exit codes: 0 the recording is on the branch — a rebase that finds it
# already landed and drops it exits 0 too; 1 it could not be landed (a
# conflict, an unmoved refusal, an exhausted budget); 3 the caller named
# no branch, or is not holding a recording to land.
#
# Portable bash 3.2, POSIX awk/sed — no gawk extensions. See the
# standing rule in docs/technical/decisions/.

set -euo pipefail

# A missing branch is a caller error like the two below it, and owes the
# same 3 — `${1:?}` would exit 1, the code that means the recording
# could not be landed.
if [ "$#" -lt 1 ] || [ -z "${1:-}" ]; then
  echo "usage: push_recording.sh <branch>" >&2
  exit 3
fi
BRANCH="$1"
BUDGET=5

# --- the caller's half, checked before the remote is touched -----------
#
# A caller that has not committed must hear so, not watch a no-op
# report success — the failure distribution/checks.md says looks
# ordinary. The guard fetches nothing: the remote-tracking ref the
# checkout already carries is the view the caller committed against,
# and the loop's own first pull is the one fetch an attempt pays.
if [ -n "$(git status --porcelain)" ]; then
  echo "the working tree is dirty — commit the recording first; this script only lands one." >&2
  exit 3
fi
# A range git cannot answer is a refusal too: defaulting it would guess
# at the one thing this guard exists to know — the posture take_task.sh
# already takes.
if ! AHEAD=$(git rev-list --count "refs/remotes/origin/${BRANCH}..HEAD" 2>&1); then
  echo "git rev-list --count refs/remotes/origin/${BRANCH}..HEAD failed:" >&2
  printf '%s\n' "$AHEAD" | head -n 3 >&2
  echo "whether HEAD carries a recording is the one thing this script must not guess at." >&2
  exit 3
fi
if [ "$AHEAD" = 0 ]; then
  echo "HEAD is not ahead of origin/${BRANCH} — nothing committed, nothing to land." >&2
  exit 3
fi

# rebase_in_progress — a rebase really did start and really did stop.
# REBASE_HEAD is git's marker for a rebase halted on the commit it was
# replaying; the state directories cover a halt this script did not
# anticipate. A pull whose fetch never reached the remote sets neither,
# which is what separates a conflict from an unreachable branch.
rebase_in_progress() {
  git rev-parse --verify -q REBASE_HEAD >/dev/null 2>&1 && return 0
  [ -d "$(git rev-parse --git-path rebase-merge)" ] && return 0
  [ -d "$(git rev-parse --git-path rebase-apply)" ] && return 0
  return 1
}

# --- the loop: rebase onto the branch as it then stands, then push -----
attempt=0
rebased_onto=""
reached=1
while [ "$attempt" -lt "$BUDGET" ]; do
  attempt=$((attempt + 1))

  if ! git pull --rebase origin "$BRANCH"; then
    if rebase_in_progress; then
      if ! git rebase --abort; then
        echo "rebase onto ${BRANCH} conflicted and the abort failed — the tree is still mid-rebase." >&2
        echo "the queue files the projection reads from disk may carry markers; nothing was pushed." >&2
        exit 1
      fi
      if [ -n "$(git status --porcelain)" ]; then
        echo "rebase onto ${BRANCH} conflicted and the abort left the tree unclean:" >&2
        git status --porcelain >&2
        echo "the tree is not the recording commit's; nothing was pushed." >&2
        exit 1
      fi
      echo "rebase onto ${BRANCH} conflicted — aborted back to the recording commit, nothing pushed." >&2
      echo "the same commit meets the same conflict; re-running the job re-derives the write against ${BRANCH} as it then stands." >&2
      exit 1
    fi
    # Nothing was rebased, so nothing conflicted: the fetch never
    # reached the branch. One more attempt is what that class costs.
    reached=0
    echo "attempt ${attempt} of ${BUDGET}: origin/${BRANCH} could not be read, and nothing was rebased." >&2
    continue
  fi
  reached=1

  tip=$(git rev-parse FETCH_HEAD)
  if [ -n "$rebased_onto" ] && [ "$tip" = "$rebased_onto" ]; then
    echo "the push was refused and ${BRANCH} is unmoved — that refusal was never a race." >&2
    echo "something else stands in the way: a ruleset on ${BRANCH}, a token without write, a required check." >&2
    exit 1
  fi

  push_status=0
  git push origin "HEAD:${BRANCH}" || push_status=$?
  if [ "$push_status" -eq 0 ]; then
    echo "recorded on ${BRANCH} (attempt ${attempt} of ${BUDGET})."
    exit 0
  fi
  if [ "$push_status" -eq 1 ]; then
    # The remote answered about the refs, so the next attempt's fetch
    # can say whether that answer was a race.
    rebased_onto="$tip"
  else
    # The push never completed. The tip is unmoved because nothing
    # arrived, so arming the movement test would read a silence as a
    # ruleset.
    reached=0
    rebased_onto=""
    echo "attempt ${attempt} of ${BUDGET}: the push to ${BRANCH} did not complete — the remote was not reached." >&2
  fi
done

if [ "$reached" -eq 0 ]; then
  echo "the recording could not be landed on ${BRANCH} in ${attempt} attempts — the remote was not reached on the last of them." >&2
else
  echo "the push to ${BRANCH} was refused on all ${attempt} attempts — the branch outran the budget." >&2
fi
echo "re-running the job re-derives the same write from the same event against ${BRANCH} as it then stands." >&2
exit 1
