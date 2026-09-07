#!/usr/bin/env bash
# rederive_labels.sh — after a merge records an approval, label the tasks
# and reports it affected from the queue as it then stands.
#
# Usage: rederive_labels.sh <owner/repo> <spec-file|task-id|report-id>...
#          [--minted <task-id|report-id>...]
#   Run from a checkout of the authority branch, *after* the flip has been
#   committed to it — the whole point is to read the queue's new state.
#   Where the tree and that branch can disagree — a recording whose push
#   was refused — set `AUTHORITY_REF` to the branch and the queue is read
#   from there instead (below).
#   `gh` must be on PATH and authenticated (GH_TOKEN in CI; a stub in the
#   test suite). Named no id, it does nothing and says so.
#
#   Every id after `--minted` is one the same job already answered for —
#   mirror_issues.sh's `tasks=`/`reports=` output, minted now or found to
#   be its own. Its Issue exists, so not finding one is this pass's own
#   defect and exits non-zero. The same miss on any other id is a task
#   that was never mirrored, which is a finding and stays a notice.
#
#   The flag is also who may pay to ask again. A caller that mints and
#   labels in one job names its mints behind `--minted`, and only an id
#   it names can spend a re-read — the flag with nothing behind it
#   entitles nobody. A caller that says nothing buys the unconditional
#   re-read: the saving is declared, never assumed, so a minting caller
#   wired without the flag wastes seconds and can never lose a mirror.
#
#   A `report-NNNN` argument is the same projection one kind over: the
#   file on disk says whether triage has ended it, and the mirror is made
#   to agree. It is here so a report mirror that drifted is repaired by
#   the same pass that repairs a task's — a report is never re-derived
#   from a forge event, because none corresponds to a judgement, so this
#   is the only reader that can heal one.
#
# Why this exists: `status:ready` was unreachable. The mirror derived it
# from the spec statuses in the merged pull request's *diff*, where they
# are still `draft` — because that same merge is what approves them. So
# every task merged this way sat on the backlog label with its specs
# approved, and no later event ever corrected it.
#
# **Reading the diff is right for what the merge carried; it is wrong for
# what the merge caused** (docs/product/stage-3-github-issues/labels.md).
# So this reads neither a diff nor a patch: it reads the files, which is
# the only source that reflects the flip the merge just triggered.
#
# The question it answers — "what label does this task deserve now" — is
# the same one mirror_issues.sh asks of a pull request's diff. Same
# question, different input, which is why it is a script of its own rather
# than a branch inside either caller.
#
# Exit codes: 0 done (including nothing to do); 1 a mirror this job
# answered for was never found; 2 the forge refused a call this pass
# cannot go on without; 3 usage error; 4 `AUTHORITY_REF` was set and
# could not be read, so no label was written.
#
# One code per condition, because 1 is what a red `writrun approve`
# shows a maintainer: "the forge would not create the label" and "this
# pass left a minted mirror unlabelled" are different mornings. 4 is a
# fifth: the queue this pass projects could not be found at all, and a
# fallback to the working tree is exactly what it must not do.
#
# Portable bash 3.2, POSIX awk/sed — no gawk extensions, no associative
# arrays. See the standing rule in docs/technical/decisions/.

set -euo pipefail

if [ "$#" -lt 1 ]; then
  echo "usage: rederive_labels.sh <owner/repo> [<spec-file>|<task-id>|<report-id>]... [--minted <id>...]" >&2
  exit 3
fi
REPO="$1"
shift

TAB=$(printf '\t')

