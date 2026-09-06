#!/usr/bin/env bash
. "$(dirname "$0")/../../doctor_lib.sh"

# Every forge finding names the setting and what is expected of it. What
# breaks the recording push exits non-zero; what the methodology merely
# recommends is reported as a recommendation and does not
# (docs/product/adoption/doctor.md).
make_repo 3
cd "$TARGET" || exit 1

forge_healthy
gh_reply "api repos/{owner}/{repo} --jq .allow_squash_merge" "false"
check "squash merging off is named" 1 "squash merging is off" -- "$WRITRUN" doctor

forge_healthy
gh_reply "api repos/{owner}/{repo} --jq .has_issues" "false"
check "Issues disabled is named at stage 3" 1 "Issues are disabled" -- "$WRITRUN" doctor

forge_healthy
gh_reply "api repos/{owner}/{repo}/rules/branches/main --jq .[].type" "required_signatures"
check "a rule that blocks the recording push is named" 1 \
  "the rule required_signatures (require signed commits) is on for main" -- "$WRITRUN" doctor

forge_healthy
gh_reply "api repos/{owner}/{repo}/rules/branches/main --jq .[].type" "pull_request"
gh_reply "api repos/{owner}/{repo} --jq .owner.type" "Organization"
check "the pull-request rule is left alone on an organization" 0 \
  "Stage 2 — the forge: all clear." -- "$WRITRUN" doctor

forge_healthy
gh_reply "api repos/{owner}/{repo}/rules/branches/main --jq .[].type" ""
gh_reply "api repos/{owner}/{repo}/rules/branches/main --jq .[].ruleset_id" ""
"$WRITRUN" doctor > "$WORK/unprotected.out" 2>&1
check "an unprotected main is a recommendation" 0 \
  "advises  main is governed by no ruleset" -- cat "$WORK/unprotected.out"
check "a recommendation alone exits 0" 0 "" -- "$WRITRUN" doctor

finish
