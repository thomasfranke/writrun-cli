#!/usr/bin/env bash
# check_state.sh — verifies the task/spec/report lifecycle transitions a
# diff makes.
#
# Usage:
#   check_state.sh [<diff-range>]      # default: main...HEAD
#
# The rules, all derivable from the range alone:
#
#   A. A change may not move a spec draft -> approved. That transition is a
#      human gate (docs/product/stage-1-tasks-and-specs/gates.md); a pull request
#      approving its own spec is the exact thing the gate exists to stop.
#   B. A spec may not reach `implemented` from `draft`. Work is authorized
#      by approval, so an implemented spec was approved at some point first.
#   C. A task's `completed` date may not be written while any spec in its
#      spec_ref is not `implemented` — and the diff that implements a
#      task's last spec writes the date, or the task can never reach done.
#   D. A spec may not enter the tree already `implemented`. No legitimate
#      path produces a spec born past both gates.
#   E. (Stage 2+) A branch may not move a task between the machinery's
#      five working states — backlog, ready, in-progress, in-review,
#      done. That line has one writer, and it is not a branch
#      (docs/product/stage-2-pull-requests/statuses.md).
#   F. (Stage 2+) A branch may not edit `taken_by` — same single writer.
#   G. A hand move touching `blocked` is legal only between `blocked`
#      and `backlog` or `ready`; an in-flight task cannot be blocked.
#   H. `dropped` is terminal: reachable by hand from any non-terminal
#      state, left by nothing.
#   I. A branch may add a `provenance` entry and may never edit one it
#      found. This is the single exception to rules E and F's premise —
#      the one machine field a branch writes — and it exists because no
#      forge event carries a token count: only the session that spent
#      them knows. The permission is shaped so it cannot widen into "a
#      branch may edit front matter": appending is a different act from
#      editing, and every entry the base already held must still be
#      there, unchanged and in order.
#   J. A report that triage has ended never returns to `open`, and never
#      changes from one end to another. Its status is the route triage
#      took, not a lifecycle: the judgement was made, and a second
#      sighting is a second observation with its own date and its own id
#      — never the first one's file re-routed
#      (docs/product/concepts/report.md). This is the one rule here with
#      no stage condition. The status has a human or an agent writer at
#      every stage, because no forge event corresponds to a judgement, so
#      there is no version of this file the machinery owns instead.
#   K. (Stage 2+) The `tracked` route never rides. A report reaching
#      `tracked`, and the task that route mints (`origin: report`, newly
#      added), travel on a `report/` branch **and** in a change carrying
#      nothing outside `work/` — or not at all. Two conditions, because
#      the name alone is a rule agents keep: a refused implementing
#      branch renamed to `report/…` clears a check that reads only the
#      name, which is report-0003's failure reached *through* the gate.
#      What a reporting change touches is visible in its diff. It is the one
#      route that puts work in the queue, and what enters the queue
#      passes a gate: the reporting change's own pull request, whose
#      squash-merge is the assent that the finding deserves the work.
#      Riding an unrelated change mints a task that arrives `ready` at
#      merge with nobody having weighed it — the mirror born closed, the
#      evaluation the open Issue exists to invite silently never held
#      (docs/product/concepts/report.md#recording-rides-any-change--routing-to-the-queue-does-not).
#      The three routes that create no work — `authored`, `fixed`,
#      `declined` — keep the exemption and ride anything. The stage
#      condition is the rule's own premise: the gate is a pull request's
#      squash-merge, and a branchless project has none to be held to — it
#      takes the route on `main`, because that is the only place it has.
#
# A transition is read from the front matter at the two ends of the range
# — the file as the base knew it against the file as it is now — never
# grepped out of the diff text. A spec in this methodology's own
# repository legitimately quotes `status: draft` at column 0 in its body
# (the docs' own examples do), and a quoted line must never read as a
# transition.
#
# Rules B, C and D are what stop a contributor from routing around rule A
# by skipping intermediate states entirely. One case is deliberately not
# judged here: a spec added as `approved`. Locally, the legitimate flip
# (recorded after a maintainer approved the PR) and self-approval look
# identical — only the forge knows whether the review exists, so CI's
# `writrun check` settles that one against the PR's actual reviews, and
# locally it passes.
#
# Exit codes: 0 clean; 1 a rule was violated; 3 usage error or git failed.
#
# Portable awk/sed only — no gawk extensions. See the standing rule in
# docs/technical/decisions/.

set -euo pipefail

DIFF_RANGE="${1:-main...HEAD}"

