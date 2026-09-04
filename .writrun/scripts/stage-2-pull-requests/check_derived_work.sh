#!/usr/bin/env bash
# check_derived_work.sh — an authoring change names its derived work, or
# states that none was (docs/product/stage-1-tasks-and-specs/authoring.md#declaring-derived-work).
#
# Usage: check_derived_work.sh <diff-range>
#   The PR body arrives via $PR_BODY — through the environment, never
#   inline interpolation: it is attacker-controlled text on a fork PR.
#
# Exit 0: nothing owed, or the declaration is present.
# Exit 1: a permanent doc changed with neither derived tasks in the diff
#         nor "none" declared under "## Derived work" in the PR body.

set -euo pipefail
RANGE="${1:?usage: check_derived_work.sh <diff-range>}"

# Only an authoring change owes a declaration. Permanent is structural —
# everything under docs/; the queue lives in work/. An implementing change
# touches permanent docs too — as loop closure — and is identified the
# same way the deltas check identifies it: some spec reached `implemented`.
# docs/writrun-instructions.md is process metadata, not project truth —
# no task derives from it and no declaration is owed for editing it.
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

git_read "git diff --name-only ${RANGE} -- docs" \
  diff --name-only "$RANGE" -- docs
# The `|| true` that stays is grep's, not git's: no match is an answer.
perm=$(printf '%s\n' "$GIT_OUT" \
  | grep -vxF 'docs/writrun-instructions.md' || true)
if [ -z "$perm" ]; then
  echo "No permanent doc changed — nothing to declare."
  exit 0
fi

# "Some spec reached implemented" is read from the front matter at the
# range's two ends, never grepped out of the diff text — a spec body
# quoting `status: implemented` at column 0 must not turn an authoring
# change into loop closure.
case "$RANGE" in
  *...*)
    left="${RANGE%%...*}"
    right="${RANGE##*...}"
    # The same rule as the diff below: a merge-base that could not be
    # computed is not a base of "nothing", it is an unanswered question.
    if ! BASE=$(git merge-base "${left:-HEAD}" "${right:-HEAD}" 2>&1); then
      echo "git merge-base ${left:-HEAD} ${right:-HEAD} failed:" >&2
      printf '%s\n' "$BASE" | head -n 2 >&2
      exit 3
    fi
    ;;
  *..*) BASE="${RANGE%%..*}" ;;
  *)    BASE="$RANGE" ;;
esac
fm_field() {
  awk -v f="$1" '
    NR == 1 { if ($0 != "---") exit; next }
    /^---$/ { exit }
    sub("^" f ": *", "") { sub(/[[:space:]]*$/, ""); print; exit }
  '
}

git_read "git diff --name-only ${RANGE} -- work/specs" \
  diff --name-only "$RANGE" -- 'work/specs/*.md'
for s in $GIT_OUT; do
  [ -f "$s" ] || continue
  [ "$(fm_field status < "$s")" = "implemented" ] || continue
  [ "$(git show "${BASE}:$s" 2>/dev/null | fm_field status)" = "implemented" ] && continue
  echo "Implementing change — permanent doc edits are loop closure."
  exit 0
done

# Authoring. The diff is the authority on the first half of the rule; the
# PR body's Derived-work section carries the second.
git_read "git diff --name-only --diff-filter=A ${RANGE} -- work/tasks" \
  diff --name-only --diff-filter=A "$RANGE" -- 'work/tasks/task-*.md'
added="$GIT_OUT"
if [ -n "$added" ]; then
  echo "Derived work present:"
  echo "$added"
  exit 0
fi

section=$(printf '%s\n' "${PR_BODY:-}" \
  | awk '/^## Derived work/{f=1; next} /^## /{f=0} f')
if printf '%s\n' "$section" \
  | grep -qiE '(^|[^[:alnum:]])none([^[:alnum:]]|$)'; then
  echo "Derived work explicitly declared as none."
  exit 0
fi

echo "This change edits a permanent doc but neither adds a task" >&2
echo "nor declares 'none' under '## Derived work' in the PR body." >&2
echo "An empty declaration and a forgotten one look identical —" >&2
echo "see docs/product/stage-1-tasks-and-specs/authoring.md#declaring-derived-work." >&2
exit 1
