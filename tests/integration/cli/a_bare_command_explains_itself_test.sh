#!/usr/bin/env bash
. "$(dirname "$0")/../../cli_lib.sh"

# A command run with nothing to go on says what it is for before it does
# anything (spec-0039, steps 2 and 3).
#
# The repository here is dirty on purpose. `init` refuses at its first
# check, which is after the frame has already explained — so the case
# reads the explanation without reaching a question, a stage or the
# network. WRITRUN_TTY_IN is the terminal the explanation is written
# for; without one nothing is printed.
git_q() { git -c user.name=suite -c user.email=suite@test -c commit.gpgsign=false "$@"; }

TARGET="$WORK/target"
mkdir -p "$TARGET"
(
  cd "$TARGET" || exit 1
  git_q init -q
  printf '# A project\n' > README.md
  git_q add -A
  git_q commit -q -m "initial import"
  printf 'uncommitted\n' >> README.md
)
printf '\033' > "$WORK/keys"
export WRITRUN_TTY_IN="$WORK/keys"
cd "$TARGET" || exit 1

"$WRITRUN" init > "$WORK/bare.out" 2>&1
check "the bare run opens with what the command is for" 0 \
  "^init — install WritRun into this repository" -- cat "$WORK/bare.out"
check "it says why you would reach for it" 0 "why you would" -- cat "$WORK/bare.out"
check "it says what it never does" 0 "what it never" -- cat "$WORK/bare.out"
check "it says what comes next" 0 "what comes next" -- cat "$WORK/bare.out"

"$WRITRUN" init --stage 1 > "$WORK/named.out" 2>&1
check "a run that names what it needs explains nothing" 1 "" \
  -- grep -q "why you would" "$WORK/named.out"

unset WRITRUN_TTY_IN
"$WRITRUN" init > "$WORK/noterm.out" 2>&1
check "without a terminal nothing is explained" 1 "" \
  -- grep -q "why you would" "$WORK/noterm.out"

check "--help answers it anywhere" 0 "what it does" -- "$WRITRUN" init --help

finish
