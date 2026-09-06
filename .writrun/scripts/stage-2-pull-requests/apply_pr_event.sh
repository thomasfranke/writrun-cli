#!/usr/bin/env bash
# apply_pr_event.sh — turns one pull-request forge event into the status
# writes it implies, via flip_task_status.sh. The workflow wires events to
# this; the edge table lives in the flip script and in
# product/stage-2-pull-requests/statuses.md.
#
# Usage: apply_pr_event.sh <event>
#   <event>: opened | reopened | ready_for_review | converted_to_draft |
#            review_requested | changes_requested | closed
#
# Context arrives in env, never argv interpolation — a head branch name,
# a title and a login are a fork's to write, and all three are data here:
#   PR_HEAD_REF   the head branch (task/NNNN-* names one task)
#   PR_TITLE      the title, whose every [TASK-NNNN] tag names another
#   PR_AUTHOR     the pull request author's login (forge-authenticated)
#   PR_DRAFT      true|false
#   PR_MERGED     true|false (closed only)
#   PR_STATE      open|closed — the pull request's own state, which an
#                 `edited` event carries for a closed pull request too
#   PR_NUMBER     this pull request's own number, so the survivor query
#                 can drop its own row from a listing that lags
#   GH_REPO       owner/repo, for the survivor query
#   GH_TOKEN      lets `gh` ask the forge about surviving pull requests
#
# **The event reaches every task the pull request carries.** A branch
# name holds one id and a title holds as many as the work carries, so
# the carried set comes from ql_carried_of — the same helper the merge
# half and the mirror projection already ask, never a second parser
# beside them (technical/settings/titles.md#pr_title_style). A pull
# request carrying none by either route exits without writing.
#
# **The ceiling bounds claiming, not releasing.** Above QL_CARRIED_MAX
# distinct tasks the helper answers with its over-ceiling sentinel, and
# every event that expands flight writes nothing and exits non-zero. The
# close is exempt: it hands work back, and a claim edited over the
# ceiling after the recording landed would otherwise strand every task
# that recording moved.
#
# **What that widens, said plainly.** This runs on pull_request_target,
# so a fork's pull request reaches it, and the title is the fork's to
# write. Before the carried set, such a pull request could claim the one
# task its head branch spelled; it can now claim every task its title
# lists, and each claim is a status write pushed to the default branch.
# The kind of exposure is unchanged — both routes were always the fork's
# to choose — and the amount is bounded: above QL_CARRIED_MAX distinct
# tasks the helper answers with its over-ceiling sentinel, and this
# script writes nothing and exits non-zero, naming the count, the
# ceiling, and the heal — editing the title back under the ceiling
# re-records what the work carries, and closing and reopening the pull
# request re-fires the rest of the event's edges
# (report-0028; spec-0069; spec-0077).
#
# On close-without-merge, the forge is asked whether another open pull
# request still works the task — by every route the reader counts: each
# open pull request's head branch and leading title tags pass through
# ql_carried_of, the same helper that derived this event's own carried
# set, so the question reaches exactly as far as the reader does. With a
# survivor, the newest one's author and draftness are re-recorded
# instead of landing the task — never a silent skip, or taken_by strands
# on the closed PR's author. The question is per task, because a
# survivor for one carried task says nothing about another; the closing
# pull request's own row answers for none of them, wherever the
# listing's lag still shows it — the event in hand is better evidence
# than the cache. When the forge cannot answer, the task lands: a queue
# that briefly forgets a survivor heals at that survivor's next event,
# while a task stranded in-flight with no PR heals never.
#
# Exits 0 in every no-op case (nothing carried, no legal edge, merged
# close); 1 when a carried task's write failed, after the rest have been
# attempted, and when a claim over the ceiling reaches an event that
# would expand flight — refused whole, nothing written; 3 only for usage
# errors. Mutates the working tree;
# the caller commits, so one event's writes land as one commit — and a
# non-zero exit is what stops a half-applied one from being pushed under
# a green run.
#
# Portable bash 3.2, POSIX awk/sed. See the standing rule in
# docs/technical/decisions/.

