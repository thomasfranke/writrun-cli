#!/usr/bin/env bash
# record_provenance.sh — appends one entry to a task's provenance ledger.
#
# Usage:
#   record_provenance.sh <task-id> by=<agent|human> login=<login> \
#                        [model=<id>] [input=N] [output=N] \
#                        [cache_read=N] [cache_write=N]
#
#   record_provenance.sh task-0026 by=agent model=claude-opus-5 \
#     login=octocat input=562 output=175853 \
#     cache_read=37266324 cache_write=366590
#
# **This is the one machine field a branch writes**, and the shape of the
# permission is what keeps that from becoming "a branch may edit front
# matter": this script only ever *appends*. It never rewrites an entry it
# found, and `writrun-check-task-state` refuses a diff that does
# (docs/technical/README.md#task-schema).
#
# Run from the repository root. The entry is written in the schema's key
# order regardless of the order the arguments arrive in, because the
# canonical form is what the line-based readers were promised.
#
# **A project that declares no ledger is asked for nothing.** When
# `stage_1.provenance_ledger` is not `true` this writes nothing and says
# so — recording nothing is a complete state, not a failure
# (docs/product/concepts/provenance.md#the-adopter-decides-whether-to-keep-it).
#
# Exit codes: 0 appended, or nothing to append; 1 the task or the entry
# is unusable; 3 usage error.
#
# Portable bash 3.2, POSIX awk/sed — no gawk extensions.

set -euo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
. "$HERE/../stage-2-pull-requests/queue_lib.sh"
READ_SETTING="$HERE/../stage-2-pull-requests/read_setting.sh"

usage() {
  echo "usage: record_provenance.sh <task-id> by=<agent|human> login=<login> [model=<id>] [input=N] [output=N] [cache_read=N] [cache_write=N]" >&2
  exit 3
}

TASK_ID="${1:-}"
[ -n "$TASK_ID" ] || usage
shift

by=""; model=""; login=""; input=""; output=""; cache_read=""; cache_write=""
for arg in "$@"; do
  case "$arg" in
    by=*)          by="${arg#by=}" ;;
    model=*)       model="${arg#model=}" ;;
    login=*)       login="${arg#login=}" ;;
    input=*)       input="${arg#input=}" ;;
    output=*)      output="${arg#output=}" ;;
    cache_read=*)  cache_read="${arg#cache_read=}" ;;
    cache_write=*) cache_write="${arg#cache_write=}" ;;
    *) echo "record_provenance.sh: '$arg' is outside the ledger's vocabulary" >&2; usage ;;
  esac
done

# The declaration first: a project that keeps no ledger is not asked to
# keep one, and an agent step that runs unconditionally must be able to
# run here and come away clean.
if [ "$(bash "$READ_SETTING" stage_1.provenance_ledger)" != "true" ]; then
  echo "stage_1.provenance_ledger is not true — this project keeps no ledger; nothing written."
  exit 0
fi

TASK_FILE=$(ql_task_file "$TASK_ID")
if [ -z "$TASK_FILE" ]; then
  echo "record_provenance.sh: ${TASK_ID} resolves to no file under work/tasks/" >&2
  exit 1
fi

# What the schema holds an entry to, checked here as well as by
# check_front_matter.sh — not because one of the two is redundant, but
# because a writer that can emit a file its own checker rejects is a
# writer nobody can run unattended.
# **A category is not a model** — the same tripwire check_front_matter.sh
# carries, for the same reason: `model: ai` satisfies every shape check
# and answers nothing a quarter later. The list is written out here as
# well because the checker is a skill that runs standalone; the two are
# held together by this comment and by the test that walks both.
MODEL_CATEGORIES="ai llm agent model assistant bot claude gpt gemini llama opus sonnet haiku fable"

