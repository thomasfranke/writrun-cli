#!/usr/bin/env bash
# rederive_labels.sh — after a merge records an approval, label the tasks
# and reports it affected from the queue as it then stands.
#
# Usage: rederive_labels.sh <owner/repo> <spec-file|task-id|report-id>...
#   Run from a checkout of the authority branch, *after* the flip has been
#   committed to it — the whole point is to read the queue's new state.
#   `gh` must be on PATH and authenticated (GH_TOKEN in CI; a stub in the
#   test suite). With no arguments, it does nothing and says so.
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
# Exit codes: 0 done (including nothing to do); 3 usage error.
#
# Portable bash 3.2, POSIX awk/sed — no gawk extensions, no associative
# arrays. See the standing rule in docs/technical/decisions/.

set -euo pipefail

REPO="${1:?usage: rederive_labels.sh <owner/repo> <spec-file>...}"
shift

TAB=$(printf '\t')

if printf '' | base64 -d >/dev/null 2>&1; then B64_FLAG="-d"; else B64_FLAG="-D"; fi
b64_decode() { base64 "$B64_FLAG"; }

fm() {   # fm <file> <field>
  awk -v f="$2" '
    NR == 1 { if ($0 != "---") exit; next }
    /^---$/ { exit }
    sub("^" f ": *", "") { sub(/[[:space:]]*$/, ""); print; exit }
  ' "$1"
}

# queue_file <dir> <prefix> <id> — the file whose id is <id>, whatever its
# subject slug and whatever width its number was written at.
queue_file() {
  local dir="$1" prefix="$2" want f n
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
# Three ends were acted on and one was not, which is the whole
# distinction the close carries
# (docs/product/stage-3-github-issues/labels.md#the-report-mirror).
report_close_for() {
  case "$(fm "$1" status)" in
    tracked|authored|fixed) printf 'completed' ;;
    declined)               printf 'not_planned' ;;
  esac
}

if [ "$#" -eq 0 ]; then
  echo "No approval was recorded by this merge — no label to re-derive."
  exit 0
fi

# The mirrors, fetched once. Identity is the tag in the title, the same
# way every other lookup here resolves it. The row shape is the one every
# reader of this list requests — body included and unused here — because
# one shape across all three readers is worth more than one saved field.
ISSUES=$(gh api "repos/${REPO}/issues?labels=writrun:task&state=all&per_page=100" \
  --paginate \
  --jq '.[] | [.number, .state, ((.labels // []) | map(.name) | join(",")), (.title | @base64), ((.body // "") | @base64)] | @tsv')

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

# The report mirrors, fetched once and only if a report is named. A
# project that has never recorded one must not pay a forge call per merge
# for a list that would always come back empty.
REPORT_ISSUES=""
REPORT_ISSUES_READ=""
report_issues() {
  [ -z "$REPORT_ISSUES_READ" ] || return 0
  REPORT_ISSUES_READ=yes
  REPORT_ISSUES=$(gh api "repos/${REPO}/issues?labels=writrun:report&state=all&per_page=100" \
    --paginate \
    --jq '.[] | [.number, .state, ((.labels // []) | map(.name) | join(",")), (.title | @base64), ((.body // "") | @base64)] | @tsv')
}

# project_report <report-id> — the whole answer for one report: find its
# mirror, and make it agree with the file.
project_report() {
  local rid="$1" rnum rf want closing found n istate labels tb bb t tn kept args l
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
  found=""
  while IFS="$TAB" read -r n istate labels tb bb; do
    [ -n "$n" ] || continue
    t=$(printf '%s' "$tb" | b64_decode)
    tn=$(num_of_id "$(id_of_title "$t" report)")
    [ -n "$tn" ] || continue
    if [ "$tn" -eq "$rnum" ] 2>/dev/null; then
      found="${n}${TAB}${istate}${TAB}${labels}"; break
    fi
  done <<EOF
$REPORT_ISSUES
EOF

  if [ -z "$found" ]; then
    echo "${rid}: no mirrored Issue."
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
      || { printf '%s\n' "$out" >&2; exit 1; }
  fi
}

ensure_origin_label() {   # ensure_origin_label <label>
  case "$1" in
    origin:rule)   ensure_label "origin:rule" "0075ca" "Derived from an authored rule" ;;
    origin:report) ensure_label "origin:report" "d73a4a" "Born from a report of work an existing rule authorizes" ;;
  esac
}

seen=""
seen_reports=""
for sf in "$@"; do
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
      [ -f "$sf" ] || continue
      tref=$(fm "$sf" task_ref) ;;
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

  found=""
  while IFS="$TAB" read -r n istate labels tb bb; do
    [ -n "$n" ] || continue
    t=$(printf '%s' "$tb" | b64_decode)
    tn=$(num_of_id "$(id_of_title "$t")")
    [ -n "$tn" ] || continue
    if [ "$tn" -eq "$tnum" ] 2>/dev/null; then
      found="${n}${TAB}${istate}${TAB}${labels}"; break
    fi
  done <<EOF
$ISSUES
EOF

  if [ -z "$found" ]; then
    echo "${tref}: no mirrored Issue."
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
