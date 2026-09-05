#!/usr/bin/env bash
# take_lib.sh — the fixture behind the take integration cases
# (tests/integration/take/): an adopted repository carrying this
# repository's own kit, a bare origin standing in for the forge, and a
# stubbed `gh`. The suite never reaches a real forge: the kit's
# WRITRUN_PR_LIST / WRITRUN_PR_FILES seams supply the pull requests, and
# every `gh` invocation is the stub's (spec-0008, tests required).

. "$(dirname "${BASH_SOURCE[0]}")/cli_lib.sh"

git_q() { git -c user.name=suite -c user.email=suite@test -c commit.gpgsign=false "$@"; }

ORIGIN="$WORK/origin.git"
TARGET="$WORK/target"

# The forge, stubbed. GH_PR_CREATE_EXIT makes `gh pr create` fail, which
# is the kit's exit-3 path; GH_LOG records every invocation so a case can
# assert what did and did not reach the forge.
GH_LOG="$WORK/gh.log"
export GH_LOG

make_gh() {
  mkdir -p "$WORK/bin"
  : > "$GH_LOG"
  cat > "$WORK/bin/gh" <<'STUB'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "$GH_LOG"
case "$1 $2" in
  "auth status") exit 0 ;;
  "pr create")   exit "${GH_PR_CREATE_EXIT:-0}" ;;
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

# spec_file <number> <slug> <task-ref> <status>
spec_file() {
  cat > "$TARGET/work/specs/spec-$1-$2.md" <<EOF
---
id: spec-$1
task_ref: $3
status: $4
created: 2026-01-01T00:00:00Z
---

# spec-$1 — the contract

- **Goal:** something the task implements.
EOF
}

# make_repo [auto_push] [auto_pr] — an adopted repository on `main`, its
# kit copied from this repository, pushed to a bare origin. The two
# conduct flags default to the composed-and-waiting shape the confirmed
# flow exercises.
make_repo() {
  local auto_push="${1:-true}" auto_pr="${2:-false}"
  mkdir -p "$TARGET/.writrun" "$TARGET/work/tasks" "$TARGET/work/specs" "$TARGET/work/reports"
  cp -R "$REPO_ROOT/.writrun/scripts"     "$TARGET/.writrun/scripts"
  cp -R "$REPO_ROOT/.writrun/skills"      "$TARGET/.writrun/skills"
  cp -R "$REPO_ROOT/.writrun/templates"   "$TARGET/.writrun/templates"
  cp -R "$REPO_ROOT/.writrun/conventions" "$TARGET/.writrun/conventions"

  cat > "$TARGET/.writrun/settings.json" <<EOF
{
  "stage": 2,
  "stage_2": {
    "agent_coauthor": false,
    "auto_commit": false,
    "auto_pr": ${auto_pr},
    "auto_push": ${auto_push},
    "pr_title_style": "bracketed"
  }
}
EOF

  task_file 0001 a-thing ready spec-0001 "A thing to do"
  spec_file 0001 a-thing task-0001 approved
  task_file 0002 another-thing ready spec-0002 "Another thing to do"
  spec_file 0002 another-thing task-0002 approved
  task_file 0003 a-third-thing backlog spec-0003 "A third thing to do"
  spec_file 0003 a-third-thing task-0003 draft
  printf '# A project\n' > "$TARGET/README.md"

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

# take — one `writrun take`, its whole reporting kept in TAKE_OUT so a
# case can assert on more than one line of it.
TAKE_OUT="$WORK/take.out"
take() {
  "$WRITRUN" take "$@" > "$TAKE_OUT" 2>&1
  local code=$?
  cat "$TAKE_OUT"
  return $code
}

# TITLE is one title the fixture's declared style accepts.
TITLE="[Feat][Ci] Debounce the mirror updates"

make_gh
