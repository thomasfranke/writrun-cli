#!/usr/bin/env bash
# read_usage.sh — proposes provenance entries from the agent platform's
# own usage data. It reads; it never writes and never stores.
#
# Usage:
#   read_usage.sh [--transcripts DIR] [--login LOGIN] [task-id]
#
#   read_usage.sh                       every task the transcripts name
#   read_usage.sh task-0026             one task
#   read_usage.sh --login octocat task-0026
#
# It prints one line per proposal, in exactly the argument form
# record_provenance.sh takes, so composing the two is a pipe and not a
# format:
#
#   read_usage.sh task-0026 \
#     | xargs -L1 bash .writrun/scripts/stage-1-tasks-and-specs/record_provenance.sh
#
# `xargs` rather than a `while read` loop, because the loop's `$rest` has
# to word-split to become arguments and zsh does not split unquoted
# parameters — the whole tail arrives as one argument, and the writer
# rejects it. Loudly, so nothing is lost; but a composition this file
# hands out has to run in the shell the reader is standing in.
#
# **The join already exists.** The platform stamps every message with the
# git branch it ran on, and this methodology's branch convention puts the
# task id in that branch name (`task/NNNN-…`), so no correlation has to be
# invented — it is read.
#
# **It is a proposal, never a record.** The numbers live in the task file
# once somebody writes them there; this directory is one vendor's, on one
# machine, absent from CI and from every other contributor, which is the
# whole reason the ledger exists in the repository at all. Nothing here is
# wired into any check, and nothing here is the only place a number lives.
#
# **Silence where there is nothing to read.** No transcript directory, no
# files, no branch that names a task: this prints nothing and exits 0. A
# helper that failed instead would make one vendor's directory a
# precondition for the methodology.
#
# The default transcript directory is Claude Code's, addressed the way it
# addresses itself — the working directory's path with every character
# that is not a letter or a digit turned into `-`. Not `/` alone: a
# repository checked out under a path holding a space, a `.` or a `_`
# folds those too, and a mapping that handled only the separator would
# resolve to a directory that does not exist — which the `-d` guard below
# reads as "nothing to propose" and reports as silence.
# $WRITRUN_TRANSCRIPTS overrides it, which is also how the tests reach a
# fixture.
#
# Exit codes: 0 always, except 3 for a usage error.
#
# Portable bash 3.2, POSIX awk/sed — no gawk extensions.

set -euo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
. "$HERE/../stage-2-pull-requests/queue_lib.sh"

DIR="${WRITRUN_TRANSCRIPTS:-}"
LOGIN=""
WANT=""

# `need_value` rather than `${2:-}`: an option written last has no $2 to
# default, and `shift 2` on a one-element list is a shell error that
# `set -e` turns into a silent exit 1 — a usage error reported as a
# crash, and not as the exit 3 the header promises.
need_value() {   # need_value <flag> <argc>
  [ "$2" -ge 2 ] || { echo "read_usage.sh: $1 takes a value" >&2; exit 3; }
}

while [ $# -gt 0 ]; do
  case "$1" in
    --transcripts) need_value "$1" $#; DIR="$2"; shift 2 ;;
    --login)       need_value "$1" $#; LOGIN="$2"; shift 2 ;;
    -h|--help)     sed -n '2,51p' "$0"; exit 0 ;;
    -*)            echo "read_usage.sh: unknown option '$1'" >&2; exit 3 ;;
    *)             WANT="$1"; shift ;;
  esac
done

if [ -z "$DIR" ]; then
  DIR="${HOME}/.claude/projects/$(pwd | sed 's|[^A-Za-z0-9]|-|g')"
fi

