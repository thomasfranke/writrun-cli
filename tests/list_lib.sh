#!/usr/bin/env bash
# list_lib.sh — the fixture behind the list integration cases
# (tests/integration/list/): an adopted repository carrying this
# repository's own copy of the selection skill's lister, a queue each
# case writes, and the forge stubbed.
#
# The lister is copied, never restated: it is the eligibility authority
# the binary wraps, so a case that wrote its own would be checking the
# fixture instead of the command (docs/about.md).

. "$(dirname "${BASH_SOURCE[0]}")/cli_lib.sh"

LISTER=".writrun/skills/writrun-select-next-task/list_tasks.sh"
# The lister sources the queue reader from the scripts folder, so the
# fixture carries both — the kit's own layout, not this file's choice.
QUEUE_LIB=".writrun/scripts/stage-2-pull-requests/queue_lib.sh"

git_q() { git -c user.name=suite -c user.email=suite@test -c commit.gpgsign=false "$@"; }

TARGET="$WORK/target"

# make_repo — an adopted repository with the real lister, an empty
# queue, and one commit, so a case can ask git whether anything changed.
make_repo() {
  mkdir -p "$TARGET/.writrun/skills/writrun-select-next-task" \
           "$TARGET/.writrun/scripts/stage-2-pull-requests" \
           "$TARGET/work/tasks" "$TARGET/work/specs" "$TARGET/work/reports"
  cp "$REPO_ROOT/$LISTER" "$TARGET/$LISTER"
  cp "$REPO_ROOT/$QUEUE_LIB" "$TARGET/$QUEUE_LIB"
  printf '# Tasks\n' > "$TARGET/work/tasks/README.md"
  printf '# Specs\n' > "$TARGET/work/specs/README.md"
  printf '# Reports\n' > "$TARGET/work/reports/README.md"
  (
    cd "$TARGET" || exit 1
    git_q init -q
    printf '# A project\n' > README.md
    git_q add -A
    git_q commit -q -m "initial import"
  )
}

# task <id> <status> <priority> <spec_ref> <title> [blocked_reason]
task() {
  local id="$1" st="$2" pr="$3" sp="$4" tt="$5" br="${6:-null}"
  cat > "$TARGET/work/tasks/$id-fixture.md" <<EOF
---
id: $id
status: $st
blocked_reason: $br
spec_ref: [$sp]
depends_on: []
priority: $pr
created: 2026-01-01T00:00:00Z
completed: null
---

# $tt
EOF
}

# spec <id> <status>
spec() {
  cat > "$TARGET/work/specs/$1-fixture.md" <<EOF
---
id: $1
status: $2
---

# $1 — a fixture spec
EOF
}

# report <id> <status> <title>
report() {
  cat > "$TARGET/work/reports/$1-fixture.md" <<EOF
---
id: $1
status: $2
---

# $3
EOF
}

# The forge, stubbed the way the tier demands (technical/testing/
# tiers.md): the open pull request list is a file a case writes, and its
# absence is a forge that cannot be reached.
mkdir -p "$WORK/bin"
cat > "$WORK/bin/gh" <<'EOF'
#!/usr/bin/env bash
case "$1" in
  pr)  [ -f "$GH_PR_LIST" ] || exit 1; cat "$GH_PR_LIST" ;;
  api) [ -f "$GH_PR_LIST" ] || exit 1 ;;
  *)   exit 0 ;;
esac
EOF
chmod +x "$WORK/bin/gh"
export PATH="$WORK/bin:$PATH"
export GH_PR_LIST="$WORK/pr_list"

# forge_offline — the forge answers nothing; the lister says so.
forge_offline() { rm -f "$GH_PR_LIST"; }

# forge_online [<number> <branch> <author> <title>] — the forge answers,
# with that pull request open, or with none.
forge_online() {
  : > "$GH_PR_LIST"
  [ "$#" -eq 0 ] && return 0
  printf '%s\t%s\t%s\t%s\n' "$1" "$2" "$3" "$4" >> "$GH_PR_LIST"
}

forge_offline