err_tmp=$(mktemp "${TMPDIR:-/tmp}/check_state.XXXXXX")
if ! CHANGED=$(git diff --name-only "$DIFF_RANGE" 2>"$err_tmp"); then
  echo "git diff --name-only ${DIFF_RANGE} failed:" >&2
  head -n 2 "$err_tmp" >&2
  rm -f "$err_tmp"
  exit 3
fi
rm -f "$err_tmp"

# **A range that selects no commits is not a change that moved nothing.**
# Without branches — which is every project at level `tasks-and-specs` —
# `main...HEAD` is empty by construction, so this check would print OK
# having read nothing and vouched for it. Only the range forms are
# tested: a bare ref means "against the working tree", where "no commits
# selected" says nothing (spec-0013).
RANGE_COMMITS=""
case "$DIFF_RANGE" in
  *..*) RANGE_COMMITS=$(git rev-list --count "$DIFF_RANGE" 2>/dev/null || true) ;;
esac

# The base side of the range — what `git diff` itself compares against:
# the merge base for the three-dot form, the left rev for two-dot, the
# rev itself when the diff is against the working tree.
case "$DIFF_RANGE" in
  *...*)
    left="${DIFF_RANGE%%...*}"
    right="${DIFF_RANGE##*...}"
    BASE=$(git merge-base "${left:-HEAD}" "${right:-HEAD}")
    ;;
  *..*) BASE="${DIFF_RANGE%%..*}" ;;
  *)    BASE="$DIFF_RANGE" ;;
esac

status=0

# The stage gates rules E and F: below Stage 2 no machinery exists to own
# the status line, so hand moves are the contract, not a violation. The
# reader resolves next to this script (the settings file it reads is the
# working directory's); absent either, the documented default applies.
READ_SETTING="$(cd "$(dirname "$0")" && pwd)/../../scripts/stage-2-pull-requests/read_setting.sh"
STAGE=$(bash "$READ_SETTING" stage 2>/dev/null || printf '3')
case "$STAGE" in 1|2|3) ;; *) STAGE=3 ;; esac

ADDED=$(git diff --name-only --diff-filter=A "$DIFF_RANGE" 2>/dev/null || true)

# Renames, as "<path now><tab><path at the base>". A queue file is never
# renamed — identity is never order (AGENTS.md) — but a check that reads
# the base side at the *current* path answers "the file did not exist
# there" for one that did, and every rule downstream then judges a
# transition nobody made: a report long since `tracked`, moved to a better
# slug, would read as one reaching `tracked` in this range. The base path
# is the file's own, so it is where the base is read.
RENAMED=$(git diff --name-status -M "$DIFF_RANGE" 2>/dev/null \
  | awk -F'\t' '$1 ~ /^R/ { print $3 "\t" $2 }' || true)