[ -d "$DIR" ] || exit 0
set -- "$DIR"/*.jsonl
[ -e "$1" ] || exit 0

# An empty WANT_NUM disables the filter below, so an argument that names
# no number must never reach it: `read_usage.sh task-abc` would otherwise
# propose an entry for *every* task in the queue, and the composition the
# header documents would write the typo onto all of them.
WANT_NUM=""
if [ -n "$WANT" ]; then
  WANT_NUM=$(ql_task_num "$WANT")
  [ -n "$WANT_NUM" ] \
    || { echo "read_usage.sh: '${WANT}' names no task number" >&2; exit 3; }
fi

# One pass over every transcript. A record counts when it carries a usage
# object and ran on a branch that names a task; the group is the session
# and the model, because "one entry per session that worked it, naming the
# specific model" is what the schema asks for — and a session that ran on
# two models is normal, not exceptional.
#
# The counts are read as the *first* occurrence of each key inside the
# usage object. That is not incidental: the object repeats every one of
# them inside its `iterations` array, and summing what a naive scan finds
# would multiply the true number by the iteration count.
awk '
  # The pattern arrives as a *string*, never as a /literal/: an awk regex
  # literal in an argument position evaluates to `$0 ~ /re/` — a 0 or a 1
  # — so a helper written the obvious way sums the number of matching
  # lines and calls it a token count.
  function first(hay, re,    m) {
    if (match(hay, re) == 0) return 0
    m = substr(hay, RSTART, RLENGTH)
    sub(/^[^0-9]*/, "", m)
    return m + 0
  }
  {
    if (index($0, "\"usage\":{") == 0) next
    if (match($0, /"gitBranch":"task\/[^"]*"/) == 0) next
    branch = substr($0, RSTART, RLENGTH)
    sub(/^"gitBranch":"/, "", branch); sub(/"$/, "", branch)
    num = branch
    sub(/^task\//, "", num); sub(/[^0-9].*$/, "", num); sub(/^0+/, "", num)
    if (num == "") next

    if (match($0, /"sessionId":"[^"]*"/) == 0) next
    session = substr($0, RSTART, RLENGTH)
    sub(/^"sessionId":"/, "", session); sub(/"$/, "", session)

    if (match($0, /"model":"[^"]*"/) == 0) next
    model = substr($0, RSTART, RLENGTH)
    sub(/^"model":"/, "", model); sub(/"$/, "", model)
    if (model == "" || model == "<synthetic>") next

    usage = substr($0, index($0, "\"usage\":{"))
    key = num SUBSEP session SUBSEP model
    if (!(key in seen)) { seen[key] = 1; order[++n] = key }
    in_t[key]  += first(usage, "[{,]\"input_tokens\":[0-9]+")
    out_t[key] += first(usage, "[{,]\"output_tokens\":[0-9]+")
    cr_t[key]  += first(usage, "\"cache_read_input_tokens\":[0-9]+")
    cw_t[key]  += first(usage, "\"cache_creation_input_tokens\":[0-9]+")
  }
  END {
    for (i = 1; i <= n; i++) {
      key = order[i]
      split(key, p, SUBSEP)
      printf "%s\t%s\t%s\t%d\t%d\t%d\t%d\n", p[1], p[2], p[3], in_t[key], out_t[key], cr_t[key], cw_t[key]
    }
  }
' "$@" | sort -t'	' -k1,1n -k2,2 -k3,3 | while IFS='	' read -r num session model input output cread cwrite; do
  if [ -n "$WANT_NUM" ] && [ "$num" != "$WANT_NUM" ]; then continue; fi

  file=$(ql_task_file "task-${num}")
  [ -n "$file" ] || continue
  id=$(ql_fm_field id "$file")

  who="$LOGIN"
  if [ -z "$who" ]; then
    who=$(ql_fm_field taken_by "$file")
    if [ "$who" = "null" ]; then who=""; fi
  fi
  if [ -z "$who" ]; then
    echo "read_usage.sh: ${id} has no taken_by and no --login — skipped (session ${session})" >&2
    continue
  fi

  printf '%s by=agent model=%s login=%s input=%s output=%s cache_read=%s cache_write=%s\n' \
    "$id" "$model" "$who" "$input" "$output" "$cread" "$cwrite"
done