set -euo pipefail

EVENT="${1:?usage: apply_pr_event.sh <event>}"

# **The event name is checked before anything the pull request carries.**
# A name this script does not know is a usage error whatever the title
# says, and the caller reads the exit code to decide whether to retry: 3
# says "you called me wrong", 1 says "the claim was too big". With the
# ceiling refusal standing first, an unknown event riding an over-ceiling
# title exited 1 and told the caller to retitle a pull request whose
# title was never the fault. One list, named once, so the dispatch below
# and this guard cannot drift apart.
EVENTS='opened reopened ready_for_review converted_to_draft review_requested changes_requested edited closed'
case " $EVENTS " in
  *" $EVENT "*) ;;
  *) echo "usage: apply_pr_event.sh <${EVENTS// /|}>" >&2; exit 3 ;;
esac

HERE="$(dirname "$0")"
FLIP="$HERE/flip_task_status.sh"
. "$HERE/queue_lib.sh"

# The carried set — validated as data by the helper, which keeps only
# digits from either source.
CARRIED=$(ql_carried_from_env)
if [ -z "$CARRIED" ]; then
  echo "head '${PR_HEAD_REF:-}' and title '${PR_TITLE:-}' carry no task — nothing to record"
  exit 0
fi
OVER=""
case "$CARRIED" in
  over-ceiling:*)
    OVER="${CARRIED#over-ceiling:}"
    # The ids are still read, because one arm below needs them and does
    # not claim on their authority. Re-read through the same helper with
    # the ceiling lifted to the count it just reported — never a second
    # parser, and never a leak: the assignment lives in the substitution's
    # own subshell.
    CARRIED=$(QL_CARRIED_MAX="$OVER" ql_carried_from_env)
    ;;
esac

# **The refusal bounds claiming, not releasing.** Every event but the
# close expands what the queue says is in flight, and above the ceiling
# the whole set is refused — never the first eight of it, since a partial
# write riding a green run is the failure this script's exit contract
# exists to prevent. The run goes red on the author's own pull request,
# and the heal is theirs.
#
# The close is the one arm that hands work back. It lands a task in
# flight, or re-records it from a surviving pull request the forge
# confirms — both are the queue converging on what the forge already
# says, and `land` against a resting task is an echo. Refusing here too
# would strand every task the pull request had already recorded under a
# shorter title: the close is the last event that can release them, and a
# task stranded in-flight with no pull request heals never (spec-0069's
# Outcome). The `edited` arm below now re-records a retitle, which is the
# cause this exemption was standing in for — but it only adds, so the
# close is still the only writer that hands work back.
if [ -n "$OVER" ] && [ "$EVENT" != "closed" ]; then
  echo "the head branch and title claim ${OVER} distinct tasks — the ceiling is ${QL_CARRIED_MAX}." >&2
  echo "Nothing was recorded. Retitle the pull request to what the work carries:" >&2
  echo "the edit re-fires the recording, and a refused title claimed nothing to undo." >&2
  exit 1
fi

draftness() {   # $1: true|false -> draft|ready
  [ "$1" = "true" ] && printf 'draft' || printf 'ready'
}

# flip_one <task> <mode> [args...] — one task's write, and one task's
# answer never abandons the tasks after it. The flip already exits 0 for
# an id resolving to no file and for an event no edge applies to; a
# louder exit is still one task's, so it is reported and the loop goes on.
#
# **Going on is not the same as passing.** The exit is remembered and
# this script ends on it: the caller commits whatever the tree holds, so
# a swallowed failure would push a half-applied event under a green run,
# and the task that did not move stays `ready` with its work in flight —
# a state no later event of this pull request heals, because `ready` has
# no edge to `in-review`. Loud, then, exactly as the run that cannot push
# is loud rather than shrugged at.
FAILED=0
flip_one() {
  local task="$1" mode="$2" code=0
  shift 2
  bash "$FLIP" "$mode" "$task" "$@" || code=$?
  if [ "$code" -ne 0 ]; then
    echo "flip ${mode} ${task} exited ${code} — the tasks after it still move" >&2
    FAILED=1
  fi
}

