#!/usr/bin/env bash
. "$(dirname "$0")/../../doctor_lib.sh"

# `rules/branches/main` answers for rulesets and for nothing else, so a
# branch a classic protection rule governs used to be examined as if
# nothing governed it — ungoverned, and nothing refusing the recording
# push, both wrong at once (report-0048). The classic rule is read
# through the branch object and its own payload
# (docs/product/adoption/doctor.md).
make_repo 3
cd "$TARGET" || exit 1

# No ruleset targets main: everything below is the classic rule's doing.
no_rulesets() {
  gh_reply "api repos/{owner}/{repo}/rules/branches/main --jq .[].type" ""
  gh_reply "api repos/{owner}/{repo}/rules/branches/main --jq .[].ruleset_id" ""
}

forge_healthy
no_rulesets
gh_reply "api repos/{owner}/{repo}/branches/main --jq .protection.enabled" "true"
gh_reply "api repos/{owner}/{repo}/branches/main/protection" \
  '{"required_pull_request_reviews":{"required_approving_review_count":0},"enforce_admins":{"enabled":true}}'
check "a classic rule requiring a pull request stops the recording push" 1 \
  "the branch protection rule over main enables pull_request (require a pull request before merging)" \
  -- "$WRITRUN" doctor

forge_healthy
no_rulesets
gh_reply "api repos/{owner}/{repo}/branches/main --jq .protection.enabled" "true"
gh_reply "api repos/{owner}/{repo}/branches/main/protection" \
  '{"required_linear_history":{"enabled":true},"allow_force_pushes":{"enabled":false}}'
"$WRITRUN" doctor > "$WORK/classic.out" 2>&1
check "a classic rule a fast-forward meets governs main and stops nothing" 0 \
  "✓  main is governed by a protection rule" -- cat "$WORK/classic.out"
check "and it exits 0" 0 "" -- "$WRITRUN" doctor

finish
