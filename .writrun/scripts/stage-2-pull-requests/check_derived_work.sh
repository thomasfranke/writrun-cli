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
# Exit 3: the range could not be read, or a changed path arrived in a
#         shape this check cannot probe. Never reported as a clean run —
#         an input that could not be read is not an empty one.

set -euo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
. "$HERE/queue_lib.sh"
RANGE="${1:?usage: check_derived_work.sh <diff-range>}"

# Only an authoring change owes a declaration. Permanent is where the
# project has committed itself, and that is no longer structural: it is
# everything under docs/ less two exemptions, each a different thing.
# The queue lives in work/ and was never permanent.
# docs/writrun-instructions.md is process metadata, not project truth —
# no task derives from it and no declaration is owed for editing it. And
# a chapter that declares itself a draft is not a rule yet, so it is not
# permanent either; that one is the filter below, where its reasoning
# lives. An implementing change touches permanent docs too — as loop
# closure — and is identified the same way the deltas check identifies
# it: some spec reached `implemented`.

# The range's two ends, derived once — in queue_lib.sh, one copy across
# the stage-2 gates (spec-0086). The bare shape hands back an empty
# HEADREF: its diff compares the ref to the *working tree*, and a probe
# against HEAD: would answer a question that diff did not ask — a
# chapter existing only in the checkout would resolve at neither ref,
# leave `perm` in silence, and turn a refusal into a pass (spec-0084).
# The probes below honour the sentinel by reading the checkout.
ql_range_ends "$RANGE"
BASE="$QL_BASE"
HEADREF="$QL_HEADREF"

# Both flags guard the same hazard, and it is this check's worst one: a
# path the loop below cannot probe drops out of `perm` in silence, and a
# dropped path turns a refusal into a pass.
#
#   - `core.quotePath=false`, because git otherwise renders a chapter
#     whose name holds non-ASCII bytes as `"docs/caf\303\251.md"` —
#     quotes and escapes included — and neither `git cat-file` probe can
#     resolve that literal string, so the chapter is a rule at neither
#     end and leaves the set.
#   - `--no-renames`, because a detected rename is reported as its
#     destination alone. Renaming a rule chapter and adding the marker in
#     one change would then present a single path that is absent at the
#     base and a draft at the head — the silent withdrawal the marker is
#     explicitly not a way out of. Read as an addition and a deletion,
#     both ends of the rename are asked the question.
ql_git_read "git -c core.quotePath=false diff --name-only --no-renames ${RANGE} -- docs" \
  -c core.quotePath=false diff --name-only --no-renames "$RANGE" -- docs
# The `|| true` that stays is grep's, not git's: no match is an answer.
perm=$(printf '%s\n' "$QL_GIT_OUT" \
  | grep -vxF 'docs/writrun-instructions.md' || true)

# **A chapter that declares itself a draft is not a rule**, so a change
# that only touches drafts commits the project to nothing and has nothing
# to declare (authoring.md#a-chapter-that-is-not-a-rule-yet).
#
# The question asked per path is "was this ever a rule in this change",
# and it needs both ends. Reading only the version the change lands would
# let a rule be withdrawn silently — add the line, and the chapter stops
# being permanent with nothing shown to a reviewer, which would make
# demotion cheaper than deletion. Reading only the base would refuse the
# case the whole rule exists for: a chapter born a draft.
#
# So a path drops out only when it is a rule at neither end. A file
# absent at an end is a rule at that end in no sense, which is what makes
# a deleted draft free and a deleted rule still owed.
#
# The exclusion above stays its own list: `docs/writrun-instructions.md`
# is process metadata, a different exemption with a different reason, and
# one list meaning two things is the next reader's trap.
kept=""
while IFS= read -r f; do
  [ -n "$f" ] || continue

  # Quoting off, git still quotes a path holding a quote, a backslash or
  # a control character. Such a path cannot be probed, and a gate refuses
  # what it cannot read rather than dropping it — the sibling promise
  # check draws this line in the same place for the same reason.
  case "$f" in
    '"'*)
      echo "cannot read the changed path ${f} — refusing rather than skipping it" >&2
      exit 3
      ;;
  esac

  rule_at_base=false
  if git cat-file -e "${BASE}:$f" 2>/dev/null; then
    ql_doc_is_draft "$f" "$BASE" || rule_at_base=true
  fi
  # An empty HEADREF is the working tree, read from the checkout the way
  # ql_doc_is_draft's no-ref mode already reads it. The sentinel is
  # never interpolated into a ref:path — `git cat-file -e ":$f"` reads
  # the *index*, a third state this diff never compared.
  rule_at_head=false
  if [ -n "$HEADREF" ]; then
    if git cat-file -e "${HEADREF}:$f" 2>/dev/null; then
      ql_doc_is_draft "$f" "$HEADREF" || rule_at_head=true
    fi
  elif [ -f "$f" ]; then
    ql_doc_is_draft "$f" || rule_at_head=true
  fi
  if [ "$rule_at_base" = false ] && [ "$rule_at_head" = false ]; then
    continue
  fi
  kept="${kept}${f}"$'\n'
done <<EOF
$perm
EOF
perm=$(printf '%s' "$kept" | sed '/^$/d')

if [ -z "$perm" ]; then
  echo "No permanent doc changed — nothing to declare."
  exit 0
fi

# "Some spec reached implemented" is read from the front matter at the
# range's two ends, never grepped out of the diff text — a spec body
# quoting `status: implemented` at column 0 must not turn an authoring
# change into loop closure.

ql_git_read "git diff --name-only ${RANGE} -- 'work/specs/*.md'" \
  diff --name-only "$RANGE" -- 'work/specs/*.md'
for s in $QL_GIT_OUT; do
  [ -f "$s" ] || continue
  [ "$(ql_fm_field_in status < "$s")" = "implemented" ] || continue
  [ "$(git show "${BASE}:$s" 2>/dev/null | ql_fm_field_in status)" = "implemented" ] && continue
  echo "Implementing change — permanent doc edits are loop closure."
  exit 0
done

# Authoring. The diff is the authority on the first half of the rule; the
# PR body's Derived-work section carries the second.
ql_git_read "git diff --name-only --diff-filter=A ${RANGE} -- 'work/tasks/task-*.md'" \
  diff --name-only --diff-filter=A "$RANGE" -- 'work/tasks/task-*.md'
added="$QL_GIT_OUT"
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