# --- which tree the queue is read from ------------------------------------
#
# `AUTHORITY_REF` names the branch this pass must project. Set it and the
# queue is read from that ref; leave it unset and the queue is the
# working tree, which is what every caller running *from* the authority
# branch already has.
#
# **Why the script reads the ref rather than the workflow resetting the
# tree.** Both were on the table. The workflow shape is one line and
# leaves this script honest about reading "the tree" — but it cannot
# answer the case it exists for: the mirror steps run under
# `!cancelled()`, so a reset step that itself failed would be followed by
# the labeller reading the unreset tree, which is the original fault with
# an extra step in front of it. Reading the ref puts the failure where
# the label is written: unreadable means no label, said out loud, here.
#
# The fault: `writrun-approve.yml`'s mirror steps run inside the recording
# job, after the commit and push, and the gate lets them run when the
# push failed. The tree at that moment still carries the commit `main`
# refused, so a labeller reading it projects a queue state that exists
# nowhere but the runner — and writes it onto the mirror. A mirror
# *behind* the queue catches up at the next recording; one *ahead* of it
# asserts a state `main` refused, and the next successful recording has
# no reason to revisit a label that already reads what it is about to
# write.
#
# **One answer to "which tree", not two.** The ref is materialised once,
# into a directory `queue_file` resolves against, so every reader below
# — `fm`, `label_for`, `close_for`, the report projection — reads the
# same tree by construction. A second reader taught the ref separately is
# the divergence again, smaller.
AUTHORITY_REF="${AUTHORITY_REF:-}"
QUEUE_ROOT="."
QUEUE_TMP=""
cleanup_queue() { [ -n "$QUEUE_TMP" ] && rm -rf "$QUEUE_TMP"; return 0; }
trap cleanup_queue EXIT

if [ -n "$AUTHORITY_REF" ]; then
  # A stale remote-tracking ref is the same class of wrong in the other
  # direction — a tree older than the queue. A successful push updates
  # the ref it pushed to, so the ordinary case is already current; the
  # fetch is for the case where this runner never pushed at all. It is
  # best-effort: a fetch that fails leaves the ref the runner has, and
  # the verify below is what decides whether that ref can be read.
  case "$AUTHORITY_REF" in
    */*)
      _rem=${AUTHORITY_REF%%/*}
      _brn=${AUTHORITY_REF#*/}
      # The listing is captured before it is searched, never piped into
      # `grep -q`. A quiet grep exits on its first match and closes the
      # pipe; under `pipefail` the writer can then die on SIGPIPE and the
      # pipeline reports 141, so the `if` reads false and the fetch this
      # block exists for is skipped — leaving exactly the stale ref it
      # was added to refresh. The race is won on a short listing and lost
      # on a long one, and which happens is the platform's to decide
      # (docs/technical/decisions/).
      _remotes=$(git remote 2>/dev/null || printf '')
      if printf '%s\n' "$_remotes" | grep -qxF "$_rem"; then
        git fetch --quiet "$_rem" "$_brn" >/dev/null 2>&1 || true
      fi
      ;;
  esac
  if ! git rev-parse --verify --quiet "${AUTHORITY_REF}^{commit}" >/dev/null 2>&1; then
    echo "AUTHORITY_REF '${AUTHORITY_REF}' names no commit this checkout can read." >&2
    echo "  The labels this pass would write are the authority branch's to" >&2
    echo "  decide, and falling back to the working tree is how a mirror gets" >&2
    echo "  ahead of the queue. No label was written." >&2
    exit 4
  fi
  QUEUE_TMP=$(mktemp -d "${TMPDIR:-/tmp}/writrun-queue.XXXXXX")
  QUEUE_ROOT="$QUEUE_TMP"
  # A ref with no `work/` at all is a queue with nothing in it, not a ref
  # that could not be read — the empty directories below give every
  # lookup its existing "no file on this branch" answer.
  if git cat-file -e "${AUTHORITY_REF}:work" 2>/dev/null; then
    if ! git archive "$AUTHORITY_REF" work | tar -xf - -C "$QUEUE_ROOT"; then
      echo "AUTHORITY_REF '${AUTHORITY_REF}' holds a work/ this pass could not read." >&2
      echo "  No label was written." >&2
      exit 4
    fi
  fi
  mkdir -p "$QUEUE_ROOT/work/tasks" "$QUEUE_ROOT/work/reports"
fi

if printf '' | base64 -d >/dev/null 2>&1; then B64_FLAG="-d"; else B64_FLAG="-D"; fi
b64_decode() { base64 "$B64_FLAG"; }

. "$(dirname "$0")/../stage-2-pull-requests/queue_lib.sh"

# fm <file> <field> — the shared front-matter reader, argument order
# kept as every call site here already speaks it; one body, in the
# stage-2 lib, for the reason its header gives.
fm() { ql_fm_field "$2" "$1"; }

