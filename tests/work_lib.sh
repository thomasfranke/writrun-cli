#!/usr/bin/env bash
# work_lib.sh — the fixture behind the work integration cases
# (tests/integration/work/): an adopted repository carrying this
# repository's own kit, a queue each case shapes, a stubbed `gh`, and a
# stub agent that records the invocation it was launched with.
#
# The agent is a stub because the whole point of the command is that it
# hands control to a program writrun knows nothing about: what the
# suite checks is the argument the launch carried — the brief the
# methodology's own script assembled — and that nothing else moved
# (spec-0007, tests required).

. "$(dirname "${BASH_SOURCE[0]}")/cli_lib.sh"

git_q() { git -c user.name=suite -c user.email=suite@test -c commit.gpgsign=false "$@"; }

TARGET="$WORK/target"

# The forge, stubbed: the lister asks it for the open pull requests, and
# a case asserts on this log that nothing was ever opened.
GH_LOG="$WORK/gh.log"
export GH_LOG

# What the launched agent was given: its working directory, then each
# argument whole.
AGENT_LOG="$WORK/agent.log"
export AGENT_LOG

make_stubs() {
  mkdir -p "$WORK/bin"
  : > "$GH_LOG"
  : > "$AGENT_LOG"

  cat > "$WORK/bin/gh" <<'STUB'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "$GH_LOG"
exit 0
STUB
  chmod +x "$WORK/bin/gh"

  # AGENT_EXIT makes the agent fail, which is the exit `work` passes up.
  cat > "$WORK/bin/agent" <<'STUB'
#!/usr/bin/env bash
{
  printf 'cwd=%s\n' "$(pwd -P)"
  printf 'argc=%s\n' "$#"
  for a in "$@"; do printf -- '--- argument ---\n%s\n' "$a"; done
} >> "$AGENT_LOG"
exit "${AGENT_EXIT:-0}"
STUB
  chmod +x "$WORK/bin/agent"

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

One paragraph of brief, and nothing the command reads.
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

# make_repo — an adopted repository on `main` carrying this
# repository's own kit, two available tasks and one the queue holds
# back. Committed, so a case can assert that nothing moved after.
make_repo() {
  mkdir -p "$TARGET/.writrun" "$TARGET/work/tasks" "$TARGET/work/specs" "$TARGET/work/reports"
  cp -R "$REPO_ROOT/.writrun/scripts"     "$TARGET/.writrun/scripts"
  cp -R "$REPO_ROOT/.writrun/skills"      "$TARGET/.writrun/skills"
  cp -R "$REPO_ROOT/.writrun/templates"   "$TARGET/.writrun/templates"
  cp -R "$REPO_ROOT/.writrun/conventions" "$TARGET/.writrun/conventions"
  printf 'v0.0.03\n' > "$TARGET/.writrun/VERSION"
  printf '# Tasks\n'   > "$TARGET/work/tasks/README.md"
  printf '# Specs\n'   > "$TARGET/work/specs/README.md"
  printf '# Reports\n' > "$TARGET/work/reports/README.md"

  cat > "$TARGET/.writrun/settings.json" <<'EOF'
{
  "stage": 2,
  "stage_2": {
    "agent_coauthor": false,
    "auto_commit": false,
    "auto_pr": false,
    "auto_push": false,
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
  printf '# AGENTS.md — entry point\n\nThe rules this project works by.\n' > "$TARGET/AGENTS.md"

  (
    cd "$TARGET" || exit 1
    git_q init -q
    git_q symbolic-ref HEAD refs/heads/main
    git_q add -A
    git_q commit -q -m "the kit and the queue"
  )
}

# configure_agent — the adopter's answer to which agent to launch.
configure_agent() { git_q -C "$TARGET" config writrun.agent "$WORK/bin/agent"; }

# work — one `writrun work`, its whole reporting kept in WORK_OUT so a
# case can assert on more than one line of it.
WORK_OUT="$WORK/work.out"
work() {
  "$WRITRUN" work "$@" > "$WORK_OUT" 2>&1
  local code=$?
  cat "$WORK_OUT"
  return $code
}

# launched_argument — everything the stub agent was given, as one blob.
launched_argument() { cat "$AGENT_LOG"; }

make_stubs
