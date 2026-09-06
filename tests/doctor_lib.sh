#!/usr/bin/env bash
# doctor_lib.sh — the fixture behind the doctor integration cases
# (tests/integration/doctor/): an adopted repository carrying this
# repository's own copy of the three scripts doctor runs, a project half
# that satisfies every stage-1 assumption, and the forge stubbed.
#
# The scripts are copied, never restated: the settings reader and the
# two checks are the authorities the binary wraps, so a case that wrote
# its own would be checking the fixture instead of the command
# (docs/about.md).

. "$(dirname "${BASH_SOURCE[0]}")/cli_lib.sh"

READER=".writrun/scripts/stage-2-pull-requests/read_setting.sh"
SETTINGS_CHECK=".writrun/scripts/stage-2-pull-requests/check_settings.sh"
FRONT_MATTER=".writrun/skills/writrun-check-front-matter/check_front_matter.sh"

git_q() { git -c user.name=suite -c user.email=suite@test -c commit.gpgsign=false "$@"; }

TARGET="$WORK/target"

# make_repo [stage] — an adopted repository nothing in stages 0–3 has a
# finding about: the three documents, the docs/ and work/ split, an
# AGENTS.md with the fence intact and the four gates answered, the kit's
# tag recorded, canonical settings, and one commit so a case can ask git
# whether anything changed.
make_repo() {
  local stage="${1:-3}"
  mkdir -p "$TARGET/.writrun/scripts/stage-2-pull-requests" \
           "$TARGET/.writrun/skills/writrun-check-front-matter" \
           "$TARGET/docs/product" "$TARGET/docs/technical" \
           "$TARGET/work/tasks" "$TARGET/work/specs" "$TARGET/work/reports"
  cp "$REPO_ROOT/$READER" "$TARGET/$READER"
  cp "$REPO_ROOT/$SETTINGS_CHECK" "$TARGET/$SETTINGS_CHECK"
  cp "$REPO_ROOT/$FRONT_MATTER" "$TARGET/$FRONT_MATTER"

  printf '# About\n\nWhat this is.\n'   > "$TARGET/docs/about.md"
  printf '# Product\n'                  > "$TARGET/docs/product/README.md"
  printf '# Rules\n'                    > "$TARGET/docs/product/rules.md"
  printf '# Technical\n'                > "$TARGET/docs/technical/README.md"
  printf '# Boundaries\n'               > "$TARGET/docs/technical/boundaries.md"
  printf '# Tasks\n'                    > "$TARGET/work/tasks/README.md"
  printf '# Specs\n'                    > "$TARGET/work/specs/README.md"
  printf '# Reports\n'                  > "$TARGET/work/reports/README.md"
  printf 'v0.0.03\n'                    > "$TARGET/.writrun/VERSION"

  agents
  settings "$stage"

  (
    cd "$TARGET" || exit 1
    git_q init -q
    printf '# A project\n' > README.md
    git_q add -A
    git_q commit -q -m "initial import"
  )
}

# agents [who] — AGENTS.md with the fence intact and the four gates
# answered. Passing a placeholder leaves the docs gate unanswered.
agents() {
  local who="${1:-Thomas reviews before merge.}"
  cat > "$TARGET/AGENTS.md" <<EOF
# AGENTS.md — entry point for AI agents

A project.

## WritRun — working the queue

<!-- writrun:begin
     This section is WritRun's flow. -->

### Human gates

| Transition | Who |
|---|---|
| Writing or changing anything under \`docs/\` | $who |
| An authored rule is finished, so derivation may start | Thomas declares it. |
| Spec \`draft → approved\` | Thomas only, via the merged PR. |
| Task with empty \`spec_ref\` and insufficient brief | Stop and ask for a spec. |
| Everything else | Agent, autonomously. |

<!-- writrun:end -->
EOF
}

# settings <stage> — .writrun/settings.json in the shape check_settings.sh
# holds the file to.
settings() {
  cat > "$TARGET/.writrun/settings.json" <<EOF
{
  "stage": $1,
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
}

# The forge, stubbed the way the tier demands (technical/testing/
# tiers.md): every invocation is answered from a file named after its
# own arguments, and an argument list with no file is a read the forge
# refuses. Every call is recorded, so a case can prove a read was never
# attempted.
mkdir -p "$WORK/gh" "$WORK/bin"
export GH_DIR="$WORK/gh"
cat > "$WORK/bin/gh" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "$GH_DIR/calls"
if [ ! -f "$GH_DIR/authenticated" ]; then
  echo "gh: To get started with GitHub CLI, please run: gh auth login" >&2
  exit 1
fi
[ "$1" = "auth" ] && exit 0
key=$(printf '%s' "$*" | tr -cs 'A-Za-z0-9' '_')
if [ -f "$GH_DIR/$key" ]; then cat "$GH_DIR/$key"; exit 0; fi
echo "gh: HTTP 404 for $*" >&2
exit 1
EOF
chmod +x "$WORK/bin/gh"
export PATH="$WORK/bin:$PATH"

# gh_reply <arguments> <output> — what gh prints for that invocation.
gh_reply() {
  local key
  key=$(printf '%s' "$1" | tr -cs 'A-Za-z0-9' '_')
  printf '%s\n' "$2" > "$GH_DIR/$key"
}

# forge_healthy — every forge assumption met: squash on, workflow
# permissions read-and-write, main governed by a ruleset that names a
# bypass actor and none of the four blocking rules, Issues on.
forge_healthy() {
  : > "$GH_DIR/authenticated"
  : > "$GH_DIR/calls"
  gh_reply "api repos/{owner}/{repo} --jq .allow_squash_merge" "true"
  gh_reply "api repos/{owner}/{repo} --jq .has_issues" "true"
  gh_reply "api repos/{owner}/{repo} --jq .owner.type" "User"
  gh_reply "api repos/{owner}/{repo}/actions/permissions/workflow --jq .default_workflow_permissions" "write"
  gh_reply "api repos/{owner}/{repo}/rules/branches/main --jq .[].type" "deletion"
  gh_reply "api repos/{owner}/{repo}/rules/branches/main --jq .[].ruleset_id" "42"
  gh_reply "api repos/{owner}/{repo}/rulesets/42 --jq (.bypass_actors // [])[].actor_type" "Integration"
}

# workflow <name> <body> — one file in the target's
# `.github/workflows`, which stage 2 reads only where the repository's
# Actions default is `read`.
workflow() {
  mkdir -p "$TARGET/.github/workflows"
  printf '%s\n' "$2" > "$TARGET/.github/workflows/$1"
}

# no_workflows — the directory removed, so the repository pushes from
# nowhere.
no_workflows() { rm -rf "$TARGET/.github"; }

# forge_offline — gh answers nothing at all, the way an unauthenticated
# one does.
forge_offline() {
  rm -f "$GH_DIR/authenticated"
  : > "$GH_DIR/calls"
}

# gh_calls — every invocation the run made, one per line.
gh_calls() { cat "$GH_DIR/calls" 2>/dev/null; }

forge_healthy