# queue_file <dir> <prefix> <id> — the file whose id is <id>, whatever its
# subject slug and whatever width its number was written at.
# The one place a queue path is resolved, so `$QUEUE_ROOT` is the one
# answer to which tree this pass reads.
queue_file() {
  local dir="$QUEUE_ROOT/$1" prefix="$2" want f n
  want=$(printf '%s' "$3" | tr '[:upper:]' '[:lower:]' \
    | sed -E "s/^${prefix}-0*([0-9]+)$/\1/")
  [ -n "$want" ] || return 0
  case "$want" in *[!0-9]*) return 0 ;; esac
  for f in "$dir"/"$prefix"-*.md; do
    [ -f "$f" ] || continue
    n=$(basename "$f" .md | tr '[:upper:]' '[:lower:]' \
      | sed -E "s/^${prefix}-0*([0-9]+).*/\1/")
    case "$n" in ''|*[!0-9]*) continue ;; esac
    if [ "$n" -eq "$want" ]; then printf '%s' "$f"; return 0; fi
  done
  return 0
}

# label_for <task-file> — the label this task deserves from the queue on
# disk, or nothing when the mirror should close or stay closed instead.
#
# The file is the truth: the machinery writes every working state onto
# the authority branch as its forge event lands, so the label restates
# the stored status, one to one
# (docs/product/stage-3-github-issues/labels.md). The two terminal
# states return nothing here — a closed mirror carries no status label,
# and closing is close_for's answer.
label_for() {
  case "$(fm "$1" status)" in
    backlog)     printf 'status:backlog' ;;
    ready)       printf 'status:ready' ;;
    in-progress) printf 'status:in-progress' ;;
    in-review)   printf 'status:in-review' ;;
    blocked)     printf 'status:blocked' ;;
  esac
}

# origin_label_for <task-file> — the label the task's stored `origin`
# projects. Nothing for a file that arrives without the field, which the
# front-matter check should have refused: a gap to leave rather than a
# value to guess.
origin_label_for() {
  case "$(fm "$1" origin)" in
    rule)   printf 'origin:rule' ;;
    report) printf 'origin:report' ;;
  esac
}

# close_for <task-file> — the close reason a terminal status implies, or
# nothing for a task still in the pipeline.
close_for() {
  case "$(fm "$1" status)" in
    done)    printf 'completed' ;;
    dropped) printf 'not_planned' ;;
  esac
}

# report_label_for <report-file> — the label a report deserves from the
# queue on disk, or nothing when it is triaged and the mirror should be
# closed instead. `status:proposed` never appears here: it is the one
# state no file can hold, because it means "a pull request offers this
# and the authority branch does not have it yet".
report_label_for() {
  case "$(fm "$1" status)" in
    open) printf 'status:open' ;;
  esac
}

# report_close_for <report-file> — the close reason triage's end implies.
# Four ends were acted on and one was not, which is the whole
# distinction the close carries
# (docs/product/stage-3-github-issues/labels.md#the-report-mirror).
report_close_for() {
  case "$(fm "$1" status)" in
    tracked|authored|fixed|routed) printf 'completed' ;;
    declined)                      printf 'not_planned' ;;
  esac
}

# The flag is not work. A merge that recorded nothing is still passed
# `--minted` with the mint's two empty outputs behind it, and it must
# still pay no round trip to the forge.
have_work=""
for sf in "$@"; do
  case "$sf" in --minted) ;; *) have_work=yes; break ;; esac
done
if [ -z "$have_work" ]; then
  echo "No approval was recorded by this merge — no label to re-derive."
  exit 0
fi

# Everything after --minted is the mint's, so the flag's *position* is
# the whole of what it says: a line carrying it twice was built wrong,
# and the second one silently renaming nothing is how a miswired caller
# stays miswired — with every id behind the misplaced flag turned from a
# notice about the repository into a red step. Judged beside the check
# above so a wrong line is refused before the forge is asked anything.
seen_minted=""
for sf in "$@"; do
  [ "$sf" = "--minted" ] || continue
  [ -z "$seen_minted" ] || {
    echo "rederive_labels.sh: --minted given twice — everything after it is the mint's, so a second one names no second set" >&2
    exit 3
  }
  seen_minted=yes
done