case "$by" in
  agent|human) ;;
  "") echo "record_provenance.sh: by= is required — an entry names a person or an agent" >&2; exit 1 ;;
  *)  echo "record_provenance.sh: by='${by}' — an entry names a person or an agent and nothing else" >&2; exit 1 ;;
esac
printf '%s' "$login" | grep -qE '^[A-Za-z0-9-]+(\[bot\])?$' \
  || { echo "record_provenance.sh: login='${login}' is not a bare forge login" >&2; exit 1; }
if [ "$by" = agent ]; then
  [ -n "$model" ] \
    || { echo "record_provenance.sh: an agent's entry names its model" >&2; exit 1; }
  case " $MODEL_CATEGORIES " in
    *" $(printf '%s' "$model" | tr '[:upper:]' '[:lower:]') "*)
      echo "record_provenance.sh: model='${model}' is a category, not a model id — a category answers nothing a quarter later" >&2
      exit 1 ;;
  esac
else
  [ -z "$model" ] \
    || { echo "record_provenance.sh: a human entry carries no model" >&2; exit 1; }
  [ -z "${input}${output}${cache_read}${cache_write}" ] \
    || { echo "record_provenance.sh: a human entry carries no counts" >&2; exit 1; }
fi
for c in "$input" "$output" "$cache_read" "$cache_write"; do
  if [ -n "$c" ]; then
    printf '%s' "$c" | grep -qE '^[0-9]+$' \
      || { echo "record_provenance.sh: count '${c}' is not a bare non-negative integer" >&2; exit 1; }
  fi
done

# The schema's key order, whatever order the arguments arrived in.
entry="by: ${by}"
if [ -n "$model" ];       then entry="${entry}, model: ${model}"; fi
entry="${entry}, login: ${login}"
if [ -n "$input" ];       then entry="${entry}, input: ${input}"; fi
if [ -n "$output" ];      then entry="${entry}, output: ${output}"; fi
if [ -n "$cache_read" ];  then entry="${entry}, cache_read: ${cache_read}"; fi
if [ -n "$cache_write" ]; then entry="${entry}, cache_write: ${cache_write}"; fi
LINE="  - {${entry}}"

if grep -qxF "$LINE" "$TASK_FILE"; then
  echo "unchanged: ${TASK_FILE} already carries this exact entry."
  exit 0
fi

# Two shapes to append into: the empty ledger written inline, which
# becomes a block list, and a block list, which gains a line after its
# last entry. Everything outside the front matter is untouched — a body
# that quotes `provenance:` at column 0 is prose, not the field.
awk -v line="$LINE" '
  NR == 1 && $0 == "---" { infm = 1; print; next }
  infm && /^---$/ {
    if (inl && !written) { print line; written = 1 }
    infm = 0; closed = 1; print; next
  }
  infm && /^provenance:[[:space:]]*\[\][[:space:]]*$/ {
    print "provenance:"; print line; written = 1; next
  }
  infm && /^provenance:[[:space:]]*$/ { print; inl = 1; next }
  infm && inl && /^  - / { print; next }
  infm && inl { print line; written = 1; inl = 0; print; next }
  { print }
  # **At the end, not at the closing fence.** A file whose first line is
  # not `---` reaches no fence rule at all — guarded there, awk copies
  # the file through unchanged, exits 0, and the caller is told the entry
  # was appended when nothing was. That is the silent wrong answer this
  # whole script is written against. `closed` is checked beside it
  # because front matter that never closes is not front matter: the
  # writer refuses the file rather than deciding for itself where it ends.
  END { if (!written || !closed) exit 9 }
' "$TASK_FILE" > "$TASK_FILE.tmp" || {
  rm -f "$TASK_FILE.tmp"
  echo "record_provenance.sh: ${TASK_FILE} has no provenance field in closed front matter — migrate it first" >&2
  exit 1
}
mv "$TASK_FILE.tmp" "$TASK_FILE"
echo "appended to ${TASK_FILE}: ${LINE#  - }"
