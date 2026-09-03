#!/usr/bin/env bash
# check_queue_impact.sh — names the queued tasks a doc change touches.
#
# Usage: check_queue_impact.sh <diff-range>
#
# Deterministic, not judgement: a non-completed task's doc_ref names the
# doc file its brief derives from. If this change edits that file, the
# person reviewing the rule change is the right one to re-check the brief —
# staleness is caught where it is born
# (docs/product/stage-1-tasks-and-specs/conflicts.md#when-the-doc-moves-ahead-of-the-queue).
#
# Always exits 0 — file-level overlap is a signal, never a failure; whether
# the brief survived the edit is the reviewer's call. The amend path is the
# special flow in the README. Warnings land on stdout (::warning for the
# forge's annotations) and in $GITHUB_STEP_SUMMARY when set.

set -euo pipefail
RANGE="${1:?usage: check_queue_impact.sh <diff-range>}"
: "${GITHUB_STEP_SUMMARY:=/dev/null}"

# docs/writrun-instructions.md is process metadata — nothing derives
# from it, so it can never put queued work at risk.
# git_read <label> <git-args...> — runs git and leaves its stdout in
# GIT_OUT. On failure it prints what git said and exits 3, because a
# check that could not read its input must never report the empty result
# as a clean one: `$(git … || true)` yields exactly the same empty string
# whether nothing matched or nothing ran, and two of these checks are
# gates (spec-0013).
#
# **Never call this inside a command substitution.** The `exit` would end
# only the subshell, and the caller would go on reading the empty value
# this exists to prevent — the very shape of the bug being removed.
GIT_OUT=""
git_read() {
  local label="$1" err
  shift
  err=$(mktemp "${TMPDIR:-/tmp}/writrun-git.XXXXXX")
  if ! GIT_OUT=$(git "$@" 2>"$err"); then
    echo "${label} failed:" >&2
    head -n 2 "$err" >&2
    rm -f "$err"
    exit 3
  fi
  rm -f "$err"
}

# This check is advisory by contract — it never fails a change — but an
# advisory that could not look must say it did not look. A failed read is
# the one case where it exits non-zero: reporting "no permanent doc
# changed" without having read the diff is the lie, not the exit code.
git_read "git diff --name-only ${RANGE} -- docs" \
  diff --name-only "$RANGE" -- docs
# The `|| true` that stays is grep's, not git's: no match is an answer.
docs_changed=$(printf '%s\n' "$GIT_OUT" \
  | grep -vxF 'docs/writrun-instructions.md' || true)
if [ -z "$docs_changed" ]; then
  echo "No permanent doc changed."
  exit 0
fi

hits=0
for t in work/tasks/*.md; do
  [ -f "$t" ] || continue
  case "$(basename "$t")" in README.md|readme.md) continue ;; esac
  st=$(sed -n 's/^status: *//p' "$t" | head -n1)
  case "$st" in done|dropped) continue ;; esac
  ref=$(sed -n 's/^doc_ref: *//p' "$t" | head -n1)
  [ -n "$ref" ] && [ "$ref" != "null" ] || continue
  file="docs/${ref%%#*}"
  if printf '%s\n' "$docs_changed" | grep -qxF "$file"; then
    id=$(sed -n 's/^id: *//p' "$t" | head -n1)
    echo "::warning file=${file}::${id} (${st}) derives its brief from this doc — re-check the task and its specs against the amended rule."
    echo "- **${id}** (\`${st}\`) references \`${ref}\`, which this change edits." >> "$GITHUB_STEP_SUMMARY"
    hits=1
  fi
done

if [ "$hits" = "1" ]; then
  echo "Queued work references docs this change edits — see the warnings."
else
  echo "No non-completed task references the docs this change edits."
fi
