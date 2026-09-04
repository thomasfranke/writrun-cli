#!/usr/bin/env bash
# provenance_rollup.sh — sums the provenance ledgers into the two answers
# the field exists for: what the work cost, and what share of it an agent
# did.
#
# Usage:
#   provenance_rollup.sh [--milestone M] [--tasks DIR]
#
#   provenance_rollup.sh                       the whole queue
#   provenance_rollup.sh --milestone v0.1-core one milestone
#
# **Counts, never money.** The entries hold the platform's own numbers and
# so does this: a conversion belongs at report time, against the rate card
# published on the day the question is asked, because a stored currency
# figure becomes a lie about the past the next time a price changes
# (docs/product/concepts/provenance.md#what-an-entry-holds). The four
# columns stay separate for the same reason a ledger keeps them: cache
# reads outweigh the rest by around two orders of magnitude, and a total
# that folded them together would misreport spend rather than round it.
#
# **A task without an agent entry is not a gap.** It is a task worked by
# hand, or a task nobody has worked yet, and the summary says which by
# counting both sides — a share done by agents means something only
# because the share done by people is written down beside it.
#
# Reads only. Nothing here is wired into a check, and nothing here decides
# anything (product/concepts/provenance.md).
#
# Exit codes: 0 always, except 3 for a usage error.
#
# Portable bash 3.2, POSIX awk/sed — no gawk extensions.

set -euo pipefail

MILESTONE=""
TASK_DIR="work/tasks"

usage() {
  echo "usage: provenance_rollup.sh [--milestone M] [--tasks DIR]" >&2
  exit 3
}

# `need_value` rather than `${2:-}`: an option written last has no $2 to
# default, and `shift 2` on a one-element list is a shell error that
# `set -e` turns into a silent exit 1 — a usage error reported as a
# crash, and not as the exit 3 the header promises.
need_value() {   # need_value <flag> <argc>
  [ "$2" -ge 2 ] || { echo "provenance_rollup.sh: $1 takes a value" >&2; usage; }
}

while [ $# -gt 0 ]; do
  case "$1" in
    --milestone) need_value "$1" $#; MILESTONE="$2"; shift 2 ;;
    --tasks)     need_value "$1" $#; TASK_DIR="$2"; shift 2 ;;
    -h|--help)   sed -n '2,28p' "$0"; exit 0 ;;
    *)           usage ;;
  esac
done

[ -d "$TASK_DIR" ] || { echo "provenance_rollup.sh: no such directory: ${TASK_DIR}" >&2; exit 3; }

# An unmatched glob arrives at awk as the literal pattern, and awk dies
# on a file by that name — so a project before its first task, which is
# a queue with nothing in it and not an error, would get exit 2 and
# `awk: can't open file` where the header promises 0. /dev/null reads as
# no records, and the END block prints the empty report it should.
set -- "$TASK_DIR"/task-*.md
[ -e "$1" ] || set -- /dev/null

# One awk over every task file: the front matter alone, the ledger read
# entry by entry. A body line spelling `provenance:` at column 0 is prose.
awk -v want="$MILESTONE" '
  function val(line, key,   m) {
    if (match(line, "[{,] *" key ": *[^,}]+") == 0) return ""
    m = substr(line, RSTART, RLENGTH)
    sub("^[{,] *" key ": *", "", m)
    return m
  }
  # Everything a file contributes is held aside until its front matter
  # closes, because the milestone the row belongs to is a field like any
  # other and may be read after the ledger. Committing as we go would
  # make --milestone depend on the order of the schema.
  FNR == 1 {
    infm = ($0 == "---"); inl = 0
    id = ""; ms = ""; f_in = 0; f_out = 0; f_cr = 0; f_cw = 0
    f_models = ""; f_entries = 0; f_agent = 0
    next
  }
  infm && /^---$/ {
    infm = 0; inl = 0
    if (want != "" && ms != want) next
    tasks++
    if (f_agent) with++; else without++
    if (f_entries > 0) {
      if (!(id in entries)) order[++n] = id
      entries[id] += f_entries
      t_in[id] += f_in; t_out[id] += f_out; t_cr[id] += f_cr; t_cw[id] += f_cw
      models[id] = f_models
      s_in += f_in; s_out += f_out; s_cr += f_cr; s_cw += f_cw
      k = split(f_models, mm, " ")
      for (j = 1; j <= k; j++)
        if (index(" " all_models " ", " " mm[j] " ") == 0)
          all_models = all_models (all_models == "" ? "" : " ") mm[j]
    }
    next
  }
  infm {
    if ($0 ~ /^provenance:[[:space:]]*$/) { inl = 1; next }
    if (inl && $0 ~ /^  - /) {
      f_entries++
      if (val($0, "by") == "agent") {
        f_agent = 1
        model = val($0, "model")
        if (model != "" && index(" " f_models " ", " " model " ") == 0)
          f_models = f_models (f_models == "" ? "" : " ") model
      }
      f_in += val($0, "input") + 0;      f_out += val($0, "output") + 0
      f_cr += val($0, "cache_read") + 0; f_cw += val($0, "cache_write") + 0
      next
    }
    inl = 0
    if (sub(/^id: */, "")) { id = $0; next }
    if (sub(/^milestone: */, "")) { ms = $0; next }
    next
  }
  END {
    for (a = 1; a <= n; a++) for (b = a + 1; b <= n; b++)
      if (order[b] < order[a]) { tmp = order[a]; order[a] = order[b]; order[b] = tmp }

    printf "%-12s %8s %9s %13s %12s  %s\n", "TASK", "INPUT", "OUTPUT", "CACHE_READ", "CACHE_WRITE", "MODELS"
    for (a = 1; a <= n; a++) {
      k = order[a]
      printf "%-12s %8d %9d %13d %12d  %s\n", k, t_in[k], t_out[k], t_cr[k], t_cw[k], models[k]
    }
    printf "%-12s %8d %9d %13d %12d  %s\n", "TOTAL", s_in + 0, s_out + 0, s_cr + 0, s_cw + 0, all_models
    print ""
    printf "%d task(s) in scope%s: %d with agent participation, %d without.\n", \
      tasks + 0, (want == "" ? "" : " (milestone " want ")"), with + 0, without + 0
    print "Counts, never money — convert at report time against the published rate card."
  }
' "$@"
