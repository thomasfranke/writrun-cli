#!/usr/bin/env bash
. "$(dirname "$0")/../../cli_lib.sh"

# The rewrite half of the adoption, against the real template rather
# than a fixture: a repository whose history is Conventional has its own
# vocabulary written where the checks read it, `commit_types` and
# `commit_scopes` in the adopter's settings. The parity case next door
# adopts a repository with no Conventional history, so the shipped
# defaults stand there and this path is never taken; a drift in the real
# file's line shape would fail the adoption mid-apply and nothing else
# would notice.
#
# It is also where the vocabulary having one home is proved: a second
# copy inside the kit's own home is what a refresh used to revert.

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

SETTINGS="writrun/settings.json"

# Ranked by frequency, most used first — feat twice, fix once.
check "the settings carry the project's types" 0 "" \
  -- grep -qF '"commit_types": "feat fix"' "$SETTINGS"
check "the settings carry the project's scopes" 0 "" \
  -- grep -qF '"commit_scopes": "api cli"' "$SETTINGS"
check "no shipped type survives the rewrite" 1 "" \
  -- grep -qF '"commit_types": "docs' "$SETTINGS"

# The vocabulary has one home, and the door is not it. A kit file
# carrying a second copy is what the next refresh would revert, leaving
# the two halves disagreeing — so the door must carry none at all.
DOOR=".writrun/scripts/stage-2-pull-requests/check_observance.sh"
check "the door declares no vocabulary of its own" 1 "" \
  -- grep -qE '^(TYPES|SCOPES)="' "$DOOR"
check "the door reads the vocabulary from the settings" 0 "" \
  -- grep -qF 'stage_2.commit_types' "$DOOR"

# The end of it: the installed hook accepts what the kit now declares
# and refuses what it does not.
check "the hook accepts a subject the rewritten vocabulary allows" 0 "" \
  -- git_q commit -q --allow-empty -m "fix(cli): a subject the vocabulary allows"
check "the hook refuses a type the rewritten vocabulary dropped" 1 "outside the vocabulary" \
  -- git_q commit -q --allow-empty -m "docs(api): a type the project does not use"
check "the hook refuses a scope the rewritten vocabulary dropped" 1 "outside the vocabulary" \
  -- git_q commit -q --allow-empty -m "feat(product): a scope the project does not use"

finish
