#!/usr/bin/env bash
. "$(dirname "$0")/../../init_lib.sh"

# Without --stage the stage is arrow-selected (spec-0002, steps): a
# down arrow and enter land stage 2. --yes answers the confirmation,
# so the selection is the one form reading the keys.
make_target "$TARGET"
printf '\033[B\r' > "$WORK/keys"
export WRITRUN_TTY_IN="$WORK/keys"
cd "$TARGET" || exit 1

check "the arrow-selected stage adopts" 0 "" \
  -- "$WRITRUN" init --yes
check "the selected stage landed in the settings" 0 '"stage": 2' \
  -- cat .writrun/settings.json

finish
