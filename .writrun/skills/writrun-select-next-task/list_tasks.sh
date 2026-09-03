#!/usr/bin/env bash
# list_tasks.sh — prints the tasks eligible for work, and what holds the
# rest back.
#
#   bash .writrun/skills/writrun-select-next-task/list_tasks.sh
#
# Eligibility is steps 2–4 of the selection algorithm
# (docs/technical/README.md#task-selection-algorithm): `ready`, every
# `depends_on` done, every `spec_ref` approved or implemented. Those
# are the gates, and they bind everyone.
#
# The order printed is steps 5–6 — priority, then `created`, then `id`. For
# an agent that order is binding. For a person it is a suggestion: any task
# in the Available list may be taken, and taking a lower one bypasses
# nothing. That is the whole reason this prints a list instead of an answer.
#
# A task already in flight is reported as such, and reported as *paused*
# when an amendment has put one of its specs' approval back in question —
# derived from this checkout and the open pull requests together, never
# from a stored field (docs/product/stage-2-pull-requests/statuses.md#an-amendment-under-an-open-pull-request).
#
# **An open report is named, never selected.** It enters no ordering and
# moves no exit code — 0 still means a task is available and 1 still
# means none is, which is what every caller branching on the status
# depends on. Naming is not selecting: what the section asks for is
# triage, and `work/reports/` never becomes a second queue
# (docs/technical/selection.md#an-open-report-is-named-never-selected).
#
# The three directories are taken together or not at all — the caller
# rule `technical/distribution.md#running-the-checks` states for
# `check_front_matter.sh`, holding here for the same reason. A caller
# naming the first two and not the third reads tasks from the tree it
# named and reports from the working directory: it reports nothing
# waiting where plenty is, or names reports that are not the named
# tree's. Nothing enforces it — `${3:-work/reports}` cannot tell a
# default from a mismatch.
#
# Exit codes: 0 something is available; 1 nothing is; 3 no work/tasks/.
#
# Portable awk/sed only — no gawk extensions.

set -uo pipefail

TASK_DIR="${1:-work/tasks}"
SPEC_DIR="${2:-work/specs}"
REPORT_DIR="${3:-work/reports}"

[ -d "$TASK_DIR" ] || { echo "No such directory: $TASK_DIR" >&2; exit 3; }

field() { sed -n "s/^$2: *//p" "$1" | head -n1; }
title() { sed -n 's/^# *//p' "$1" | head -n1; }

# list_field <file> <name> — a [a, b] front-matter list as bare words.
list_field() {
  field "$1" "$2" | tr -d '[]' | tr ',' ' '
}

# A queue file is named <id>.md or <id>-<subject>.md — the subject slug
# makes a listing readable and is never identity, so both resolve here.
queue_file() {
  find "$1" \( -iname "$2.md" -o -iname "$2-*.md" \) 2>/dev/null | head -n1
}

# num_file <dir> <prefix> <number> — the queue file whose id is that
# number, whatever width it was written at. A branch names a number, not
# an id, so `spec/0001-name` and a historical `spec/001-name` must both
# reach the one spec-0001 file that exists.
num_file() {
  local want f n
  want=$(printf '%s' "$3" | sed -n 's/^0*\([0-9][0-9]*\)$/\1/p')
  [ -n "$want" ] || return 0
  for f in "$1"/"$2"-*.md; do
    [ -f "$f" ] || continue
    n=$(basename "$f" .md | tr '[:upper:]' '[:lower:]' \
      | sed -n "s/^$2-0*\([0-9][0-9]*\).*/\1/p")
    [ -n "$n" ] || continue
    if [ "$n" -eq "$want" ] 2>/dev/null; then printf '%s' "$f"; return 0; fi
  done
  return 0
}

# id_num <spec-0007> — an id's digits with the zero padding dropped, so
# a comparison never depends on the width the id was written at. The
# companion to num_file, for the times the file is not what is wanted.
id_num() {
  printf '%s' "$1" | tr '[:upper:]' '[:lower:]' \
    | sed -n 's/^[a-z][a-z]*-0*\([0-9][0-9]*\)$/\1/p'
}

# spec_num_of <path> — the spec a repository path belongs to, as a bare
# number; nothing for a path that is not a spec file.
spec_num_of() {
  printf '%s' "$1" | tr '[:upper:]' '[:lower:]' \
    | sed -n 's|^work/specs/spec-0*\([0-9][0-9]*\).*|\1|p'
}