# fetch_mirrors <label> — one kind's mirror list. Identity is the tag in
# the title, the same way every other lookup here resolves it. The row
# shape is the one every reader of this list requests — body included and
# unused here — because one shape across all three readers is worth more
# than one saved field.
fetch_mirrors() {
  gh api "repos/${REPO}/issues?labels=${1}&state=all&per_page=100" \
    --paginate \
    --jq '.[] | [.number, .state, ((.labels // []) | map(.name) | join(",")), (.title | @base64), ((.body // "") | @base64)] | @tsv'
}

# The task mirrors, read once up front — and read again when an id
# misses. One read for the whole run is the right shape and its timing
# was the fault: the same job mints mirrors seconds before this step, so
# the list can be older than the mirror it is asked about
# (work/reports/report-0021-rederive-labels-sh.md).
ISSUES=$(fetch_mirrors 'writrun:task')

# The title is lowercased before the match rather than matched with a
# character class per letter — the answer was already lowercased on the
# way out, so the two are the same and one of them is legible.
id_of_title() {   # id_of_title <title> [kind] — default task
  local kind="${2:-task}"
  printf '%s' "$1" | tr '[:upper:]' '[:lower:]' | sed -n \
    -e "s/^\[\(${kind}-[0-9][0-9]*\)\].*/\1/p" \
    -e "s/^\(${kind}-[0-9][0-9]*\)[[:space:]].*/\1/p" \
    | head -n1
}
num_of_id() {
  printf '%s' "$1" | tr '[:upper:]' '[:lower:]' | sed -n \
    -e 's/^task-0*\([0-9][0-9]*\)$/\1/p' \
    -e 's/^report-0*\([0-9][0-9]*\)$/\1/p'
}

# The ids this job already answered for, collected before anything is
# projected. The same id can arrive twice — once from the commit range,
# once from the mint — so reading the flag as the loop walks past it
# would let the earlier arrival settle a question the later one answers.
MINTED_TASKS=""
MINTED_REPORTS=""
minting=""
for sf in "$@"; do
  if [ "$sf" = "--minted" ]; then minting=yes; continue; fi
  [ -n "$minting" ] || continue
  mid=$(num_of_id "$(basename "$sf" .md | sed -E 's/^((task|report)-[0-9]+).*/\1/')")
  [ -n "$mid" ] || continue
  case "$(basename "$sf")" in
    report-*) MINTED_REPORTS="$MINTED_REPORTS $mid" ;;
    task-*)   MINTED_TASKS="$MINTED_TASKS $mid" ;;
  esac
done
# Whether the flag was given at all, kept apart from what it named: the
# flag with nothing behind it entitles nobody, and its absence entitles
# everybody.
MINTED_DECLARED="$minting"

# minted_num <kind> <number> — did this job's mint step answer for this
# id. The one membership test, over the numbers the lists hold.
minted_num() {
  case "$1" in
    report) case " $MINTED_REPORTS " in *" $2 "*) return 0 ;; esac ;;
    *)      case " $MINTED_TASKS " in *" $2 "*) return 0 ;; esac ;;
  esac
  return 1
}

# One re-read answers every id still unresolved, so the budget is the
# run's and not each id's: staleness seen once is paid for once, and a
# run that sees none pays nothing. Two re-reads, spaced — the observed
# gap between a create and a read that could not see it was about four
# seconds. The wait is overridable because the suite must not spend it.
#
# Who may spend it is the caller's declaration. Behind a `--minted`
# flag, only the ids it names: a miss on any other id cannot be this
# run's staleness, and the list in hand already answers it. No flag at
# all entitles every miss, so a caller that mints and stays silent
# spends seconds it did not need to rather than losing a mirror.
REFRESH_BUDGET=2
REFRESH_WAIT="${WRITRUN_MIRROR_REFRESH_WAIT:-3}"
task_refreshes=0
report_refreshes=0

# refresh_mirrors <kind> — spend one of the run's re-reads on one list,
# or say there is none left to spend. A read that fails is spent all the
# same: the list in hand stands, the id stays unresolved, and that is an
# answer every caller here already has.
refresh_mirrors() {
  local out
  case "$1" in
    report)
      [ "$report_refreshes" -lt "$REFRESH_BUDGET" ] || return 1
      report_refreshes=$((report_refreshes + 1))
      sleep "$REFRESH_WAIT"
      if out=$(fetch_mirrors 'writrun:report'); then REPORT_ISSUES="$out"; fi ;;
    *)
      [ "$task_refreshes" -lt "$REFRESH_BUDGET" ] || return 1
      task_refreshes=$((task_refreshes + 1))
      sleep "$REFRESH_WAIT"
      if out=$(fetch_mirrors 'writrun:task'); then ISSUES="$out"; fi ;;
  esac
}

