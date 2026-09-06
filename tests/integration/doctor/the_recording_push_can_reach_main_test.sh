#!/usr/bin/env bash
. "$(dirname "$0")/../../doctor_lib.sh"

# Stage 2 asks whether the recording push can reach main, not how the
# forge is configured to let it: two repositories configured differently
# can both be right, and a check that names one configuration fails the
# other for no defect (docs/product/adoption/doctor.md; spec-0019).
make_repo 3
cd "$TARGET" || exit 1

RECORDS='name: record
on: [pull_request]
permissions:
  contents: write
jobs:
  record:
    runs-on: ubuntu-latest
    steps:
      - run: git push origin "HEAD:${BASE_REF}"'

SILENT='name: record
on: [pull_request]
permissions:
  contents: read
jobs:
  record:
    runs-on: ubuntu-latest
    steps:
      - run: git push origin "HEAD:${BASE_REF}"'

ELSEWHERE='name: pages
on: [push]
jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - run: git push origin HEAD:gh-pages'

# The right to write, granted per workflow rather than to every one at
# once.
forge_healthy
gh_reply "api repos/{owner}/{repo}/actions/permissions/workflow --jq .default_workflow_permissions" "read"
workflow record.yml "$RECORDS"
check "a read default with the right raised per workflow holds" 0 \
  "Stage 2 — the forge: all clear." -- "$WRITRUN" doctor

workflow record.yml "$SILENT"
check "a pushing workflow that raises nothing is named" 1 \
  ".github/workflows/record.yml pushes to main and raises no" -- "$WRITRUN" doctor

# A push to a branch that is not main is another flow's, and what it may
# write is not stage 2's business.
workflow record.yml "$ELSEWHERE"
check "a workflow pushing to another branch is left alone" 0 \
  "Stage 2 — the forge: all clear." -- "$WRITRUN" doctor

no_workflows
check "a read default with nothing that pushes holds" 0 \
  "Stage 2 — the forge: all clear." -- "$WRITRUN" doctor

# An empty bypass list denies nothing where the ruleset enables no rule
# a fast-forward push meets.
forge_healthy
gh_reply "api repos/{owner}/{repo}/rules/branches/main --jq .[].type" "deletion"
gh_reply "api repos/{owner}/{repo}/rulesets/42 --jq (.bypass_actors // [])[].actor_type" ""
check "an empty bypass list with nothing to bypass holds" 0 \
  "Stage 2 — the forge: all clear." -- "$WRITRUN" doctor

gh_reply "api repos/{owner}/{repo}/rules/branches/main --jq .[].type" "update"
check "a rule that refuses the push with no bypass actor names it" 1 \
  "ruleset 42 governs main, enables update (restrict updates) and names no bypass actor" \
  -- "$WRITRUN" doctor

finish
