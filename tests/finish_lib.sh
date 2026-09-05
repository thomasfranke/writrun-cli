#!/usr/bin/env bash
# finish_lib.sh — the fixture behind the finish integration cases
# (tests/integration/finish/): an adopted repository carrying this
# repository's own kit, a task branch with one commit on it, a bare
# origin standing in for the forge, and a stubbed `gh`. The suite never
# reaches a real forge: every `gh` invocation is the stub's, and
# GH_LOG records them so a case can assert what did and did not reach it
# (spec-0010, tests required).

. "$(dirname "${BASH_SOURCE[0]}")/cli_lib.sh"

git_q() { git -c user.name=suite -c user.email=suite@test -c commit.gpgsign=false "$@"; }

ORIGIN="$WORK/origin.git"
TARGET="$WORK/target"

# The forge, stubbed. GH_PR_VIEW is the pull request `gh pr view`
# answers with; GH_PR_VIEW_FAILS makes the read fail the way a branch
# with no pull request does; GH_PR_READY_EXIT fails the act itself.
GH_LOG="$WORK/gh.log"
export GH_LOG
export GH_PR_VIEW='{"number":7,"title":"[TASK-0001] A thing to do","state":"OPEN","isDraft":true}'

make_gh() {
  mkdir -p "$WORK/bin"
  : > "$GH_LOG"
  cat > "$WORK/bin/gh" <<'STUB'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "$GH_LOG"
case "$1 $2" in
  "auth status") exit 0 ;;
  "pr view")
    if [ -n "${GH_PR_VIEW_FAILS:-}" ]; then
      echo "no pull requests found for branch" >&2
      exit 1
    fi
    printf '%s\n' "$GH_PR_VIEW"
    exit 0 ;;
  "pr ready") exit "${GH_PR_READY_EXIT:-0}" ;;
esac
exit 0
STUB
  chmod +x "$WORK/bin/gh"
  export PATH="$WORK/bin:$PATH"
}

# task_file <number> <slug> <status> <spec-ref> <title>
task_file() {
  cat > "$TARGET/work/tasks/task-$1-$2.md" <<EOF
---
id: task-$1
status: $3
blocked_reason: null
taken_by: null
spec_ref: [$4]
doc_ref: null
origin: rule
priority: medium
depends_on: []
milestone: null
created: 2026-01-0${1#000}T00:00:00Z
queued: 2026-01-0${1#000}T00:00:00Z
completed: null
merged: null
provenance: []
---

# $5

One paragraph of brief.
EOF
}

# spec_file <number> <slug> <task-ref> <status> <promised-doc> <outcome>
# A promised doc of `none` promises nothing; an outcome of `none` leaves
# the generator's placeholder in place.
spec_file() {
  local promised="- none — no behaviour change"
  local outcome="_(fill after execution)_"
  [ "$5" = none ] || promised="- \`$5\` — the change"
  [ "$6" = none ] || outcome="$6"
  cat > "$TARGET/work/specs/spec-$1-$2.md" <<EOF
---
id: spec-$1
task_ref: $3
status: $4
created: 2026-01-01T00:00:00Z
---

# spec-$1 — the contract

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

$promised

## Proposed technical changes

- none — no machinery change

## Outcome

$outcome
EOF
}

# make_repo — an adopted repository on `main`, its kit copied from this
# repository, pushed to a bare origin. task-0001 is the one every case
# finishes; task-0002 promises a doc it never touched, and task-0003
# carries no spec at all.
make_repo() {
  mkdir -p "$TARGET/.writrun" "$TARGET/work/tasks" "$TARGET/work/specs" "$TARGET/work/reports"
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

  task_file 0001 a-thing in-progress spec-0001 "A thing to do"
  spec_file 0001 a-thing task-0001 approved none "What was built, and nothing diverged."
  task_file 0002 another-thing in-progress spec-0002 "Another thing to do"
  spec_file 0002 another-thing task-0002 approved product/promised.md "What was built."
  task_file 0003 a-third-thing in-progress "" "A third thing to do"
  printf '# A project\n' > "$TARGET/README.md"
  mkdir -p "$TARGET/docs/product"
  printf '# Product\n' > "$TARGET/docs/product/README.md"

  (
    cd "$TARGET" || exit 1
    git_q init -q
    git_q symbolic-ref HEAD refs/heads/main
    git_q add -A
    git_q commit -q -m "the kit and the queue"
  )
  git_q init --bare -q "$ORIGIN"
  git_q -C "$TARGET" remote add origin "$ORIGIN"
  git_q -C "$TARGET" push -q -u origin main
}

# on_branch <branch> — the task branch the work happened on: one commit
# that touches no permanent doc, so the delta contract of a spec
# promising nothing is honoured.
on_branch() {
  git_q -C "$TARGET" checkout -q -b "$1"
  printf 'the work\n' >> "$TARGET/README.md"
  git_q -C "$TARGET" add -A
  git_q -C "$TARGET" commit -q -m "the work"
}

# finish_cmd — one `writrun finish`, its whole reporting kept in
# FINISH_OUT so a case can assert on more than one line of it.
FINISH_OUT="$WORK/finish.out"
finish_cmd() {
  "$WRITRUN" finish "$@" > "$FINISH_OUT" 2>&1
  local code=$?
  cat "$FINISH_OUT"
  return $code
}

# field <field> <file> — one front-matter value, the front matter alone.
field() {
  awk -v f="$1" '
    NR == 1 { if ($0 != "---") exit; next }
    /^---$/ { exit }
    sub("^" f ": *", "") { sub(/[[:space:]]*$/, ""); print; exit }
  ' "$2"
}

make_gh