spec_status() {
  local f
  f=$(queue_file "$SPEC_DIR" "$1")
  [ -n "$f" ] || { echo "missing"; return; }
  field "$f" status
}

task_status() {
  local f
  f=$(queue_file "$TASK_DIR" "$1")
  [ -n "$f" ] || { echo "missing"; return; }
  field "$f" status
}

rank() {
  case "$1" in high) echo 1 ;; medium) echo 2 ;; low) echo 3 ;; *) echo 4 ;; esac
}

# --- what is already in flight -------------------------------------------
#
# The queue files cannot answer this. A task someone started an hour ago is
# briefly still resting in `main` until the recording commit lands, so a lister
# that reads only files hands the same task to the next person who asks.
#
# This is not a claim and WritRun has no claim mechanism — reserving work is
# a tracker's job (see docs/about.md's non-goals). It is the one real-time
# signal a forge can be asked for: an open pull request means somebody is
# already working on that task.
#
# Set WRITRUN_PR_LIST to bypass `gh` — lines of
# "number<TAB>branch<TAB>author<TAB>title". The title is optional: a line
# without it still resolves through the branch, which is what a pull
# request opened before the tag convention looks like.
# Absent `gh`, a remote, or authentication, this degrades to files only and
# says so rather than implying nobody is working.

taken=""       # id|#number|author
idless=""      # numbers of the open pull requests carrying no task id
pr_source="none"

pr_lines=""
PR_FETCH_LIMIT=200
if [ -n "${WRITRUN_PR_LIST:-}" ]; then
  pr_lines="$WRITRUN_PR_LIST"
  pr_source="supplied"
elif command -v gh >/dev/null 2>&1; then
  # gh defaults to 30, and a silently truncated list reports a taken task
  # as free — the exact lie this section exists to prevent. The limit is
  # raised, and *hitting* it is reported at the end rather than passed
  # off as a complete answer.
  if pr_lines=$(gh pr list --state open --limit "$PR_FETCH_LIMIT" \
        --json number,headRefName,author,title \
        --jq '.[] | "\(.number)\t\(.headRefName)\t\(.author.login)\t\(.title)"' 2>/dev/null); then
    pr_source="gh"
  fi
fi

if [ "$pr_source" != "none" ]; then
  while IFS="$(printf '\t')" read -r num branch author ptitle; do
    [ -n "$branch" ] || continue

    # The title is the authority on which tasks a pull request carries: a
    # branch name holds one id, and a PR may carry several. Every task
    # tagged in the leading run is in flight; only when the title carries
    # no tag does the branch marker answer instead.
    tagged=""
    rest="${ptitle:-}"
    while :; do
      rest=$(printf '%s' "$rest" | sed 's/^[[:space:]]*//')
      tg=$(printf '%s' "$rest" \
        | sed -n 's/^\[[Tt][Aa][Ss][Kk]-0*\([0-9][0-9]*\)\].*/\1/p')
      [ -n "$tg" ] || break
      tf=$(num_file "$TASK_DIR" task "$tg")
      if [ -n "$tf" ]; then tid=$(field "$tf" id)
      else tid=$(printf 'task-%04d' "$tg"); fi
      tagged="${tagged} ${tid}"
      rest=$(printf '%s' "$rest" | sed 's/^\[[Tt][Aa][Ss][Kk]-[0-9][0-9]*\]//')
    done
    if [ -n "$tagged" ]; then
      for tid in $tagged; do
        taken="${taken}$(printf '%s' "$tid" | tr '[:upper:]' '[:lower:]')|#${num}|${author}"$'\n'
      done
      continue
    fi

    # A branch is named after the task being worked, or historically after
    # the spec, which resolves back through its task_ref.
    ref=""
    tn=$(printf '%s' "$branch" | sed -n 's|^[a-z]*/\{0,1\}task-0*\([0-9][0-9]*\).*|\1|p')
    if [ -z "$tn" ]; then
      sn=$(printf '%s' "$branch" | sed -n 's|^[a-z]*/\{0,1\}spec-0*\([0-9][0-9]*\).*|\1|p')
      if [ -z "$sn" ]; then
        # `spec/0003-name` — the bare number form the branch convention uses.
        sn=$(printf '%s' "$branch" | sed -n 's|^spec/0*\([0-9][0-9]*\).*|\1|p')
      fi
      if [ -n "$sn" ]; then
        sf=$(num_file "$SPEC_DIR" spec "$sn")
        [ -n "$sf" ] && ref=$(field "$sf" task_ref)
      fi
      if [ -z "$ref" ]; then
        tn=$(printf '%s' "$branch" | sed -n 's|^task/0*\([0-9][0-9]*\).*|\1|p')
      fi
    fi
    if [ -z "$ref" ] && [ -n "$tn" ]; then
      tf=$(num_file "$TASK_DIR" task "$tn")
      if [ -n "$tf" ]; then ref=$(field "$tf" id)
      else ref=$(printf 'task-%04d' "$tn"); fi
    fi
    if [ -n "$ref" ]; then
      taken="${taken}$(printf '%s' "$ref" | tr '[:upper:]' '[:lower:]')|#${num}|${author}"$'\n'
    else
      # It works no task — but it may still be amending one's spec, which
      # is what suspends that task. Kept for the file-list pass below.
      idless="${idless}${num}"$'\n'
    fi
  done <<EOF
