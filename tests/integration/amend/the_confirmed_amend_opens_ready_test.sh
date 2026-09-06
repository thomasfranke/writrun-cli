#!/usr/bin/env bash
. "$(dirname "$0")/../../amend_lib.sh"

# Step 2 and step 4, on the confirmed path: the spec reads `draft` on
# the pushed branch, the branch reached the origin, and the pull request
# opened **ready** — an amendment announces no work, so it is not a
# draft (spec-0011, step 4; conventions/prs.md).
make_repo
in_flight_pr
cd "$TARGET" || exit 1

check "the amendment is opened" 0 "pull request open and ready for review" \
  -- amend_cmd spec-0011 --title "$TITLE" --yes

check "the spec reads draft on the branch" 0 "draft" \
  -- field status work/specs/spec-0011-amend-command.md
check "the amendment is the branch's own commit" 0 "docs(specs): return spec-0011 to draft" \
  -- git log -1 --format=%s
git -C "$TARGET" log -1 --format=%B > "$WORK/message"
check "the commit carries no agent credit" 1 "" \
  -- grep -qiE "co-authored-by|generated with|claude" "$WORK/message"
check "the branch reached the origin" 0 "" \
  -- git -C "$ORIGIN" rev-parse --verify --quiet refs/heads/docs/amend-command
check "the pull request was opened" 0 "" \
  -- grep -q "pr create" "$GH_LOG"
check "it opened ready, not as a draft" 1 "" \
  -- grep -qx -- "--draft" "$GH_ARGS"
check "the title is in the declared style, with no task tag" 0 "" \
  -- grep -qxF -- "[Docs][Specs] $TITLE" "$GH_ARGS"
check "the amended spec is the whole of the change" 0 "work/specs/spec-0011-amend-command.md" \
  -- git diff --name-only main...HEAD

finish
