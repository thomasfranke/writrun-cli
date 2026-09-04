#!/usr/bin/env bash
. "$(dirname "$0")/../../cli_lib.sh"

# The Definition of Done (spec-0002): a fresh repository adopted
# end-to-end by the command matches a hand adoption of the same tag —
# the template copied file for file. The source is a local clone of the
# real WritRun repository at the pinned tag; the target's history is
# not Conventional, so the shipped conventions stand and every kit
# file must land byte-identical.

TAG=$("$WRITRUN" --version | sed -n 's/.*pins WritRun \(v[0-9.]*\).*/\1/p')
git_q() { git -c user.name=suite -c user.email=suite@test -c commit.gpgsign=false "$@"; }

if ! git clone -q --depth 1 --branch "$TAG" \
    https://github.com/thomasfranke/writrun "$WORK/writrun" 2>&1; then
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
  git_q commit -q -m "initial import"
)

export WRITRUN_SOURCE="$WORK/writrun"
cd "$TARGET" || exit 1

check "init adopts the fresh repository end to end" 0 "Adopted WritRun $TAG" \
  -- "$WRITRUN" init --stage 1 --yes

# Every template entry, compared file for file against what a hand
# copy of the same tag would have placed.
for entry in "$WORK/writrun/template"/* "$WORK/writrun/template"/.[!.]*; do
  [ -e "$entry" ] || continue
  name=$(basename "$entry")
  check "the adopted $name matches the hand copy" 0 "" \
    -- diff -r "$entry" "$TARGET/$name"
done

check "the commit-msg hook is installed and executable" 0 "" \
  -- test -x .git/hooks/commit-msg

finish
