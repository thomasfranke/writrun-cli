#!/usr/bin/env bash
# report_lib.sh — the fixture behind the report integration cases
# (tests/integration/report/): an adopted repository carrying this
# repository's own creation skill, one report already recorded so the
# next id has a sequence to continue, a bare origin standing in for the
# forge, and a stubbed `gh`.
#
# The generator is copied, never restated: it is the authority the
# binary wraps, so a case that wrote its own file would be checking the
# fixture instead of the command (docs/about.md).

. "$(dirname "${BASH_SOURCE[0]}")/cli_lib.sh"

git_q() { git -c user.name=suite -c user.email=suite@test -c commit.gpgsign=false "$@"; }

ORIGIN="$WORK/origin.git"
TARGET="$WORK/target"

# The forge, stubbed. GH_LOG records every invocation, so a case can
# assert what did and did not reach it; `pr list` answers with no open
# pull request, which is the generator's forge view.
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
esac
exit 0
STUB
  chmod +x "$WORK/bin/gh"
  export PATH="$WORK/bin:$PATH"
}

# make_repo — an adopted repository on `main` carrying the creation
# skill and its templates, with report-0001 already recorded and one
# ready task in the queue, pushed to a bare origin.
make_repo() {
  mkdir -p "$TARGET/.writrun" "$TARGET/work/tasks" "$TARGET/work/specs" "$TARGET/work/reports"
  cp -R "$REPO_ROOT/.writrun/scripts"     "$TARGET/.writrun/scripts"
  cp -R "$REPO_ROOT/.writrun/skills"      "$TARGET/.writrun/skills"
  cp -R "$REPO_ROOT/.writrun/templates"   "$TARGET/.writrun/templates"
  cp -R "$REPO_ROOT/.writrun/conventions" "$TARGET/.writrun/conventions"

  cat > "$TARGET/.writrun/settings.json" <<'EOF'
{
  "stage": 2,
  "stage_2": {
    "agent_coauthor": false,
    "auto_commit": false,
    "auto_pr": false,
    "auto_push": true,
    "pr_title_style": "bracketed"
  }
}
EOF

  cat > "$TARGET/work/reports/report-0001-an-earlier-finding.md" <<'EOF'
---
id: report-0001
status: open
task_ref: []
doc_ref: null
created: 2026-01-01T00:00:00Z
triaged: null
---

# An earlier finding

One paragraph of what was seen.
EOF

  cat > "$TARGET/work/tasks/task-0001-a-thing.md" <<'EOF'
---
id: task-0001
status: ready
blocked_reason: null
taken_by: null
spec_ref: []
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

# A thing to do

One paragraph of brief.
EOF

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

# report — one `writrun report`, its whole reporting kept in REPORT_OUT
# so a case can assert on more than one line of it.
REPORT_OUT="$WORK/report.out"
report() {
  "$WRITRUN" report "$@" > "$REPORT_OUT" 2>&1
  local code=$?
  cat "$REPORT_OUT"
  return $code
}

# RECORDED is the file the next minted id lands in — report-0001 is
# already there, so the sequence continues at 0002.
RECORDED="work/reports/report-0002-a-finding.md"

# TITLE and BODY are one observation: what was seen, and the words the
# reporter wrote.
TITLE="The fixtures table omits two fixtures"
BODY="Found while adding a suite: the table names five, the tree holds seven."

make_gh
