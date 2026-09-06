#!/usr/bin/env bash
. "$(dirname "$0")/../../amend_lib.sh"

# The task and spec counters run independently and are routinely one
# apart — here task-0012 rides spec-0011, and spec-0012 is a different
# subject entirely. `amend task-0012` kept only the digits and resolved
# to spec-0012, returning it to draft and pushing it without a word.
#
# Under --yes, which the project's own rules make first-class ("every
# question has a flag that answers it, so automation never meets a
# prompt"), no composition is shown in time for a person to catch it. So
# the refusal has to be the command's, before anything is composed.
make_repo
spec_file 0012 release-distribution task-0013 approved
git_q -C "$TARGET" add -A
git_q -C "$TARGET" commit -q -m "spec-0012 lands"
git_q -C "$TARGET" push -q origin main
in_flight_pr
cd "$TARGET" || exit 1

check "the task id is refused, naming what it would have hit" 1 "different file" \
  -- amend_cmd task-0012 --title "$TITLE" --yes

check "the id it names is the one that was typed" 0 "" \
  -- grep -qF "task-0012" "$AMEND_OUT"
check "spec-0012 was not returned to draft" 0 "approved" \
  -- field status work/specs/spec-0012-release-distribution.md
check "spec-0011 was not touched either" 0 "approved" \
  -- field status work/specs/spec-0011-amend-command.md
check "nothing reached the forge" 1 "" \
  -- grep -q "pr create" "$GH_LOG"
check "no branch was cut" 1 "" \
  -- git rev-parse --verify --quiet refs/heads/docs/release-distribution
check "the working tree is untouched" 0 "" \
  -- test -z "$(git status --porcelain)"

# The bare number still resolves — it declares no kind, and a person who
# types `amend 11` means the spec they are amending.
check "a bare number still amends the spec" 0 "Amended spec-0011" \
  -- amend_cmd 11 --title "$TITLE" --yes

finish
