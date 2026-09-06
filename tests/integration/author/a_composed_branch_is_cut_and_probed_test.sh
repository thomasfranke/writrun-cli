#!/usr/bin/env bash
. "$(dirname "$0")/../../author_lib.sh"

# A change that is not on an authoring branch is cut onto the composed
# one on the word, and never before it (spec-0009, step 3) — and the
# composed name is asked about on the forge first, because cutting it
# fresh and pushing would land on whatever is already under it, which is
# somebody's open pull request (take_task.sh refuses the same way).
make_repo
authoring_change scratch
cd "$TARGET" || exit 1

check "the composed branch is cut and opened" 0 "pull request open and ready for review" \
  -- author --title "$TITLE" --slug derived-work --yes
check "HEAD is on the branch that was composed" 0 "docs/derived-work" \
  -- git rev-parse --abbrev-ref HEAD
check "the composed branch reached the origin" 0 "" \
  -- git -C "$ORIGIN" rev-parse --verify --quiet refs/heads/docs/derived-work
check "the pull request was opened off it" 0 "" \
  -- grep -q -- "--head docs/derived-work" "$GH_LOG"
check "the branch it was cut from is untouched" 0 "" \
  -- git rev-parse --verify --quiet refs/heads/scratch

# The same composition again, from a second local branch. The local
# branch is deleted first, so the only thing left saying the name is
# taken is the forge — the refusal this asserts is the remote one, not
# the local one standing in front of it.
: > "$GH_LOG"
git_q -C "$TARGET" checkout -q main
git_q -C "$TARGET" branch -q -D docs/derived-work
on_branch scratch-again
rule_written
committed "docs(product): a second rule"

check "a composed branch already on the forge is refused" 1 "already on the forge" \
  -- author --title "$TITLE" --slug derived-work --yes
check "nothing reached the forge" 1 "" \
  -- grep -q . "$GH_LOG"
check "no branch was cut" 1 "" \
  -- git rev-parse --verify --quiet refs/heads/docs/derived-work
check "HEAD never moved" 0 "scratch-again" \
  -- git rev-parse --abbrev-ref HEAD

finish