# find_mirror <kind> <id-number> — the mirror row for one id in the list
# already in hand, or nothing.
#
# The oldest *open* match wins, not the first: the list arrives newest
# first, and a duplicate pair — two runs racing one mint, report-0038 —
# would put the younger mirror first and take this pass's label while
# the reconciler retires it (decision 0073). The survivor there is the
# oldest open mirror, so the row labelled here is the same one. With
# one match, or none open, the answer is what it always was.
find_mirror() {
  local kind="$1" want="$2" rows n istate labels tb bb t tn first="" best="" best_n=""
  case "$kind" in
    report) rows="$REPORT_ISSUES" ;;
    *)      rows="$ISSUES" ;;
  esac
  while IFS="$TAB" read -r n istate labels tb bb; do
    [ -n "$n" ] || continue
    t=$(printf '%s' "$tb" | b64_decode)
    tn=$(num_of_id "$(id_of_title "$t" "$kind")")
    [ -n "$tn" ] || continue
    if [ "$tn" -eq "$want" ] 2>/dev/null; then
      [ -n "$first" ] || first="${n}${TAB}${istate}${TAB}${labels}"
      if [ "$istate" = "open" ]; then
        if [ -z "$best_n" ] || [ "$n" -lt "$best_n" ] 2>/dev/null; then
          best_n="$n"
          best="${n}${TAB}${istate}${TAB}${labels}"
        fi
      fi
    fi
  done <<EOF
$rows
EOF
  printf '%s' "${best:-$first}"
}

# resolve_mirror <kind> <id-number> — the same lookup, retried against a
# re-read list when it misses, with the answer left in FOUND.
#
# A miss on a minted id is not a conclusion. The list was read before
# this id was looked up, and the mirror can have been minted in between —
# by the very job running this. The re-read is what keeps the rule the
# pass exists for: every task the merge touched wears the label its file
# names (docs/product/stage-3-github-issues/labels.md).
#
# Entitlement is asked before anything is spent: where a `--minted` flag
# was given, a miss on an id outside it is the answer, returned from the
# list in hand — no re-read, no wait.
FOUND=""
resolve_mirror() {
  FOUND=$(find_mirror "$1" "$2")
  if [ -n "$MINTED_DECLARED" ] && ! minted_num "$1" "$2"; then
    return 0
  fi
  while [ -z "$FOUND" ] && refresh_mirrors "$1"; do
    FOUND=$(find_mirror "$1" "$2")
  done
}

# The report mirrors, fetched once and only if a report is named. A
# project that has never recorded one must not pay a forge call per merge
# for a list that would always come back empty.
REPORT_ISSUES=""
REPORT_ISSUES_READ=""
report_issues() {
  [ -z "$REPORT_ISSUES_READ" ] || return 0
  REPORT_ISSUES_READ=yes
  REPORT_ISSUES=$(fetch_mirrors 'writrun:report')
}

