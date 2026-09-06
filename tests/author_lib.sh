#!/usr/bin/env bash
# author_lib.sh — the fixture behind the author integration cases
# (tests/integration/author/): an adopted repository carrying this
# repository's own kit, an authoring change already committed on a local
# branch, a bare origin standing in for the forge, and a stubbed `gh`.
# The suite never reaches a real forge: every `gh` invocation is the
# stub's, GH_LOG records them so a case can assert what did and did not
# reach it, and the pull request's body is kept whole in GH_BODY
# (spec-0009, tests required).

. "$(dirname "${BASH_SOURCE[0]}")/cli_lib.sh"

git_q() { git -c user.name=suite -c user.email=suite@test -c commit.gpgsign=false "$@"; }

ORIGIN="$WORK/origin.git"
TARGET="$WORK/target"

# The forge, stubbed. GH_AUTH_EXIT makes `gh auth status` fail, which is
# the path where nothing may be pushed; GH_PR_CREATE_EXIT fails the
# opening itself, which is the one state the act must not leave behind.
GH_LOG="$WORK/gh.log"
GH_BODY="$WORK/pr-body.md"
export GH_LOG
export GH_BODY_FILE="$GH_BODY"

make_gh() {
  mkdir -p "$WORK/bin"
  : > "$GH_LOG"
  : > "$GH_BODY"
  # The body is a whole document, so it is written to its own file and
  # squashed to one line in the log — a multi-line record would make
  # every `grep -q` in the cases read a different thing than it says.
  cat > "$WORK/bin/gh" <<'STUB'
#!/usr/bin/env bash
{ printf '%s' "$*" | tr '\n' ' '; printf '\n'; } >> "$GH_LOG"
case "$1 $2" in
  "auth status") exit "${GH_AUTH_EXIT:-0}" ;;
  "pr create")
    while [ "$#" -gt 0 ]; do
      if [ "$1" = "--body" ] && [ "$#" -ge 2 ]; then
        printf '%s\n' "$2" > "${GH_BODY_FILE:-/dev/null}"
        shift
      fi
      shift
    done
    echo "https://forge.test/pull/9"
    exit "${GH_PR_CREATE_EXIT:-0}" ;;
esac
exit 0
STUB
  chmod +x "$WORK/bin/gh"
  export PATH="$WORK/bin:$PATH"
}

# task_file <number> <slug> <status> <spec-ref> <title>
#
# The directory is made here, not once in make_repo: git tracks files
# and not folders, so checking a branch out from one that carried the
# queue takes the emptied `work/tasks/` with it.
task_file() {
  mkdir -p "$TARGET/work/tasks"
  cat > "$TARGET/work/tasks/task-$1-$2.md" <<EOF
---
id: task-$1
status: $3
blocked_reason: null
taken_by: null
spec_ref: [$4]
doc_ref: product/rules.md
origin: rule
priority: medium
depends_on: []
milestone: null
created: 2026-01-01T00:00:00Z
queued: null
completed: null
merged: null
provenance: []
---

# $5

One paragraph of brief, and no technical detail.
EOF
}

# spec_file <number> <slug> <task-ref> <status> <title>
spec_file() {
  mkdir -p "$TARGET/work/specs"
  cat > "$TARGET/work/specs/spec-$1-$2.md" <<EOF
---
id: spec-$1
task_ref: $3
status: $4
created: 2026-01-01T00:00:00Z
---

# spec-$1 — $5

**References:** [$3](../tasks/$3-$2.md)

- **Goal:** something the task implements.

## Scope

In: the thing. Out: everything else.

## Steps

1. Do the thing.

## Acceptance criteria (EARS)

- When the thing is done, the system shall say so.

## Edge cases

- none

## Tests required

Integration over the fixture.

## Definition of Done

- [ ] Suite green.

## Proposed product changes

- none — no behaviour change

## Proposed technical changes

- none — no machinery change

## Outcome

_(fill after execution)_
EOF
}

# make_repo — an adopted repository on `main`, its kit copied from this
# repository, its permanent docs already there, pushed to a bare origin.
# The queue starts empty: what an authoring change derives is what it
# adds.
make_repo() {
  mkdir -p "$TARGET/.writrun" "$TARGET/work/tasks" "$TARGET/work/specs" \
    "$TARGET/work/reports" "$TARGET/docs/product"
  cp -R "$REPO_ROOT/.writrun/scripts"     "$TARGET/.writrun/scripts"
  cp -R "$REPO_ROOT/.writrun/skills"      "$TARGET/.writrun/skills"
  cp -R "$REPO_ROOT/.writrun/templates"   "$TARGET/.writrun/templates"
  cp -R "$REPO_ROOT/.writrun/conventions" "$TARGET/.writrun/conventions"

  cat > "$TARGET/.writrun/settings.json" <<'EOF'
{
  "stage": 2,
  "stage_1": {
    "decisions_style": "chronological",
    "product_layout": "by-feature",
    "provenance_ledger": false,
    "spec_required": "when-warranted"
  },
  "stage_2": {
    "agent_coauthor": false,
    "auto_commit": false,
    "auto_pr": false,
    "auto_push": true,
    "pr_title_style": "bracketed"
  }
}
EOF

  printf '# A project\n' > "$TARGET/README.md"
  printf '# Product documentation\n\nWhat the project does.\n' > "$TARGET/docs/product/README.md"
  printf '# Rules\n\nThe rules every part of the project obeys.\n' > "$TARGET/docs/product/rules.md"

  (
    cd "$TARGET" || exit 1
    git_q init -q
    git_q symbolic-ref HEAD refs/heads/main
    git_q add -A
    git_q commit -q -m "the kit, the docs and the empty queue"
  )
  git_q init --bare -q "$ORIGIN"
  git_q -C "$TARGET" remote add origin "$ORIGIN"
  git_q -C "$TARGET" push -q -u origin main
}

# on_branch <name> — a local branch that never reached the forge, which
# is where authoring starts.
on_branch() { git_q -C "$TARGET" checkout -q -b "$1"; }

# rule_written — the permanent doc the rule was written into.
rule_written() {
  printf '\n## The declaration is the section\n\nA rule that derives work names it.\n' \
    >> "$TARGET/docs/product/rules.md"
}

# work_derived — the task and the draft spec the rule derives.
work_derived() {
  task_file 0001 derived-work backlog spec-0001 "Declare the derived work"
  spec_file 0001 derived-work task-0001 draft "The declaration is the section"
}

# committed <subject> — everything staged, in one commit.
committed() {
  git_q -C "$TARGET" add -A
  git_q -C "$TARGET" commit -q -m "$1"
}

# authoring_change — the whole starting state every green case shares:
# a rule written, work derived from it, committed on a local `docs/`
# branch that the forge has never seen.
authoring_change() {
  on_branch "${1:-docs/derived-work}"
  rule_written
  work_derived
  committed "docs(product): the declaration is the section"
}

# author — one `writrun author`, its whole reporting kept in AUTHOR_OUT
# so a case can assert on more than one line of it.
AUTHOR_OUT="$WORK/author.out"
author() {
  "$WRITRUN" author "$@" > "$AUTHOR_OUT" 2>&1
  local code=$?
  cat "$AUTHOR_OUT"
  return $code
}

# TITLE is one title the fixture's declared style accepts, carrying no
# task tag — the shape an authoring title has.
TITLE="[DOCS] The declaration is the section"

make_gh