# Rule K's input: the name of the branch the change is on. CI knows it and
# passes it as data — a fork controls the string and this script only ever
# prefix-matches it — and a local run falls back to the checkout. An empty
# HEAD_REF is an unset one: a push event hands the workflow exactly that.
#
# A name that cannot be read is not a pass. Detached HEAD with no
# environment leaves the rule with nothing to judge, and a check that
# silently drops a rule it could not run is the one failure mode worse
# than the rule not existing — so the skip is announced.
#
# Below Stage 2 the rule has nothing to hold the route to: it exists to
# make a pull request's squash-merge the assent, and a branchless project
# has no pull request to be the vehicle. The route runs on `main` there,
# legally, so the rule stands down — silently, because a stage the rule
# does not apply at is not a rule that could not be run.
#
# Rule K has two conditions and they fail differently. (a), the branch
# name, needs a name to read and can be left with none. (b), what the
# change carries, needs only the diff — which is always there — so it
# runs wherever the rule itself applies, including where (a) stood down.
# A change carrying code is not a reporting change whatever it is
# called, and that is the half the rename cannot clear.
HEAD_BRANCH=""
REPORT_BRANCH=skip
CARRIES_OUTSIDE=""
if [ "$STAGE" -ge 2 ]; then
  HEAD_BRANCH="${HEAD_REF:-}"
  if [ -z "$HEAD_BRANCH" ]; then
    HEAD_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || true)
    if [ "$HEAD_BRANCH" = "HEAD" ]; then HEAD_BRANCH=""; fi
  fi
  case "$HEAD_BRANCH" in
    "")        echo "Rule K: the branch half skipped — no branch name is readable:"
               echo "HEAD_REF is unset and HEAD is detached. The diff half still runs,"
               echo "so a change carrying code is still refused; set HEAD_REF to the"
               echo "head branch to have the name checked too." ;;
    report/*)  REPORT_BRANCH=true ;;
    *)         REPORT_BRANCH=false ;;
  esac

  # Everything the range touches that is not under `work/`. A reporting
  # change is the report plus the pair the route mints and nothing else
  # (docs/product/stage-1-tasks-and-specs/authoring.md), so a path
  # outside `work/` is implementation riding along — visible in the diff
  # without trusting anyone's choice of branch name.
  CARRIES_OUTSIDE=$(printf '%s\n' "$CHANGED" | sed '/^$/d' | grep -v '^work/' || true)
fi

# rides_outside — the refusal's shared tail: what the change carries that
# a reporting one may not, at most five paths and a count for the rest,
# because a refusal nobody reads to the end is a refusal that taught
# nothing.
rides_outside() {
  local n
  n=$(printf '%s\n' "$CARRIES_OUTSIDE" | sed '/^$/d' | wc -l | tr -d ' ')
  # awk, never `head`: `head` closes the pipe at five and the writer
  # upstream dies of SIGPIPE, which `pipefail` turns into an exit 141
  # that swallows the count, the explanation and every rule after this
  # one. awk reads to the end and prints five.
  printf '%s\n' "$CARRIES_OUTSIDE" | sed '/^$/d' \
    | awk 'NR <= 5 { print "    " $0 }' >&2
  [ "$n" -gt 5 ] && echo "    … and $((n - 5)) more" >&2
  return 0
}

# is_added <file> — was the file created by this diff?
is_added() { printf '%s\n' "$ADDED" | grep -qxF "$1"; }

# fm_field <field> — reads a file on stdin, returns the field from the
# front-matter block only: a quoted `status:` line in a body never counts.
fm_field() {
  awk -v f="$1" '
    NR == 1 { if ($0 != "---") exit; next }
    /^---$/ { exit }
    sub("^" f ": *", "") { sub(/[[:space:]]*$/, ""); print; exit }
  '
}

# fm_now / fm_base — the field in the working tree, and at the base end
# of the range (empty when the file did not exist there — a failing
# `git show` is that case, not an error).
fm_now()  { fm_field "$2" < "$1"; }
fm_base() { git show "${BASE}:$(base_path "$1")" 2>/dev/null | fm_field "$2" || true; }

# base_path <file> — where this file lived at the base end of the range:
# its own path, unless the range renamed it.
base_path() {
  printf '%s\n' "$RENAMED" \
    | awk -F'\t' -v n="$1" '$1 == n { print $2; seen = 1; exit } END { if (!seen) print n }'
}

# fm_ledger — the provenance entries, one per line, in order, from the
# front-matter block alone. The whole line is printed: rule I compares
# entries as written, because "unchanged" is about the text a reviewer
# reads, not about a value some parser recovered from it.
fm_ledger() {
  awk '
    NR == 1 { if ($0 != "---") exit; next }
    /^---$/ { exit }
    /^provenance:[[:space:]]*$/ { inl = 1; next }
    inl && /^  - / { print; next }
    inl { inl = 0 }
  '
}
ledger_now()  { fm_ledger < "$1"; }
ledger_base() { git show "${BASE}:$(base_path "$1")" 2>/dev/null | fm_ledger || true; }

for f in $CHANGED; do
  # Selection reads the filename's id, not the file's contents — the
  # shape schemas.md#front-matter-is-canonical already fixes for a task
  # (`task-NNNN-<subject>.md`) and a spec (the id agreeing with the
  # filename). A directory glob alone would read the queue's own
  # `README.md` as a task with no status and refuse it (report-0013);
  # narrowing to the reserved prefix lets a non-task file sit beside the
  # ones this check judges without being judged as one.
  case "$f" in
    work/specs/spec-*.md)
      [ -f "$f" ] || continue
      new=$(fm_now "$f" status)
      old=$(fm_base "$f" status)

      # D — a spec never enters the tree already implemented. Born
      # `approved` is deliberately not judged here: only the forge can
      # tell the recorded flip from self-approval, and CI does.
      if is_added "$f" && [ "$new" = "implemented" ]; then
        echo "FORBIDDEN: ${f} enters the tree already 'implemented'." >&2
        echo "  Work is authorized by approval and recorded after it; a spec" >&2
        echo "  born implemented skipped both gates. Add it as draft." >&2
        status=1
      fi

      # A — draft -> approved is never a contributor's to make.
      if [ "$old" = "draft" ] && [ "$new" = "approved" ]; then
        echo "FORBIDDEN: ${f} moves draft -> approved." >&2
        echo "  That transition is a human gate. Leave the spec in draft;" >&2
        echo "  approval is recorded when the change is approved." >&2
        status=1
      fi

      # B — implemented is only reachable from approved.
      if [ "$old" = "draft" ] && [ "$new" = "implemented" ]; then
        echo "FORBIDDEN: ${f} moves draft -> implemented, skipping approval." >&2
        echo "  A spec is authorized to be implemented only once approved." >&2
        status=1
      fi
      ;;

    work/reports/report-*.md)
      [ -f "$f" ] || continue
      new=$(fm_now "$f" status)
      old=$(fm_base "$f" status)

      # A report the diff creates is not judged: recording rides any
      # change, and one that arrives already triaged is the ordinary
      # case, not a skipped gate. Triage is not a human gate — the file
      # stays, the body carries the reason, and the mirror shows it
      # (docs/product/concepts/report.md).
      case "$old" in
        tracked|authored|fixed|declined)
          if [ "$new" = "open" ]; then
            echo "FORBIDDEN: ${f} returns ${old} -> open." >&2
            echo "  Triage ended this report; nothing reopens one. The same" >&2
            echo "  thing seen again is a second observation, so it is a" >&2
            echo "  second report — ids are never reused, and a recurrence" >&2
            echo "  sharing a file loses the date of the first sighting." >&2
            status=1
          elif [ -n "$new" ] && [ "$new" != "$old" ]; then
            echo "FORBIDDEN: ${f} moves ${old} -> ${new} — one end to another." >&2
            echo "  A report's status is the route triage took, and triage" >&2
            echo "  ran once. Re-routing it rewrites a judgement instead of" >&2
            echo "  recording a new one; record a second report." >&2
            status=1
          fi
          ;;
      esac

      # K — the tracked route never rides. Recording rides any change and
      # so do the ends that leave the queue untouched; this is the one
      # route that adds to it, so it travels on its own reporting change
      # and the maintainer's squash-merge of *that* pull request is the
      # assent (docs/product/concepts/report.md).
      #
      # Two conditions, and neither alone. (a) catches the honest case, a
      # session flipping a report on the implementing branch it already
      # stands on — and it is free. (b) is the one a rename cannot clear,
      # because an implementing change carries code whatever its branch is
      # called. (b) alone would pass a completion change whose spec
      # promised no doc delta: `work/`-only, and a `tracked` flip riding
      # *that* is still a ride.
      if [ "$new" = "tracked" ] && [ "$old" != "tracked" ]; then
        if [ "$REPORT_BRANCH" = false ]; then
          echo "FORBIDDEN: ${f} reaches 'tracked' on '${HEAD_BRANCH}'." >&2
          echo "  The tracked route is the one that puts work in the queue, and" >&2
          echo "  what enters the queue passes a gate: a reporting change of its" >&2
          echo "  own, on a report/ branch, whose merge is the assent that the" >&2
          echo "  finding deserves the work. Leave the report open here and route" >&2
          echo "  it on its own branch; fixed and declined still ride." >&2
          status=1
        elif [ -n "$CARRIES_OUTSIDE" ]; then
          echo "FORBIDDEN: ${f} reaches 'tracked' in a change carrying more than reporting:" >&2
          rides_outside
          echo "  A reporting change is the report and the pair the route mints," >&2
          echo "  and nothing else — the branch prefix is how such a change is" >&2
          echo "  named, never what makes it one. Route the report on a change" >&2
          echo "  that carries only work/; fixed and declined still ride." >&2
          status=1
        fi
      fi
      ;;

    work/tasks/task-*.md)
      [ -f "$f" ] || continue

      new=$(fm_now "$f" status)
      old=$(fm_base "$f" status)

      # A task the diff creates is not exempt from the single writer:
      # born in flight, born done, or born with a holder would be a
      # branch writing the machinery's line by arriving instead of by
      # editing. It enters as backlog (or blocked, with its reason);
      # the recording moves it from there.
      if [ "$STAGE" -ge 2 ] && is_added "$f"; then
        case "$new" in
          backlog|blocked) ;;
          *)
            echo "FORBIDDEN: ${f} enters the tree already '${new}'." >&2
            echo "  A task is born backlog (or blocked, with its reason); every" >&2
            echo "  other state is the machinery's to write after the merge." >&2
            status=1
            ;;
        esac
        tb_new=$(fm_now "$f" taken_by)
        if [ -n "$tb_new" ] && [ "$tb_new" != "null" ]; then
          echo "FORBIDDEN: ${f} enters the tree with taken_by '${tb_new}'." >&2
          echo "  Who has a task is the forge's record, machinery-written." >&2
          status=1
        fi
      fi

      # K, the other half — the task the tracked route mints travels with
      # the report that justified it. Judged on the task rather than only
      # on the report because the two are separable: a report tracked in
      # one change and its task added in another would pass a rule that
      # watched the status line alone.
      if is_added "$f" && [ "$(fm_now "$f" origin)" = "report" ] \
        && { [ "$REPORT_BRANCH" = false ] || [ -n "$CARRIES_OUTSIDE" ]; }; then
        if [ "$REPORT_BRANCH" = false ]; then
          echo "FORBIDDEN: ${f} is born of a report on '${HEAD_BRANCH}'." >&2
        else
          echo "FORBIDDEN: ${f} is born of a report in a change carrying more than reporting:" >&2
          rides_outside
        fi
        echo "  A task derived from a report is a reporting change on its own" >&2
        echo "  report/ branch — the pull request presents the report, the task" >&2
        echo "  and the spec together, and its merge is the judgement that the" >&2
        echo "  finding deserves the work. Derive it in a reporting change" >&2
        echo "  of its own — renaming this branch is not what clears the gate." >&2
        status=1
      fi

      # E — the five working states have one writer, and it is the
      # machinery on the authority branch, never a branch. Only from
      # Stage 2 up: with no forge there is no machinery, and statuses
      # stay hand-moved.
      if [ "$STAGE" -ge 2 ] && [ -n "$old" ] && [ "$new" != "$old" ]; then
        case "$old" in backlog|ready|in-progress|in-review|done)
          case "$new" in backlog|ready|in-progress|in-review|done)
            echo "FORBIDDEN: ${f} moves ${old} -> ${new} on a branch." >&2
            echo "  The working states are the machinery's, written on the" >&2
            echo "  authority branch from forge events — a branch never edits" >&2
            echo "  the status line (statuses.md). Leave it; the forge writes it." >&2
            status=1
            ;;
          esac ;;
        esac
      fi

      # F — taken_by is the same single writer's.
      if [ "$STAGE" -ge 2 ] && ! is_added "$f"; then
        tb_new=$(fm_now "$f" taken_by)
        tb_old=$(fm_base "$f" taken_by)
        if [ -n "$tb_old" ] && [ "$tb_new" != "$tb_old" ]; then
          echo "FORBIDDEN: ${f} edits taken_by ('${tb_old}' -> '${tb_new}') on a branch." >&2
          echo "  Who has a task is the forge's record, machinery-written." >&2
          status=1
        fi
      fi

      # G — blocked pairs with backlog/ready only. An in-flight task has
      # an open pull request; what stalls it is visible there, and the
      # status table draws no such edge.
      if [ "$new" = "blocked" ] && [ -n "$old" ] && [ "$old" != "blocked" ]; then
        case "$old" in
          backlog|ready) ;;
          *)
            echo "FORBIDDEN: ${f} moves ${old} -> blocked." >&2
            echo "  blocked is reachable from backlog or ready only (statuses.md)." >&2
            status=1
            ;;
        esac
      fi
      if [ "$old" = "blocked" ] && [ "$new" != "blocked" ] && [ -n "$new" ]; then
        case "$new" in
          backlog|ready) ;;
          *)
            echo "FORBIDDEN: ${f} moves blocked -> ${new}." >&2
            echo "  A released task returns to backlog or ready (statuses.md)." >&2
            status=1
            ;;
        esac
      fi

      # H — dropped is terminal.
      if [ "$old" = "dropped" ] && [ "$new" != "dropped" ] && [ -n "$new" ]; then
        echo "FORBIDDEN: ${f} moves dropped -> ${new} — dropped is terminal." >&2
        echo "  A dropped task stays dropped; new work is a new task." >&2
        status=1
      fi

      # I — the ledger is append-only. An entry is a fact about work
      # that happened; rewriting one rewrites the past, and the reason
      # this field exists at all is that the field beside it (taken_by)
      # keeps erasing itself.
      if ! is_added "$f"; then
        pv_old=$(ledger_base "$f")
        if [ -n "$pv_old" ]; then
          pv_new=$(ledger_now "$f")
          kept=$(printf '%s\n' "$pv_new" | head -n "$(printf '%s\n' "$pv_old" | wc -l | tr -d ' ')")
          if [ "$kept" != "$pv_old" ]; then
            echo "FORBIDDEN: ${f} edits a provenance entry it found rather than adding one." >&2
            echo "  The ledger is append-only: an entry records work that happened," >&2
            echo "  and every entry the base held stays as written, in order." >&2
            echo "  Add your entry at the end and leave the others alone." >&2
            status=1
          fi
        fi
      fi

      # C — the completed date is the worker's declaration of finishing,
      # and it may only be written once every referenced spec is
      # implemented. The reverse binds too: the diff that implements a
      # task's last spec writes the date, or no merge can ever move the
      # task to done.
      cd_new=$(fm_now "$f" completed)
      cd_old=$(fm_base "$f" completed)
      refs=$(fm_now "$f" spec_ref | tr -d '[]' | tr ',' ' ')
      all_implemented=true
      for ref in $refs; do
        [ -n "$ref" ] || continue
        # <id>.md or <id>-<subject>.md — the subject slug is not identity.
        spec=$(find work/specs \
          \( -iname "${ref}.md" -o -iname "${ref}-*.md" \) 2>/dev/null | head -n1)
        if [ -z "$spec" ]; then
          echo "BROKEN: ${f} lists ${ref} in spec_ref, which resolves to no file." >&2
          status=1
          all_implemented=false
          continue
        fi
        spec_status=$(fm_now "$spec" status)
        if [ "$spec_status" != "implemented" ]; then
          all_implemented=false
          # An added file's base read is empty — the same "no date yet"
          # as null, or a task born with its date would bypass this.
          if [ "$cd_new" != "null" ] && [ -n "$cd_new" ] \
            && { [ "$cd_old" = "null" ] || [ -z "$cd_old" ]; }; then
            echo "INCONSISTENT: ${f} writes its completed date but ${ref} is '${spec_status}'." >&2
            echo "  Fill the spec's Outcome and set it to implemented in this change." >&2
            status=1
          fi
        fi
      done
      ;;
  esac
done

# The other half of rule C, keyed on the specs rather than the task: the
# diff that implements a task's *last* unimplemented spec writes the
# task's completed date, or no merge can ever move the task to done. The
# task file itself may be untouched by the diff, so this pass starts
# from the changed specs.
checked_tasks=""
for f in $CHANGED; do
  case "$f" in work/specs/spec-*.md) ;; *) continue ;; esac
  [ -f "$f" ] || continue
  [ "$(fm_now "$f" status)" = "implemented" ] || continue
  [ "$(fm_base "$f" status)" != "implemented" ] || continue
  tref=$(fm_now "$f" task_ref)
  [ -n "$tref" ] || continue
  case " $checked_tasks " in *" $tref "*) continue ;; esac
  checked_tasks="$checked_tasks $tref"
  tf=$(find work/tasks \
    \( -iname "${tref}.md" -o -iname "${tref}-*.md" \) 2>/dev/null | head -n1)
  [ -n "$tf" ] || continue
  cd_now=$(fm_now "$tf" completed)
  [ "$cd_now" = "null" ] || [ -z "$cd_now" ] || continue
  all_done=true
  for ref in $(fm_now "$tf" spec_ref | tr -d '[]' | tr ',' ' '); do
    [ -n "$ref" ] || continue
    sp=$(find work/specs \
      \( -iname "${ref}.md" -o -iname "${ref}-*.md" \) 2>/dev/null | head -n1)
    [ -n "$sp" ] || { all_done=false; break; }
    [ "$(fm_now "$sp" status)" = "implemented" ] || { all_done=false; break; }
  done
  if [ "$all_done" = "true" ]; then
    echo "INCONSISTENT: this diff implements ${tf}'s last spec but leaves its completed date null." >&2
    echo "  The date is the declaration the merge turns into done — write it." >&2
    status=1
  fi
done

if [ "$status" -eq 0 ]; then
  if [ "$RANGE_COMMITS" = "0" ]; then
    echo "The range ${DIFF_RANGE} selects no commits — nothing was checked." >&2
    echo "That is not a pass. A check that read nothing has vouched for" >&2
    echo "nothing, and reporting it as clean is the failure this refusal" >&2
    echo "exists to prevent. Name the range the change actually spans." >&2
    exit 3
  fi
  echo "OK — no forbidden lifecycle transition in ${DIFF_RANGE}"
fi

exit "$status"