$pr_lines
EOF
fi

# --- what an open pull request proposes for a spec ------------------------
#
# An amendment cut while a task rides an open pull request suspends that
# task: a spec whose approval is in question authorizes nothing, so the
# work waits until the amendment merges
# (docs/product/stage-2-pull-requests/statuses.md#an-amendment-under-an-open-pull-request).
# The pause is derived and never stored (decision 0059), and it is
# visible only in the union of two halves — this checkout, where an
# already-landed amendment shows the spec back in `draft`, and the open
# pull requests, where one still riding shows nothing on the branch at
# all. For the whole window an amendment is open, the files alone report
# a healthy queue.
#
# Only pull requests carrying no task id are read, and that is the whole
# rule rather than a shorthand for "not this task's own": an amendment is
# a `queue/` change that deliberately carries no id (decision 0059),
# while a pull request carrying one is work on that task — and a task's
# own pull request touches its own spec at the end of every
# implementation, which would otherwise report every finished task as
# suspended by itself. Two shapes therefore fall outside this half and
# are held by the gate instead, which makes no such exclusion: an
# amendment cut on an id-carrying branch, and one named after the spec
# (`spec/NNNN-...`), which resolves through the spec's task_ref back to
# the task above and reads as work on it.
#
# A file list is all the forge is asked for, so this half reads "touches
# one of its specs" where the rule reads "proposes returning one of its
# specs to `draft`" — a name cannot answer the narrower question, and
# spec-0037 chose the name. The widening is deliberate and it errs toward
# the pause: a `queue/` change that only fixes a typo in the spec of a
# task in flight reads as an amendment here, while the gate, which has
# both ends of the diff, correctly says nothing is suspended. A pause
# that turns out to be nothing costs a look; a missed one costs work
# built on an authorization that was being withdrawn.
#
# Set WRITRUN_PR_FILES to bypass `gh` — lines of "number<TAB>path". The
# forge half degrades exactly as the list above does: with no answer, a
# spec already returned to draft here is still caught, which is the whole
# Stage 1 story.

amend=""            # <spec-number>|#<pr-number>
files_partial=""