# project_report <report-id> — the whole answer for one report: find its
# mirror, and make it agree with the file.
project_report() {
  local rid="$1" rnum rf want closing found n istate labels kept args l
  rnum=$(num_of_id "$rid")
  [ -n "$rnum" ] || return 0
  rf=$(queue_file work/reports report "$rid")
  if [ -z "$rf" ]; then
    echo "${rid}: no report file on this branch — nothing to derive from."
    return 0
  fi
  want=$(report_label_for "$rf")
  closing=$(report_close_for "$rf")
  if [ -z "$want" ] && [ -z "$closing" ]; then
    echo "${rid} is $(fm "$rf" status) — not a status this step can project."
    return 0
  fi

  report_issues
  resolve_mirror report "$rnum"
  found="$FOUND"

  if [ -z "$found" ]; then
    unresolved report "$rid"
    return 0
  fi
  n=$(printf '%s' "$found" | cut -f1)
  istate=$(printf '%s' "$found" | cut -f2)
  labels=$(printf '%s' "$found" | cut -f3)

  if [ -n "$closing" ]; then
    if [ "$istate" != "open" ]; then
      echo "${rid}: mirror #${n} is already closed."
      return 0
    fi
    kept=$(printf '%s\n' "$labels" | tr ',' '\n' | grep -v '^status:' | sed '/^$/d' || true)
    args=()
    while IFS= read -r l; do
      [ -n "$l" ] || continue
      args+=(-f "labels[]=$l")
    done <<EOF
$kept
EOF
    gh api -X PUT "repos/${REPO}/issues/${n}/labels" ${args[@]+"${args[@]}"} >/dev/null
    gh api -X PATCH "repos/${REPO}/issues/${n}" \
      -f state=closed -f "state_reason=${closing}" >/dev/null
    echo "${rid} → mirror #${n} closed as ${closing}"
    return 0
  fi

  if [ "$istate" != "open" ]; then
    echo "${rid}: mirror #${n} is closed — no label is written."
    return 0
  fi
  ensure_label "status:open" "0e8a16" "Recorded and awaiting triage"
  set_status "$n" "$labels" "status:open"
  echo "${rid} → status:open (re-derived from the queue)"
}

set_status() {   # set_status <issue> <labels-csv> <status-label> [extra-label]
  local kept l args
  kept=$(printf '%s\n' "$2" | tr ',' '\n' | grep -v '^status:' | sed '/^$/d' || true)
  args=()
  while IFS= read -r l; do
    [ -n "$l" ] || continue
    args+=(-f "labels[]=$l")
  done <<EOF
$kept
EOF
  args+=(-f "labels[]=$3")
  [ -n "${4:-}" ] && args+=(-f "labels[]=$4")
  gh api -X PUT "repos/${REPO}/issues/${1}/labels" "${args[@]}" >/dev/null
}

ensure_label() {   # ensure_label <name> <color> <description>
  local out
  if ! out=$(gh api -X POST "repos/${REPO}/labels" \
      -f "name=$1" -f "color=$2" -f "description=$3" 2>&1); then
    printf '%s\n' "$out" | grep -q "HTTP 422" \
      || { printf '%s\n' "$out" >&2; exit 2; }
  fi
}

ensure_origin_label() {   # ensure_origin_label <label>
  case "$1" in
    origin:rule)   ensure_label "origin:rule" "0075ca" "Derived from an authored rule" ;;
    origin:report) ensure_label "origin:report" "d73a4a" "Born from a report of work an existing rule authorizes" ;;
  esac
}

# minted <kind> <id> — minted_num over an id string, converting first.
minted() {
  local n
  n=$(num_of_id "$2")
  [ -n "$n" ] || return 1
  minted_num "$1" "$n"
}

# unresolved <kind> <id> — what a lookup that found nothing means, which
# is decided by who named the id. A mirror the mint answered for exists,
# so not finding it is this pass's own defect and the step must fail:
# minted and never labelled is the one outcome no later event corrects.
# Any other id may simply never have been mirrored, and that is a finding
# to report rather than a fault to raise.
FAILED=""
unresolved() {
  local tag
  if minted "$1" "$2"; then
    tag=$(printf '%s' "$2" | tr '[:lower:]' '[:upper:]')
    echo "${2}: no mirrored Issue, and this job's mint answered for one." >&2
    echo "  Its [${tag}] mirror exists and this pass left it unlabelled." >&2
    FAILED=yes
  else
    echo "${2}: no mirrored Issue."
  fi
}

