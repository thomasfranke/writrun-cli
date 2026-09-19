#!/usr/bin/env bash
# preflight.sh — the completion gates, in the order they must run.
#
#   bash .writrun/scripts/stage-1-tasks-and-specs/preflight.sh \
#     [task-id[,task-id…]] [diff-range]
#
# **The two arguments are told apart by what a task list looks like**,
# never by the range's punctuation. An argument whose every
# comma-separated entry is a task id — `task-0034`, `0034`,
# `task/0034-…` — is the task list; anything else is the range. So a
# bare `origin/main` is a range, and it means to the three stages what
# it means to every gate they call: the ref against the **working
# tree**, the one shape whose head end is the checkout rather than a
# commit. That shape is how a flow that writes the completion edits and
# commits nothing gets them judged; `origin/main..` and `origin/main...`
# keep meaning `HEAD`
# (docs/technical/distribution/preflight.md#a-bare-ref-reaches-the-working-tree).
#
# The trade is stated rather than hidden: a ref spelled like a task id —
# a branch literally named `0034` — reads as a task list, and is named
# as a range by writing `0034..` instead.
#
# Three checks, and they are CI's own three (`writrun-check.yml`), called
# unmodified and in one order:
#
#   1. check_front_matter.sh — the whole-queue sweep, the only interface
#      it has; the range plays no part in it.
#   2. check_promised_deltas.sh — derives the specs the range moved to
#      `implemented` and runs check_deltas.sh on exactly that set.
#   3. check_state.sh — the lifecycle transitions the range makes.
#
# **The order is the rule this encodes.** The state gate exists to reject
# the transitions the completion edits make, so running it before them
# passes without reading anything — the warning that used to live in
# prose ("run it after step 4") is the delta stage's own derivation here:
# a range that moved no spec to `implemented` says so, loudly, and a task
# whose `completed` is still null is named in the summary.
#
# It adds no rule of its own. A stage that fails stops the run and exits
# with **that check's** code, printed under the stage's name; preflight's
# own failures — a malformed argument, an explicit task id resolving to
# no file — exit **4**, a code no stage uses, so a caller retrying on
# preflight's word never mistakes a stage's 3 for preflight asking for
# different arguments.
#
# Portable bash 3.2, POSIX awk/sed. See the standing rule in
# docs/technical/decisions/.

set -uo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
CHECK_FRONT_MATTER="$HERE/../../skills/writrun-check-front-matter/check_front_matter.sh"
CHECK_PROMISED_DELTAS="$HERE/../stage-2-pull-requests/check_promised_deltas.sh"
CHECK_STATE="$HERE/../../skills/writrun-check-task-state/check_state.sh"

own_failure() { echo "PREFLIGHT: $*" >&2; exit 4; }

# is_task_id <word> — the three spellings a task id has: `task-NNNN`,
# the branch form `task/NNNN-…`, and the bare number the queue and every
# [TASK-NNNN] tag print. Nothing else is one, and that is the whole of
# the classification below.
is_task_id() {
  local rest
  case "$1" in
    task/[0-9]*) return 0 ;;
    task-*)      rest="${1#task-}" ;;
    *)           rest="$1" ;;
  esac
  case "$rest" in
    ""|*[!0-9]*) return 1 ;;
    *)           return 0 ;;
  esac
}

# is_task_list <argument> — every comma-separated entry is a task id.
# One entry that is not makes the whole argument a range, because a
# task list with a ref in it is not a list this run could honour.
is_task_list() {
  local one
  [ -n "$1" ] || return 1
  for one in $(printf '%s' "$1" | tr ',' ' '); do
    is_task_id "$one" || return 1
  done
  return 0
}

IDS_ARG=""
RANGE=""
for arg in "$@"; do
  [ -n "$arg" ] || continue
  case "$arg" in
    -*) own_failure "unknown option '$arg' — usage: preflight.sh [task-id[,task-id…]] [diff-range]" ;;
  esac
  if is_task_list "$arg"; then
    [ -z "$IDS_ARG" ] || own_failure "two task lists given ('$IDS_ARG' and '$arg')"
    IDS_ARG="$arg"
  else
    [ -z "$RANGE" ] || own_failure "two diff ranges given ('$RANGE' and '$arg')"
    RANGE="$arg"
  fi
done

# Run from a subdirectory and every path below would resolve against the
# wrong root, so it re-roots first.
TOP=$(git rev-parse --show-toplevel 2>/dev/null) \
  || own_failure "not a git repository"
cd "$TOP" || own_failure "cannot enter ${TOP}"

# The shared front-matter reader and the task resolvers — one copy each,
# in the stage-2 lib, for the reason its header gives; the stages ship
# as one tree, so the path is always there to source.
. "$(dirname "$0")/../stage-2-pull-requests/queue_lib.sh"

# --- which tasks this run is about ------------------------------------
#
# Named, or inferred from the branch. Inferring nothing is not an error:
# a reporting or docs branch carries no task marker, and all three stages
# still have work to do — only the completion warning has nothing to
# attach to.

BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo HEAD)
if [ -z "$IDS_ARG" ]; then
  inferred=$(printf '%s' "$BRANCH" | sed -n 's|^task/0*\([0-9][0-9]*\).*|\1|p')
  IDS_ARG="$inferred"
  ids_named=""
else
  ids_named=yes
fi

TASK_FILES=""
for one in $(printf '%s' "$IDS_ARG" | tr ',' ' '); do
  [ -n "$one" ] || continue
  f=$(ql_task_file "$one")
  if [ -z "$f" ]; then
    # An id the caller typed is a claim about the queue; an id inferred
    # from a branch name is a guess, and a guess that misses is silence.
    [ -n "$ids_named" ] && own_failure "task id '${one}' resolves to no file under work/tasks/"
    continue
  fi
  TASK_FILES="${TASK_FILES}${f}"$'\n'
done

# --- the base the transitions are read against ------------------------

if ! FETCH_ERR=$(git fetch origin main 2>&1); then
  echo "Could not fetch origin main — reading a possibly stale base:"
  printf '%s\n' "$FETCH_ERR" | head -n 2
  echo
fi

if [ -z "$RANGE" ]; then
  if git rev-parse --verify --quiet refs/remotes/origin/main >/dev/null 2>&1; then
    RANGE="origin/main...HEAD"
  elif git rev-parse --verify --quiet refs/heads/main >/dev/null 2>&1; then
    echo "No origin/main in this checkout — the range is the local main."
    echo
    RANGE="main...HEAD"
  else
    own_failure "no origin/main and no main to read the change against — name a range"
  fi
fi

# --- the completion warning -------------------------------------------
#
# A run made before the completion edits passes the state gate by having
# nothing to judge. That is not a green light, and it is said out loud
# rather than inferred from a quiet run.
#
# **It reads the end the stages read**, never the checkout regardless of
# range. The bare shape's head *is* the checkout, so there the two are
# the same file; under a two-ended range the head is a commit, and a
# task whose `completed` was written but not committed has none there —
# which is exactly the state the stages are about to read as empty. Read
# from the checkout instead, this warning would fall silent over a run
# whose stages saw nothing, and print PREFLIGHT OK on the one input it
# never had (report-0046).
ql_range_ends "$RANGE"
WARN_HEAD="$QL_HEADREF"
[ -n "$WARN_HEAD" ] && WHERE="${WARN_HEAD}" || WHERE="the working tree"

WARNING=""
while IFS= read -r f; do
  [ -n "$f" ] || continue
  id=$(ql_fm_field id "$f")
  if [ -z "$WARN_HEAD" ]; then
    done_at=$(ql_fm_field completed "$f")
  else
    # A task file the head does not carry — created on this branch and
    # not yet committed — reads as no completed date, which is what it
    # is at that end. `git show` failing is that answer, never an error.
    done_at=$(git show "${WARN_HEAD}:${f}" 2>/dev/null | ql_fm_field_in completed)
  fi
  if [ -z "$done_at" ] || [ "$done_at" = null ]; then
    WARNING="${WARNING}${id} has no completed date in ${WHERE}, so this run precedes the completion edits and does not stand for them; run it again after them."$'\n'
  fi
done <<TASKS
${TASK_FILES}
TASKS

[ -n "$WARNING" ] && printf '%s\n' "$WARNING"

# --- the three stages, in order ---------------------------------------

stage() {   # stage <n> <name> <cmd...>
  local n="$1" name="$2"; shift 2
  local out code log
  echo "== ${n}/3 ${name} =="
  # Streamed and captured in one pass. Capturing alone prints nothing
  # until the stage is over, and the whole-queue sweep takes long enough
  # that the silence reads as a hang; the capture is still needed,
  # because stage 2 names the deltas it checked out of its own output.
  log=$(mktemp "${TMPDIR:-/tmp}/writrun-preflight.XXXXXX")
  "$@" 2>&1 | tee "$log"; code=${PIPESTATUS[0]}
  out=$(cat "$log"); rm -f "$log"
  echo
  if [ "$code" -ne 0 ]; then
    if [ "$n" -lt 3 ]; then
      echo "PREFLIGHT STOPPED at ${n}/3 ${name} — exit ${code}. The stages after it did not run." >&2
    else
      echo "PREFLIGHT STOPPED at ${n}/3 ${name} — exit ${code}." >&2
    fi
    [ -n "$WARNING" ] && printf '%s' "$WARNING" >&2
    exit "$code"
  fi
  STAGE_OUT="$out"
}

STAGE_OUT=""
stage 1 "front matter" bash "$CHECK_FRONT_MATTER"
stage 2 "promised deltas" bash "$CHECK_PROMISED_DELTAS" "$RANGE"
DELTAS=$(printf '%s' "$STAGE_OUT" | sed -n 's/^Checking //p' | head -n1)
[ -n "$DELTAS" ] || DELTAS="none — no spec reached 'implemented' in this range"
stage 3 "task state" bash "$CHECK_STATE" "$RANGE"

echo "PREFLIGHT OK — range ${RANGE}; deltas checked: ${DELTAS}"
[ -n "$WARNING" ] && printf '%s' "$WARNING"
exit 0
