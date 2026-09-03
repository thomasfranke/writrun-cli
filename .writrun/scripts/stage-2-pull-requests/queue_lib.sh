#!/usr/bin/env bash
# queue_lib.sh — the helpers both halves of the transition machine share:
# front-matter reads and writes, the resting derivation, the task-file
# resolver, and the carried-ids parser. Sourced, never executed; the
# sourcing script owes `set -euo pipefail` itself.
#
# One copy on purpose: flip_task_status.sh and record_task_status.sh
# each carried private clones of these until a review caught the clones
# drifting — and caught the resolver pipeline dying under pipefail when
# the last find candidate failed the id filter.
#
# Portable bash 3.2, POSIX awk/sed — no gawk extensions. See the
# standing rule in docs/technical/decisions/.

# ql_fm_field <field> <file> — the field's value from the front-matter
# block alone; a body line spelling `status:` at column 0 never counts.
ql_fm_field() {
  awk -v f="$1" '
    NR == 1 { if ($0 != "---") exit; next }
    /^---$/ { exit }
    sub("^" f ": *", "") { sub(/[[:space:]]*$/, ""); print; exit }
  ' "$2"
}

# ql_set_field <file> <field> <value> — front matter only.
ql_set_field() {
  awk -v field="$2" -v value="$3" '
    NR == 1 && $0 == "---" { infm = 1; print; next }
    infm && /^---$/        { infm = 0; print; next }
    infm && index($0, field ":") == 1 { print field ": " value; next }
    { print }
  ' "$1" > "$1.tmp" && mv "$1.tmp" "$1"
}

# ql_task_num <anything> — the task number, zero-padding stripped;
# empty when the input names none.
#
# The zeros go in a step of their own, after the prefix rather than with
# it: `0034` is how every queue file and every [TASK-NNNN] tag spells the
# id, so it is what a person retypes — and stripping the padding only
# when a `task-` prefix carried it made that spelling resolve to nothing.
ql_task_num() {
  printf '%s' "$1" | sed -E 's/^task-//; s/^task\///; s/^0+//; s/[^0-9].*$//'
}

# ql_task_file <task-id-or-number> — the work/tasks file whose id is
# that number, whatever width it was written at; empty when none. The
# filter is an `if` on purpose: a trailing failed `[ … ] &&` would end
# the while loop non-zero and, under pipefail, kill the caller with no
# output at all.
ql_task_file() {
  local num
  num=$(ql_task_num "$1")
  [ -n "$num" ] || return 0
  find work/tasks \( -iname "task-*${num}.md" -o -iname "task-*${num}-*.md" \) 2>/dev/null \
    | while IFS= read -r c; do
        if [ "$(ql_task_num "$(basename "$c" .md)")" = "$num" ]; then
          printf '%s\n' "$c"
        fi
      done | head -n1
  return 0
}

# ql_spec_file <spec-id> — same resolution for a spec.
ql_spec_file() {
  find work/specs \( -iname "$1.md" -o -iname "$1-*.md" \) 2>/dev/null | head -n1
  return 0
}

# ql_resting <task-file> — where a task out of flight belongs: ready, or
# backlog if any spec it references is draft. An empty spec_ref is ready
# by construction — no approval event exists for it, and backlog must
# not be a trap.
ql_resting() {
  local refs ref spec st
  refs=$(ql_fm_field spec_ref "$1" | tr -d '[]' | tr ',' ' ')
  for ref in $refs; do
    [ -n "$ref" ] || continue
    spec=$(ql_spec_file "$ref")
    [ -n "$spec" ] || continue
    st=$(ql_fm_field status "$spec")
    if [ "$st" = "draft" ]; then printf 'backlog'; return 0; fi
  done
  printf 'ready'
}

# ql_carried_of <head-branch> <title> — the task ids whose work a pull
# request carries: the head branch's own (task/NNNN-*) plus every
# [TASK-NNNN] tag leading the title, deduplicated. Both arguments are a
# fork's to write, so only digits survive.
#
# Taking the pair as arguments is what lets a caller ask the question of
# *another* pull request — the amendment check has to, to name the one it
# suspends — while the env-reading form below stays the shape CI uses.
ql_carried_of() {
  local carried="" num rest tg
  case "${1:-}" in
    task/[0-9]*)
      num=$(ql_task_num "$1")
      [ -n "$num" ] && carried="task-$num"
      ;;
  esac
  rest="${2:-}"
  while :; do
    rest=$(printf '%s' "$rest" | sed 's/^[[:space:]]*//')
    tg=$(printf '%s' "$rest" | sed -n 's/^\[[Tt][Aa][Ss][Kk]-0*\([0-9][0-9]*\)\].*/\1/p')
    [ -n "$tg" ] || break
    case " $carried " in
      *" task-$tg "*) ;;
      *) carried="${carried:+$carried }task-$tg" ;;
    esac
    rest=$(printf '%s' "$rest" | sed 's/^\[[Tt][Aa][Ss][Kk]-[0-9][0-9]*\]//')
  done
  printf '%s' "$carried"
}

# ql_carried_from_env — the same question about the pull request CI is
# running on, read from env as data (PR_HEAD_REF, PR_TITLE).
ql_carried_from_env() {
  ql_carried_of "${PR_HEAD_REF:-}" "${PR_TITLE:-}"
}
