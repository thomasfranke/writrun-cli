#!/usr/bin/env bash
# session_card.sh — the settings this session obeys, rendered.
#
#   bash .writrun/scripts/stage-1-tasks-and-specs/session_card.sh
#
# Everything an agent obeys per session that is a *value* — the stage,
# the conduct flags, the title style, the vocabularies, the constants —
# on one card, so the conventions files are opened for their reasoning
# and not re-read every session for data `settings.json` states in a few
# hundred bytes.
#
# **It computes nothing and decides nothing.** Every line is read from
# `settings.json` (through read_setting.sh, defaults included — its
# --origin flag is what lets a default be marked as one), from
# check_observance.sh's TYPES=/SCOPES= lines (the machine half of the
# vocabulary, and its single source), or is a methodology constant the
# contract already fixes. A second parser of either file would be a
# second answer.
#
# It replaces reading, so growing is regressing: the card is ~30 lines.
#
# Exit codes: 0 always — a project with no settings file is pre-adoption,
# which is a state and not an error — except 3 when the vocabularies
# cannot be read, because a card missing them must not look complete.
#
# Portable bash 3.2, POSIX awk/sed. See the standing rule in
# docs/technical/decisions/.

set -uo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
READ_SETTING="$HERE/../stage-2-pull-requests/read_setting.sh"
OBSERVANCE="$HERE/../stage-2-pull-requests/check_observance.sh"

TOP=$(git rev-parse --show-toplevel 2>/dev/null) && cd "$TOP"

# value/origin, from the one reader. A key nobody declared prints its
# documented default, marked as one.
VAL=""; ORIGIN=""
read_key() {
  local out
  out=$(bash "$READ_SETTING" "$1" --origin 2>/dev/null)
  VAL=$(printf '%s' "$out" | cut -f1)
  ORIGIN=$(printf '%s' "$out" | cut -f2)
  [ -n "$ORIGIN" ] || ORIGIN=default
}

show() {   # show <label> <address> <meaning>
  read_key "$2"
  printf '  %-18s %s (%s)%s\n' "${1}:" "$VAL" "$ORIGIN" "${3:+  — $3}"
}

TYPES=$(sed -n 's/^TYPES="\(.*\)"$/\1/p' "$OBSERVANCE" | head -n1)
SCOPES=$(sed -n 's/^SCOPES="\(.*\)"$/\1/p' "$OBSERVANCE" | head -n1)
if [ -z "$TYPES" ] || [ -z "$SCOPES" ]; then
  echo "Could not read the TYPES/SCOPES lines from ${OBSERVANCE}." >&2
  echo "They are the machine half of the vocabulary, and a card without them" >&2
  echo "would look complete while stating nothing about what a title may say." >&2
  exit 3
fi

read_key stage
STAGE="$VAL"; STAGE_ORIGIN="$ORIGIN"
case "$STAGE" in
  1) STAGE_MEANING="tasks and specs — files only, no forge" ;;
  2) STAGE_MEANING="pull requests — the forge writes the queue's statuses" ;;
  3) STAGE_MEANING="the GitHub Issues mirror, on top of Stage 2" ;;
  *) STAGE_MEANING="a stage this card does not know; check_settings.sh names it" ;;
esac

read_key stage_2.pr_title_style
STYLE="$VAL"; STYLE_ORIGIN="$ORIGIN"

echo "WritRun — the settings this session obeys (.writrun/settings.json)"
echo
printf 'stage: %s (%s) — %s\n' "$STAGE" "$STAGE_ORIGIN" "$STAGE_MEANING"
echo
if [ "$STAGE" = 1 ]; then
  echo "conduct — these bind from Stage 2 up, and this project is at Stage 1:"
else
  echo "conduct:"
fi
show auto_commit    stage_2.auto_commit    "commit without asking"
show auto_push      stage_2.auto_push      "push without asking; the act that makes work public"
show auto_pr        stage_2.auto_pr        "open and update pull requests without asking"
show agent_coauthor stage_2.agent_coauthor "every commit an agent writes carries Co-Authored-By naming the model"
echo
printf 'pr_title_style: %s (%s) — one example per kind:\n' "$STYLE" "$STYLE_ORIGIN"
case "$STYLE" in
  bracketed)
    echo '  implementing   [TASK-0012][Fix][Ci] Debounce the mirror updates'
    echo '  authoring      [Docs][Product] The merge is the assenting act'
    echo '  reporting      [Chore][Queue] Record what the session observed' ;;
  conventional)
    echo '  implementing   [TASK-0012] fix(ci): debounce the mirror updates'
    echo '  authoring      docs(product): the merge is the assenting act'
    echo '  reporting      chore(queue): record what the session observed' ;;
  *)
    echo "  (a style this card does not know — check_settings.sh names it)" ;;
esac
echo
echo 'commit subject — a constant, whatever the title style:'
echo '  type(scope): imperative summary'
printf '  types:   %s\n' "$TYPES"
printf '  scopes:  %s\n' "$SCOPES"
echo
echo 'branches and the tag — constants:'
echo '  docs/<short-name>        authoring       title carries no task tag'
echo '  report/<short-name>      reporting       title carries no task tag'
echo '  task/NNNN-<short-name>   implementing    [TASK-NNNN] leads the title, one bracket per task'
echo
echo 'declarations:'
show spec_required     stage_1.spec_required
show decisions_style   stage_1.decisions_style
show product_layout    stage_1.product_layout
show provenance_ledger stage_1.provenance_ledger
exit 0
