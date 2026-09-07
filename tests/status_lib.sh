#!/usr/bin/env bash
# status_lib.sh — the fixture behind the status integration cases
# (tests/integration/status/): an adopted repository carrying this
# repository's own copy of the kit, a queue each case shapes, and a main
# branch to read a range against.
#
# The completion checks are copied, never restated: `preflight.sh` and
# the three checks it calls are the authority the binary wraps, so a
# case that stubbed them would be checking the fixture instead of the
# command (docs/about.md, spec-0013).

. "$(dirname "${BASH_SOURCE[0]}")/cli_lib.sh"

PREFLIGHT=".writrun/scripts/stage-1-tasks-and-specs/preflight.sh"

git_q() { git -c user.name=suite -c user.email=suite@test -c commit.gpgsign=false "$@"; }

TARGET="$WORK/target"

# make_repo — an adopted repository at the tag this client pins, one
# task with its approved spec, one open report, main, and a task branch
# checked out. Every case starts from a queue the checks pass on and
# breaks exactly what it is about.
make_repo() {
  mkdir -p "$TARGET/work/tasks" "$TARGET/work/specs" "$TARGET/work/reports" \
           "$TARGET/docs/product/queue" "$TARGET/docs/technical/testing"
  cp -R "$REPO_ROOT/.writrun" "$TARGET/.writrun"
  printf '%s\n' "$PINNED" > "$TARGET/.writrun/VERSION"
  printf '# Tasks\n'   > "$TARGET/work/tasks/README.md"
  printf '# Specs\n'   > "$TARGET/work/specs/README.md"
  printf '# Reports\n' > "$TARGET/work/reports/README.md"
  printf '# status\n'  > "$TARGET/docs/product/queue/status.md"
  printf '# suites\n'  > "$TARGET/docs/technical/testing/suites.md"

  cat > "$TARGET/work/tasks/task-0014-fixture.md" <<'EOF'
---
id: task-0014
status: in-progress
blocked_reason: null
taken_by: null
spec_ref: [spec-0013]
doc_ref: product/queue/status.md
origin: rule
priority: medium
depends_on: []
milestone: null
created: 2026-09-03T23:27:24Z
queued: 2026-09-04T05:05:03Z
completed: null
merged: null
provenance: []
---

# Answer where the work stands from the current branch
EOF

  cat > "$TARGET/work/specs/spec-0013-fixture.md" <<'EOF'
---
id: spec-0013
task_ref: task-0014
status: approved
created: 2026-09-03T23:27:24Z
---

# spec-0013 — Read the branch, the checks and the queue
EOF

  (
    cd "$TARGET" || exit 1
    git_q init -q
    git_q symbolic-ref HEAD refs/heads/main
    printf '# A project\n' > README.md
    git_q add -A
    git_q commit -q -m "initial import"
    git_q checkout -q -b task/0014-status-command
    printf 'the work\n' >> README.md
    git_q add -A
    git_q commit -q -m "the work"
  )
}

# report <id> <status> — one report in the queue, at that status.
report() {
  local triaged="null" refs="[]"
  [ "$2" = open ] || triaged="2026-09-05T12:00:00Z"
  [ "$2" = tracked ] && refs="[task-0014]"
  cat > "$TARGET/work/reports/$1-fixture.md" <<EOF
---
id: $1
status: $2
task_ref: $refs
doc_ref: technical/testing/suites.md
created: 2026-09-05T11:00:00Z
triaged: $triaged
---

# Something that was noticed
EOF
}

# break_the_first_check — a queue file the front-matter sweep refuses.
# The value is quoted, which is exactly the shape the line-based readers
# cannot see; stage 1 of preflight is that sweep, so this is what
# `finish` would stop at first.
break_the_first_check() {
  local f="$TARGET/work/tasks/task-0014-fixture.md"
  sed 's/^status: in-progress$/status: "in-progress"/' "$f" > "$f.broken"
  mv "$f.broken" "$f"
}

# kit_tag <tag> — what the kit records it was installed at.
kit_tag() { printf '%s\n' "$1" > "$TARGET/.writrun/VERSION"; }

# on_branch <name> — the branch the answer is read from.
on_branch() { (cd "$TARGET" && git_q checkout -q "$@"); }
