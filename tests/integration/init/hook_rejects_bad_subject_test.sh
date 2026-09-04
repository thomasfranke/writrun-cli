#!/usr/bin/env bash
. "$(dirname "$0")/../../init_lib.sh"

# The installed hook rejects a commit whose subject violates the
# convention, naming the fault (spec-0002, acceptance criteria). The
# target's history is Conventional, so the extracted vocabulary — not
# the shipped one — is what the hook enforces.
make_target "$TARGET" "feat: begin" "fix(api): mend it"
cd "$TARGET" || exit 1

check "init adopts, extracting the vocabulary" 0 "extracted from the commit history" \
  -- "$WRITRUN" init --stage 1 --yes

check "a shapeless subject is rejected naming the grammar" 1 "not a Conventional subject" \
  -- git_q commit --allow-empty -m "did some stuff"
check "a type outside the vocabulary is rejected naming it" 1 "outside the vocabulary" \
  -- git_q commit --allow-empty -m "chore: tidy"
check "a subject inside the vocabulary passes" 0 "" \
  -- git_q commit --allow-empty -m "feat(api): add a thing"

finish
