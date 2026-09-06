#!/usr/bin/env bash
. "$(dirname "$0")/../../author_lib.sh"

# What author opens is a rule and the work derived from it. A diff with
# no permanent doc in it is not that, and neither is one carrying code
# beside the doc — one kind per change (spec-0009, step 1 and edge
# cases). Both are refused before a check runs and before the forge is
# named.
make_repo

on_branch report/only-the-queue
work_derived
committed "chore(queue): record what was noticed"
cd "$TARGET" || exit 1

check "a diff with no docs change is refused" 1 "touches no docs/ path" \
  -- author --title "$TITLE" --yes
check "no check ran" 1 "" \
  -- grep -q "all canonical" "$AUTHOR_OUT"
check "the forge was never reached" 1 "" \
  -- grep -q . "$GH_LOG"
check "no branch reached the origin" 1 "" \
  -- git -C "$ORIGIN" rev-parse --verify --quiet refs/heads/report/only-the-queue

git_q -C "$TARGET" checkout -q main
on_branch docs/and-some-code
rule_written
work_derived
mkdir -p "$TARGET/internal"
printf 'package internal\n' > "$TARGET/internal/thing.go"
committed "docs(product): the declaration is the section"

check "a mixed diff is refused" 1 "one kind per change" \
  -- author --title "$TITLE" --yes
check "still nothing on the forge" 1 "" \
  -- grep -q . "$GH_LOG"

finish
