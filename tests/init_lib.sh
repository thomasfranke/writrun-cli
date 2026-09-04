#!/usr/bin/env bash
# init_lib.sh — the fixture behind the init integration cases
# (tests/integration/init/): a local WritRun repository as the pinned
# source (spec-0002, tests required), and a target repository to adopt.
#
# The source is built here rather than cloned from the network, so the
# suite is hermetic; the e2e tier is where the real upstream template
# is the source. WRITRUN_SOURCE points the binary's fetch at it, and
# WRITRUN_TTY_IN feeds key bytes to the forms where a case needs a
# terminal (a decline, an arrow selection).

. "$(dirname "${BASH_SOURCE[0]}")/cli_lib.sh"

# The tag the binary pins, read from the binary itself so the fixture
# and the release can never disagree.
TAG=$("$WRITRUN" --version | sed -n 's/.*pins WritRun \(v[0-9.]*\).*/\1/p')

git_q() { git -c user.name=suite -c user.email=suite@test -c commit.gpgsign=false "$@"; }

# make_source <dir> — a WritRun repository with a template/ carrying
# the files init touches, committed and tagged $TAG.
make_source() {
  local src="$1"
  mkdir -p "$src"
  (
    cd "$src" || exit 1
    git_q init -q
    mkdir -p template/.writrun/conventions \
             template/.writrun/scripts/stage-2-pull-requests \
             template/docs/product template/docs/technical \
             template/work/tasks template/work/specs template/work/reports

    cat > template/AGENTS.md <<'EOF'
# AGENTS.md — entry point for AI agents

<!-- TODO: one paragraph. -->

## WritRun — working the queue

<!-- writrun:begin
     This section is WritRun's flow. -->

### Picking work

The flow's text.

<!-- writrun:end -->
EOF

    cat > template/.writrun/settings.json <<'EOF'
{
  "stage": 1,
  "stage_2": {
    "auto_commit": false
  }
}
EOF

    cat > template/.writrun/conventions/commits.md <<'EOF'
# Commits

- **Types**: `docs`, `feat`, `fix`, `refactor`, `chore`.
- **Scopes** (optional — omit when a change genuinely spans the
  repository): `about`, `product`, `technical`.
- Example: `docs(product): add a chapter`.
EOF

    cat > template/.writrun/scripts/stage-2-pull-requests/check_observance.sh <<'EOF'
#!/usr/bin/env bash
# check_observance.sh — the door.
TYPES="docs feat fix refactor chore"
SCOPES="about product technical"
exit 0
EOF
    chmod +x template/.writrun/scripts/stage-2-pull-requests/check_observance.sh

    printf '%s\n' "$TAG" > template/.writrun/VERSION
    printf '# This project uses WritRun\n' > template/WRITRUN.md
    printf '# Product skeleton\n' > template/docs/product/README.md
    printf '# Technical skeleton\n' > template/docs/technical/README.md
    printf '# Tasks\n' > template/work/tasks/README.md
    printf '# Specs\n' > template/work/specs/README.md
    printf '# Reports\n' > template/work/reports/README.md

    git_q add .
    git_q commit -q -m "the kit"
    git_q tag "$TAG"
  )
}

# make_target <dir> [subject…] — a repository to adopt: one commit per
# subject ("initial import" alone by default), a clean tree.
make_target() {
  local dir="$1"; shift
  mkdir -p "$dir"
  (
    cd "$dir" || exit 1
    git_q init -q
    printf '# A project\n' > README.md
    git_q add .
    if [ "$#" -eq 0 ]; then
      git_q commit -q -m "initial import"
    else
      git_q commit -q -m "$1"; shift
      for s in "$@"; do git_q commit -q --allow-empty -m "$s"; done
    fi
  )
}

# The forge, stubbed the way the tier demands (technical/testing/
# tiers.md): authenticated, squash on, write permissions, issues on —
# so a stage-2 or -3 adoption's checks read canned answers, never the
# network.
mkdir -p "$WORK/bin"
cat > "$WORK/bin/gh" <<'EOF'
#!/usr/bin/env bash
case "$*" in
  "auth status") exit 0 ;;
  *".allow_squash_merge") echo true ;;
  *".default_workflow_permissions") echo write ;;
  *".has_issues") echo true ;;
  *) exit 0 ;;
esac
EOF
chmod +x "$WORK/bin/gh"
export PATH="$WORK/bin:$PATH"

SOURCE="$WORK/source"
TARGET="$WORK/target"
make_source "$SOURCE"
export WRITRUN_SOURCE="$SOURCE"