seen=""
seen_reports=""
for sf in "$@"; do
  [ "$sf" = "--minted" ] && continue
  # A spec file names its task; a task file or bare task id names itself;
  # a report file or bare report id is the other kind entirely — the
  # callers pass whichever the merge put in front of them.
  case "$(basename "$sf")" in
    report-*)
      rref=$(basename "$sf" .md | sed -E 's/^(report-[0-9]+).*/\1/')
      rnum=$(num_of_id "$rref")
      [ -n "$rnum" ] || continue
      case " $seen_reports " in *" $rnum "*) continue ;; esac
      seen_reports="$seen_reports $rnum"
      project_report "$rref"
      continue ;;
    task-*)
      tref=$(basename "$sf" .md | sed -E 's/^(task-[0-9]+).*/\1/') ;;
    *)
      # Through $QUEUE_ROOT like every other read: a spec file named on
      # the command line is still a queue file, and reading it from the
      # tree while the tasks come from the ref is the two-answers
      # divergence in miniature.
      [ -f "$QUEUE_ROOT/$sf" ] || continue
      tref=$(fm "$QUEUE_ROOT/$sf" task_ref) ;;
  esac
  [ -n "$tref" ] || continue
  tnum=$(num_of_id "$tref")
  [ -n "$tnum" ] || continue
  # A pull request may approve several specs of one task; label it once.
  case " $seen " in *" $tnum "*) continue ;; esac
  seen="$seen $tnum"

  tf=$(queue_file work/tasks task "$tref")
  if [ -z "$tf" ]; then
    echo "${tref}: no task file on this branch — nothing to derive from."
    continue
  fi

  want=$(label_for "$tf")
  closing=$(close_for "$tf")
  if [ -z "$want" ] && [ -z "$closing" ]; then
    echo "${tref} is $(fm "$tf" status) — its label is not this step's to write."
    continue
  fi

  resolve_mirror task "$tnum"
  found="$FOUND"

  if [ -z "$found" ]; then
    unresolved task "$tref"
    continue
  fi

  num=$(printf '%s' "$found" | cut -f1)
  istate=$(printf '%s' "$found" | cut -f2)
  labels=$(printf '%s' "$found" | cut -f3)

  # The `origin:` label never changes and never comes off, so this pass
  # only ever adds one: a mirror minted before the field existed gains it
  # here, once, from the stored field
  # (docs/product/stage-3-github-issues/labels.md). One already worn is
  # left exactly as it is.
  olbl=""
  if ! printf '%s\n' "$labels" | tr ',' '\n' | grep -q '^origin:'; then
    olbl=$(origin_label_for "$tf")
  fi

  # Closing wins. A mirror closed by the same merge is out of the
  # pipeline, and every label names a place inside it.
  if [ "$istate" != "open" ]; then
    echo "${tref}: mirror #${num} is closed — no label is written."
    continue
  fi

  # Past every path that writes nothing, and only here: creating the
  # label in the repository is itself a write, and a mirror this pass
  # decided to leave alone must not leave a label behind it.
  if [ -n "$olbl" ]; then ensure_origin_label "$olbl"; fi

  # A terminal status closes the mirror: the close and its reason are
  # the state, and no status label survives it.
  if [ -n "$closing" ]; then
    kept=$(printf '%s\n' "$labels" | tr ',' '\n' | grep -v '^status:' | sed '/^$/d' || true)
    args=()
    while IFS= read -r l; do
      [ -n "$l" ] || continue
      args+=(-f "labels[]=$l")
    done <<EOF
$kept
EOF
    [ -n "$olbl" ] && args+=(-f "labels[]=$olbl")
    gh api -X PUT "repos/${REPO}/issues/${num}/labels" ${args[@]+"${args[@]}"} >/dev/null
    gh api -X PATCH "repos/${REPO}/issues/${num}" \
      -f state=closed -f "state_reason=${closing}" >/dev/null
    echo "${tref} → mirror #${num} closed as ${closing}"
    continue
  fi

  case "$want" in
    status:ready)
      ensure_label "status:ready" "0e8a16" "Ready for development — waiting for someone to take it" ;;
    status:backlog)
      ensure_label "status:backlog" "fbca04" "In the queue, with a spec it references not yet approved" ;;
    status:in-progress)
      ensure_label "status:in-progress" "bfd4f2" "Someone is working on it; leave the worker alone" ;;
    status:in-review)
      ensure_label "status:in-review" "d93f0b" "A pull request is open and waiting on review" ;;
    status:blocked)
      ensure_label "status:blocked" "b60205" "Stalled by something outside the queue — blocked_reason says what" ;;
  esac
  set_status "$num" "$labels" "$want" "$olbl"
  echo "${tref} → ${want} (re-derived from the queue)"
done

# Failing here rather than at the miss: one mirror this pass cannot find
# is no reason to leave the rest of the run unlabelled, and the step
# still must not report success.
[ -z "$FAILED" ] || exit 1
exit 0
