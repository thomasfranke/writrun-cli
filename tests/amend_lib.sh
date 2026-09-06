#!/usr/bin/env bash
# amend_lib.sh — the fixture behind the amend integration cases
# (tests/integration/amend/): an adopted repository carrying this
# repository's own kit, a bare origin standing in for the forge, and a
# stubbed `gh`. The suite never reaches a real forge: the kit's
# WRITRUN_PR_LIST seam supplies the open pull requests, every `gh`
# invocation is the stub's, and GH_LOG / GH_ARGS record them so a case
# can assert what did and did not reach it (spec-0011, tests required).

. "$(dirname "${BASH_SOURCE[0]}")/cli_lib.sh"

git_q() { git -c user.name=suite -c user.email=suite@test -c commit.gpgsign=false "$@"; }

ORIGIN="$WORK/origin.git"
TARGET="$WORK/target"

# The forge, stubbed. GH_PR_LIST is what `gh pr list` answers when no
# WRITRUN_PR_LIST seam is set; GH_PR_LIST_EXIT makes the read fail the
# way an unreachable forge does; GH_PR_CREATE_EXIT fails the opening.
#
# GH_LOG holds one line per invocation, arguments joined — a body's own
# newlines survive it, so a case greps the composed body there. GH_ARGS
# holds one argument per line, which is how a case asks whether a flag
# was passed at all rather than whether the word appears somewhere.
#
# GH_BODY holds the body `pr create` was handed, alone — what the kit's
# own check_amendment_reference.sh reads out of $PR_BODY, so a case can
# hand the composed body straight to the gate it was composed for.
GH_LOG="$WORK/gh.log"
GH_ARGS="$WORK/gh.args"
GH_BODY="$WORK/gh.body"
export GH_LOG GH_ARGS GH_BODY

make_gh() {
  mkdir -p "$WORK/bin"
  : > "$GH_LOG"
  : > "$GH_ARGS"
  : > "$GH_BODY"
  cat > "$WORK/bin/gh" <<'STUB'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "$GH_LOG"
printf '%s\n' "$@" >> "$GH_ARGS"
prev=""
for a in "$@"; do
  [ "$prev" = "--body" ] && printf '%s\n' "$a" > "$GH_BODY"
  prev="$a"
done
case "$1 $2" in
  "auth status") exit 0 ;;
  "pr list")
    if [ -n "${GH_PR_LIST_EXIT:-}" ] && [ "${GH_PR_LIST_EXIT}" != 0 ]; then
      echo "could not reach the forge" >&2
      exit "$GH_PR_LIST_EXIT"
    fi
    printf '%s' "${GH_PR_LIST:-}"
    exit 0 ;;
  "pr create")
    if [ "${GH_PR_CREATE_EXIT:-0}" != 0 ]; then
      echo "gh: validation failed" >&2
      exit "$GH_PR_CREATE_EXIT"
    fi
    echo "https://forge/pull/99"
    exit 0 ;;
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
created: 2026-01-01T00:00:00Z
queued: 2026-01-01T00:00:00Z
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

## Outcome

_(fill after execution)_
EOF
}

# make_repo — an adopted repository on `main`, its kit copied from this
# repository, pushed to a bare origin. spec-0011 is the approved one
# every case amends, with task-0012 in flight on it; spec-0013 is a
# draft, which amend must refuse.
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

  spec_file 0011 amend-command task-0012 approved
  task_file 0012 amend-command in-progress spec-0011 "Amend a spec"
  spec_file 0013 another-thing task-0014 draft
  task_file 0014 another-thing ready spec-0013 "Another thing"
  printf '# A project\n' > "$TARGET/README.md"

  (
    cd "$TARGET" || exit 1
    git_q init -q
    git_q symbolic-ref HEAD refs/heads/main
    git_q add -A
    git_q commit -q -m "the kit and the queue"
  )
  # The binary commits as itself, with no -c overrides of its own, so the
  # identity has to be the repository's.
  git -C "$TARGET" config user.name suite
  git -C "$TARGET" config user.email suite@test
  git -C "$TARGET" config commit.gpgsign false
  git_q init --bare -q "$ORIGIN"
  git_q -C "$TARGET" remote add origin "$ORIGIN"
  git_q -C "$TARGET" push -q -u origin main
}

# in_flight_pr — the seam the kit itself reads: one open pull request
# working task-0012, on the branch and under the title a take would have
# given it.
in_flight_pr() {
  export WRITRUN_PR_LIST="$(printf '42\ttask/0012-amend-command\tsomeone\t[TASK-0012] Amend a spec')"
}

# amend_cmd — one `writrun amend`, its whole reporting kept in AMEND_OUT
# so a case can assert on more than one line of it.
AMEND_OUT="$WORK/amend.out"
amend_cmd() {
  "$WRITRUN" amend "$@" > "$AMEND_OUT" 2>&1
  local code=$?
  cat "$AMEND_OUT"
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

# TITLE is the sentence the composed title carries.
TITLE="Reopen the amendment gate"

make_gh