# Nothing but an in-flight task can be suspended, so a queue with none
# asks the forge nothing — the fetch is per pull request, and paying for
# it to answer a question nobody has is how a lister becomes too slow to
# run.
any_in_flight=""
grep -lE '^status: (in-progress|in-review)[[:space:]]*$' "$TASK_DIR"/*.md \
  >/dev/null 2>&1 && any_in_flight=yes

if [ -n "$any_in_flight" ] && [ -n "${WRITRUN_PR_FILES:-}" ]; then
  while IFS="$(printf '\t')" read -r num path; do
    [ -n "$num" ] && [ -n "$path" ] || continue
    printf '%s' "$idless" | grep -qxF "$num" || continue
    sn=$(spec_num_of "$path")
    [ -n "$sn" ] && amend="${amend}${sn}|#${num}"$'\n'
  done <<EOF
$WRITRUN_PR_FILES
EOF
elif [ -n "$any_in_flight" ] && [ "$pr_source" = "gh" ] && [ -n "$idless" ]; then
  for num in $idless; do
    # `{owner}/{repo}` is resolved by gh from the checkout's remote, the
    # same endpoint and pagination the id check reads.
    if ! files=$(gh api "repos/{owner}/{repo}/pulls/${num}/files" --paginate \
          --jq '.[].filename' 2>/dev/null); then
      files_partial=yes
      continue
    fi
    while IFS= read -r path; do
      [ -n "$path" ] || continue
      sn=$(spec_num_of "$path")
      [ -n "$sn" ] && amend="${amend}${sn}|#${num}"$'\n'
    done <<EOF
$files
EOF
  done
fi

taken_by() {
  printf '%s' "$taken" | while IFS='|' read -r id num who; do
    [ "$id" = "$(printf '%s' "$1" | tr '[:upper:]' '[:lower:]')" ] && {
      printf '%s by @%s' "$num" "$who"; break
    }
  done
}

# suspension <task-file> — why a task in flight cannot advance, or
# nothing. Both halves of the union in one string: a spec that is no
# longer approved here, and a spec an open pull request proposes to
# change. Partial authorization is not authorization — one spec of
# several is enough.
suspension() {
  local why="" s ss sn pr
  for s in $(list_field "$1" spec_ref); do
    [ -n "$s" ] || continue
    ss=$(spec_status "$s")
    case "$ss" in
      approved|implemented) ;;
      *) why="${why}${s} is ${ss} here; " ;;
    esac
    sn=$(id_num "$s")
    [ -n "$sn" ] || continue
    for pr in $(printf '%s' "$amend" \
        | awk -F'|' -v n="$sn" '$1 == n { print $2 }' | sort -u); do
      why="${why}${s} is amended by ${pr}; "
    done
  done
  [ -n "$why" ] && printf '%s' "${why%; }"
  return 0
}

available=""
held=""
resumable=""
inflight=""
open_reports=""

for f in "$REPORT_DIR"/*.md; do
  [ -f "$f" ] || continue
  case "$(basename "$f")" in README.md|readme.md) continue ;; esac

  id=$(field "$f" id)
  [ -n "$id" ] || continue
  [ "$(field "$f" status)" = "open" ] || continue
  open_reports="${open_reports}${id}|$(title "$f")"$'\n'
done

for f in "$TASK_DIR"/*.md; do
  [ -f "$f" ] || continue
  case "$(basename "$f")" in README.md|readme.md) continue ;; esac

  id=$(field "$f" id)
  [ -n "$id" ] || continue
  st=$(field "$f" status)
  pr=$(field "$f" priority)
  cr=$(field "$f" created)
  tt=$(title "$f")

  # Step 0 — an in-flight task with no open pull request is work someone
  # abandoned without the forge hearing about it; resume it before
  # selecting anything new. One with an open pull request is theirs,
  # however stale — named as in flight, never hidden, and taking it over
  # is a human decision (technical/README.md#task-selection-algorithm).
  if [ "$st" = "in-progress" ] || [ "$st" = "in-review" ]; then
    who=$(taken_by "$id")
    pause=$(suspension "$f")
    if [ -n "$who" ]; then
      inflight="${inflight}${id}|${who}|${pause}|${tt}"$'\n'
    else
      resumable="${resumable}${id}|${pause}|${tt}"$'\n'
    fi
    continue
  fi

  # Done and dropped are not held back — they are out of the running
  # entirely, and listing them as obstacles would grow with the project
  # until it buried the part that needs attention.
  [ "$st" = "done" ] && continue
  [ "$st" = "dropped" ] && continue

  if [ "$st" != "ready" ] && [ "$st" != "backlog" ]; then
    reason="$st"
    [ "$st" = "blocked" ] && reason="blocked: $(field "$f" blocked_reason)"
    held="${held}${id}|${reason}"$'\n'
    continue
  fi

  # The stored status summarizes the specs; the algorithm re-reads both
  # and stops loudly on a mismatch rather than trusting either side
  # alone (technical/README.md#task-selection-algorithm).
  why=""
  for d in $(list_field "$f" depends_on); do
    [ -n "$d" ] || continue
    ds=$(task_status "$d")
    [ "$ds" = "done" ] || why="${why}waiting on ${d} (${ds}); "
  done
  specs_hold=""
  for s in $(list_field "$f" spec_ref); do
    [ -n "$s" ] || continue
    ss=$(spec_status "$s")
    case "$ss" in
      approved|implemented) ;;
      *) specs_hold="${specs_hold}${s} is ${ss}; " ;;
    esac
  done
  if [ "$st" = "ready" ] && [ -n "$specs_hold" ]; then
    why="${why}MISMATCH — stored ready but ${specs_hold}"
  elif [ "$st" = "backlog" ] && [ -z "$specs_hold" ] && [ -z "$why" ]; then
    why="MISMATCH — stored backlog but every spec is approved; "
  else
    why="${why}${specs_hold}"
  fi

  if [ -n "$why" ]; then
    held="${held}${id}|${why%; }"$'\n'
    continue
  fi

  # Eligible on paper, but a pull request for it is already open. Not
  # "held back" — nothing is wrong with the task; someone is on it.
  who=$(taken_by "$id")
  if [ -n "$who" ]; then
    # Four fields, the same shape step 0 packs: a record one field short
    # would slide the title into the pause slot and print it as a reason
    # nobody derived. The pause is derived here too — this task reads
    # `ready` only because its status commit has not landed yet, and the
    # amendment that suspends it does not wait for that.
    pause=$(suspension "$f")
    inflight="${inflight}${id}|${who}|${pause}|${tt}"$'\n'
  else
    available="${available}$(rank "$pr")|${cr}|${id}|${pr}|${tt}"$'\n'
  fi
done

if [ -n "$resumable" ]; then
  echo "In progress — resume before selecting anything new:"
  printf '%s' "$resumable" | while IFS='|' read -r id pause tt; do
    [ -n "$id" ] || continue
    printf '  %-10s %s\n' "$id" "$tt"
    [ -n "$pause" ] && printf '  %-10s paused — %s; the work waits on the re-approval\n' "" "$pause"
  done
  echo
fi

if [ -n "$available" ]; then
  echo "Available — any of these may be taken:"
  printf '%s' "$available" | sed '/^$/d' | sort -t'|' -k1,1n -k2,2 -k3,3 \
    | while IFS='|' read -r _ _ id pr tt; do
        printf '  %-10s %-7s %s\n' "$id" "$pr" "$tt"
      done
  echo
  echo "Order is a suggestion for a person and binding for an agent."
else
  echo "Nothing is available."
fi

if [ -n "$inflight" ]; then
  echo
  echo "In flight — an open pull request already exists:"
  printf '%s' "$inflight" | sed '/^$/d' | sort | while IFS='|' read -r id who pause tt; do
    printf '  %-10s %-16s %s\n' "$id" "$who" "$tt"
    [ -n "$pause" ] && printf '  %-10s %-16s paused — %s; the work waits on the re-approval\n' "" "" "$pause"
  done
fi

if [ -n "$held" ]; then
  echo
  echo "Held back:"
  printf '%s' "$held" | sed '/^$/d' | sort | while IFS='|' read -r id why; do
    printf '  %-10s %s\n' "$id" "$why"
  done
fi

if [ -n "$open_reports" ]; then
  echo
  echo "Open reports — waiting to be triaged, never selected:"
  printf '%s' "$open_reports" | sed '/^$/d' | sort | while IFS='|' read -r id tt; do
    printf '  %-12s %s\n' "$id" "$tt"
  done
fi

if [ "$pr_source" = "none" ]; then
  echo
  echo "Note: could not reach GitHub, so nothing above accounts for work"
  echo "already in flight, nor for an amendment suspending a task from an"
  echo "open pull request — only a spec already returned to draft on this"
  echo "checkout was seen. Check open pull requests before starting."
elif [ -n "$files_partial" ]; then
  echo
  echo "Note: an open pull request would not yield its file list, so a task"
  echo "it suspends is not named above. Check open pull requests before"
  echo "starting."
elif [ "$pr_source" = "gh" ] \
    && [ "$(printf '%s\n' "$pr_lines" | grep -c .)" -ge "$PR_FETCH_LIMIT" ]; then
  echo
  echo "Note: the open pull request list hit the fetch limit (${PR_FETCH_LIMIT}),"
  echo "so the in-flight section may be incomplete. Check open pull requests"
  echo "before starting."
fi

[ -n "$available" ]