# flip_all <mode> [args...] — the same write, once per carried task.
flip_all() {
  local mode="$1" task
  shift
  for task in $CARRIED; do
    flip_one "$task" "$mode" "$@"
  done
}

case "$EVENT" in
  opened|reopened)
    flip_all take "${PR_AUTHOR:?}" "$(draftness "${PR_DRAFT:-true}")"
    ;;
  ready_for_review)
    flip_all review
    ;;
  converted_to_draft|changes_requested)
    flip_all rework
    ;;
  review_requested)
    # GitHub fires this on drafts too (CODEOWNERS auto-requests); a
    # draft's task is being worked, not reviewed.
    if [ "${PR_DRAFT:-true}" = "true" ]; then
      echo "review requested on a draft — not an in-review signal"
      exit 0
    fi
    flip_all review
    ;;
  edited)
    # **The title is one of the two routes into the carried set, and it
    # is writable after the recording.** Without this arm a pull request
    # recorded under one tag and retitled to nine was stranded: no event
    # re-read the title, so the close was the first writer to see it —
    # and above the ceiling the close was refused too
    # (docs/technical/decisions/pull-requests/0069-*).
    #
    # `edited` also fires on body and base changes, which say nothing
    # about what is carried. `changes.title.from` is set by the forge
    # only when the title moved, so an empty one is the whole test, and
    # the cheap path costs no forge call and no file read. A pull
    # request's title cannot be empty, so empty never means "was empty".
    if [ -z "${PR_TITLE_FROM:-}" ]; then
      echo "the title did not change — nothing to re-record"
      exit 0
    fi
    # **A closed pull request claims nothing.** The forge fires `edited`
    # on closed and merged pull requests as readily as on open ones, and
    # nothing above this line reads the state: without the guard, adding
    # a tag to the title of a pull request that closed last week took the
    # task it named — `in-progress` with `taken_by`, on work no pull
    # request is doing. That is the stranding this arm exists to end,
    # written by the arm itself.
    #
    # The close wins the ordered case here, which is what spec-0077's
    # edge case asks for. It cannot win a truly simultaneous one: an
    # `edited` payload minted while the pull request was still open says
    # `open` however late its run lands, and no state a script reads
    # afterwards is the event's own. That residue stays as the spec
    # records it, bounded by the close being the last event either way.
    if [ "${PR_MERGED:-false}" = "true" ] || [ "${PR_STATE:-open}" = "closed" ]; then
      echo "the pull request is closed — a retitle after the release claims nothing"
      exit 0
    fi
    # **Re-recording adds; it does not release.** Only the tasks the old
    # title did not claim are taken. A task already in flight keeps the
    # status its own events gave it — `take` against an in-review task
    # would knock it back to in-progress, and a title edit is not a
    # review event.
    #
    # A task the old title claimed and the new one does not is left in
    # flight, and no later event releases it: the close reads the title
    # as it then stands, so a dropped tag is invisible there too. That is
    # a stranding this arm does not answer — releasing it needs the close
    # arm's survivor query, since a second pull request may still carry
    # the task — and it is recorded as such in spec-0077's Outcome rather
    # than half-answered here.
    #
    # The old set is read through the same helper, never a second parser.
    # An over-ceiling old title is read as claiming *nothing*, and not
    # re-read with the ceiling lifted: the refusal is whole, so no event
    # under that title ever wrote a status, and its tags name tasks it
    # did not put in flight. Counting them as already claimed would make
    # the edit that brings a nine-tag title back under the ceiling record
    # nothing at all — the stranding this arm exists to end, surviving
    # its own fix.
    #
    # **What that leaves the title unable to answer, the queue answers.**
    # "No event under that title" is not "no event": a pull request
    # opened as one tag, moved to `in-progress` by a changes-requested
    # review, retitled to nine and then back to one reaches here with the
    # nine-tag title as `WAS` and its first task reading as newly added.
    # `take` would knock that task out of the state the review gave it —
    # the move the rule above forbids, reached by the corollary beside
    # it. So the cheap test's survivors are checked against the file, and
    # a task this author already has in flight is skipped. In flight
    # under *someone else* still goes through: that is two pull requests
    # on one task, and `take`'s newest-wins edge is the rule for it.
    WAS=$(PR_TITLE="$PR_TITLE_FROM" ql_carried_from_env)
    case "$WAS" in over-ceiling:*) WAS="" ;; esac
    ADDED=""
    HELD=""
    for task in $CARRIED; do
      case " $WAS " in
        *" $task "*) continue ;;
      esac
      t_file=$(ql_task_file "$task")
      if [ -n "$t_file" ] \
         && [ "$(ql_fm_field taken_by "$t_file")" = "${PR_AUTHOR:?}" ]; then
        case "$(ql_fm_field status "$t_file")" in
          in-progress|in-review)
            echo "already in flight under ${PR_AUTHOR}: ${t_file} — a retitle does not move it"
            HELD=1
            continue ;;
        esac
      fi
      ADDED="$ADDED $task"
    done
    if [ -z "$ADDED" ]; then
      # Two reasons reach here and the line names the one that applies:
      # a reader deciding whether a recording is missing is reading this.
      if [ -n "${HELD:-}" ]; then
        echo "the retitle claims no task not already in flight — nothing to record"
      else
        echo "the retitle claims no task the old title did not — nothing to record"
      fi
      exit 0
    fi
    for task in $ADDED; do
      flip_one "$task" take "${PR_AUTHOR:?}" "$(draftness "${PR_DRAFT:-true}")"
    done
    ;;
  closed)
    if [ "${PR_MERGED:-false}" = "true" ]; then
      echo "closed by merging — the merge recording owns this move"
      exit 0
    fi
    if [ -n "$OVER" ]; then
      # Said, not hidden — but on stdout and green: the release is
      # correct and the claim it rides was already met red at the event
      # that made it.
      echo "the head branch and title claim ${OVER} distinct tasks — the ceiling is ${QL_CARRIED_MAX}."
      echo "A close releases rather than claims, so the whole set is still let go."
    fi
    # **One listing answers every carried task.** Asking the question per
    # task must not make the call per task: a pull request carrying six
    # tags would fetch the same list six times, and no two of those
    # answers could differ. The filter moves to the reader instead.
    #
    # `--limit` is given because `gh`'s default is 30 and the filter is
    # client-side: a survivor sitting below that line comes back invisible,
    # and an invisible survivor lands a task whose work is still open —
    # the exact failure this query exists to prevent, produced by the
    # query itself.
    #
    # `@tsv`, because the title is a field now: it has spaces in it and
    # may carry a newline, and a space-joined row would hand title text
    # to whichever field reads next. @tsv escapes both, so one pull
    # request is one line and a tab is the only separator — a claim the
    # reader below has to keep, since `read` on its own does not.
    OPEN_PRS=""
    if [ -n "${GH_TOKEN:-}" ] && command -v gh >/dev/null 2>&1; then
      OPEN_PRS=$(gh pr list --repo "${GH_REPO:?}" --state open --limit 200 \
        --json number,headRefName,author,isDraft,title \
        --jq '.[] | [.number, .headRefName, .author.login, .isDraft, .title] | @tsv' \
        2>/dev/null || printf '')
    fi
    # **The survivor index, built once.** One line per open pull request
    # — number, login, draftness, and the carried set ql_carried_of
    # answers for its head branch and title — computed before the loop
    # over carried tasks, so a close carrying six tags still costs the
    # helper one pass. Two kinds of row never reach the helper. The
    # closing pull request's own, dropped by number: the event in hand
    # proves it closed, wherever the listing's lag still shows it, and a
    # closed pull request must answer for nothing. And any row that
    # cannot carry a task — a head branch not under task/ and a title
    # whose first character is neither `[` nor blank — dropped by one
    # cheap test first, because the helper forks subshells and a 200-row
    # listing would otherwise buy several hundred forks for rows that
    # answer nothing.
    #
    # **The row is split by the shared reader, because `read` cannot
    # split it.** A tab is an IFS *whitespace* character, so `IFS=$TAB
    # read` folds a run of tabs into one separator and every empty field
    # vanishes with it — and `author.login` is empty for a deleted
    # account, which `gh` emits verbatim. That one absence would shift
    # the title into the draftness field and leave the title empty, and
    # the row would then be dropped by the guard below while the pull
    # request it names is still working the task. `ql_row_fields` peels
    # one field at a time, which keeps an empty field empty, and it
    # returns 1 for a row that does not hold the five fields asked for —
    # a row this reader cannot answer is skipped, never read short.
    TAB=$(printf '\t')
    INDEX=""
    while IFS= read -r o_row; do
      ql_row_fields 5 "$o_row" || continue
      o_num="$QL_F1"; o_head="$QL_F2"; o_login="$QL_F3"
      o_draft="$QL_F4"; o_title="$QL_F5"
      [ -n "$o_num" ] || continue
      [ "$o_num" = "${PR_NUMBER:-}" ] && continue
      # The guard is never narrower than the helper it saves work for.
      # ql_carried_of strips leading whitespace before it looks for a
      # tag, so a title opening with whitespace goes through to it. A row
      # let through wrongly costs one fork; a row dropped wrongly lands a
      # task whose work is still open.
      case "$o_head" in
        task/*) ;;
        *) case "$o_title" in '['*|[[:space:]]*) ;; *) continue ;; esac ;;
      esac
      o_carried=$(ql_carried_of "$o_head" "$o_title")
      [ -n "$o_carried" ] || continue
      INDEX="${INDEX}${o_num}${TAB}${o_login}${TAB}${o_draft}${TAB}${o_carried}
"
    done <<EOT
$OPEN_PRS
EOT
    for TASK in $CARRIED; do
      # A surviving open PR on the same task supersedes the landing. The
      # match is membership in the row's carried set — field-wise, never
      # a substring of the line, so a login or a number that spells a
      # task id stays a login or a number. Both sides arrive normalized
      # through ql_task_num, which is why nothing is stripped here. The
      # newest wins, so the highest number is kept rather than the last
      # line read.
      survivor=$(printf '%s\n' "$INDEX" | awk -F'\t' -v t="$TASK" '
        {
          n = split($4, c, " ")
          for (i = 1; i <= n; i++)
            if (c[i] == t) {
              if ($1 + 0 > best) { best = $1 + 0; out = $2 " " $3 }
              break
            }
        }
        END { if (out != "") print out }')
      if [ -n "$survivor" ]; then
        s_login=${survivor%% *}
        s_draft=${survivor##* }
        echo "a surviving pull request still works ${TASK} — recording it instead"
        flip_one "$TASK" take "$s_login" "$(draftness "$s_draft")"
      else
        flip_one "$TASK" land
      fi
    done
    ;;
  *)
    # Unreachable: the guard at the top admits only the names above. Kept
    # so a name added to $EVENTS and forgotten here fails loudly instead
    # of falling through the case and exiting 0 over an unwritten queue.
    echo "apply_pr_event.sh: '${EVENT}' is in \$EVENTS but has no arm" >&2
    exit 3
    ;;
esac

exit "$FAILED"
