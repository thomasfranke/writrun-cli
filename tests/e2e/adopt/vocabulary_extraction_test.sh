#!/usr/bin/env bash
. "$(dirname "$0")/../../cli_lib.sh"

# The rewrite half of the adoption, against the real template rather
# than a fixture: a repository whose history is Conventional has its own
# vocabulary written onto both halves of the kit's statement — the prose
# lists in conventions/commits.md and the TYPES/SCOPES lines in
# check_observance.sh. The parity case next door adopts a repository
# with no Conventional history, so the shipped defaults stand there and
# this path is never taken; a drift in the real files' line shapes would
# fail the adoption mid-apply and nothing else would notice.

TAG=$("$WRITRUN" --version | sed -n 's/.*pins WritRun \(v[0-9.]*\).*/\1/p')
git_q() { git -c user.name=suite -c user.email=suite@test -c commit.gpgsign=false "$@"; }

if ! git clone -q --depth 1 --branch "$TAG" \
    https://github.com/thomasfranke/writrun "$WORK/kit" 2>&1; then
  echo "FAIL  cloning WritRun $TAG — the e2e tier needs the network"
  fail=$((fail + 1))
  finish
fi

TARGET="$WORK/target"
mkdir -p "$TARGET"
(
  cd "$TARGET" || exit 1
  git_q init -q
  printf '# A project\n' > README.md
  git_q add .
  git_q commit -q -m "feat(api): begin the thing"
  git_q commit -q --allow-empty -m "fix(api): repair the thing"
  git_q commit -q --allow-empty -m "feat(cli): add another"
)

export WRITRUN_SOURCE="$WORK/kit"
cd "$TARGET" || exit 1

check "init adopts and reports the extracted vocabulary" 0 "types feat, fix" \
  -- "$WRITRUN" init --stage 1 --yes

OBSERVANCE=".writrun/scripts/stage-2-pull-requests/check_observance.sh"

# Ranked by frequency, most used first — feat twice, fix once.
check "the door's types are the project's" 0 "" \
  -- grep -qxF 'TYPES="feat fix"' "$OBSERVANCE"
check "the door's scopes are the project's" 0 "" \
  -- grep -qxF 'SCOPES="api cli"' "$OBSERVANCE"

COMMITS=".writrun/conventions/commits.md"
check "the prose types match the door" 0 "" \
  -- grep -qF -e '- **Types**: `feat`, `fix`.' "$COMMITS"
check "the prose scopes match the door" 0 "" \
  -- grep -qF '`api`, `cli`.' "$COMMITS"
check "no shipped type survives the rewrite" 1 "" \
  -- grep -qF -e '- **Types**: `docs`' "$COMMITS"

# The example is the file's own demonstration; left as shipped it would
# be a subject the hook this same run installed refuses.
check "the example is respelled in the project's vocabulary" 0 "" \
  -- grep -qF -e '- Example: `feat(api): ' "$COMMITS"

# The end of it: the installed hook accepts what the kit now declares
# and refuses what it does not.
check "the hook accepts a subject the rewritten vocabulary allows" 0 "" \
  -- git_q commit -q --allow-empty -m "fix(cli): a subject the vocabulary allows"
check "the hook refuses a type the rewritten vocabulary dropped" 1 "outside the vocabulary" \
  -- git_q commit -q --allow-empty -m "docs(api): a type the project does not use"
check "the hook refuses a scope the rewritten vocabulary dropped" 1 "outside the vocabulary" \
  -- git_q commit -q --allow-empty -m "feat(product): a scope the project does not use"

finish
